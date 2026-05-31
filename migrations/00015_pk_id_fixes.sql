-- +goose Up

-- company_branding: promote id to PK, company_id becomes UNIQUE FK
ALTER TABLE company_branding DROP CONSTRAINT company_branding_pkey;
ALTER TABLE company_branding ADD PRIMARY KEY (id);
ALTER TABLE company_branding ADD CONSTRAINT company_branding_company_id_key UNIQUE (company_id);

-- agent_knowledge_bases: replace composite PK with surrogate id
ALTER TABLE agent_knowledge_bases DROP CONSTRAINT agent_knowledge_bases_pkey;
ALTER TABLE agent_knowledge_bases ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE agent_knowledge_bases ADD PRIMARY KEY (id);
ALTER TABLE agent_knowledge_bases ADD CONSTRAINT agent_knowledge_bases_agent_kb_key UNIQUE (agent_id, knowledge_base_id);

-- +goose Down
ALTER TABLE agent_knowledge_bases DROP CONSTRAINT IF EXISTS agent_knowledge_bases_agent_kb_key;
ALTER TABLE agent_knowledge_bases DROP CONSTRAINT IF EXISTS agent_knowledge_bases_pkey;
ALTER TABLE agent_knowledge_bases DROP COLUMN IF EXISTS id;
ALTER TABLE agent_knowledge_bases ADD PRIMARY KEY (agent_id, knowledge_base_id);

ALTER TABLE company_branding DROP CONSTRAINT IF EXISTS company_branding_company_id_key;
ALTER TABLE company_branding DROP CONSTRAINT IF EXISTS company_branding_pkey;
ALTER TABLE company_branding ADD PRIMARY KEY (company_id);
