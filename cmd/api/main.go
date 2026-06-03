// Command api is the HTTP entrypoint: it assembles the fx application from the
// platform layer and the domain modules, then runs it.
package main

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/orchestration"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/realtime"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// @title                      API
// @version                    1.0
// @description                SaaS multi-tenant white-label de atendimento automatizado e omnichannel (WhatsApp/Instagram) com IA + RAG.
// @host                       localhost:8080
// @BasePath                   /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description.markdown
// options assembles the fx application (platform layer + domain modules). It is
// shared by main and the wiring-validation test (fx.ValidateApp).
func options() fx.Option {
	return fx.Options(
		shared.Module,
		// Domain modules.
		tenant.Module,
		iam.Module,
		billing.Module,
		agent.Module,
		channel.Module,
		automation.Module,
		knowledge.Module,
		conversation.Module,
		webhook.Module,
		orchestration.Module,
		handover.Module,
		realtime.Module,

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.Named("fx")}
		}),
	)
}

func main() {
	app := fx.New(options())
	app.Run()
}
