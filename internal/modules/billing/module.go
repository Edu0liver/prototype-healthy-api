// Package billing wires the billing module into fx: plan/subscription reads,
// quota enforcement, metering (hot counter + durable outbox) and the meter
// worker that persists usage to Postgres.
package billing

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
	"github.com/Edu0liver/prototype-healthy-api/pkg/stripe"
	"go.uber.org/fx"
)

// Module is the billing module's sole public entry for fx. The *service.Service
// is exported into the graph so other modules (orchestration, knowledge,
// channel, agent, iam) can enforce quotas and record usage.
var Module = fx.Module("billing",
	fx.Provide(
		fx.Annotate(repository.New, fx.As(new(service.Repository))),
		// Adapt the generic Stripe client to the service's gateway interface.
		func(c stripe.Client) service.StripeGateway { return c },
		service.New,
		service.NewMeterWorker,
		http.NewHandler,
	),
	fx.Invoke(http.RegisterRoutes),
	fx.Invoke(service.RegisterMeter),
)
