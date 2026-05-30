// Package iam wires authentication, RBAC and user management into fx.
package iam

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/mailer"
	"go.uber.org/fx"
)

// Module is the iam module's sole public entry for fx.
var Module = fx.Module("iam",
	fx.Provide(
		fx.Annotate(repositories.New, fx.As(new(services.Repository))),
		// Adapt the platform mailer to the module's narrow interface.
		func(m *mailer.Mailer) services.Mailer { return m },
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
