-- +goose Up
CREATE TABLE companies (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    slug        text UNIQUE NOT NULL,
    status      text NOT NULL DEFAULT 'active',  -- active | suspended
    plan        text NOT NULL DEFAULT 'free',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE company_branding (
    company_id        uuid PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    logo_url          text,
    favicon_url       text,
    primary_color     text,
    secondary_color   text,
    email_sender_name text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE company_domains (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    domain      text UNIQUE NOT NULL,
    is_primary  boolean NOT NULL DEFAULT false,
    verified_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_company_domains_company ON company_domains (company_id);

-- +goose Down
DROP TABLE IF EXISTS company_domains;
DROP TABLE IF EXISTS company_branding;
DROP TABLE IF EXISTS companies;
