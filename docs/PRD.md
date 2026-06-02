# PRD — Plataforma Multi-Tenant de Atendimento Automatizado e Omnicanal (White-Label)

> **Documento de Requisitos de Produto (Source of Truth)**
> Versão: 1.0 · Estado: Baseline para Desenvolvimento
> Stack de referência: **Golang · PostgreSQL + pgvector · Redis · OpenAI API · Evolution API V2**

---

## Índice

1. [Visão Geral e Objetivos do Produto](#1-visão-geral-e-objetivos-do-produto)
2. [Arquitetura do Sistema e Fluxo de Dados](#2-arquitetura-do-sistema-e-fluxo-de-dados)
3. [Modelagem de Dados (ERD)](#3-modelagem-de-dados-erd)
4. [Requisitos Funcionais Detalhados](#4-requisitos-funcionais-detalhados)
5. [Requisitos Não-Funcionais](#5-requisitos-não-funcionais)
6. [Mapeamento de Integrações (Evolution API V2)](#6-mapeamento-de-integrações-evolution-api-v2)
7. [Glossário e Anexos](#7-glossário-e-anexos)

---

## 1. Visão Geral e Objetivos do Produto

### 1.1 Proposta de Valor

A plataforma é um motor de atendimento conversacional **multi-tenant** e **white-label** que permite a qualquer empresa (designada por *Company* / *Tenant*) automatizar o seu atendimento ao cliente em canais de mensagens (WhatsApp e Instagram) recorrendo a agentes de Inteligência Artificial alimentados por modelos GPT da OpenAI, com capacidade de **RAG** (*Retrieval-Augmented Generation*) sobre bases de conhecimento próprias e **transição fluida para atendimento humano** (*handover*).

O produto resolve três problemas centrais:

- **Custo e escala do atendimento:** automatiza a primeira linha de contacto, respondendo com base no conhecimento curado pela própria empresa, reduzindo o volume que chega a operadores humanos.
- **Fragmentação de canais:** unifica WhatsApp e Instagram sob uma mesma camada de orquestração, conversas e auditoria.
- **Controlo e personalização:** cada tenant configura os seus próprios agentes (prompt de sistema), as suas bases de conhecimento e a sua identidade visual (white-label), sem partilhar dados nem configuração com outros tenants.

### 1.2 Personas

| Persona | Descrição | Necessidades-chave |
| --- | --- | --- |
| **Administrador da Empresa (Tenant Admin)** | Gere a conta da empresa, faturação, branding e utilizadores. | Configurar white-label, criar agentes, ligar canais, gerir bases RAG. |
| **Operador / Agente Humano** | Atende conversas que transitam da IA para humano. | Lista de conversas em espera, histórico completo, responder em tempo real. |
| **Gestor de Conhecimento** | Mantém as bases de conhecimento (RAG). | Upload de ficheiros, edição de textos, ver estado da indexação. |
| **Super-Admin (Operador da Plataforma)** | Equipa que opera o SaaS. | Gestão de tenants, observabilidade, limites e quotas globais. |
| **Contacto Final (End-User)** | O cliente da empresa que envia mensagens. | Respostas rápidas, corretas e, quando necessário, falar com um humano. |

### 1.3 Objetivos Macro (Outcomes)

- **O1 — Isolamento absoluto entre tenants.** Nenhum dado, configuração, conversa, embedding ou ficheiro de uma empresa é acessível por outra, em nenhuma camada.
- **O2 — Orquestração omnicanal coerente.** A mesma lógica de agente, RAG e handover aplica-se a WhatsApp e Instagram através de uma camada de abstração de canal.
- **O3 — Respostas de IA fundamentadas.** As respostas dos agentes são ancoradas no conhecimento do tenant (RAG), minimizando alucinações.
- **O4 — Handover sem fricção.** A transição IA → humano (e humano → IA) é instantânea e fiável, com a IA a cessar imediatamente as respostas automáticas quando a conversa está sob controlo humano.
- **O5 — Alto desempenho e concorrência.** O backend em Golang processa milhares de webhooks concorrentes com baixa latência e sem condições de corrida entre mensagens da mesma conversa.
- **O6 — White-label completo.** Domínio, logótipo e identidade visual configuráveis por tenant.

### 1.4 Métricas de Sucesso (KPIs)

| KPI | Definição | Alvo inicial |
| --- | --- | --- |
| Taxa de resolução automática | % de conversas fechadas sem handover humano | ≥ 60% |
| Latência de resposta da IA (p95) | Tempo entre receção da mensagem e envio da resposta | ≤ 6 s |
| Tempo até handover | Tempo entre o gatilho de handover e a notificação ao operador | ≤ 2 s |
| Fidelidade do RAG | % de respostas com citação correta da fonte (avaliação humana) | ≥ 90% |
| Disponibilidade da ligação ao canal | Uptime das instâncias de mensagens | ≥ 99,5% |
| Isolamento (incidentes de fuga entre tenants) | Nº de incidentes confirmados | 0 |

### 1.5 Âmbito (Scope)

**Dentro do âmbito (v1):**

- Gestão multi-tenant e white-label.
- Canais WhatsApp (via Evolution API V2) e Instagram (via adaptador — ver §6.7).
- Agentes GPT com prompt de sistema configurável.
- Bases de conhecimento RAG com pgvector.
- Handover IA ↔ humano gerido via Redis.
- Painel do tenant com logs em tempo real, gestão de RAG e configuração de agentes.

**Fora do âmbito (v1):**

- Faturação/billing automático e gateways de pagamento.
- Construtor visual de fluxos (flow builder *drag-and-drop*).
- Análise de sentimento avançada e relatórios de BI complexos.
- Canais adicionais (e-mail, SMS, Telegram, voz).

---

## 2. Arquitetura do Sistema e Fluxo de Dados

### 2.1 Modelo Conceptual de Alto Nível

A plataforma segue uma arquitetura de **serviços stateless em Golang** assentes em três pilares de persistência/estado:

- **PostgreSQL (+ pgvector):** *source of truth* relacional e armazenamento vetorial dos embeddings.
- **Redis:** estado efémero — sessões, *locks* distribuídos, estado de transição das conversas e *buffers* de mensagens.
- **OpenAI API:** geração de embeddings e respostas conversacionais (GPT).
- **Evolution API V2:** gateway de mensagens WhatsApp.

```txt
                         ┌──────────────────────────────────────────────┐
                         │              CLIENTES / CANAIS                 │
                         │   WhatsApp (Evolution API V2)   Instagram (Meta)│
                         └───────────────┬───────────────────────┬────────┘
                                         │ webhooks               │ webhooks
                                         ▼                        ▼
                  ┌───────────────────────────────────────────────────────┐
                  │                  API GATEWAY / INGRESS                  │
                  │        (TLS, routing por tenant, rate limiting)         │
                  └───────────────┬───────────────────────────┬────────────┘
                                  │                            │
                ┌─────────────────▼──────────┐   ┌─────────────▼─────────────┐
                │   WEBHOOK INGESTION SVC     │   │      ADMIN / BFF API       │
                │  (Golang, valida + enfileira)│   │ (painel do tenant, CRUD)  │
                └───────────────┬─────────────┘   └─────────────┬─────────────┘
                                │ enqueue (Redis Stream/fila)    │
                                ▼                                │
                ┌─────────────────────────────┐                 │
                │   ORCHESTRATION WORKERS      │                 │
                │  (pool de goroutines)        │                 │
                │  • debounce/aggregation      │                 │
                │  • lock distribuído por conv │                 │
                │  • verifica estado handover  │                 │
                │  • RAG retrieval             │                 │
                │  • chamada OpenAI            │                 │
                │  • envio via Evolution/Meta  │                 │
                └──┬───────────┬──────────┬────┘                 │
                   │           │          │                      │
          ┌────────▼──┐  ┌─────▼─────┐ ┌──▼─────────┐   ┌────────▼────────┐
          │  Redis    │  │ Postgres  │ │  OpenAI    │   │  WebSocket Hub   │
          │ (estado)  │  │ +pgvector │ │   API      │   │ (logs realtime)  │
          └───────────┘  └───────────┘ └────────────┘   └─────────────────┘
```

### 2.2 Serviços (Bounded Contexts)

| Serviço | Responsabilidade | Estado |
| --- | --- | --- |
| **Ingestion Service** | Receber webhooks da Evolution/Meta, validar assinatura/origem, garantir idempotência e enfileirar eventos. | Stateless |
| **Orchestration Worker** | Núcleo do processamento: aplica debounce, adquire lock, avalia estado de handover, executa RAG, chama a OpenAI e despacha a resposta. | Stateless (estado em Redis/PG) |
| **Admin/BFF API** | API REST/JSON para o painel do tenant: CRUD de agentes, canais, automações, RAG, utilizadores, branding. | Stateless |
| **RAG Ingestion Service** | Processa uploads: extração de texto, *chunking*, geração de embeddings e escrita em pgvector. | Stateless (jobs assíncronos) |
| **Realtime Service** | Servidor WebSocket que faz *push* de logs de conversa e eventos de estado ao painel. | Stateful (ligações) |

Todos os serviços são **horizontalmente escaláveis**; o estado partilhado vive em Redis e PostgreSQL, nunca na memória do processo.

### 2.3 Estratégia de Isolamento Multi-Tenant

O isolamento é **lógico** (base de dados partilhada, schema partilhado, discriminador por `company_id`) reforçado em três camadas:

1. **Camada de aplicação (obrigatória):** todo o acesso a dados passa por um repositório que injeta `company_id` em **todas** as queries. O `company_id` é resolvido a partir do token de autenticação (painel) ou do mapeamento `instance → channel → company` (webhooks). Nenhuma query de domínio é executada sem `company_id` no `WHERE`.
2. **Camada de base de dados (defesa em profundidade):** ativação de **Row-Level Security (RLS)** no PostgreSQL. Cada sessão de BD define `SET app.current_company_id = '<uuid>'`, e as *policies* RLS filtram automaticamente as linhas:

   ```sql
   ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
   CREATE POLICY tenant_isolation ON conversations
     USING (company_id = current_setting('app.current_company_id')::uuid);
   ```

3. **Camada de armazenamento de ficheiros:** os objetos (uploads RAG, logótipos) são prefixados por `tenant/<company_id>/...` no bucket, com políticas de acesso por prefixo.

> **Regra invariante:** qualquer endpoint, worker ou job que toque em dados de domínio **tem** de ter um `company_id` resolvido e validado antes de qualquer operação de leitura/escrita. Falha na resolução = rejeição imediata (`403`/descarte do evento).

### 2.4 White-Label

- Cada `Company` possui um registo de *branding*: `custom_domain`, `logo_url`, `primary_color`, `secondary_color`, `favicon_url`, `email_sender_name`.
- O *resolver* de domínio mapeia o `Host` HTTP recebido → `company_id` (tabela `company_domains`), permitindo que cada tenant sirva o painel sob o seu próprio domínio/subdomínio.
- O frontend consome um endpoint público `GET /branding?host=<domain>` que devolve o tema, aplicado em *runtime* (sem rebuild).

### 2.5 Concorrência em Golang

O desenho de concorrência é central ao desempenho. Princípios:

- **Pool de workers fixo + filas:** os webhooks são consumidos por um *worker pool* dimensionado (configurável via `WORKER_CONCURRENCY`). Usa-se um *Redis Stream* (ou fila) como buffer durável entre ingestão e processamento, garantindo *back-pressure* e durabilidade.
- **`context.Context` em toda a cadeia:** cada evento carrega um `context` com *timeout* e *cancellation*, propagado às chamadas a Postgres, Redis, OpenAI e Evolution.
- **Lock distribuído por conversa:** antes de processar mensagens de uma conversa, o worker adquire um lock `lock:conv:{conversation_id}` (via Redis, padrão *Redlock*). Isto serializa o processamento por conversa — evita respostas duplicadas e condições de corrida quando o contacto envia várias mensagens em rajada.
- **Debounce / agregação de mensagens:** mensagens recebidas dentro de uma janela curta (ex.: 3–8 s, configurável) para a mesma conversa são agregadas num único *prompt* (campo `buffer:conv:{id}`), evitando responder a cada fragmento separadamente.
- **`errgroup` para fan-out paralelo:** operações independentes (ex.: *retrieval* em múltiplas bases RAG, *fetch* de perfil do contacto) executam em paralelo com `golang.org/x/sync/errgroup`, com cancelamento agregado em caso de erro.
- **Connection pooling:** `pgxpool` para PostgreSQL e um pool de clientes Redis; HTTP clients com *keep-alive* e *timeouts* explícitos para OpenAI e Evolution.

### 2.6 Fluxo de Dados — Mensagem Recebida (Inbound)

```txt
1.  Contacto envia mensagem no WhatsApp.
2.  Evolution API V2 dispara webhook MESSAGES_UPSERT → Ingestion Service.
3.  Ingestion valida origem (apikey/header), extrai `instance` do payload.
4.  Resolve instance → channel → company_id → automação ativa → agente.
5.  Idempotência: verifica `message_id` em Redis (SETNX). Se já visto → descarta.
6.  Persiste a mensagem inbound em `messages` (Postgres) e faz push via WebSocket.
7.  Enfileira evento de processamento (Redis Stream).
8.  Orchestration Worker consome o evento:
    a.  Adquire lock `lock:conv:{conversation_id}` (Redlock).
    b.  Lê estado em `conv:state:{id}`:
        - Se "human"  → NÃO responde. Apenas regista a mensagem. Liberta lock. FIM.
        - Se "closed" → reabre/aplica regras. (configurável)
        - Se "ai"     → continua.
    c.  Aplica debounce: agrega mensagens da janela.
    d.  RAG: gera embedding da pergunta → pesquisa pgvector (filtrada por
        company_id + knowledge_base_ids do agente) → top-K chunks.
    e.  Monta prompt: system_prompt do agente + contexto RAG + histórico recente.
    f.  Chama OpenAI (com function calling, incl. `transfer_to_human`).
    g.  Se a IA decidir transferir → muda estado para "human" e notifica operador.
        Caso contrário → envia a resposta via Evolution `POST /message/sendText/{instance}`.
    h.  Persiste a resposta outbound em `messages` + push WebSocket.
    i.  Liberta o lock.
```

### 2.7 Fluxo de Dados — Ingestão de Conhecimento (RAG)

```txt
1.  Gestor faz upload de ficheiro/texto no painel → Admin API.
2.  Ficheiro guardado em storage (prefixo tenant/<company_id>/).
3.  Cria registo em `documents` (status = "pending").
4.  RAG Ingestion Service (job assíncrono):
    a.  Extrai texto (PDF, DOCX, TXT, MD, HTML).
    b.  Faz chunking (ex.: ~500–1000 tokens, com overlap).
    c.  Para cada chunk → gera embedding (OpenAI text-embedding-3-small, 1536 dims).
    d.  Insere em `document_chunks` (company_id, knowledge_base_id, embedding, content).
    e.  Atualiza `documents.status` = "indexed" (ou "failed" com erro).
5.  Push WebSocket informa o painel do progresso/estado.
```

---

## 3. Modelagem de Dados (ERD)

### 3.1 Diagrama de Entidades (textual)

```txt
companies ──1:N── users
companies ──1:N── company_domains
companies ──1:1── company_branding
companies ──1:N── channels
companies ──1:N── agents
companies ──1:N── automations
companies ──1:N── knowledge_bases
companies ──1:N── conversations

agents ──N:M── knowledge_bases        (via agent_knowledge_bases)
channels ──1:1(ativa)── automations    (1 automação ativa por canal)
automations ──N:1── agents             (uma automação aponta para 1 agente)
automations ──N:1── channels
knowledge_bases ──1:N── documents
documents ──1:N── document_chunks
channels ──1:N── conversations
conversations ──1:N── messages
conversations ──N:1── contacts
contacts ──N:1── channels
```

### 3.2 Restrições de Negócio Críticas (refletidas no schema)

- **Um único agente por canal de cada vez:** garantido por uma *unique constraint* parcial sobre automações ativas por canal — ver `automations` abaixo. Em alternativa simplificada, `channels.active_agent_id` é um único FK.
- **Agente ↔ Base de Conhecimento N:M:** tabela de junção `agent_knowledge_bases`.
- **Isolamento por tenant:** `company_id` (NOT NULL) presente em todas as tabelas de domínio, com FK para `companies` e RLS ativa.

### 3.3 Definição das Tabelas (PostgreSQL)

> Tipos: `uuid` (default `gen_random_uuid()`), `timestamptz`, `jsonb`. Todas as tabelas de domínio incluem `created_at`, `updated_at` e `company_id`.

#### `companies`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| name | text NOT NULL | |
| slug | text UNIQUE NOT NULL | identificador legível |
| status | text | `active`, `suspended` |
| plan | text | nível de subscrição |
| created_at / updated_at | timestamptz | |

#### `company_branding` (white-label)

| Coluna | Tipo | Notas |
| --- | --- | --- |
| company_id | uuid PK FK→companies | 1:1 |
| logo_url | text | |
| favicon_url | text | |
| primary_color | text | hex |
| secondary_color | text | hex |
| email_sender_name | text | |

#### `company_domains`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| domain | text UNIQUE NOT NULL | mapeamento Host→tenant |
| is_primary | boolean | |
| verified_at | timestamptz | |

#### `users`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| email | citext NOT NULL | UNIQUE por (company_id, email) |
| password_hash | text | argon2id |
| role | text | `admin`, `operator`, `knowledge_manager` |
| status | text | `active`, `invited`, `disabled` |

#### `channels`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| type | text NOT NULL | `whatsapp` \| `instagram` |
| name | text | rótulo amigável |
| evolution_instance_name | text | nome da instância na Evolution (WhatsApp) |
| evolution_instance_id | text | id devolvido pela Evolution |
| evolution_apikey_enc | text | apikey da instância (cifrada at-rest) |
| external_account_id | text | nº WhatsApp ou conta Instagram |
| status | text | `disconnected`, `connecting`, `connected`, `error` |
| active_agent_id | uuid FK→agents NULL | reflexo do agente associado (ver regra) |
| metadata | jsonb | |

*Constraint:* `UNIQUE (company_id, type, external_account_id)` para evitar duplicação do mesmo número/conta.

#### `agents`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| name | text NOT NULL | |
| system_prompt | text NOT NULL | instrução primária (ajuste fino pelo tenant) |
| model | text | ex.: `gpt-4o`, `gpt-4o-mini` |
| temperature | numeric(3,2) | |
| max_output_tokens | int | |
| handover_enabled | boolean | permite `transfer_to_human` |
| handover_keywords | text[] | gatilhos textuais opcionais |
| status | text | `active`, `draft` |

#### `automations`

Representa a ligação operacional **canal → agente** e respetivas regras (horário, *fallback*, gatilhos de handover). É o ponto onde a regra "um agente por canal" é imposta.

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| channel_id | uuid FK→channels | |
| agent_id | uuid FK→agents | |
| is_active | boolean NOT NULL | |
| business_hours | jsonb | janelas de funcionamento |
| fallback_message | text | mensagem quando IA falha |
| created_at / updated_at | timestamptz | |

*Constraint crítica (um agente ativo por canal):*

```sql
CREATE UNIQUE INDEX uniq_active_automation_per_channel
  ON automations (channel_id)
  WHERE is_active = true;
```

Esta *partial unique index* garante que, em qualquer momento, **existe no máximo uma automação ativa por canal** — logo, no máximo **um agente** a operar esse canal de cada vez.

#### `knowledge_bases`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| name | text NOT NULL | |
| description | text | |
| embedding_model | text | ex.: `text-embedding-3-small` |
| chunk_size | int | |
| chunk_overlap | int | |

#### `agent_knowledge_bases` (junção N:M)

| Coluna | Tipo | Notas |
| --- | --- | --- |
| agent_id | uuid FK→agents | PK composta |
| knowledge_base_id | uuid FK→knowledge_bases | PK composta |
| company_id | uuid FK→companies | redundância para RLS/integridade |

*PK:* `(agent_id, knowledge_base_id)`.

#### `documents`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| knowledge_base_id | uuid FK→knowledge_bases | |
| source_type | text | `file` \| `text` |
| filename | text | |
| storage_path | text | prefixo tenant/<company_id>/... |
| status | text | `pending`, `processing`, `indexed`, `failed` |
| error | text | |
| token_count | int | |

#### `document_chunks` (vetorial — pgvector)

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| knowledge_base_id | uuid FK→knowledge_bases | |
| document_id | uuid FK→documents | |
| chunk_index | int | ordem no documento |
| content | text NOT NULL | texto do chunk |
| embedding | `vector(1536)` | embedding OpenAI |
| metadata | jsonb | página, título, etc. |

#### `contacts`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| channel_id | uuid FK→channels | |
| remote_jid | text | ex.: `5511...@s.whatsapp.net` |
| push_name | text | |
| profile_pic_url | text | |

*Constraint:* `UNIQUE (channel_id, remote_jid)`.

#### `conversations`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| channel_id | uuid FK→channels | |
| contact_id | uuid FK→contacts | |
| agent_id | uuid FK→agents NULL | agente que serviu a conversa |
| state | text NOT NULL | `ai` \| `human` \| `closed` (espelho do Redis) |
| assigned_user_id | uuid FK→users NULL | operador em handover |
| last_message_at | timestamptz | |
| opened_at / closed_at | timestamptz | |

#### `messages`

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies | |
| conversation_id | uuid FK→conversations | |
| direction | text | `inbound` \| `outbound` |
| sender_type | text | `contact` \| `ai` \| `human` |
| content | text | |
| media | jsonb | url, mimetype, base64 ref |
| external_message_id | text | id da Evolution/Meta (idempotência) |
| status | text | `received`, `sent`, `delivered`, `read`, `failed` |
| created_at | timestamptz | |

*Constraint:* `UNIQUE (company_id, external_message_id)` quando não nulo.

#### `webhook_events` (auditoria/idempotência durável)

| Coluna | Tipo | Notas |
| --- | --- | --- |
| id | uuid PK | |
| company_id | uuid FK→companies NULL | |
| channel_id | uuid FK→channels NULL | |
| event_type | text | ex.: `MESSAGES_UPSERT` |
| external_id | text | id do evento/mensagem |
| payload | jsonb | corpo bruto recebido |
| processed_at | timestamptz | |

### 3.4 Índices

- **Vetorial (pgvector):** índice HNSW para similaridade por cosseno.

  ```sql
  CREATE INDEX idx_chunks_embedding_hnsw
    ON document_chunks
    USING hnsw (embedding vector_cosine_ops);
  ```

  Alternativa: `ivfflat` com `lists` ajustado ao volume; HNSW recomendado para *recall*/latência em cargas de leitura.
- **Filtragem tenant + RAG:** `CREATE INDEX ON document_chunks (company_id, knowledge_base_id);` — a pesquisa vetorial é sempre pré-filtrada por estas colunas para garantir isolamento e relevância.
- **Conversas/mensagens:** `CREATE INDEX ON messages (conversation_id, created_at DESC);` e `CREATE INDEX ON conversations (company_id, state, last_message_at DESC);`.
- **Resolução de canal:** `CREATE UNIQUE INDEX ON channels (company_id, evolution_instance_name);` e índice por `evolution_instance_name` para o *lookup* a partir do webhook.

> **Nota sobre pesquisa filtrada + HNSW:** para combinar filtro por `company_id`/`knowledge_base_id` com pesquisa vetorial mantendo *recall*, considerar índices parciais por base de conhecimento de grande volume, ou *pre-filtering* aplicado na query com `WHERE company_id = $1 AND knowledge_base_id = ANY($2)` antes do operador `<=>`.

### 3.5 Estruturas de Estado no Redis

| Chave | Tipo | TTL | Propósito |
| --- | --- | --- | --- |
| `conv:state:{conversation_id}` | string | — | Estado da conversa: `ai` \| `human` \| `closed`. **Fonte de verdade operacional** para o handover. |
| `lock:conv:{conversation_id}` | string (Redlock) | curto (ex.: 30 s) | Lock distribuído para serializar o processamento por conversa. |
| `buffer:conv:{conversation_id}` | list | janela debounce | Acumula fragmentos de mensagens antes de processar. |
| `dedupe:msg:{external_message_id}` | string (SETNX) | ~24 h | Idempotência de webhooks. |
| `channel:status:{channel_id}` | string | curto | Cache do estado de ligação (connected/connecting/...). |
| `session:{user_id}` | hash | sessão | Sessão do utilizador do painel (se aplicável). |
| `ratelimit:company:{company_id}` | counter | janela | Controlo de quota/rate por tenant. |
| `presence:typing:{conversation_id}` | string | curto | Controlo de envio de presença ("a escrever..."). |

> **Sincronização Redis ↔ Postgres:** o estado de handover vive primariamente em Redis (leitura quente, baixa latência) e é **espelhado** de forma assíncrona para `conversations.state` para auditoria e recuperação. Em caso de *cache miss* (Redis indisponível ou chave expirada), o worker recorre a `conversations.state` no Postgres como *fallback* e repovoa o Redis.

---

## 4. Requisitos Funcionais Detalhados

> Critérios de aceitação em formato **Dado / Quando / Então**.

### 4.1 Módulo: Gestão de Tenants e White-Label

**RF-WL-01 — Configuração de branding.**
O Tenant Admin pode definir logótipo, favicon, cores e nome do remetente.

- *Dado* que sou Tenant Admin autenticado,
- *Quando* atualizo o logótipo e as cores no painel,
- *Então* o painel passa a refletir o novo tema sem necessidade de *rebuild* e os valores persistem em `company_branding`.

**RF-WL-02 — Domínio personalizado.**

- *Dado* um domínio personalizado registado e verificado,
- *Quando* um utilizador acede via esse `Host`,
- *Então* o sistema resolve o `company_id` correto e serve o branding desse tenant.

**RF-WL-03 — Isolamento de dados.**

- *Dado* dois tenants A e B,
- *Quando* um utilizador de A consulta qualquer recurso (conversas, agentes, RAG),
- *Então* nunca são devolvidos dados de B, mesmo com manipulação de IDs (verificado por RLS + camada de aplicação).

### 4.2 Módulo: Canais

**RF-CH-01 — Criar canal WhatsApp e ligar via QR Code.**

- *Dado* que crio um canal do tipo `whatsapp`,
- *Quando* solicito a ligação,
- *Então* o sistema cria a instância na Evolution (`POST /instance/create`), obtém o QR Code (`GET /instance/connect/{instance}`) e exibe-o no painel; o estado transita para `connecting` e depois `connected` após leitura.

**RF-CH-02 — Ligar via Pairing Code.**

- *Dado* que indico o número de telefone do canal,
- *Quando* opto por *Pairing Code*,
- *Então* o sistema chama `GET /instance/connect/{instance}?number=<E164>` e exibe o `pairingCode` devolvido para introdução no telemóvel.

**RF-CH-03 — Associar um único agente ao canal.**

- *Dado* um canal ligado,
- *Quando* associo um agente (cria/ativa uma automação),
- *Então* o sistema garante que **não existe outra automação ativa** para esse canal; se já existir, a operação exige desativar a anterior (a *partial unique index* rejeita duas ativas).

**RF-CH-04 — Monitorizar estado de ligação.**

- *Dado* um canal,
- *Quando* a Evolution dispara `CONNECTION_UPDATE`,
- *Então* o estado do canal é atualizado em `channels.status` e em cache Redis, e refletido em tempo real no painel.

**RF-CH-05 — Desligar / remover canal.**

- *Quando* o Admin desliga o canal,
- *Então* o sistema chama o *logout*/*delete* da instância na Evolution e marca o canal como `disconnected`.

### 4.3 Módulo: Agentes

**RF-AG-01 — Criar/editar agente e prompt de sistema.**

- *Dado* que sou Admin,
- *Quando* defino `name`, `system_prompt`, `model`, `temperature`,
- *Então* o agente é persistido e fica disponível para associação a canais e bases RAG.

**RF-AG-02 — Associar múltiplas bases RAG ao agente (N:M).**

- *Dado* um agente,
- *Quando* seleciono várias bases de conhecimento,
- *Então* o agente consome, em simultâneo, os chunks de todas as bases associadas no *retrieval* (filtrado por `company_id`).

**RF-AG-03 — Ajuste fino do prompt em produção.**

- *Quando* edito o `system_prompt`,
- *Então* a alteração aplica-se às novas mensagens processadas (sem necessidade de redeploy).

### 4.4 Módulo: RAG (Bases de Conhecimento)

**RF-RAG-01 — Criar base de conhecimento.**

- *Quando* crio uma base,
- *Então* é persistida com `embedding_model`, `chunk_size`, `chunk_overlap`.

**RF-RAG-02 — Upload de ficheiros e textos.**

- *Dado* uma base,
- *Quando* faço upload de PDF/DOCX/TXT/MD ou colo texto,
- *Então* o ficheiro é guardado (prefixo do tenant), criado um `document` com `status=pending`, e iniciada a indexação assíncrona.

**RF-RAG-03 — Indexação (chunking + embeddings).**

- *Quando* a indexação corre,
- *Então* o texto é segmentado, cada chunk recebe um embedding via OpenAI e é inserido em `document_chunks`; no fim, `documents.status=indexed`. Em erro, `status=failed` com mensagem.

**RF-RAG-04 — Pesquisa semântica isolada por tenant.**

- *Dado* uma pergunta de um contacto,
- *Quando* o worker executa o *retrieval*,
- *Então* a pesquisa vetorial filtra **sempre** por `company_id` e pelas `knowledge_base_id` do agente, devolvendo os top-K chunks mais relevantes por similaridade de cosseno.

**RF-RAG-05 — Gestão do ciclo de vida.**

- *Quando* removo um documento,
- *Então* os respetivos `document_chunks` são eliminados e deixam de ser usados no *retrieval*.

### 4.5 Módulo: Handover (Transição para Humano)

**RF-HO-01 — IA transfere para humano.**

- *Dado* `handover_enabled = true`,
- *Quando* a IA deteta intenção de falar com humano (via *function calling* `transfer_to_human` ou *keyword* configurada),
- *Então* o estado da conversa passa a `human` em Redis (`conv:state:{id}`), um operador é notificado, e a IA **cessa imediatamente** qualquer resposta automática nessa conversa.

**RF-HO-02 — IA não responde em estado humano.**

- *Dado* uma conversa em estado `human`,
- *Quando* o contacto envia novas mensagens,
- *Então* o worker regista as mensagens mas **não** chama a OpenAI nem envia respostas automáticas; as mensagens aparecem em tempo real para o operador.

**RF-HO-03 — Operador assume e responde.**

- *Dado* uma conversa em `human`,
- *Quando* o operador envia uma mensagem pelo painel,
- *Então* a mensagem é despachada via Evolution e registada com `sender_type=human`.

**RF-HO-04 — Devolver à IA / fechar.**

- *Quando* o operador devolve a conversa à IA,
- *Então* o estado volta a `ai` e a automação retoma o atendimento; ao fechar, o estado passa a `closed`.

**RF-HO-05 — Consistência do estado.**

- *Dado* uma falha transitória do Redis,
- *Quando* o worker não encontra `conv:state:{id}`,
- *Então* recorre a `conversations.state` no Postgres como *fallback* e repovoa o Redis, garantindo que nunca responde automaticamente sobre uma conversa que estava em `human`.

### 4.6 Módulo: Logs e Auditoria em Tempo Real

**RF-LOG-01 — Visualização em tempo real.**

- *Dado* que estou no painel,
- *Quando* chegam/saem mensagens numa conversa,
- *Então* vejo-as em tempo real via WebSocket, com `sender_type` (contacto/IA/humano), timestamp e estado de entrega.

**RF-LOG-02 — Auditoria completa.**

- *Quando* abro o histórico de uma conversa,
- *Então* vejo a sequência completa e imutável de mensagens, incluindo eventos de handover e qual agente/operador atuou.

**RF-LOG-03 — Filtros.**

- *Quando* filtro por canal, estado (`ai`/`human`/`closed`) ou período,
- *Então* a lista de conversas reflete o filtro, sempre restrita ao meu tenant.

### 4.7 Módulo: Utilizadores e Permissões

**RF-USR-01 — Convidar utilizadores.**

- *Quando* o Admin convida por e-mail com um `role`,
- *Então* o utilizador é criado com `status=invited` e recebe acesso conforme o papel (`admin`, `operator`, `knowledge_manager`).

**RF-USR-02 — Controlo de acesso por papel.**

- *Então* operadores acedem a conversas/handover; gestores de conhecimento acedem a RAG; apenas admins gerem canais, agentes, branding e utilizadores.

---

## 5. Requisitos Não-Funcionais

### 5.1 Segurança

- **Isolamento multi-tenant** (ver §2.3): camada de aplicação + RLS no PostgreSQL + prefixação de storage. Testes automatizados de *tenant leakage* na pipeline de CI.
- **Autenticação e autorização:** JWT de curta duração + *refresh tokens*; RBAC por `role`. Todas as rotas do painel exigem `company_id` derivado do token.
- **Segredos at-rest:** apikeys das instâncias Evolution e chaves de integração cifradas (ex.: AES-GCM com chave gerida por KMS/secret manager). Nunca em logs.
- **Webhooks autenticados:** validação de origem dos webhooks (apikey/header partilhado + *allowlist* de IP quando possível) e verificação de assinatura para o canal Instagram (Meta).
- **Transporte:** TLS obrigatório em todos os endpoints; HSTS.
- **Proteção de dados pessoais (RGPD/LGPD):** dados de contactos e conversas tratados como pessoais; suporte a eliminação por pedido do titular e por tenant; retenção configurável.
- **Rate limiting e *abuse prevention*:** por tenant e por instância, em Redis.
- **Sanitização de prompt:** mitigação de *prompt injection* via separação clara de instruções de sistema, contexto RAG e input do utilizador; validação das respostas de *function calling*.

### 5.2 Escalabilidade

- **Serviços stateless** escaláveis horizontalmente atrás de *load balancer*; estado em Redis/Postgres.
- **Workers desacoplados** da ingestão via fila/stream durável (Redis Streams), permitindo escalar *consumers* de forma independente do *throughput* de webhooks.
- **PostgreSQL:** *connection pooling* (pgxpool + PgBouncer), réplicas de leitura para queries do painel e *retrieval* pesado; particionamento de `messages`/`document_chunks` por tenant ou por tempo quando o volume justificar.
- **pgvector:** índices HNSW dimensionados; possibilidade de mover *retrieval* para réplicas de leitura.
- **Caching:** estado de canal, branding e configurações de agente em cache com invalidação por evento.

### 5.3 Latência das Respostas de IA

- **Alvo:** p95 ≤ 6 s entre receção e envio da resposta (excluindo tempo de debounce intencional).
- **Técnicas:**
  - *Streaming* da resposta da OpenAI sempre que o canal o permita, reduzindo o tempo até ao primeiro envio.
  - *Retrieval* paralelo às operações preparatórias com `errgroup`.
  - Cache de embeddings de perguntas frequentes (opcional).
  - *Timeouts* explícitos por etapa com *fallback message* configurável (`automations.fallback_message`) em caso de falha/timeout da OpenAI.
- **Debounce** configurável por automação para equilibrar naturalidade vs. latência percebida.

### 5.4 Resiliência da Ligação à API de Mensagens

- **Retentativas com backoff exponencial + jitter** em chamadas à Evolution e à OpenAI, respeitando idempotência.
- **Circuit breaker** por instância/integração: em falhas sucessivas, abre o circuito e degrada graciosamente (enfileira para reenvio, alerta o operador).
- **Idempotência de webhooks:** `dedupe:msg:{external_message_id}` (Redis SETNX) + `UNIQUE (company_id, external_message_id)` em `messages` + tabela durável `webhook_events`.
- **Reconexão de instâncias:** ao receber `CONNECTION_UPDATE` com desligamento, o sistema sinaliza o canal como `disconnected` e, conforme política, tenta reconectar / solicita novo QR; notifica o Admin.
- **Dead-letter:** eventos que falham repetidamente vão para uma DLQ para inspeção manual, sem bloquear o pipeline.
- **Entrega de saída:** confirmação de envio via resposta da Evolution e atualização de `messages.status` com base nos eventos `SEND_MESSAGE`/`MESSAGES_UPDATE`.

### 5.5 Observabilidade

- **Logs estruturados** (JSON) com `company_id`, `conversation_id`, `trace_id` (sem dados sensíveis).
- **Métricas** (Prometheus): latência por etapa, taxa de handover, erros de integração, profundidade da fila, custo OpenAI por tenant.
- **Tracing distribuído** (OpenTelemetry) cobrindo ingestão → worker → OpenAI → Evolution.
- **Alertas:** instância desligada, fila a crescer, *circuit breaker* aberto, taxa de erro elevada.

### 5.6 Manutenibilidade e Qualidade

- **Código limpo em Go:** *layering* (handlers → services → repositories), interfaces para integrações externas (testabilidade/mocks), `context` propagado.
- **Migrações versionadas** (ex.: `golang-migrate`).
- **Testes:** unitários, de integração (Postgres/Redis em contentores), e testes de isolamento multi-tenant obrigatórios.

---

## 6. Mapeamento de Integrações (Evolution API V2)

> Base de referência: documentação oficial Evolution API V2 (`https://doc.evolution-api.com/v2/api-reference`). Autenticação por **API Key** no header `apikey`. `{{server-url}}` = URL do servidor Evolution; `{instance}` = nome da instância (mapeado para um `channel`).

### 6.1 Autenticação

Todas as chamadas autenticam com o header:

```txt
apikey: <API_KEY_GLOBAL_OU_DA_INSTANCIA>
Content-Type: application/json
```

A `apikey` de cada instância (devolvida na criação) é armazenada cifrada em `channels.evolution_apikey_enc`.

### 6.2 Criação de Instância (vincular canal WhatsApp)

**Endpoint:** `POST {{server-url}}/instance/create`

Payload (essencial):

```json
{
  "instanceName": "tenant-<company_slug>-<channel_id>",
  "integration": "WHATSAPP-BAILEYS",
  "qrcode": true,
  "number": "5511999999999",
  "webhook": {
    "url": "https://api.suaplataforma.com/api/v1/webhooks/evolution",
    "byEvents": false,
    "base64": true,
    "headers": { "authorization": "Bearer <token-interno>" },
    "events": [
      "QRCODE_UPDATED",
      "CONNECTION_UPDATE",
      "MESSAGES_UPSERT",
      "MESSAGES_UPDATE",
      "SEND_MESSAGE"
    ]
  }
}
```

Resposta relevante (`201`): `instance.instanceName`, `instance.instanceId`, `instance.status` e `hash.apikey` (apikey da instância).
**Ação no sistema:** persistir `evolution_instance_name`, `evolution_instance_id` e `evolution_apikey_enc` no `channel`.

> `integration` pode ser `WHATSAPP-BAILEYS` (QR/pairing) ou `WHATSAPP-BUSINESS` (Cloud API).

### 6.3 Ligação: QR Code e Pairing Code

**Endpoint:** `GET {{server-url}}/instance/connect/{instance}`

- **QR Code:** sem `number` → a resposta inclui `code` (conteúdo do QR) e `count`. O QR também chega de forma assíncrona via evento `QRCODE_UPDATED` (em base64).
- **Pairing Code:** com query `?number=<E164>` → a resposta inclui `pairingCode` (ex.: `WZYEH1YY`) para introdução manual no telemóvel.

Resposta (exemplo):

```json
{ "pairingCode": "WZYEH1YY", "code": "2@y8eK+bjtEjUWy9/FOM...", "count": 1 }
```

**Mapeamento de UI:** o painel apresenta o QR (render do `code`/base64) ou o `pairingCode`, conforme o método escolhido pelo utilizador (RF-CH-01 / RF-CH-02).

### 6.4 Estado da Ligação

**Endpoint:** `GET {{server-url}}/instance/connectionState/{instance}` — devolve o estado atual (ex.: `open`, `connecting`, `close`). Usado para *polling* de confirmação e para sincronizar `channels.status`. O estado também é empurrado via evento `CONNECTION_UPDATE`.

### 6.5 Configuração de Webhook (alternativa/posterior à criação)

**Endpoint:** `POST {{server-url}}/webhook/set/{instance}`

```json
{
  "enabled": true,
  "url": "https://api.suaplataforma.com/api/v1/webhooks/evolution",
  "webhookByEvents": false,
  "webhookBase64": true,
  "events": [
    "QRCODE_UPDATED",
    "CONNECTION_UPDATE",
    "MESSAGES_UPSERT",
    "MESSAGES_UPDATE",
    "SEND_MESSAGE"
  ]
}
```

Consulta: `GET {{server-url}}/webhook/find/{instance}`.

### 6.6 Receção de Mensagens (Webhook → Ingestion)

A Evolution faz `POST` para a `url` configurada. Eventos relevantes:

| Evento | Uso no sistema |
| --- | --- |
| `MESSAGES_UPSERT` | **Mensagem recebida** do contacto → cria/atualiza conversa, persiste mensagem, dispara o pipeline. |
| `CONNECTION_UPDATE` | Atualiza estado de ligação do canal. |
| `QRCODE_UPDATED` | Novo QR (base64) → push para o painel durante a ligação. |
| `SEND_MESSAGE` | Confirmação de envio → atualiza `messages.status`. |
| `MESSAGES_UPDATE` | Atualização de estado (delivered/read) → atualiza `messages.status`. |

Estrutura típica do *envelope* recebido (campos-chave a extrair): `instance` (resolve o canal/tenant), `data.key.id` (id da mensagem, para idempotência), `data.key.remoteJid` (contacto), `data.key.fromMe` (ignorar se `true` em `MESSAGES_UPSERT`), `data.pushName`, e o conteúdo em `data.message` (ex.: `conversation` ou `extendedTextMessage.text`; media em estruturas próprias, com base64 se `webhookBase64=true`).

> **Idempotência:** usar `data.key.id` como `external_message_id` (Redis SETNX + UNIQUE no Postgres).

### 6.7 Envio de Mensagens (Outbound)

| Tipo | Endpoint |
| --- | --- |
| Texto | `POST {{server-url}}/message/sendText/{instance}` |
| Mídia (imagem/doc/vídeo) | `POST {{server-url}}/message/sendMedia/{instance}` |
| Áudio | `POST {{server-url}}/message/sendWhatsAppAudio/{instance}` |
| Presença ("a escrever...") | `POST {{server-url}}/chat/sendPresence/{instance}` |
| Marcar como lida | `POST {{server-url}}/chat/markMessageAsRead/{instance}` |

Envio de texto:

```json
POST /message/sendText/{instance}
{
  "number": "5511999999999",
  "text": "Olá! Em que posso ajudar?",
  "delay": 1200,
  "linkPreview": false
}
```

Resposta (`201`) inclui `key.id` (id da mensagem enviada), `key.remoteJid`, `status` (ex.: `PENDING`) — persistir como `external_message_id` da mensagem outbound.

> **Naturalidade:** opcionalmente, enviar presença ("a escrever...") antes da resposta e usar `delay` para simular cadência humana.

### 6.8 Gestão do Ciclo de Vida da Instância

| Operação | Endpoint |
| --- | --- |
| Listar instâncias | `GET {{server-url}}/instance/fetchInstances` |
| Reiniciar | `POST {{server-url}}/instance/restart/{instance}` |
| *Logout* (desligar sessão) | `DELETE {{server-url}}/instance/logout/{instance}` |
| Eliminar instância | `DELETE {{server-url}}/instance/delete/{instance}` |

### 6.9 Considerações sobre o Canal Instagram

A Evolution API V2 é orientada a WhatsApp (motores Baileys e WhatsApp Business Cloud API). **O Instagram não é coberto nativamente** pelos endpoints WhatsApp da Evolution. Recomenda-se desenhar uma **camada de abstração de canal** (interface `ChannelAdapter` em Go) com duas implementações:

- `WhatsAppEvolutionAdapter` → integra a Evolution API V2 (esta secção).
- `InstagramAdapter` → integra a **Meta Instagram Messaging API** (Graph API / webhooks da Meta) ou outro provedor equivalente.

Ambos os adaptadores expõem o mesmo contrato (`SendText`, `SendMedia`, `OnInboundMessage`, `ConnectionState`), permitindo que o orquestrador, o RAG e o handover funcionem de forma idêntica em qualquer canal. Esta decisão isola a dependência específica de cada provedor e mantém o núcleo omnicanal coeso.

> **Recomendação:** validar, na fase de descoberta técnica, se o Instagram será integrado diretamente via Meta ou via um eventual canal/integração de terceiros, e ajustar o `InstagramAdapter` em conformidade. O contrato do adaptador não muda.

### 6.10 Integração OpenAI (núcleo de IA)

- **Embeddings:** `text-embedding-3-small` (1536 dims) para indexação RAG e *query*.
- **Conversação:** *Chat Completions* (ou *Responses API*) com `model` configurável por agente, `temperature` e `max_output_tokens`.
- **Function calling:** ferramenta `transfer_to_human` (e outras ações futuras) declarada ao modelo; quando invocada, dispara o fluxo de handover (RF-HO-01).
- **Montagem do prompt:** `system_prompt` do agente + chunks RAG (com metadados de fonte) + histórico recente da conversa + mensagem agregada do contacto.
- **Resiliência:** *retries* com backoff, *timeouts*, *fallback message* em falha.

---

## 7. Glossário e Anexos

### 7.1 Glossário

| Termo | Definição |
| --- | --- |
| **Tenant / Company** | Empresa cliente da plataforma, com isolamento total de dados. |
| **White-label** | Capacidade de personalizar domínio, logótipo e visual por tenant. |
| **Agente** | Persona de IA com prompt de sistema, modelo e bases RAG associadas. |
| **Canal** | Ligação a um número WhatsApp ou conta Instagram. |
| **Automação** | Ligação operacional canal↔agente com regras (impõe "1 agente por canal"). |
| **RAG** | *Retrieval-Augmented Generation* — respostas ancoradas em conhecimento próprio. |
| **Handover** | Transição do atendimento de IA para humano (e vice-versa). |
| **Instância (Evolution)** | Sessão de WhatsApp gerida pela Evolution API, mapeada 1:1 a um canal. |
| **Chunk** | Fragmento de texto indexado com o seu embedding vetorial. |

### 7.2 Regras Invariantes (Resumo)

1. Toda a operação de domínio tem um `company_id` resolvido e validado (isolamento).
2. No máximo **uma automação ativa por canal** → no máximo **um agente** por canal de cada vez.
3. Um agente pode consumir **múltiplas** bases RAG (N:M); uma base pode servir múltiplos agentes.
4. Em estado `human`, a IA **nunca** responde automaticamente.
5. Webhooks são processados de forma **idempotente** e **serializada por conversa**.
6. A pesquisa vetorial é **sempre** pré-filtrada por `company_id` e pelas bases do agente.

### 7.3 Pressupostos e Questões em Aberto

- **Instagram:** confirmar provedor/abordagem de integração (Meta direta vs. terceiros) — ver §6.9.
- **Modelos OpenAI:** confirmar os modelos exatos por plano/tenant e política de custos.
- **Retenção de dados:** definir período de retenção de conversas e media por tenant (RGPD/LGPD).
- **Dimensão do embedding:** assumido 1536 (`text-embedding-3-small`); ajustar `vector(N)` se mudar o modelo.

---

*Fim do documento — versão baseline. Atualizar a versão e o changelog a cada alteração relevante de âmbito, modelo de dados ou contratos de integração.*
