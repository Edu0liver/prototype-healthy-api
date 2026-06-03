// Package knowledge wires the RAG module into fx.
package knowledge

import (
	billingsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/pkg/openai"
	"github.com/Edu0liver/prototype-healthy-api/pkg/storage"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module is the knowledge module's sole public entry for fx.
var Module = fx.Module("knowledge",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		// Adapt platform integrations to the module's narrow interfaces.
		func(s storage.Storage) service.Storage { return s },
		func(c openai.Client) service.Embedder { return c },
		func(repo service.Repository, db *database.DB, store service.Storage, embed service.Embedder, log *zap.Logger, b *billingsvc.Service) *service.Service {
			return service.New(repo, db, store, embed, log).WithBilling(b)
		},
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
