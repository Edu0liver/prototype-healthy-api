-- +goose Up
CREATE TABLE agents (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name              text NOT NULL,
    system_prompt     text NOT NULL,
    model             text NOT NULL DEFAULT 'gpt-4o-mini',
    temperature       numeric(3,2) NOT NULL DEFAULT 0.70,
    max_output_tokens int NOT NULL DEFAULT 1024,
    handover_enabled  boolean NOT NULL DEFAULT true,
    handover_keywords jsonb NOT NULL DEFAULT '[]'::jsonb,
    status            text NOT NULL DEFAULT 'draft',  -- active | draft
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_agents_company ON agents (company_id);

-- +goose Down
DROP TABLE IF EXISTS agents;
