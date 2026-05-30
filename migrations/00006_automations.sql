-- +goose Up
CREATE TABLE automations (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    channel_id       uuid NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    agent_id         uuid NOT NULL REFERENCES agents(id) ON DELETE RESTRICT,
    is_active        boolean NOT NULL DEFAULT true,
    business_hours   jsonb NOT NULL DEFAULT '{}'::jsonb,
    fallback_message text,
    debounce_seconds int NOT NULL DEFAULT 8,   -- debounce window (PRD 2.5 / PROMPT 4)
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_automations_company ON automations (company_id);
CREATE INDEX idx_automations_channel ON automations (channel_id);

-- Invariant: at most one active automation per channel -> one active agent per channel.
CREATE UNIQUE INDEX uniq_active_automation_per_channel
    ON automations (channel_id)
    WHERE is_active = true;

-- +goose Down
DROP TABLE IF EXISTS automations;
