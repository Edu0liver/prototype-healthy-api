// Package orchestration wires the inbound processing worker into fx. It has no
// HTTP surface — it consumes the Redis Stream produced by the webhook module.
package orchestration

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/orchestration/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/orchestration/services"
	"go.uber.org/fx"
)

// Module is the orchestration module's sole public entry for fx.
var Module = fx.Module("orchestration",
	fx.Provide(
		repositories.New,
		services.New,
		services.NewWorker,
	),
	fx.Invoke(services.Register),
)
