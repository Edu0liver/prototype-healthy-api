-- +goose Up
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id    uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    email         citext NOT NULL,
    password_hash text,                              -- argon2id; null while invited
    name          text,
    role          text NOT NULL DEFAULT 'operator',  -- admin | operator | knowledge_manager
    status        text NOT NULL DEFAULT 'invited',   -- active | invited | disabled
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (company_id, email)
);
CREATE INDEX idx_users_company ON users (company_id);

-- +goose Down
DROP TABLE IF EXISTS users;
