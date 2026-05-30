-- +goose Up
CREATE TABLE knowledge_bases (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            text NOT NULL,
    description     text,
    embedding_model text NOT NULL DEFAULT 'text-embedding-3-small',
    chunk_size      int NOT NULL DEFAULT 800,
    chunk_overlap   int NOT NULL DEFAULT 100,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_kb_company ON knowledge_bases (company_id);

CREATE TABLE agent_knowledge_bases (
    agent_id          uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    knowledge_base_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    company_id        uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    PRIMARY KEY (agent_id, knowledge_base_id)
);
CREATE INDEX idx_akb_company ON agent_knowledge_bases (company_id);
CREATE INDEX idx_akb_kb ON agent_knowledge_bases (knowledge_base_id);

CREATE TABLE documents (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    knowledge_base_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    source_type       text NOT NULL,   -- file | text
    filename          text,
    storage_path      text,            -- tenant/<company_id>/...
    status            text NOT NULL DEFAULT 'pending', -- pending|processing|indexed|failed
    error             text,
    token_count       int NOT NULL DEFAULT 0,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_documents_company ON documents (company_id);
CREATE INDEX idx_documents_kb ON documents (knowledge_base_id);

CREATE TABLE document_chunks (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    knowledge_base_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    document_id       uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index       int NOT NULL,
    content           text NOT NULL,
    embedding         vector(1536),
    metadata          jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at        timestamptz NOT NULL DEFAULT now()
);
-- Tenant + RAG pre-filter (PRD 3.4): vector search is always filtered by these first.
CREATE INDEX idx_chunks_company_kb ON document_chunks (company_id, knowledge_base_id);
CREATE INDEX idx_chunks_document ON document_chunks (document_id);
-- HNSW cosine index for similarity search.
CREATE INDEX idx_chunks_embedding_hnsw
    ON document_chunks USING hnsw (embedding vector_cosine_ops);

-- +goose Down
DROP TABLE IF EXISTS document_chunks;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS agent_knowledge_bases;
DROP TABLE IF EXISTS knowledge_bases;
