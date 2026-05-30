# API — Contexto para IA

## Produto

SaaS multi-tenant white-label em Go para atendimento automatizado com IA + RAG.
Cada empresa (`company`) tem usuários, canais de mensageria e base de conhecimento própria.

## Stack

- **Go 1.26** · **Gin v1.12** · **uber/fx v1.23** · **zap v1.27**
- **GORM v1.31** + **pgx/v5** · PostgreSQL + pgvector
- **goose v3** (migrations SQL)
- **go-redis/v9** · **jwt/v5** · **resend-go/v2**
- **validator/v10** + cpfcnpj · **ulule/limiter/v3**
- **UUID v7** para todas as PKs (`github.com/google/uuid v1.6.0`)

## Estrutura obrigatória de módulo
Um arquivo por caso de uso (ex.: `create_user.go`, `update_user.go`, `delete_user.go`) com o respectivo `_test.go` ao lado, em `dto/`, `http/` e `service/`.
```
/internal/modules/<nome>/
├── dto/            # 1 arquivo por caso de uso (request/response structs, json + validator tags)
│   ├── create_<x>.go
│   └── update_<x>.go
├── http/          # 1 handler por caso de uso + _test, routes.go, integration_test.go
│   ├── create_<x>.go
│   ├── create_<x>_test.go
│   ├── routes.go
│   └── integration_test.go
├── infra/
│   ├── models/    # GORM structs (sem json tags) = entidades
│   └── repository/  # <nome>_repository.go + <nome>_repository_test.go
├── service/        # 1 arquivo por caso de uso + _test + interfaces.go + errors.go
│   ├── create_<x>.go
│   └── create_<x>_test.go
├── module.go      # fx.Module — único ponto público do módulo para o fx
└── <NOME>.md      # Deve ser referênciado no CLAUDE.md após criação de todo módulo, ao invés de ser citado diretamente nele
```

## Camada transversal (Go standard layout)
fx em `internal/shared/module.go` (package `shared`). Providers mapeiam `config.Config` → opções de cada pacote `/pkg`.

- **`/pkg`** — libs genéricas, SEM import de `internal/` (reutilizáveis): `httputil`, `resilience` (retry+breaker), `metrics` (Prometheus), `crypto` (AES-256-GCM, `New(keyHex)`), `token` (JWT, `New(secret,ttls)`), `storage` (`NewLocal(root)`), `openai` (`New(openai.Config)`, retry), `evolution` (`New(evolution.Config)`, retry).
- **`/internal/shared`** — glue específico da app: `config`, `logger` (zap), `database` (GORM+pgx; `TenantScope`, tipos `JSONMap`/`JSONStringArray`/`Vector`), `redisx` (lock/stream/buffer/state), `appctx` (identidade no ctx), `mailer` (Resend), `events` (Pub/Sub realtime), `jobs` (`InboundJob`), `middleware` (auth/tenant-RLS/RBAC/rate-limit), `httpserver` (+swagger UI `/swagger`), `channeladapter`, `testsupport`.

> Regra: `/pkg` nunca importa `internal/`. Clientes que precisavam de `config` recebem structs próprias (ex.: `openai.Config`).

## Swagger
Anotações `// @...` nos handlers + info geral em `cmd/api/main.go`. `make swag` gera `docs/`. UI em `/docs/index.html`.

### Isolamento multi-tenant (OBRIGATÓRIO em todo módulo)
- **App-layer (primário):** toda query de domínio filtra `company_id`. Reads usam `database.MustTx(ctx).Scopes(database.TenantScope(ctx))`. Nunca executar query de domínio sem `company_id`.
- **RLS (defesa em profundidade):** a app conecta como `app_user` (NÃO-superuser) para as policies engajarem; migrations correm como superuser (`MIGRATE_DATABASE_URL`). Escopo definido via `SET LOCAL app.current_company_id` dentro da transação (`DB.Tenant` / `middleware.Tenant`).
- Registo de tenants (`companies`/`company_domains`/`company_branding`) e `webhook_events` ficam fora da RLS — usar `db.System`.

## Módulos
- [tenant](internal/modules/tenant/TENANT.md) — empresas + white-label (branding, domínios, Host→tenant).
- [iam](internal/modules/iam/IAM.md) — auth (JWT), RBAC, convites, gestão de utilizadores.
- [agent](internal/modules/agent/AGENT.md) — CRUD de agentes de IA (prompt, modelo, handover).
- [channel](internal/modules/channel/CHANNEL.md) — canais WhatsApp (Evolution) / Instagram; QR/pairing; estado.
- [automation](internal/modules/automation/AUTOMATION.md) — binding canal↔agente; invariante 1-agente-por-canal.
- [knowledge](internal/modules/knowledge/KNOWLEDGE.md) — RAG: bases, ingestão (chunk+embed), retrieval pgvector, agente↔base N:M.
- [conversation](internal/modules/conversation/CONVERSATION.md) — contactos/conversas/mensagens; histórico do painel; persistência do pipeline.
- [webhook](internal/modules/webhook/WEBHOOK.md) — receção Evolution: idempotência, routing, enqueue, handover passivo `fromMe`.
- [orchestration](internal/modules/orchestration/ORCHESTRATION.md) — worker: lock+debounce, áudio→Whisper, RAG, OpenAI+function-calling, humanização, envio.
- [handover](internal/modules/handover/HANDOVER.md) — controlo humano: take/reply/return/close (operador).
- [realtime](internal/modules/realtime/REALTIME.md) — WebSocket `/ws` via Redis Pub/Sub (`platform/events`).

---

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
| ------ | ---------- |
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.
