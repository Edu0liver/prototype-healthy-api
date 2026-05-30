-- +goose Up
CREATE TABLE channels (
    id                       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id               uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type                     text NOT NULL,            -- whatsapp | instagram
    name                     text,
    evolution_instance_name  text,                     -- WhatsApp instance name
    evolution_instance_id    text,
    evolution_apikey_enc     text,                     -- AES-GCM ciphertext
    external_account_id      text,                     -- phone number / instagram account
    status                   text NOT NULL DEFAULT 'disconnected', -- disconnected|connecting|connected|error
    active_agent_id          uuid REFERENCES agents(id) ON DELETE SET NULL,
    metadata                 jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now(),
    UNIQUE (company_id, type, external_account_id)
);
CREATE INDEX idx_channels_company ON channels (company_id);
-- Channel resolution from webhook (instance -> channel/tenant).
CREATE UNIQUE INDEX uniq_channels_company_instance
    ON channels (company_id, evolution_instance_name)
    WHERE evolution_instance_name IS NOT NULL;
CREATE INDEX idx_channels_instance_name
    ON channels (evolution_instance_name)
    WHERE evolution_instance_name IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS channels;
