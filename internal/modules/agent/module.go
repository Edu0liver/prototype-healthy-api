// Package agent wires the agent module into fx.
package agent

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/services"
	"go.uber.org/fx"
)

// Module is the agent module's sole public entry for fx.
var Module = fx.Module("agent",
	fx.Provide(
		fx.Annotate(repositories.New, fx.As(new(services.Repository))),
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
