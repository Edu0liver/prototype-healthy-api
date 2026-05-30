package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/jobs"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/redisx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Worker consumes the inbound Redis Stream with a pool of goroutines, one
// consumer each, and dispatches jobs to the pipeline. It serializes per
// conversation via the pipeline's Redlock.
type Worker struct {
	svc    *Service
	rdb    *redisx.Client
	cfg    *config.Config
	log    *zap.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewWorker builds the worker.
func NewWorker(svc *Service, rdb *redisx.Client, cfg *config.Config, log *zap.Logger) *Worker {
	return &Worker{svc: svc, rdb: rdb, cfg: cfg, log: log}
}

// Register hooks the worker into the fx lifecycle.
func Register(lc fx.Lifecycle, w *Worker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return w.start(ctx) },
		OnStop:  func(ctx context.Context) error { return w.stop(ctx) },
	})
}

func (w *Worker) start(ctx context.Context) error {
	if err := w.rdb.EnsureGroup(ctx, w.cfg.Worker.StreamName, w.cfg.Worker.ConsumerGroup); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	n := w.cfg.Worker.Concurrency
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		consumer := fmt.Sprintf("consumer-%s-%d", shortID(), i)
		w.wg.Add(1)
		go w.consume(runCtx, consumer)
	}
	w.log.Info("orchestration workers started", zap.Int("concurrency", n))
	return nil
}

func (w *Worker) stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	done := make(chan struct{})
	go func() { w.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}
	w.log.Info("orchestration workers stopped")
	return nil
}

func (w *Worker) consume(ctx context.Context, consumer string) {
	defer w.wg.Done()
	stream := w.cfg.Worker.StreamName
	group := w.cfg.Worker.ConsumerGroup
	for {
		if ctx.Err() != nil {
			return
		}
		streams, err := w.rdb.ReadGroup(ctx, stream, group, consumer, 1, 5*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// redis.Nil (no entries within the block window) is expected; on a
			// real error back off briefly to avoid a busy loop.
			if err != redis.Nil {
				time.Sleep(500 * time.Millisecond)
			}
			continue
		}
		for _, st := range streams {
			for _, msg := range st.Messages {
				job := jobs.FromMap(msg.Values)
				if err := w.svc.Process(ctx, job); err != nil {
					w.log.Error("pipeline error", zap.String("conversation_id", job.ConversationID), zap.Error(err))
				}
				// Ack regardless: errors are logged/audited; the buffer + dedupe
				// keep the system consistent without redelivery storms.
				_ = w.rdb.Ack(ctx, stream, group, msg.ID)
			}
		}
	}
}

func shortID() string { return uuid.NewString()[:8] }
