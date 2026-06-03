package service

import (
	"context"
	"sync"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// MeterWorker drains the usage outbox stream and persists each event to the
// per-tenant ledger (usage_events). It is the durable counterpart to the hot
// Redis counter; a crash before persistence loses at most the in-flight entries
// the stream has not yet acked (at-least-once delivery).
type MeterWorker struct {
	svc    *Service
	rdb    *redisx.Client
	db     *database.DB
	log    *zap.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMeterWorker builds the worker.
func NewMeterWorker(svc *Service, rdb *redisx.Client, db *database.DB, log *zap.Logger) *MeterWorker {
	return &MeterWorker{svc: svc, rdb: rdb, db: db, log: log}
}

// RegisterMeter hooks the meter worker into the fx lifecycle.
func RegisterMeter(lc fx.Lifecycle, w *MeterWorker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return w.start(ctx) },
		OnStop:  func(ctx context.Context) error { return w.stop(ctx) },
	})
}

func (w *MeterWorker) start(ctx context.Context) error {
	if err := w.rdb.EnsureGroup(ctx, usageStream, usageGroup); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.wg.Add(1)
	go w.consume(runCtx, "meter-"+uuid.NewString()[:8])
	w.log.Info("billing meter worker started")
	return nil
}

func (w *MeterWorker) stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	done := make(chan struct{})
	go func() { w.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}
	w.log.Info("billing meter worker stopped")
	return nil
}

func (w *MeterWorker) consume(ctx context.Context, consumer string) {
	defer w.wg.Done()
	for {
		if ctx.Err() != nil {
			return
		}
		streams, err := w.rdb.ReadGroup(ctx, usageStream, usageGroup, consumer, 16, 5*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err != redis.Nil {
				time.Sleep(500 * time.Millisecond)
			}
			continue
		}
		for _, st := range streams {
			for _, msg := range st.Messages {
				if err := w.persist(ctx, eventFromMap(msg.Values)); err != nil {
					w.log.Error("billing: persist usage failed", zap.Error(err))
				}
				// Ack regardless: the hot counter already reflects the usage and
				// retries would double-count; the ledger is best-effort durable.
				_ = w.rdb.Ack(ctx, usageStream, usageGroup, msg.ID)
			}
		}
	}
}

func (w *MeterWorker) persist(ctx context.Context, e Event) error {
	if e.CompanyID == uuid.Nil {
		return nil
	}
	return w.db.Tenant(ctx, e.CompanyID, func(ctx context.Context) error {
		return w.svc.repo.InsertUsageEvent(ctx, &models.UsageEvent{
			ID:             uuid.New(),
			CompanyID:      e.CompanyID,
			Kind:           e.Kind,
			Quantity:       e.Quantity,
			ConversationID: e.ConversationID,
			AgentID:        e.AgentID,
			Model:          e.Model,
			Metadata:       database.JSONMap{},
		})
	})
}
