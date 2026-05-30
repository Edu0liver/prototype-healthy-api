// Package handover wires operator handover endpoints into fx.
package handover

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/service"
	"go.uber.org/fx"
)

// Module is the handover module's sole public entry for fx.
var Module = fx.Module("handover",
	fx.Provide(
		repository.New,
		service.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
