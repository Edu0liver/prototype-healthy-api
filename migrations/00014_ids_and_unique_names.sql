-- +goose Up
ALTER TABLE company_branding
    ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();

DELETE FROM automations
    WHERE agent_id IN (
        SELECT a.id FROM agents a
        INNER JOIN agents b ON a.company_id = b.company_id AND a.name = b.name AND a.created_at > b.created_at
    );

DELETE FROM agents a
    USING agents b
    WHERE a.company_id = b.company_id AND a.name = b.name AND a.created_at > b.created_at;

ALTER TABLE agents
    ADD CONSTRAINT agents_company_name_key UNIQUE (company_id, name);

DELETE FROM agent_knowledge_bases
    WHERE knowledge_base_id IN (
        SELECT a.id FROM knowledge_bases a
        INNER JOIN knowledge_bases b ON a.company_id = b.company_id AND a.name = b.name AND a.created_at > b.created_at
    );

DELETE FROM documents
    WHERE knowledge_base_id IN (
        SELECT a.id FROM knowledge_bases a
        INNER JOIN knowledge_bases b ON a.company_id = b.company_id AND a.name = b.name AND a.created_at > b.created_at
    );

DELETE FROM knowledge_bases a
    USING knowledge_bases b
    WHERE a.company_id = b.company_id AND a.name = b.name AND a.created_at > b.created_at;

ALTER TABLE knowledge_bases
    ADD CONSTRAINT knowledge_bases_company_name_key UNIQUE (company_id, name);

-- +goose Down
ALTER TABLE knowledge_bases DROP CONSTRAINT IF EXISTS knowledge_bases_company_name_key;
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_company_name_key;
ALTER TABLE company_branding DROP COLUMN IF EXISTS id;
