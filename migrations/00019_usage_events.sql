-- +goose Up
-- Durable, per-tenant metering ledger (source of truth for billing; Redis holds
-- the hot counter). Written asynchronously by the orchestration worker and the
-- knowledge ingest, both of which already run inside a tenant transaction
-- (db.Tenant -> SET LOCAL app.current_company_id). It is therefore a domain
-- table UNDER RLS, like conversations/messages.
--
-- Platform-side billing aggregation iterates companies and sets the tenant
-- scope per company (RLS-safe), rather than running a cross-tenant query.
CREATE TABLE usage_events (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,

    kind            text   NOT NULL,   -- ai_message | llm_tokens | audio_minutes | storage_mb | embedding_tokens
    quantity        bigint NOT NULL DEFAULT 0,

    -- Optional attribution for per-conversation / per-agent / per-model cost.
    conversation_id uuid REFERENCES conversations(id) ON DELETE SET NULL,
    agent_id        uuid REFERENCES agents(id) ON DELETE SET NULL,
    model           text,

    metadata        jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at      timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT usage_events_kind_check CHECK (
        kind IN ('ai_message', 'llm_tokens', 'audio_minutes', 'storage_mb', 'embedding_tokens')
    )
);

-- Aggregation by tenant + period, and by tenant + kind + period.
CREATE INDEX idx_usage_events_company_created ON usage_events (company_id, created_at);
CREATE INDEX idx_usage_events_company_kind_created ON usage_events (company_id, kind, created_at);

-- RLS (defense in depth) — identical policy to the other domain tables.
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON usage_events
    USING (company_id = current_setting('app.current_company_id', true)::uuid)
    WITH CHECK (company_id = current_setting('app.current_company_id', true)::uuid);

-- +goose Down
DROP TABLE IF EXISTS usage_events;
