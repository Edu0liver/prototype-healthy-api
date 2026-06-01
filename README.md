# API

SaaS multi-tenant white-label em Go com atendimento automatizado por IA (RAG + handover), canais WhatsApp (Evolution API V2) / Instagram.

## Pré-requisitos

- Go 1.26+
- Docker + Docker Compose (sobe Postgres+pgvector, Redis, Evolution e MinIO/S3)
- `goose` — `go install github.com/pressly/goose/v3/cmd/goose@latest`
- (opcional) `swag` p/ regenerar docs — `go install github.com/swaggo/swag/cmd/swag@latest`

## Rodando localmente

```bash
# 1. Copie e configure as variáveis de ambiente
cp .env.example .env
# Única var OBRIGATÓRIA p/ subir: JWT_SECRET (já preenchida no exemplo).
# A ligação ao Postgres é montada a partir de PG_HOST/PG_PORT/PG_USER/PG_PASSWORD/PG_DB
# (defaults já apontam p/ o app_user do compose). MIGRATE_DATABASE_URL (superuser) é
# usada só pelas migrations.
# OPENAI_API_KEY / EVOLUTION_API_KEY são opcionais — sem eles a app sobe,
# mas o pipeline de IA e o envio via Evolution ficam inativos.
# STORAGE_* já aponta p/ o MinIO do compose (localhost:9000, minioadmin).

# 2. Suba só a infra (Postgres + Redis + Evolution + MinIO) — NÃO o container da api
#    O MinIO é obrigatório: a app verifica/cria o bucket no arranque e NÃO sobe sem ele.
make up-db

# 3. Instale dependências
make deps

# 4. Rode as migrations (correm como superuser via MIGRATE_DATABASE_URL;
#    criam extensões, tabelas, RLS e o role app_user usado pela app)
make migrate-up

# 5. Inicie o servidor (no host, porta 8080)
make run
```

O servidor sobe em `http://localhost:8080`.

> **Atenção:** use `make up-db` (não `make up`) durante o desenvolvimento local.
> `make up` sobe também o container `api` em :8080 e colide com o `make run` do host.
> Para rodar **tudo em containers**, use `make up` e rode as migrations a partir do
> host (`make migrate-up`) antes — o `app_user` só existe depois delas.

### Endpoints utilitários

| Endpoint | Descrição |
| --- | --- |
| `GET /health` | Liveness. |
| `GET /metrics` | Métricas Prometheus. |
| `GET /docs/index.html` | UI do Swagger (gere com `make swag`). |
| `GET /ws` | WebSocket de logs em tempo real (token JWT). |

> Console do MinIO em `http://localhost:9001` (`minioadmin` / `minioadmin`); S3 API em `:9000`.

## Comandos úteis (Makefile)

| Comando | Ação |
| --- | --- |
| `make up-db` | Sobe Postgres + Redis + Evolution + MinIO (sem a api). |
| `make up` / `make down` | Sobe / derruba todo o compose. |
| `make migrate-up` / `migrate-down` / `migrate-status` | Migrations (goose). |
| `make seed` | Insere tenant demo (company + admin + agente + KB + canal). |
| `make run` | Roda a api no host. |
| `make test` | Testes com `-race`. |
| `make lint` | `go vet` + verificação `gofmt`. |
| `make swag` | Regenera `docs/` do Swagger. |
| `make ci` | `lint` + `swag-check` + `build` + `test`. |

---

## Como Executar e Testar

Fluxo de ponta a ponta para validar a automação localmente. Pressupõe os passos 1–5
de *Rodando localmente* concluídos (infra de pé, migrations aplicadas, `make run` ativo).

### 1. Popule um tenant de demonstração

```bash
make seed
```

Cria o tenant:

| Campo | Valor |
| --- | --- |
| company slug | `demo` |
| admin (email / senha) | `admin@demo.com` / `demodemo` |

> Sem o seed, crie manualmente: `POST /companies` → `POST /auth/register` → `POST /auth/login`.

### 2. Smoke test — saúde da API

```bash
curl -s http://localhost:8080/health
# {"status":"ok"}
```

### 3. Autenticação (obtenha o token)

```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.com","password":"demodemo"}'
```

