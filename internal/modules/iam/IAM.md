# Módulo: iam

Autenticação, autorização (RBAC) e gestão de utilizadores.

## Responsabilidades
- Bootstrap do primeiro admin (`POST /auth/register`, só com 0 utilizadores).
- Login multi-tenant (resolvido por `company_slug`) com JWT access + refresh — RF-USR.
- Convite por e-mail + aceitação (define password) — RF-USR-01.
- RBAC: `admin` | `operator` | `knowledge_manager` — RF-USR-02.

## Endpoints
| Método | Rota | Auth | Descrição |
| --- | --- | --- | --- |
| POST | `/auth/register` | público | Cria 1º admin da empresa (bootstrap) |
| POST | `/auth/login` | público | Login → tokens (access+refresh) |
| POST | `/auth/refresh` | público | Renova access via refresh token |
| POST | `/auth/accept-invite` | público | Define password de utilizador convidado |
| GET | `/auth/me` | autenticado | Utilizador atual |
| POST | `/users` | admin | Convida utilizador com role |
| GET | `/users` | admin | Lista utilizadores do tenant |

## Notas de design
- Password: **argon2id** (PHC encoded), comparação em tempo constante.
- JWT HS256 via `platform/token`; tipos `access`/`refresh`/`invite`. Convite é um JWT assinado (sem coluna na BD).
- Login resolve `company_id` por slug (`db.System`, tabela `companies` sem RLS), depois abre escopo de tenant para ler `users` (sob RLS + filtro app-layer).
- Isolamento: `users` tem RLS (FORCE) + filtro `company_id` no app-layer (`database.TenantScope`). `email` é único **por empresa** — toda query por email filtra company_id. Ver teste `http/integration_test.go` (TestTenantIsolation).
