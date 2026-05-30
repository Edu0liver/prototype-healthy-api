package service

import (
	"context"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
)

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
