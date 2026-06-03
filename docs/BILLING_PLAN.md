# Plano de Billing — Design (sem implementação)

> Expansão **fora do escopo v1** do PRD (§1.5). Documento de desenho para revisão antes de codar.
> Modelo de receita escolhido: **híbrido tier + overage** (assinatura com quota + cobrança por excedente) — protege margem do custo variável OpenAI/Whisper.

---

## 1. Visão geral

4 módulos novos + extensões em módulos existentes:

| # | Módulo | Responsabilidade |
|---|--------|------------------|
| 1 | `plan` | Catálogo global de planos (tiers, preços, quotas). Read-only p/ tenant. |
| 2 | `subscription` | Assinatura por company (estado, ciclo, liga a plano + Stripe). |
| 3 | `metering` | Registo de uso (`usage_events`) + agregação por período. |
| 4 | `quota` | Enforcement (middleware/check) — bloqueia ou degrada ao bater limite. |

Gateway (Stripe) entra dentro de `subscription` (webhooks + sync).

---

## 2. Modelagem de dados

### 2.1 `plans` (catálogo global — SEM `company_id`, FORA da RLS)

Reference table, gerida por super-admin. Padrão = tabelas de registo (companies/branding) que já ficam fora da RLS.

| Coluna | Tipo | Notas |
|--------|------|-------|
| id | uuid PK | |
| code | text UNIQUE | `starter`, `pro`, `enterprise` |
| name | text | |
| price_cents | int | preço base/mês |
| currency | text | `BRL`/`USD` |
| quota_ai_messages | int | msgs IA/mês incluídas (0 = ilimitado) |
| quota_tokens | bigint | tokens LLM/mês incluídos |
| quota_audio_minutes | int | min Whisper/mês |
| quota_storage_mb | int | MB ingest RAG |
| max_channels | int | add-on natural |
| max_agents | int | |
| max_kb | int | bases conhecimento |
| max_seats | int | operadores (handover) |
| overage_per_msg_cents | int | excedente por msg IA |
| overage_per_1k_tokens_cents | int | excedente por 1k tokens |
| is_active | bool | catálogo visível |

### 2.2 `subscriptions` (1 por company — SYSTEM-scoped, FORA da RLS)

Gerida pela plataforma (Stripe webhooks correm sem contexto tenant). Tenant **lê** a sua via endpoint dedicado (system read filtrado por company_id). Espelha o padrão do registo de tenants.

| Coluna | Tipo | Notas |
|--------|------|-------|
| id | uuid PK | |
| company_id | uuid FK→companies UNIQUE | 1 ativa por company |
| plan_id | uuid FK→plans | |
| status | text | `trialing`,`active`,`past_due`,`canceled`,`suspended` |
| billing_cycle | text | `monthly`,`annual` |
| current_period_start | timestamptz | âncora do reset de quota |
| current_period_end | timestamptz | |
| stripe_customer_id | text | |
| stripe_subscription_id | text | |
| cancel_at_period_end | bool | |

> `companies.plan` (text, já existe) passa a ser **denormalização cache** do `plans.code` ativo — ou deprecia-se. Decidir: manter como cache (rápido p/ quota check) vs. join. **Recomendo cache** (evita join no hot-path).

### 2.3 `usage_events` (alto volume — SOB RLS, escrita tenant-scoped)

Escrito pelo worker de orquestração (já corre em `db.Tenant(ctx, companyID)`) e pelo knowledge ingest. Tenant lê o próprio uso (dashboard). Plataforma agrega via system scope p/ faturação.

| Coluna | Tipo | Notas |
|--------|------|-------|
| id | uuid PK | |
| company_id | uuid FK→companies | RLS |
| kind | text | `ai_message`,`llm_tokens`,`audio_minutes`,`storage_mb`,`embedding_tokens` |
| quantity | bigint | unidades (tokens, ms áudio, etc.) |
| conversation_id | uuid NULL | rastreio |
| agent_id | uuid NULL | custo por agente (KPI PRD §5.5) |
| model | text | `gpt-4o`, `whisper-1` — custo por modelo |
| metadata | jsonb | |
| created_at | timestamptz | particionar por mês quando volume justificar (PRD §5.2) |

Índices: `(company_id, created_at)`, `(company_id, kind, created_at)`.

### 2.4 Migrations

Próximas: `00017_plans.sql`, `00018_subscriptions.sql`, `00019_usage_events.sql`.
RLS: adicionar `usage_events` ao array da migration de policies (novo `CREATE POLICY tenant_isolation`); `plans`/`subscriptions` ficam de fora (documentar no comentário, como já se faz p/ companies).

---

## 3. Metering — pontos de hook

### 3.1 PRÉ-REQUISITO: expor token usage

`pkg/openai.ChatResult` **não** devolve usage hoje. Necessário estender:

```go
type Usage struct{ PromptTokens, CompletionTokens, TotalTokens int }
type ChatResult struct {
    Content   string
    ToolCalls []ToolCall
    Usage     Usage // novo — parsear o campo "usage" da resposta OpenAI
}
```

Embed/Transcribe idem (Embed → tokens; Transcribe → não dá minutos, calcular pela duração do áudio decodificado ou tamanho).

