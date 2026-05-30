// Package conversation wires contacts/conversations/messages into fx. Its
// Service is also consumed by the webhook and orchestration modules.
package conversation

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/services"
	"go.uber.org/fx"
)

// Module is the conversation module's sole public entry for fx.
var Module = fx.Module("conversation",
	fx.Provide(
		fx.Annotate(repositories.New, fx.As(new(services.Repository))),
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
