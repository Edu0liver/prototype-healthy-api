-- +goose Up
CREATE TABLE contacts (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    channel_id      uuid NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    remote_jid      text NOT NULL,          -- 5511...@s.whatsapp.net
    push_name       text,
    profile_pic_url text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (channel_id, remote_jid)
);
CREATE INDEX idx_contacts_company ON contacts (company_id);

CREATE TABLE conversations (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    channel_id       uuid NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    contact_id       uuid NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    agent_id         uuid REFERENCES agents(id) ON DELETE SET NULL,
    state            text NOT NULL DEFAULT 'ai',  -- ai | human | closed (mirror of Redis)
    assigned_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    last_message_at  timestamptz,
    opened_at        timestamptz NOT NULL DEFAULT now(),
    closed_at        timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_conversations_company_state
    ON conversations (company_id, state, last_message_at DESC);
CREATE INDEX idx_conversations_contact ON conversations (contact_id);

CREATE TABLE messages (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id          uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    conversation_id     uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    direction           text NOT NULL,   -- inbound | outbound
    sender_type         text NOT NULL,   -- contact | ai | human
    content             text,
    media               jsonb,
    external_message_id text,            -- Evolution/Meta id (idempotency)
    status              text NOT NULL DEFAULT 'received', -- received|sent|delivered|read|failed
    created_at          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_messages_conversation
    ON messages (conversation_id, created_at DESC);
-- Idempotency of inbound/outbound by external id (when present).
CREATE UNIQUE INDEX uniq_messages_company_external
    ON messages (company_id, external_message_id)
    WHERE external_message_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS contacts;
