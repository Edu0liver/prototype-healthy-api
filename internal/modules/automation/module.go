// Package automation wires the automation module into fx.
package automation

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/service"
	"go.uber.org/fx"
)

// Module is the automation module's sole public entry for fx.
var Module = fx.Module("automation",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
