package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const reconcileInterval = 10 * time.Minute

// Reconciler periodically polls Evolution for channels that appear active in the
// DB and corrects any status divergence caused by missed CONNECTION_UPDATE events.
type Reconciler struct {
	db     *database.DB
	repo   Repository
	evo    evolution.Client
	cipher *crypto.Cipher
	cfg    *config.Config
	log    *zap.Logger
	stop   chan struct{}
}

// NewReconciler builds the Reconciler.
func NewReconciler(db *database.DB, repo Repository, evo evolution.Client, cipher *crypto.Cipher, cfg *config.Config, log *zap.Logger) *Reconciler {
	return &Reconciler{db: db, repo: repo, evo: evo, cipher: cipher, cfg: cfg, log: log, stop: make(chan struct{})}
}

// Start begins the background reconciliation loop (fx OnStart).
func (r *Reconciler) Start(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(reconcileInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := r.run(context.Background()); err != nil {
					r.log.Error("channel: reconcile run failed", zap.Error(err))
				}
			case <-r.stop:
				return
			}
		}
	}()
	return nil
}

// Stop signals the background goroutine to exit (fx OnStop).
func (r *Reconciler) Stop() { close(r.stop) }

type channelRef struct {
	id        uuid.UUID
	companyID uuid.UUID
	instance  string
	status    string
}

// run performs one reconciliation pass over all connected/connecting channels.
func (r *Reconciler) run(ctx context.Context) error {
	var channels []channelRef
	if err := r.db.System(ctx, func(sysCtx context.Context) error {
		chs, err := r.repo.ListAllActive(sysCtx)
		if err != nil {
			return err
		}
		for _, ch := range chs {
			channels = append(channels, channelRef{
				id:        ch.ID,
				companyID: ch.CompanyID,
				instance:  evoInstance(&ch),
				status:    ch.Status,
			})
		}
		return nil
	}); err != nil {
		return err
	}

	r.log.Info("channel: reconcile pass", zap.Int("candidates", len(channels)))
	for _, ch := range channels {
		r.reconcileOne(ctx, ch)
	}
	return nil
}

func (r *Reconciler) reconcileOne(ctx context.Context, ch channelRef) {
	evoState, err := r.evo.ConnectionState(ctx, ch.instance)
	var desired string
	if err != nil {
		desired = StatusDisconnected
	} else {
		desired = mapState(evoState)
	}
	if desired == ch.status {
		return
	}
	r.log.Warn("channel: status divergence detected",
		zap.String("channel_id", ch.id.String()),
		zap.String("db_status", ch.status),
		zap.String("evo_status", desired),
	)
	_ = r.db.Tenant(ctx, ch.companyID, func(tenCtx context.Context) error {
		fetched, ferr := r.repo.Get(tenCtx, ch.id)
		if ferr != nil {
			return ferr
		}
		fetched.Status = desired
		return r.repo.Update(tenCtx, fetched)
	})
}
