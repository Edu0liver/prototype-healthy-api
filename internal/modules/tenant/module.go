// Package tenant wires the tenant + white-label module into fx.
package tenant

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/service"
	"go.uber.org/fx"
)

// Module is the tenant module's sole public entry for fx.
var Module = fx.Module("tenant",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
