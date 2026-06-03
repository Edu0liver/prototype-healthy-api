# Módulo: orchestration

Worker (sem HTTP) que processa o pipeline de entrada. Consome o Redis Stream produzido pelo `webhook`. Núcleo do PRD §2.6 + extras do PROMPT.md.

## Worker
- Pool de `WORKER_CONCURRENCY` goroutines (fx lifecycle); cada consumer faz `XREADGROUP` (group `orchestrators`), processa e `XACK`. Shutdown gracioso.

## Pipeline (`Service.Process`) por job
1. Carrega creds do canal (instância + apikey decifrada) e **agente ativo** (join `automations`+`agents`). Sem agente → drop.
2. **Áudio** (`messageType=audioMessage`, PROMPT 5): `get-media-base64` → decode → **Whisper** → usa transcrição como texto.
3. **Debounce** (PROMPT 4): push do fragmento p/ `buffer:conv` → **Redlock** `lock:conv:{id}` (se ocupado, devolve — outro worker trata o buffer) → espera janela (`automations.debounce_seconds`) → drena+agrega fragmentos. *(verificado: 2 fragmentos → 1 chamada ao LLM)*
4. **Estado** (`conv:state`, fallback PG, RF-HO-05): `human` → **não responde** (RF-HO-02); `closed` → reabre sob IA (PRD §2.6b; normalmente já é nova conversa criada no webhook); `block:conv` ativo (handover passivo do telemóvel) → **não responde** (auto-retoma após TTL).
5. **Horário de funcionamento** (`automations.business_hours`): fora da janela → envia `fallback_message` (se houver) + mark-as-read e **para** (sem LLM). Config vazia = 24/7. Formato: `{"timezone","windows":{"mon":[{"start","end"}]}}` — dia ausente = fechado.
6. **Handover por keyword** (RF-HO-01) → `handover`.
7. **RAG**: `knowledge.Retrieve(agentID, agregado, k=10)` (filtrado company+KBs do agente) + histórico recente (15 msgs).
8. **OpenAI chat** + function calling `transfer_to_human`. Erro → `fallback_message`. Tool call → `handover`.
9. **Humanização** (PROMPT 7): `humanize` — strip markdown, links→`texto (url)`, split ≤4 msgs.
10. **Envio** via `ChannelAdapter`: presença "composing" + `delay` 2–3s por mensagem; persiste outbound `ai`.
11. **read-messages** (mark as read) + mirror `conv:state=ai`.

## Notas
- Lê tabelas de outros módulos (channels/automations/agents) via repo próprio (tenant-scoped) p/ evitar acoplamento; usa `conversation.Service` + `knowledge.Service` para persistência/retrieval.
- `InboundJob` (`platform/jobs`) é o contrato com o `webhook`.
- Sem `OPENAI_API_KEY` o pipeline corre até ao chat e aplica fallback (verificado em dev).
