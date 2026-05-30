# Módulo: channel

Canais de mensageria (WhatsApp via Evolution API V2; Instagram via adapter stub).

## Endpoints (admin, sob tenant tx)
| Método | Rota | Descrição |
| --- | --- | --- |
| POST | `/channels` | Cria canal; p/ WhatsApp provisiona instância Evolution (RF-CH-01) |
| GET | `/channels` / `/channels/:id` | Lista / detalhe |
| POST | `/channels/:id/connect` | QR (default) ou Pairing Code (`{"method":"pairing"}`) (RF-CH-01/02) |
| GET | `/channels/:id/connection-state` | Sincroniza estado com Evolution (RF-CH-04) |
| DELETE | `/channels/:id` | Logout + delete da instância → `disconnected` (RF-CH-05) |

## Notas
- `evolution_apikey_enc`: apikey da instância cifrada at-rest (AES-GCM). Coluna mapeada via tag `gorm:"column:evolution_apikey_enc"` (GORM derivaria `evolution_api_key_enc`).
- `instanceName = lumia-<channel_id>`. Webhook configurado na criação (URL/token de `config.Evolution`).
- Estados: `disconnected|connecting|connected|error` (map de `open/connecting/close`).
- `CONNECTION_UPDATE` em runtime será tratado pelo módulo `webhook` (M4).
- `active_agent_id` é reflexo escrito pelo módulo `automation`.
