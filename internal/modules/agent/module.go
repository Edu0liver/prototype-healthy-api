// Package agent wires the agent module into fx.
package agent

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	"go.uber.org/fx"
)

// Module is the agent module's sole public entry for fx.
var Module = fx.Module("agent",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
