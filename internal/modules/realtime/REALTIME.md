# Módulo: realtime

Push de eventos em tempo real para o painel via WebSocket (RF-LOG-01).

## Endpoint
| Método | Rota | Auth | Descrição |
| --- | --- | --- | --- |
| GET | `/ws?token=<access_jwt>` | token na query (browsers não definem headers em WS) | Stream de eventos do tenant |

## Arquitetura (desacoplada via Redis Pub/Sub)
- Produtores (`conversation.Service` em `AppendMessage`/`SetState`) publicam em `rt:{company_id}` através de `platform/events.Publisher`.
- Este módulo autentica o token → `company_id`, faz `SUBSCRIBE rt:{company_id}` e reencaminha cada mensagem para o socket. Read-pump deteta desconexão.

## Eventos
- `{"type":"message","conversation_id":..,"payload":{id,direction,sender_type,content,status,created_at}}`
- `{"type":"state","conversation_id":..,"payload":{state}}`

Verificado: WS recebeu `state:human → state:ai → state:closed` durante take/return/close.
