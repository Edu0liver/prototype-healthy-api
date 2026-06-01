-- +goose Up

-- System roles (global, no RLS)
CREATE TABLE roles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL UNIQUE,
    description text,
    created_at  timestamptz NOT NULL DEFAULT now()
);

INSERT INTO roles (name, description) VALUES
    ('admin',             'Full access to all resources'),
    ('company_admin',     'Full access to all resources of a company'),
    ('operator',          'Handle conversations and customers'),
    ('knowledge_manager', 'Manage knowledge bases');

-- Migrate users.role (text) → users.role_id (FK)
ALTER TABLE users ADD COLUMN role_id uuid REFERENCES roles(id);
UPDATE users SET role_id = (SELECT id FROM roles WHERE name = users.role);
ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;
ALTER TABLE users DROP COLUMN role;

-- Enforce globally unique emails (users belong to exactly one company)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_company_id_email_key;
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);

-- SECURITY DEFINER function lets app_user bypass RLS for the login lookup.
-- Returns only id+company_id for the given email; password is never exposed here.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION find_user_by_email(p_email citext)
RETURNS TABLE(id uuid, company_id uuid)
SECURITY DEFINER
SET search_path = public
LANGUAGE sql AS $$
    SELECT id, company_id FROM users WHERE email = p_email LIMIT 1;
$$;
-- +goose StatementEnd

GRANT EXECUTE ON FUNCTION find_user_by_email(citext) TO app_user;
GRANT SELECT ON roles TO app_user;

-- +goose Down
REVOKE EXECUTE ON FUNCTION find_user_by_email(citext) FROM app_user;
DROP FUNCTION IF EXISTS find_user_by_email(citext);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;

ALTER TABLE users ADD COLUMN role text;
UPDATE users SET role = (SELECT name FROM roles WHERE id = users.role_id);
ALTER TABLE users ALTER COLUMN role SET NOT NULL;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'operator';
ALTER TABLE users DROP COLUMN role_id;

ALTER TABLE users ADD CONSTRAINT users_company_id_email_key UNIQUE (company_id, email);

DROP TABLE roles;
