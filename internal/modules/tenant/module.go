// Package tenant wires the tenant + white-label module into fx.
package tenant

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/services"
	"go.uber.org/fx"
)

// Module is the tenant module's sole public entry for fx.
var Module = fx.Module("tenant",
	fx.Provide(
		fx.Annotate(repositories.New, fx.As(new(services.Repository))),
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
