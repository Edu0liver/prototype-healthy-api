# Módulo: tenant

Gestão de empresas (tenants) e white-label.

## Responsabilidades
- CRUD de `companies` (signup público) — RF-WL.
- `company_branding` (tema white-label) — RF-WL-01.
- `company_domains` + resolução Host→tenant — RF-WL-02.

## Endpoints
| Método | Rota | Auth | Descrição |
| --- | --- | --- | --- |
| POST | `/companies` | público | Signup de tenant (cria company + branding default) |
| GET | `/branding?host=` | público | Tema white-label por domínio (runtime, sem rebuild) |
| GET | `/company` | admin | Dados da própria empresa |
| PUT | `/branding` | admin | Atualiza logo/cores/favicon/remetente |
| POST | `/domains` | admin | Regista domínio personalizado |
| GET | `/domains` | admin | Lista domínios |

## Notas de design
- Tabelas `companies`/`company_branding`/`company_domains` **não** estão sob RLS (registo de tenants, lido pré-auth pelo resolver de Host e pelo endpoint público de branding). Isolamento aqui é pela chave de lookup.
- Operações públicas usam `db.System` (sem escopo de tenant); operações autenticadas usam a transação de tenant da request.
- PKs UUID v7.
