// Package service holds the billing use cases: plan/subscription reads, quota
// checks (hard resource caps + soft usage limits) and metering (hot Redis
// counter + durable outbox to Postgres).
package service

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"go.uber.org/zap"
)

// Outbox stream for durable usage-event writes (consumed by the meter worker).
const (
	usageStream = "stream:billing-usage"
	usageGroup  = "billing-meters"
)

// Service implements the billing use cases.
type Service struct {
	repo   Repository
	db     *database.DB
	rdb    *redisx.Client
	stripe StripeGateway
	log    *zap.Logger
}

// New builds the billing service.
func New(repo Repository, db *database.DB, rdb *redisx.Client, stripe StripeGateway, log *zap.Logger) *Service {
	return &Service{repo: repo, db: db, rdb: rdb, stripe: stripe, log: log}
}
