// Package iam wires authentication, RBAC and user management into fx.
package iam

import (
	billingsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
	iamhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/mailer"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"go.uber.org/fx"
)

// Module is the iam module's sole public entry for fx.
var Module = fx.Module("iam",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(iamsvc.Repository))),
		// Adapt the platform mailer to the module's narrow interface.
		func(m *mailer.Mailer) iamsvc.Mailer { return m },
		// Build the service with the real billing quota guard wired in.
		func(
			repo iamsvc.Repository,
			m iamsvc.Mailer,
			db *database.DB,
			tokens *token.Manager,
			cfg *config.Config,
			rdb *redisx.Client,
			b *billingsvc.Service,
		) *iamsvc.Service {
			return iamsvc.New(repo, db, tokens, m, cfg, rdb).WithBilling(b)
		},
		iamhttp.NewHandler,
	),
	fx.Invoke(iamhttp.RegisterRoutes),
)
