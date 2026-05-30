// Command api is the HTTP entrypoint: it assembles the fx application from the
// platform layer and the domain modules, then runs it.
package main

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/orchestration"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/realtime"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// Build metadata injected via -ldflags.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	app := fx.New(
		platform.Module,
		// Domain modules.
		tenant.Module,
		iam.Module,
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
	app.Run()
}
