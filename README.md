# API

SaaS multi-tenant white-label em Go com atendimento automatizado por IA (RAG + handover), canais WhatsApp (Evolution API V2) / Instagram, com billing híbrido (tier + overage).

## Módulos

Arquitetura modular (fx): cada módulo é um *bounded context* com `dto/` · `http/` · `service/` · `infra/{models,repository}` · `module.go`. Detalhe em `internal/modules/<nome>/<NOME>.md`.

| Módulo | Responsabilidade |
| --- | --- |
| **tenant** | Empresas + white-label (branding, domínios, resolução Host→tenant). |
| **iam** | Auth (JWT access/refresh), RBAC, convites, gestão de utilizadores. |
| **agent** | CRUD de agentes de IA (prompt de sistema, modelo, handover). |
| **channel** | Canais WhatsApp (Evolution) / Instagram; QR/pairing; estado de ligação. |
| **automation** | Binding canal↔agente (invariante: 1 agente ativo por canal); horário; fallback. |
| **knowledge** | RAG: bases, ingestão (extract→chunk→embed), retrieval pgvector, agente↔base N:M. |
| **conversation** | Contactos / conversas / mensagens; histórico do painel; persistência do pipeline. |
| **webhook** | Receção Evolution: idempotência, routing instance→tenant, enqueue, handover passivo `fromMe`. |
| **orchestration** | Worker (sem HTTP): lock+debounce, áudio→Whisper, RAG, OpenAI+function-calling, humanização, envio. |
| **handover** | Controlo humano: take/reply/return/close (operador). |
| **realtime** | WebSocket `/ws` via Redis Pub/Sub (logs/eventos em tempo real). |
| **billing** | Planos, subscrição, metering (contador Redis + outbox), quotas (hard 402 / soft no worker), Stripe. |

Camada transversal em `internal/shared` (config, logger, database+RLS, redisx, middleware, mailer, events, httpserver) e libs genéricas em `/pkg` (`httputil`, `crypto`, `token`, `openai`, `evolution`, `storage`, `stripe`, `resilience`, `metrics`) — `/pkg` nunca importa `internal/`.

## Fluxo da aplicação

**Inbound (mensagem do cliente → resposta IA):**