Resposta inclui `access_token`. Guarde-o:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"company_slug":"demo","email":"admin@demo.com","password":"demodemo"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')
```

> Todas as rotas de domínio exigem `Authorization: Bearer $TOKEN`. O `company_id` é
> derivado do token e injetado em todas as queries (isolamento multi-tenant + RLS).

### 4. Criar um agente de IA (RF-AG-01)

```bash
curl -s -X POST http://localhost:8080/agents \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
        "name": "Suporte",
        "system_prompt": "Você é um atendente prestável da Demo Co.",
        "model": "gpt-4o-mini",
        "temperature": 0.7,
        "max_output_tokens": 1024,
        "handover_enabled": true
      }'
```

### 5. Criar base de conhecimento + ingerir texto (RF-RAG-01/02/03)

```bash
# cria a base
KB=$(curl -s -X POST http://localhost:8080/knowledge-bases \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"FAQ","embedding_model":"text-embedding-3-small","chunk_size":800,"chunk_overlap":100}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])')

# ingere texto colado (chunk + embedding assíncrono — requer OPENAI_API_KEY)
curl -s -X POST http://localhost:8080/knowledge-bases/$KB/documents/text \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"filename":"horario.txt","content":"A Demo Co atende de segunda a sexta, das 9h às 18h."}'
```

### 6. Canal + automação (binding canal↔agente, RF-CH-03)

```bash
# cria canal WhatsApp
CH=$(curl -s -X POST http://localhost:8080/channels \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"type":"whatsapp","name":"Atendimento"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])')

# liga via Evolution (gera QR / pairing — requer Evolution configurada)
curl -s -X POST http://localhost:8080/channels/$CH/connect \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'

# vincula UM agente ao canal (a partial unique index garante 1 ativa por canal)
curl -s -X POST http://localhost:8080/automations \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"channel_id\":\"$CH\",\"agent_id\":\"<AGENT_ID>\",\"is_active\":true}"
```

### 7. Simular mensagem inbound (webhook Evolution → pipeline)

Sem um WhatsApp real, dispare o webhook manualmente para exercitar idempotência →
persistência → enqueue → worker (lock + debounce + RAG + OpenAI + envio):

```bash
curl -s -X POST http://localhost:8080/webhooks/evolution \
  -H 'Content-Type: application/json' \
  -H 'authorization: Bearer change-me' \
  -d '{
        "event": "messages.upsert",
        "instance": "<EVOLUTION_INSTANCE_NAME>",
        "data": {
          "key": { "id": "MSG-001", "remoteJid": "5511999999999@s.whatsapp.net", "fromMe": false },
          "pushName": "Cliente Teste",
          "message": { "conversation": "Qual o horário de atendimento?" }
        }
      }'
```

> O token do header é o `EVOLUTION_WEBHOOK_TOKEN` do `.env`. `data.key.id` é a chave de
> idempotência (reenviar o mesmo id é descartado). A resposta da IA só sai se
> `OPENAI_API_KEY` e o canal Evolution estiverem ativos.

### 8. Inspecionar conversa, mensagens e handover

```bash
# lista conversas do tenant (RF-LOG-03)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/conversations

# histórico de mensagens (RF-LOG-02)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/conversations/<CONV_ID>/messages

# operador assume → responde → devolve à IA → fecha (RF-HO-01..04)
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/conversations/<CONV_ID>/handover/take
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"content":"Olá, sou um atendente humano."}' \
  http://localhost:8080/conversations/<CONV_ID>/handover/reply
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/conversations/<CONV_ID>/handover/return
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/conversations/<CONV_ID>/handover/close
```

Em estado `human`, a IA **não** responde automaticamente (RF-HO-02): as mensagens
ficam registadas e visíveis em tempo real via `GET /ws`.

### 9. Logs em tempo real (WebSocket)

```bash
# requer um cliente WS (ex.: websocat); autentique com o access_token
websocat "ws://localhost:8080/ws?token=$TOKEN"
```

### Testes automatizados

```bash
make test        # unitários + integração (-race); Postgres em container via testsupport
make lint        # go vet + gofmt
make ci          # lint + swag-check + build + test (espelha o pipeline)
```
