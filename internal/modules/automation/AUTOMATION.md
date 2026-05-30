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
- `business_hours` e `debounce_seconds` (default 8s) consumidos pelo worker de orquestração (M5).