### 3.2 Hooks

| Ponto | Evento | Onde |
|-------|--------|------|
| Após `oa.Chat` | `ai_message` (qty 1) + `llm_tokens` (Usage.Total) | `orchestration/service/pipeline.go` |
| Após `transcribe` | `audio_minutes` (duração) | `pipeline.go` |
| Após `oa.Embed` no ingest | `embedding_tokens` + `storage_mb` | `knowledge` ingest |

Escrita: **assíncrona/best-effort** — nunca bloquear/falhar a resposta ao cliente por causa de metering. Fire-and-forget com log de erro, ou bufferizar em Redis e flush em batch (preferível p/ não onerar hot-path).

### 3.3 Contador quente (Redis)

Espelho rápido p/ quota check sem agregar Postgres a cada msg:
`usage:company:{id}:{period}:{kind}` (INCRBY, TTL até fim do período). Fonte de verdade = `usage_events` (PG); Redis = cache de leitura quente, repovoável.

---

## 4. Quota enforcement

### 4.1 Estratégia

- **Hard limits** (max_channels/agents/kb/seats): verificados no `create_*` de cada módulo (channel, agent, knowledge, iam-invite). Check síncrono contra `subscriptions.plan` quota.
- **Soft/usage limits** (msgs/tokens/áudio): no **worker de orquestração** (não no middleware HTTP — a msg chega via webhook, não via painel). Ao bater quota:
  - **Política A (overage):** continua a responder, regista excedente p/ cobrança. ← recomendado p/ tier pago.
  - **Política B (hard stop):** envia `fallback_message` ("limite atingido"), não chama LLM. ← p/ trial/free.
  - Configurável por plano (`overage_*` definido → política A; senão B).

### 4.2 Onde no pipeline

Após o check de business hours, antes do RAG/LLM:

```
quota := quotaSvc.Check(ctx, companyID, kindAIMessage)
if quota.Exceeded && !quota.OverageAllowed {
    sendFallback(limitMessage); markRead; return
}
// senão prossegue; metering regista (e marca overage se Exceeded)
```

### 4.3 Reset de período

`current_period_start/end` na subscription. Cron/job diário (ou lazy no primeiro evento do novo período) zera contadores Redis e abre novo bucket. Reusa infra de jobs existente.

---

## 5. Gateway (Stripe) — dentro de `subscription`

- **Checkout/upgrade:** endpoint cria Stripe Checkout Session → redirect. Webhook `checkout.session.completed` → ativa subscription.
- **Webhooks Stripe** (`/webhooks/stripe`, fora da RLS, validação de assinatura `Stripe-Signature`): `customer.subscription.updated/deleted`, `invoice.payment_failed` → atualiza `subscriptions.status` (`past_due`/`canceled`). Padrão idêntico ao webhook Evolution (idempotência + auditoria).
- **Overage billing:** report de uso agregado → Stripe usage records (metered billing) OU invoice item no fecho do ciclo.
- Segredo Stripe cifrado/secret manager (NFR §5.1, como apikeys Evolution).

---

## 6. Impacto em módulos existentes

| Módulo | Mudança |
|--------|---------|
| `pkg/openai` | `ChatResult.Usage` + parse do campo `usage` (pré-req metering) |
| `orchestration` | hooks metering (chat/whisper) + quota check no pipeline |
| `knowledge` | hooks metering (embed/storage) no ingest |
| `channel`/`agent`/`knowledge`/`iam` | hard-limit check no `create_*` |
| `tenant` | `companies.plan` vira cache do plano ativo; expor subscription no painel |
| `middleware` | (opcional) helper de quota reutilizando padrão `ratelimit:company` |
| `database/migrations` | 00017–00019 + RLS update |

---

## 7. Ordem de implementação sugerida

1. **`pkg/openai` Usage** (desbloqueia metering exato; mudança pequena, testável isolada).
2. **`plan` + `subscription`** (migrations 17–18, CRUD super-admin, read tenant). Sem isto não há quota a comparar.
3. **`metering`** (migration 19 + RLS, hooks no pipeline/knowledge, contador Redis). Começa a recolher dados — útil mesmo antes de cobrar.
4. **`quota`** (hard limits nos create_* + soft limit no worker).
5. **Stripe** (checkout + webhooks + overage). Último — depende de tudo acima.

> Recomendação: 1→2→3 dão um MVP de *observabilidade de uso* (dashboard de consumo por tenant) sem cobrança real. Cobrança (4→5) só depois de validar os números do metering contra a realidade.

---

## 8. Decisões em aberto (precisam de input)

- **`companies.plan`:** cache do code ativo (recomendado) vs. deprecar e fazer join sempre?
- **Política de overage default:** A (cobra excedente) ou B (hard stop) por tier?
- **Trial:** existe plano free/trial? Duração? Cartão obrigatório no signup?
- **Moeda/mercado:** só BRL ou multi-moeda desde já?
- **Metering de áudio:** por minuto (duração real) ou por mensagem de áudio (flat)?
- **Stripe vs. alternativa** (ex.: Lemon Squeezy MoR p/ impostos internacionais, ou Asaas/Pagar.me p/ BR)?
