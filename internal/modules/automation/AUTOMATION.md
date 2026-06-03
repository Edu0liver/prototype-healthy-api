# Módulo: automation

Ligação operacional canal↔agente e regras (horário, fallback, debounce). Impõe a invariante **1 agente ativo por canal**.

## Endpoints (admin, sob tenant tx)
| Método | Rota | Descrição |
| --- | --- | --- |
| POST | `/automations` | Cria binding canal→agente (RF-CH-03) |
| GET | `/automations` / `/automations/:id` | Lista / detalhe |
| PUT | `/automations/:id` | Atualiza (toggle `is_active`, agente, fallback, debounce) |
| DELETE | `/automations/:id` | Remove |

## Notas
- **Invariante 2 (PRD):** índice parcial `uniq_active_automation_per_channel` rejeita 2ª automação ativa no mesmo canal → erro `409` (`ErrActiveExists`).
- Valida que `channel_id`/`agent_id` pertencem ao tenant (queries scoped) antes de inserir.
- Ao ativar/desativar, reflete `channels.active_agent_id` (set/clear).
- `debounce_seconds` (default 8s) e `business_hours` consumidos pelo worker de orquestração (M5). `business_hours` JSON: `{"timezone":"America/Sao_Paulo","windows":{"mon":[{"start":"09:00","end":"18:00"}]}}` — fora da janela o worker envia `fallback_message` e não chama o LLM; config vazia = 24/7; dia ausente em `windows` = fechado.
