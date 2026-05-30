// Package iam wires authentication, RBAC and user management into fx.
package iam

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/mailer"
	"go.uber.org/fx"
)

// Module is the iam module's sole public entry for fx.
var Module = fx.Module("iam",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		// Adapt the platform mailer to the module's narrow interface.
		func(m *mailer.Mailer) service.Mailer { return m },
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
