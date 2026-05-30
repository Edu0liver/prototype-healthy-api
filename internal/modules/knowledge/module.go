// Package knowledge wires the RAG module into fx.
package knowledge

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/openai"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/storage"
	"go.uber.org/fx"
)

// Module is the knowledge module's sole public entry for fx.
var Module = fx.Module("knowledge",
	fx.Provide(
		fx.Annotate(repositories.New, fx.As(new(services.Repository))),
		// Adapt platform integrations to the module's narrow interfaces.
		func(s storage.Storage) services.Storage { return s },
		func(c openai.Client) services.Embedder { return c },
		services.New,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
)
