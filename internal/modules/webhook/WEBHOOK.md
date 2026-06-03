# Módulo: webhook

Receção de webhooks da Evolution API V2 → validação, idempotência, routing, persistência e enfileiramento.

## Endpoint
| Método | Rota | Auth | Descrição |
| --- | --- | --- | --- |
| POST | `/webhooks/evolution` | token partilhado (`authorization: Bearer <EVOLUTION_WEBHOOK_TOKEN>`) | Recebe todos os eventos |

Responde sempre `200` (evento auditado em `webhook_events`); erros internos logados, sem retry do provedor.

## Fluxo (PRD §2.6, §6.6)
1. Valida token. Parse envelope (`event`, `instance`, `data`).
2. **Routing sem tenant:** `resolve_channel_by_instance(instance)` (função `SECURITY DEFINER`, migration 00012) devolve `company_id`/`channel_id` — único escape à RLS para o lookup. Instância desconhecida → ignora.
3. Auditoria em `webhook_events` (system scope).
4. Por evento:
   - **MESSAGES_UPSERT:** idempotência (`dedupe:msg:{id}` SETNX 24h + unique em `messages`). `fromMe=false` → persiste inbound (contact) e **enfileira** `InboundJob` no Redis Stream (após commit). `fromMe=true` → **handover passivo** (PROMPT 6): persiste outbound `human` e seta **só** `block:conv` (TTL 30m), **não** enfileira. **Não** flipa `state=human` permanente — isso fica reservado ao handover explícito (painel/`transfer_to_human`). O block renova a cada msg do telemóvel; após o TTL sem novas msgs, a IA retoma sozinha (igual ao protótipo n8n).
   - **CONNECTION_UPDATE:** atualiza `channels.status` + cache Redis.
   - **SEND_MESSAGE / MESSAGES_UPDATE:** atualiza `messages.status` por `external_message_id`.
   - **QRCODE_UPDATED:** realtime (M6).

## Notas
- `InboundJob` (`platform/jobs`) é o contrato com o worker de orquestração (M5).
- Enfileira **após** commit da tx p/ o worker ver a mensagem.
