-- +goose Up
-- Row-Level Security (defense in depth on top of the application layer).
-- Each DB session sets `SET app.current_company_id = '<uuid>'`; policies then
-- filter every row by company_id. FORCE is required because the app connects as
-- the table owner, and owners bypass RLS unless it is forced.
--
-- NOTE: companies / company_domains / company_branding are intentionally NOT
-- under RLS — they form the tenant registry that the Host->tenant resolver and
-- the public branding endpoint must read before any company context exists.
-- webhook_events is a system-only audit table (rows may precede tenant
-- resolution) and is accessed only by platform code, never tenant-facing.

-- +goose StatementBegin
DO $$
DECLARE
    t text;
    tables text[] := ARRAY[
        'users', 'channels', 'agents', 'automations',
        'knowledge_bases', 'agent_knowledge_bases', 'documents',
        'document_chunks', 'contacts', 'conversations', 'messages'
    ];
BEGIN
    FOREACH t IN ARRAY tables LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON %I '
            'USING (company_id = current_setting(''app.current_company_id'', true)::uuid) '
            'WITH CHECK (company_id = current_setting(''app.current_company_id'', true)::uuid)',
            t
        );
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    t text;
    tables text[] := ARRAY[
        'users', 'channels', 'agents', 'automations',
        'knowledge_bases', 'agent_knowledge_bases', 'documents',
        'document_chunks', 'contacts', 'conversations', 'messages'
    ];
BEGIN
    FOREACH t IN ARRAY tables LOOP
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
        EXECUTE format('ALTER TABLE %I NO FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I DISABLE ROW LEVEL SECURITY', t);
    END LOOP;
END $$;
-- +goose StatementEnd
