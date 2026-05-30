# Módulo: agent

CRUD de agentes de IA (persona: prompt de sistema, modelo, parâmetros, handover).

## Endpoints (admin, sob tenant tx)
| Método | Rota | Descrição |
| --- | --- | --- |
| POST | `/agents` | Cria agente (RF-AG-01) |
| GET | `/agents` | Lista agentes |
| GET | `/agents/:id` | Detalhe |
| PUT | `/agents/:id` | Edita; `system_prompt` aplica-se a novas mensagens sem redeploy (RF-AG-03) |
| DELETE | `/agents/:id` | Remove |

## Notas
- `handover_keywords` persistido como `jsonb` via `database.JSONStringArray` (evita quirks de `text[]`).
- `temperature` numeric(3,2); defaults: model `gpt-4o-mini`, temp 0.7, max_tokens 1024, handover_enabled true, status `draft`.
- Associação N:M com bases RAG vive no módulo `knowledge` (M3).
