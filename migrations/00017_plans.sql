-- +goose Up
-- Global plan catalogue (tiers / quotas / prices). Like companies and the rest
-- of the tenant registry, `plans` is NOT under RLS: it has no company_id and is
-- read by every tenant and managed only by platform (super-admin) code.
--
-- Quota columns: 0 means UNLIMITED (used by enterprise). Freemium/paid tiers set
-- explicit positive limits. Overage columns of 0 mean overage is DISABLED for
-- that tier (hard-stop); a positive value opts the tier into metered overage.
CREATE TABLE plans (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                        text UNIQUE NOT NULL,           -- free | starter | pro | enterprise
    name                        text NOT NULL,
    price_cents                 int  NOT NULL DEFAULT 0,
    currency                    text NOT NULL DEFAULT 'BRL',

    -- Usage quotas (per billing period). 0 = unlimited.
    quota_ai_messages           int    NOT NULL DEFAULT 0,
    quota_tokens                bigint NOT NULL DEFAULT 0,
    quota_audio_minutes         int    NOT NULL DEFAULT 0,
    quota_storage_mb            int    NOT NULL DEFAULT 0,

    -- Hard resource caps (enforced at create_* time). 0 = unlimited.
    max_channels                int NOT NULL DEFAULT 0,
    max_agents                  int NOT NULL DEFAULT 0,
    max_kb                      int NOT NULL DEFAULT 0,
    max_seats                   int NOT NULL DEFAULT 0,

    -- Overage pricing. 0 = overage disabled (hard-stop).
    overage_per_msg_cents       int NOT NULL DEFAULT 0,
    overage_per_1k_tokens_cents int NOT NULL DEFAULT 0,

    is_active                   boolean NOT NULL DEFAULT true,
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now()
);

-- Seed the baseline catalogue. Enterprise uses 0 (unlimited) on quotas/caps.
INSERT INTO plans
    (code, name, price_cents, currency,
     quota_ai_messages, quota_tokens, quota_audio_minutes, quota_storage_mb,
     max_channels, max_agents, max_kb, max_seats,
     overage_per_msg_cents, overage_per_1k_tokens_cents)
VALUES
    ('free',       'Free',         0, 'BRL',
        100,      50000,    10,    50,
        1, 1, 1, 1,
        0, 0),
    ('starter',    'Starter',   9900, 'BRL',
        2000,     2000000,  120,   500,
        2, 3, 5, 3,
        0, 0),
    ('pro',        'Pro',      29900, 'BRL',
        10000,    10000000, 600,   2000,
        5, 10, 20, 10,
        0, 0),
    ('enterprise', 'Enterprise',   0, 'BRL',
        0, 0, 0, 0,
        0, 0, 0, 0,
        0, 0);

-- +goose Down
DROP TABLE IF EXISTS plans;
