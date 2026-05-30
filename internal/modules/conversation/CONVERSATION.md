# Módulo: conversation

Contactos, conversas e mensagens. Persistência do pipeline + histórico do painel.

## Endpoints (admin + operator, sob tenant tx) — RF-LOG
| Método | Rota | Descrição |
| --- | --- | --- |
| GET | `/conversations` | Lista com filtros `channel_id`, `state`, `since` (RF-LOG-03) |
| GET | `/conversations/:id` | Detalhe |
| GET | `/conversations/:id/messages` | Histórico completo e imutável (RF-LOG-02) |

## Service (consumido por webhook + orchestration)
- `EnsureContact(channelID, remoteJID, pushName)` — upsert por `(channel_id, remote_jid)`.
- `EnsureOpenConversation(channelID, contactID, agentID)` — reabre/abre conversa não-`closed`.
- `AppendMessage(conv, AppendInput)` — insere mensagem **idempotente** (unique `(company_id, external_message_id)` → `inserted=false`), atualiza `last_message_at`.
- `SetState` / `AssignUser` — mirror de estado/handover.
- `RecentMessages(convID, n)` — histórico p/ prompt; `MarkStatusByExternalID` — entrega.

## Notas
- `Contact.RemoteJID` mapeado via `gorm:"column:remote_jid"` (GORM derivaria `remote_j_id`).
- `state` é espelho do Redis (`ai|human|closed`).
