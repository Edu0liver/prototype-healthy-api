// Package channel wires the channel module into fx.
package channel

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/service"
	"go.uber.org/fx"
)

// Module is the channel module's sole public entry for fx.
var Module = fx.Module("channel",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
