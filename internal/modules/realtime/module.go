// Package realtime wires the WebSocket realtime bridge into fx.
package realtime

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/realtime/http"
	"go.uber.org/fx"
)

// Module is the realtime module's sole public entry for fx.
var Module = fx.Module("realtime",
	fx.Provide(http.NewHandler),
	fx.Invoke(http.RegisterRoutes),
)
