-- +goose Up
-- Webhooks arrive with no tenant context, yet must resolve instance -> company
-- before any tenant scope can be set. channels is under RLS, so app_user cannot
-- read it without the GUC. This SECURITY DEFINER function runs as the owner
-- (bypassing RLS) and is the single, narrow, read-only escape hatch for routing.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION resolve_channel_by_instance(p_instance text)
RETURNS TABLE(company_id uuid, channel_id uuid)
LANGUAGE sql
SECURITY DEFINER
SET search_path = public
AS $$
    SELECT c.company_id, c.id
      FROM channels c
     WHERE c.evolution_instance_name = p_instance
     LIMIT 1;
$$;
-- +goose StatementEnd

GRANT EXECUTE ON FUNCTION resolve_channel_by_instance(text) TO app_user;

-- +goose Down
DROP FUNCTION IF EXISTS resolve_channel_by_instance(text);
