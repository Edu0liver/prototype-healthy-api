-- +goose Up
CREATE TABLE webhook_events (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   uuid REFERENCES companies(id) ON DELETE CASCADE,
    channel_id   uuid REFERENCES channels(id) ON DELETE CASCADE,
    event_type   text NOT NULL,    -- MESSAGES_UPSERT, CONNECTION_UPDATE, ...
    external_id  text,
    payload      jsonb NOT NULL,
    processed_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_webhook_events_company ON webhook_events (company_id);
CREATE INDEX idx_webhook_events_external ON webhook_events (external_id);

-- +goose Down
DROP TABLE IF EXISTS webhook_events;
