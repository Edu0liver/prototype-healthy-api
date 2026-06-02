# Módulo: knowledge (RAG)

Bases de conhecimento, ingestão de documentos e retrieval vetorial (pgvector).

## Endpoints (admin + knowledge_manager, sob tenant tx)
| Método | Rota | Descrição |
| --- | --- | --- |
| POST/GET | `/knowledge-bases` | Cria / lista bases (RF-RAG-01) |
| GET/DELETE | `/knowledge-bases/:id` | Detalhe / remove |
| POST | `/knowledge-bases/:id/documents` | Upload de ficheiro (multipart `file`) (RF-RAG-02) |
| POST | `/knowledge-bases/:id/documents/text` | Indexa texto colado |
| GET | `/knowledge-bases/:id/documents` | Lista documentos + status |
| DELETE | `/documents/:docId` | Remove documento (chunks via cascade) (RF-RAG-05) |
| POST/DELETE | `/agents/:id/knowledge-bases/:kbId` | Liga/desliga agente↔base (N:M, RF-AG-02) |

## Pipeline de ingestão (assíncrono, RF-RAG-03)
Upload → guarda em storage (`tenant/<company_id>/kb/...`) → cria `documents` (pending) → goroutine `ingest`: `processing` → extrai texto → chunk → embeddings (OpenAI batch) → `ReplaceChunks` em `document_chunks` → `indexed` (ou `failed` com erro).
- Extração (pure-Go): txt/md/html nativo; **pdf** via `dslipak/pdf`; **docx** via parse OOXML (`archive/zip`+`encoding/xml` de `word/document.xml`). PDF escaneado/imagem → `failed` com `ErrNoTextExtracted` (sem OCR). Legacy `.doc` (binário) → `failed` com `ErrUnsupportedFormat`.
- Chunking: janela por caracteres (≈ tokens×4) com overlap.

## Retrieval (RF-RAG-04, invariante 6)
`Service.Retrieve(agentID, query, k)`: resolve KBs do agente → embed da query → `Search` pgvector. **Sempre** pré-filtrado por `company_id` + `knowledge_base_id IN (KBs do agente)`. Operador `<=>` (cosine), índice HNSW. Consumido pelo worker (M5).

## Notas
- Embedding `vector(1536)` via `database.Vector` (encoding texto `[..]`).
- Sem `OPENAI_API_KEY`, ingestão termina em `failed` (esperado em dev).
