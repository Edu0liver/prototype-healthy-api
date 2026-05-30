// Package channel wires the channel module into fx.
package channel

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/services"
	"go.uber.org/fx"
)

// Module is the channel module's sole public entry for fx.
var Module = fx.Module("channel",
	fx.Provide(
		fx.Annotate(repositories.New, fx.As(new(services.Repository))),
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