1. Evolution dispara webhook `MESSAGES_UPSERT` → **webhook**: valida token, resolve `instance→channel→company` (fora da RLS), idempotência (`dedupe:msg` SETNX + UNIQUE), persiste inbound, **enfileira** `InboundJob` no Redis Stream.
2. **orchestration** worker consome: adquire lock por conversa (Redlock) → **debounce** (agrega fragmentos) → verifica estado (`human`/`block`/horário/**quota**) → áudio→Whisper → **RAG** (pgvector filtrado por tenant) → OpenAI (function-calling `transfer_to_human`) → humaniza (split ≤4) → envia via Evolution.
3. **billing** mede o uso (msgs, tokens, áudio) fora do hot-path (contador Redis + outbox → `usage_events`).

**Handover:** `fromMe` (resposta do telemóvel) suprime a IA por TTL (auto-retoma); operador no painel faz take/reply/return/close (estado `human` permanente até devolver).

**RAG ingest:** upload → storage (prefixo do tenant) → job assíncrono extract→chunk→embed → `document_chunks` (pgvector) → métrica de embedding/storage.

**Billing:** hard limits (canais/agentes/KBs) verificados no `create_*` → **HTTP 402**; soft limits (uso) no worker → `fallback` quando excede sem overage; Stripe Checkout + webhook (`/webhooks/stripe`) gerem o ciclo da subscrição.

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
# STRIPE_* são opcionais — vazias desativam checkout/webhook (HTTP 402 nas quotas
# continua a funcionar; subscrições podem ser provisionadas manualmente).

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
| `GET /health` | Liveness (root). |
| `GET /metrics` | Métricas Prometheus (root). |
| `GET /docs/index.html` | UI do Swagger (root; gere com `make swag`). |
| `GET /api/v1/ws` | WebSocket de logs em tempo real (token JWT). |

> **Base path:** rotas de domínio e auth ficam sob **`/api/v1`** (ex.: `/api/v1/auth/login`, `/api/v1/agents`). Apenas `/health`, `/metrics` e `/docs` são servidas na raiz.

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
| `make docker-build` | Build da imagem `lumia-api:$(VERSION)` (injeta version/commit/date). |

---

## Deploy (produção)

Fluxo padronizado para subir a stack completa em containers:

```bash
# 1. Configure .env de produção. Em produção a config exige (config.Load valida):
#    JWT_SECRET ≥ 32 bytes · ENCRYPTION_KEY (hex 64 chars / AES-256) · PG_SSLMODE != disable
cp .env.example .env   # edite os segredos

# 2. Migrations PRIMEIRO (correm como superuser via MIGRATE_DATABASE_URL;
#    criam tabelas, RLS e o role app_user que a app usa). O container `api` faz
#    restart até o app_user existir.
make migrate-up

# 3. Suba toda a stack em containers (postgres, redis, evolution, minio, api)
make up        # ou: make docker-build && docker compose up -d
```

Notas de produção:

- A app conecta como `app_user` (NÃO-superuser) para as **policies RLS engajarem**; migrations correm como superuser (`MIGRATE_DATABASE_URL`).
- Segredos at-rest (apikeys Evolution, Stripe) cifrados com `ENCRYPTION_KEY` — nunca commitar `.env`.
- Webhooks autenticados: `EVOLUTION_WEBHOOK_TOKEN` (Evolution) e `STRIPE_WEBHOOK_SECRET` (assinatura Stripe).
- Observabilidade: `GET /metrics` (Prometheus) + logs estruturados (zap) com `company_id`/`conversation_id`.
- Imagem versionada via `make docker-build` (tag = `APP_VERSION`).

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
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.com","password":"demodemo"}'
```

Resposta inclui `access_token`. Guarde-o:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"company_slug":"demo","email":"admin@demo.com","password":"demodemo"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')
```

> Todas as rotas de domínio exigem `Authorization: Bearer $TOKEN`. O `company_id` é
> derivado do token e injetado em todas as queries (isolamento multi-tenant + RLS).

### 4. Criar um agente de IA (RF-AG-01)

```bash
curl -s -X POST http://localhost:8080/api/v1/agents \
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
KB=$(curl -s -X POST http://localhost:8080/api/v1/knowledge-bases \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"FAQ","embedding_model":"text-embedding-3-small","chunk_size":800,"chunk_overlap":100}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])')

# ingere texto colado (chunk + embedding assíncrono — requer OPENAI_API_KEY)
curl -s -X POST http://localhost:8080/api/v1/knowledge-bases/$KB/documents/text \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"filename":"horario.txt","content":"A Demo Co atende de segunda a sexta, das 9h às 18h."}'
```

### 6. Canal + automação (binding canal↔agente, RF-CH-03)

```bash
# cria canal WhatsApp
CH=$(curl -s -X POST http://localhost:8080/api/v1/channels \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"type":"whatsapp","name":"Atendimento"}' \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["id"])')

# liga via Evolution (gera QR / pairing — requer Evolution configurada)
curl -s -X POST http://localhost:8080/api/v1/channels/$CH/connect \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{}'

# vincula UM agente ao canal (a partial unique index garante 1 ativa por canal)
curl -s -X POST http://localhost:8080/api/v1/automations \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"channel_id\":\"$CH\",\"agent_id\":\"<AGENT_ID>\",\"is_active\":true}"
```

### 7. Simular mensagem inbound (webhook Evolution → pipeline)

Sem um WhatsApp real, dispare o webhook manualmente para exercitar idempotência →
persistência → enqueue → worker (lock + debounce + RAG + OpenAI + envio):

```bash
curl -s -X POST http://localhost:8080/api/v1/webhooks/evolution \
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
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/conversations

# histórico de mensagens (RF-LOG-02)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/conversations/<CONV_ID>/messages

# operador assume → responde → devolve à IA → fecha (RF-HO-01..04)
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/conversations/<CONV_ID>/handover/take
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"content":"Olá, sou um atendente humano."}' \
  http://localhost:8080/api/v1/conversations/<CONV_ID>/handover/reply
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/conversations/<CONV_ID>/handover/return
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/conversations/<CONV_ID>/handover/close
```

Em estado `human`, a IA **não** responde automaticamente (RF-HO-02): as mensagens
ficam registadas e visíveis em tempo real via `GET /ws`.

### 9. Logs em tempo real (WebSocket)

```bash
# requer um cliente WS (ex.: websocat); autentique com o access_token
websocat "ws://localhost:8080/api/v1/ws?token=$TOKEN"
```

### 10. Billing — subscrição, uso e checkout

```bash
# subscrição do tenant (admin). 404 se a company foi criada depois das migrations
# (o backfill só cobre companies pré-existentes) — provisione via checkout/seed.
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/billing/subscription

# consumo do período corrente vs quotas do plano (msgs/tokens/áudio/storage)
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/billing/usage

# inicia checkout Stripe (devolve checkout_url). Requer STRIPE_SECRET_KEY +
# plans.stripe_price_id populado; sem isso → 409 plan_not_purchasable / 503.
curl -s -X POST http://localhost:8080/api/v1/billing/checkout \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"plan_code":"pro"}'
```

Enforcement de quotas:

- **Hard limits** (canais/agentes/bases): ao exceder o plano, o `create_*` devolve **HTTP 402** (`quota_exceeded`). Sem subscription → *fail-open* (não bloqueia).
- **Soft limits** (msgs/tokens IA): verificados no worker antes do LLM; ao exceder sem overage, a IA não responde e envia o `fallback_message`.

O webhook `POST /api/v1/webhooks/stripe` (público, verificado por assinatura `Stripe-Signature`) gere o ciclo: `checkout.session.completed` ativa a subscrição; `customer.subscription.updated/deleted` e `invoice.payment_failed` atualizam o estado.

### Testes automatizados

```bash
make test        # unitários + integração (-race); Postgres em container via testsupport
make lint        # go vet + gofmt
make ci          # lint + swag-check + build + test (espelha o pipeline)
```
