// Package platform wires all cross-cutting infrastructure into a single fx module.
package platform

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/crypto"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/evolution"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httpserver"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/logger"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/mailer"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/middleware"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/openai"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/storage"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/token"
	"go.uber.org/fx"
)

// Module provides the platform layer: config, logging, persistence, integrations,
// HTTP server and global middleware.
var Module = fx.Module("platform",
	fx.Provide(
		config.Load,
		logger.New,
		database.New,
		redisx.New,
		crypto.New,
		token.New,
		mailer.New,
		events.New,
		middleware.New,
		channeladapter.NewRegistry,
		httpserver.NewEngine,
		// Integrations bound to their interfaces for testability.
		fx.Annotate(openai.New, fx.As(new(openai.Client))),
		fx.Annotate(evolution.New, fx.As(new(evolution.Client))),
		fx.Annotate(storage.NewLocal, fx.As(new(storage.Storage))),
	),
	fx.Invoke(
		httpserver.InstallGlobal,
		httpserver.Run,
	),
)
