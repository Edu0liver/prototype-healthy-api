# Módulo: billing

Billing híbrido **tier + overage** (default hard-stop; overage opt-in por plano). Catálogo de planos, subscrição por company, metering de uso e enforcement de quotas. Fora do escopo v1 do PRD — expansão deliberada.

## Entidades (migrations 00017–00019)

| Tabela | RLS | Notas |
|--------|-----|-------|
| `plans` | ❌ system | Catálogo global (sem `company_id`). Quotas/caps `0 = ilimitado`; overage cents `0 = desativado` (hard-stop). Seed: free/starter/pro/enterprise. |
| `subscriptions` | ❌ system | 1 por company (UNIQUE), ligada a `plans` + ids Stripe + período. Gateway webhook escreve sem contexto tenant. Backfill criou p/ companies existentes. |
| `usage_events` | ✅ tenant | Ledger durável por tenant. Escrito pelo meter worker (db.Tenant). Kinds: `ai_message`,`llm_tokens`,`audio_minutes`,`storage_mb`,`embedding_tokens`. |

## Serviço

- **`Limits(companyID)`** — plano+período flatten, cacheado em Redis (`billing:limits:{id}`, TTL 5m). Lê via `db.System` (plans/subscriptions fora da RLS) → funciona de qualquer contexto.
- **`EnsureResource(companyID, resource)`** — hard cap em create-time (`channels`/`agents`/`knowledge_bases`/`seats`). Conta tenant-scoped vs `max_*`. `>=` cap → `ErrQuotaExceeded` (**HTTP 402**). Sem subscription → fail-open.
- **`CheckUsage(companyID, kind)`** — soft limit. Lê contador quente Redis vs quota. Default hard-stop; overage opt-in se `overage_*>0`. Erro/sem subscription → fail-open (Allowed).
- **`Record(Event)`** — best-effort: `INCRBY` contador `usage:{id}:{period}:{kind}` (TTL até fim do período) + enqueue outbox stream `stream:billing-usage`. Chamar em goroutine com `context.WithoutCancel`.
- **MeterWorker** — consome o outbox (group `billing-meters`) → `usage_events` (db.Tenant). At-least-once; ack sempre (contador já reflete; retry duplicaria).

## Enforcement (onde)

- **Hard limits (HTTP 402):** `channel.Create`, `agent.Create`, `knowledge.CreateKB` chamam `EnsureResource` via interface `QuotaGuard`/`Billing` (noop default + `WithBilling` no fx — testes sem billing).
- **Soft limits (worker):** `orchestration` pipeline chama `CheckUsage(ai_message)` antes do LLM; fora da quota e sem overage → `fallback_message`/mensagem de limite + mark-read, sem chamar OpenAI.

## Metering (hooks)

| Origem | Kind | Quantidade |
|--------|------|-----------|
| `orchestration` pós-Chat | `ai_message` / `llm_tokens` | 1 / `ChatResult.Usage.TotalTokens` |
| `orchestration` pós-transcrição | `audio_minutes` | estimativa por bytes (~2KB/s; refinar c/ duração do provedor) |
| `knowledge` pós-ingest | `embedding_tokens` / `storage_mb` | tokens estimados / MB do documento |

## Gateway Stripe (PASSO 5)

- **Cliente** `/pkg/stripe` (sem import de `internal/`): `CreateCheckoutSession` (form-encoded) + `VerifyWebhook` (assinatura `t=...,v1=...`, HMAC-SHA256 de `t.payload`, tolerância 5min, compare constante). `VerifySignature` é função pura (testada).
- **Checkout** (`CreateCheckout`): `POST /billing/checkout {plan_code}` → resolve `plans.stripe_price_id` → cria Checkout Session (mode=subscription, `client_reference_id`=companyID, metadata `plan_id`) → devolve `checkout_url`. Subscrição só persiste no callback.
- **Webhook** (`POST /webhooks/stripe`, público, system-scoped, idempotente por `event.id`):
  - `checkout.session.completed` → `ActivateSubscription` (upsert: plano + stripe ids + status active) + invalida cache.
  - `customer.subscription.updated` → status/period/cancel + plano (via `stripe_price_id`).
  - `customer.subscription.deleted` → `canceled`.
  - `invoice.payment_failed` → `past_due`.
- **Migration 00020**: `plans.stripe_price_id` (UNIQUE) + `stripe_product_id`.
- **Config** (`STRIPE_*`): chaves vazias → checkout/webhook desativados (`503 stripe_disabled`); catálogo continua a funcionar.

## Endpoints (admin, tenant-scoped — exceto webhook)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/billing/subscription` | Subscrição + plano do tenant |
| GET | `/billing/usage` | Consumo do período corrente vs quotas |
| POST | `/billing/checkout` | Inicia Checkout Session Stripe (devolve URL) |
| POST | `/webhooks/stripe` | **Público**, verificado por assinatura; lifecycle da subscrição |

## Follow-ups (não implementado)

- **Subscription no signup:** novas companies não têm subscription até checkout/backfill → enforcement fail-open. Criar subscription `free` no registo da company.
- **Seats:** `max_seats` não enforced (iam.Invite). Definir "seat" (operadores ativos) antes de gatear.
- **Stripe price ids:** popular `plans.stripe_price_id` (migration 00020 deixa NULL) com as Prices reais antes de habilitar checkout por tier.
- **Overage billing:** reportar excedente agregado ao Stripe (usage records / invoice items) no fecho do ciclo.
- **Billing portal:** sessão do Stripe Customer Portal p/ o cliente gerir/cancelar.
- **Áudio preciso:** estender `openai.Transcribe` p/ devolver duração (verbose_json) em vez de estimativa por bytes.
- **Embedding tokens exatos:** `openai.Embed` não devolve usage; hoje estima-se. Estender se precisão importar.
