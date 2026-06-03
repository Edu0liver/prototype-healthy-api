// Package agent wires the agent module into fx.
package agent

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	billingsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
	"go.uber.org/fx"
)

// Module is the agent module's sole public entry for fx.
var Module = fx.Module("agent",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		// Build the service with the real billing quota guard wired in.
		func(repo service.Repository, b *billingsvc.Service) *service.Service {
			return service.New(repo).WithBilling(b)
		},
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
