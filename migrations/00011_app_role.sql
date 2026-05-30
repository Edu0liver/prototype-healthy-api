-- +goose Up
-- The application must connect as a NON-superuser, NON-owner role for RLS to
-- engage (superusers and table owners bypass RLS). Migrations keep running as
-- the superuser (`lumia`); the Go app connects as `app_user`.
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
        CREATE ROLE app_user LOGIN PASSWORD 'app_pw';
    END IF;
END $$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- Auto-grant privileges on tables/sequences created by later migrations
-- (which run as the superuser `lumia`).
ALTER DEFAULT PRIVILEGES FOR ROLE lumia IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES FOR ROLE lumia IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO app_user;

-- +goose Down
ALTER DEFAULT PRIVILEGES FOR ROLE lumia IN SCHEMA public
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM app_user;
ALTER DEFAULT PRIVILEGES FOR ROLE lumia IN SCHEMA public
    REVOKE USAGE, SELECT ON SEQUENCES FROM app_user;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM app_user;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM app_user;
REVOKE USAGE ON SCHEMA public FROM app_user;
-- Role is left in place (other databases/objects may depend on it).
