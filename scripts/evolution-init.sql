-- Bootstrap schema for the Evolution API container, which shares this Postgres
-- instance but isolates its tables under the `evolution` schema (see the
-- search_path in docker-compose.yml). Runs once on first DB initialization.
CREATE SCHEMA IF NOT EXISTS evolution;
