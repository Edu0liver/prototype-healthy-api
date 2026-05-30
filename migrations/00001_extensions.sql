-- +goose Up
-- Extensions required by the platform.
CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS citext;     -- case-insensitive email
CREATE EXTENSION IF NOT EXISTS vector;     -- pgvector embeddings

-- +goose Down
DROP EXTENSION IF EXISTS vector;
DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;
