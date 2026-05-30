// Package service holds the handover use cases (operator takes over, replies,
// returns to AI, closes). State lives in Redis (operational) mirrored to PG.
package service

import (
	"strings"
	"time"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
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
	repo     *repository.Repository
	cipher   *crypto.Cipher
	adapters *channeladapter.Registry
}

// New builds the handover service.
func New(conv *convsvc.Service, rdb *redisx.Client, repo *repository.Repository, cipher *crypto.Cipher, adapters *channeladapter.Registry) *Service {
	return &Service{conv: conv, rdb: rdb, repo: repo, cipher: cipher, adapters: adapters}
}

func stripJID(jid string) string {
	if i := strings.IndexByte(jid, '@'); i >= 0 {
		jid = jid[:i]
	}
	return jid
}
