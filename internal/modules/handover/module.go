// Package handover wires operator handover endpoints into fx.
package handover

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/services"
	"go.uber.org/fx"
)

// Module is the handover module's sole public entry for fx.
var Module = fx.Module("handover",
	fx.Provide(
		repositories.New,
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
