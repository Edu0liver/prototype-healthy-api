-- +goose Up
-- One subscription per company. Managed by platform/gateway code (Stripe or a
-- Merchant-of-Record): the gateway webhook runs WITHOUT tenant context, so this
-- table is NOT under RLS — it follows the system-scoped tenant-registry pattern
-- (companies / company_domains / webhook_events). The application enforces the
-- company_id filter at the query layer (db.System), and tenants read their own
-- subscription through a dedicated, company-filtered endpoint.
CREATE TABLE subscriptions (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id             uuid NOT NULL UNIQUE REFERENCES companies(id) ON DELETE CASCADE,
    plan_id                uuid NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,

    status                 text NOT NULL DEFAULT 'active',     -- trialing | active | past_due | canceled | suspended
    billing_cycle          text NOT NULL DEFAULT 'monthly',    -- monthly | annual

    current_period_start   timestamptz NOT NULL DEFAULT now(),
    current_period_end     timestamptz NOT NULL DEFAULT (now() + interval '1 month'),
    cancel_at_period_end   boolean NOT NULL DEFAULT false,

    -- Gateway linkage (nullable until a paid checkout completes).
    stripe_customer_id     text,
    stripe_subscription_id text,

    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now()
);

-- Webhook lookups arrive keyed by the gateway subscription id.
CREATE UNIQUE INDEX uniq_subscriptions_stripe_sub
    ON subscriptions (stripe_subscription_id)
    WHERE stripe_subscription_id IS NOT NULL;

-- Backfill: every existing company gets a subscription on the plan whose code
-- matches companies.plan (default 'free'), falling back to 'free' if unknown.
-- +goose StatementBegin
INSERT INTO subscriptions (company_id, plan_id, status)
SELECT c.id,
       COALESCE(p.id, pf.id),
       'active'
FROM companies c
CROSS JOIN (SELECT id FROM plans WHERE code = 'free') pf
LEFT JOIN plans p ON p.code = c.plan;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
