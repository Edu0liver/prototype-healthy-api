-- +goose Up
-- Allow a company to have multiple channels of the same type for the same phone number.
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_company_id_type_external_account_id_key;

-- +goose Down
ALTER TABLE channels ADD CONSTRAINT channels_company_id_type_external_account_id_key
    UNIQUE (company_id, type, external_account_id);
