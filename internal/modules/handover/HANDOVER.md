# Módulo: handover

Controlo do atendimento humano (transição IA ↔ humano). Estado vive no Redis (operacional) espelhado no Postgres.

## Endpoints (admin + operator, sob tenant tx)
| Método | Rota | Descrição |
| --- | --- | --- |
| POST | `/conversations/:id/handover/take` | Operador assume → estado `human`, atribui-se, `block:conv` (RF-HO-03) |
| POST | `/conversations/:id/handover/reply` | Envia mensagem do operador via adapter, persiste `human/outbound` (RF-HO-03) |
| POST | `/conversations/:id/handover/return` | Devolve à IA → `ai`, desbloqueia (RF-HO-04) |
| POST | `/conversations/:id/handover/close` | Fecha → `closed`, desbloqueia |

## Notas
- `take`/`return`/`close` atualizam `conv:state` + `conversations.state` e publicam evento realtime (via `conversation.Service.SetState`).
- `reply` exige estado `human` (senão `409`); resolve instância+apikey+remoteJid por join (repo próprio), decifra apikey, envia, refaz `block`.
- Handover **ativo** (function-calling/keyword) é despoletado pelo worker `orchestration`; handover **passivo** (`fromMe`) pelo `webhook`. Este módulo é o controlo manual pelo operador.
