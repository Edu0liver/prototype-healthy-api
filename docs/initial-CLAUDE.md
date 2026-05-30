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
```
/internal/modules/<nome>/
├── dtos/           # request/response structs com json + validator tags
├── http/          # handlers Gin + routes.go + integration_test.go
├── infra/
│   ├── models/    # GORM structs (sem json tags) = entidades
│   └── repositories/
├── services/       # 1 arquivo por caso de uso + interfaces.go + errors.go
├── module.go      # fx.Module — único ponto público do módulo para o fx
└── <NOME>.md      # Deve ser referênciado no CLAUDE.md após criação de todo módulo, ao invés de ser citado diretamente nele
```

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
