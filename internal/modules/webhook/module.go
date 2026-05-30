// Package webhook wires the Evolution webhook ingestion service into fx.
package webhook

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/service"
	"go.uber.org/fx"
)

// Module is the webhook module's sole public entry for fx.
var Module = fx.Module("webhook",
	fx.Provide(
		repository.New,
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
