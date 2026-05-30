// Package services holds the handover use cases (operator takes over, replies,
// returns to AI, closes). State lives in Redis (operational) mirrored to PG.
package services

import (
	"context"
	"strings"
	"time"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/crypto"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/redisx"
	"github.com/google/uuid"
)

// blockTTL keeps the AI silent while a human is active.
const blockTTL = 30 * time.Minute

// Errors.
var (
	ErrNotHuman  = httputil.Conflict("conversation is not under human control")
	ErrNoChannel = httputil.BadRequest("conversation channel cannot dispatch")
)

// Service implements handover use cases.
type Service struct {
	conv     *convsvc.Service
	rdb      *redisx.Client
	repo     *repositories.Repository
	cipher   *crypto.Cipher
	adapters *channeladapter.Registry
}

// New builds the handover service.
func New(conv *convsvc.Service, rdb *redisx.Client, repo *repositories.Repository, cipher *crypto.Cipher, adapters *channeladapter.Registry) *Service {
	return &Service{conv: conv, rdb: rdb, repo: repo, cipher: cipher, adapters: adapters}
}

// Take transfers a conversation to a human operator (RF-HO-03).
func (s *Service) Take(ctx context.Context, convID, userID uuid.UUID) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if err := s.conv.SetState(ctx, conv, convsvc.StateHuman); err != nil {
		return err
	}
	if err := s.conv.AssignUser(ctx, conv, &userID); err != nil {
		return err
	}
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateHuman)
	_ = s.rdb.Block(ctx, convID.String(), blockTTL)
	return nil
}

// Reply dispatches an operator message and persists it as sender_type=human.
func (s *Service) Reply(ctx context.Context, convID uuid.UUID, text string) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if conv.State != convsvc.StateHuman {
		return ErrNotHuman
	}
	info, err := s.repo.LoadDispatchInfo(ctx, convID)
	if err != nil {
		return err
	}
	adapter, ok := s.adapters.For(info.ChannelType)
	if !ok {
		return ErrNoChannel
	}
	apiKey, _ := s.cipher.Decrypt(info.APIKeyEnc)
	out := channeladapter.Outbound{Instance: info.Instance, APIKey: apiKey, Number: stripJID(info.RemoteJID)}
	msgID, err := adapter.SendText(ctx, out, text, 0)
	if err != nil {
		return err
	}
	_, _, err = s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
		Direction: "outbound", SenderType: "human", Content: text,
		ExternalMessageID: msgID, Status: "sent",
	})
	// Refresh the block so the AI stays silent while the operator works.
	_ = s.rdb.Block(ctx, convID.String(), blockTTL)
	return err
}

// Return hands the conversation back to the AI (RF-HO-04).
func (s *Service) Return(ctx context.Context, convID uuid.UUID) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if err := s.conv.SetState(ctx, conv, convsvc.StateAI); err != nil {
		return err
	}
	_ = s.conv.AssignUser(ctx, conv, nil)
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateAI)
	_ = s.rdb.Unblock(ctx, convID.String())
	return nil
}

// Close ends the conversation (RF-HO-04).
func (s *Service) Close(ctx context.Context, convID uuid.UUID) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if err := s.conv.SetState(ctx, conv, convsvc.StateClosed); err != nil {
		return err
	}
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateClosed)
	_ = s.rdb.Unblock(ctx, convID.String())
	return nil
}

func stripJID(jid string) string {
	if i := strings.IndexByte(jid, '@'); i >= 0 {
		jid = jid[:i]
	}
	return jid
}
