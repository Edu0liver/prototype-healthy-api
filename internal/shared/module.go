// Package shared wires all cross-cutting infrastructure into a single fx module.
// Generic clients live under /pkg (decoupled from app config); the providers
// here map the application Config into each package's own options.
package shared

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/httpserver"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/logger"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/mailer"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/Edu0liver/prototype-healthy-api/pkg/openai"
	"github.com/Edu0liver/prototype-healthy-api/pkg/storage"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"go.uber.org/fx"
)

// Module provides the shared infrastructure layer: config, logging, persistence,
// integrations, HTTP server and global middleware.
var Module = fx.Module("shared",
	fx.Provide(
		config.Load,
		logger.New,
		database.New,
		redisx.New,
		mailer.New,
		events.New,
		middleware.New,
		channeladapter.NewRegistry,
		httpserver.NewEngine,
		httpserver.NewAPIGroup,
		// Generic /pkg clients built from app config.
		provideCrypto,
		provideToken,
		fx.Annotate(provideOpenAI, fx.As(new(openai.Client))),
		fx.Annotate(provideEvolution, fx.As(new(evolution.Client))),
		fx.Annotate(provideStorage, fx.As(new(storage.Storage))),
	),
	fx.Invoke(
		httpserver.InstallGlobal,
		httpserver.Run,
	),
)

func provideCrypto(cfg *config.Config) (*crypto.Cipher, error) {
	return crypto.New(cfg.Crypto.EncryptionKey)
}

func provideToken(cfg *config.Config) *token.Manager {
	return token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
}

func provideStorage(cfg *config.Config) (*storage.MinIOStorage, error) {
	return storage.NewMinIO(storage.MinIOConfig{
		Endpoint:  cfg.Storage.Endpoint,
		AccessKey: cfg.Storage.AccessKey,
		SecretKey: cfg.Storage.SecretKey,
		Bucket:    cfg.Storage.Bucket,
		UseSSL:    cfg.Storage.UseSSL,
		Region:    cfg.Storage.Region,
	})
}

func provideOpenAI(cfg *config.Config) *openai.HTTPClient {
	return openai.New(openai.Config{
		APIKey:         cfg.OpenAI.APIKey,
		BaseURL:        cfg.OpenAI.BaseURL,
		EmbeddingModel: cfg.OpenAI.EmbeddingModel,
		WhisperModel:   cfg.OpenAI.WhisperModel,
		Timeout:        cfg.OpenAI.Timeout,
	})
}

func provideEvolution(cfg *config.Config) *evolution.HTTPClient {
	return evolution.New(evolution.Config{
		BaseURL:      cfg.Evolution.BaseURL,
		GlobalAPIKey: cfg.Evolution.GlobalAPIKey,
		Timeout:      cfg.Evolution.Timeout,
	})
}
