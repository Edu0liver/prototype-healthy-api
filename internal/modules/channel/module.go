// Package channel wires the channel module into fx.
package channel

import (
	"context"

	billingsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module is the channel module's sole public entry for fx.
var Module = fx.Module("channel",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		func(repo service.Repository, evo evolution.Client, cipher *crypto.Cipher, cfg *config.Config, log *zap.Logger, b *billingsvc.Service) *service.Service {
			return service.New(repo, evo, cipher, cfg, log).WithBilling(b)
		},
		service.NewReconciler,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
	fx.Invoke(func(lc fx.Lifecycle, r *service.Reconciler) {
		lc.Append(fx.Hook{
			OnStart: r.Start,
			OnStop:  func(_ context.Context) error { r.Stop(); return nil },
		})
	}),
)
