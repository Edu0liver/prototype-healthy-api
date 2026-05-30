# Papel e Contexto
Age como um Engenheiro de Software Principal e Diretor de Produto (CPO) com vasta experiência em arquiteturas multi-tenant SaaS de alto desempenho, sistemas distribuídos, processamento assíncrono e integrações avançadas de Inteligência Artificial.

# Objetivo
Cria um Documento de Requisitos de Produto (PRD) extremamente detalhado, completo, longo e estruturado para um sistema multi-tenant/white-label de atendimento automatizado e omnicanal. Este documento será utilizado como a principal referência ("source of truth") técnica e funcional durante todo o ciclo de desenvolvimento da aplicação.

# Arquitetura e Stack Tecnológica
A aplicação deve ser desenhada estritamente sob as seguintes definições tecnológicas:
- **Backend:** Golang (focado em alta concorrência, processamento assíncrono via goroutines, performance e código limpo).
- **Base de Dados Relacional e Vetorial:** PostgreSQL utilizando a extensão `pgvector` para armazenamento e pesquisa de embeddings.
- **Cache, Filas e Controlo de Estado:** Redis (utilizado para gestão de sessões, locks distribuídos, debounce de mensagens, buffers temporários e controlo de estado de transição de conversas).
- **Motor de IA e Transcrição:** OpenAI API (modelos GPT para agentes conversacionais; modelo de Embeddings para RAG; modelo Whisper-large-v3 para transcrição de áudio).
- **Integração de Mensagens:** Evolution API V2.

# Especificações e Regras de Negócio

## 1. Multi-Tenancy e White-Label
- Isolamento lógico absoluto de dados entre diferentes empresas ("Companies") utilizando identificadores de tenant (TenantID).
- Estrutura preparada para White-Label, permitindo customização de domínios, logótipos e aspetos visuais por empresa.

## 2. Modelagem de Entidades e Restrições de Canais
- Uma "Company" pode gerir múltiplas automações, múltiplos agentes e múltiplos canais de comunicação (WhatsApp e Instagram).
- **Regra Crítica:** Ao interligar um canal (um número de WhatsApp ou uma conta de Instagram), o utilizador da empresa só pode associar **um único agente** a esse canal específico de cada vez.

## 3. Gestão de Conhecimento e RAG (Retrieval-Augmented Generation)
- O sistema deve permitir que as empresas criem bases de dados de conhecimento através do upload de ficheiros e textos.
- Uma "Company" pode ter múltiplas bases de dados (RAG).
- **Relacionamento Flexível:** Um agente de IA pode estar associado e consumir dados de múltiplas bases de dados RAG em simultâneo (Relação N:M), injetando os documentos recuperados como ferramentas (tools) no contexto do agente LangChain/OpenAI.

## 4. Dinâmica de Mensagens: Debounce e Agrupamento
- **Problema:** Utilizadores reais tendem a enviar múltiplas mensagens curtas fragmentadas antes de concluírem o raciocínio.
- **Solução (Debounce via Redis):** O sistema deve intercetar mensagens recebidas e armazená-las numa lista temporária no Redis (chave baseada em `instance` + `remoteJid`). A aplicação em Golang deve aguardar uma janela de tempo (ex: 10 segundos). Se nenhuma nova mensagem chegar nesse período, o sistema concatena todas as mensagens da lista temporária e envia um único bloco de texto consolidado para o Agente de IA processar, apagando a lista de seguida.

## 5. Processamento Omnicanal e Multimédia (Áudios)
- O sistema deve intercetar webhooks da Evolution API V2 identificando o tipo de mensagem (`messageType`).
- Se for mensagem de texto (`conversation`), segue o fluxo normal.
- Se for mensagem de áudio (`audioMessage`), o backend deve requisitar o base64 do áudio à Evolution API (`get-media-base64`), converter/tratar em memória no Golang e submeter ao endpoint da **OpenAI (Whisper)** para transcrição. O texto transcrito segue então o fluxo normal como se fosse uma mensagem de texto.
- O sistema deve emitir o evento de leitura (`read-messages`) na Evolution API assim que as mensagens são processadas.

## 6. Handover Passivo e Controlo de Estado (Atendimento Humano)
- O sistema deve ter uma transição suave e automática (Handover) do bot para o humano.
- **Deteção via `fromMe`:** O webhook da Evolution API envia a flag `fromMe`. Se uma mensagem for enviada pelo próprio número (ou seja, um humano pegou no telemóvel ou no WhatsApp Web para responder):
  1. O sistema deve acionar automaticamente uma chave de bloqueio no Redis (`isBlocked = true` com TTL).
  2. O sistema não deve gerar resposta da IA.
  3. A mensagem do atendente humano deve ser **injetada na memória do agente de IA** (Chat Memory) para que o bot tenha contexto das ações tomadas pelo humano caso a automação seja reativada.
- O utilizador pode gerir manualmente este estado "Bloqueado / Ativo" através do painel.

## 7. Humanização das Saídas da IA (Output Parsing e Delay)
- As respostas geradas pela IA bruta devem passar por uma etapa de formatação (Output Parser via LLM Chain).
- **Regras de formatação:** - Quebrar respostas longas num máximo de 4 mensagens menores e humanizadas.
  - Remover marcações Markdown inadequadas para chat (ex: links formatados `[texto](url)` devem ser convertidos em texto simples com a URL).
- **Envio Assíncrono com Delay:** O envio destas mensagens fracionadas de volta pela Evolution API deve conter um atraso (`delay` de ~2000ms a 3000ms) entre elas, ativando também o `linkPreview: true`, para simular a digitação e comportamento humano.

## 8. Interface e Painel do Utilizador (Tenant)
O utilizador de cada empresa deve ter acesso direto a:
- Visualização e auditoria completa dos logs de conversas em tempo real.
- Configuração e gestão das bases de dados de conhecimento (RAG) com visualização dos embeddings e chunks.
- Configuração e ajuste fino do prompt de sistema (instruções primárias) de cada agente.

## 9. Integração Técnica com Evolution API V2
- O fluxo de integração deve respeitar rigorosamente a documentação oficial da Evolution API V2 (Referência: https://doc.evolution-api.com/v2/api-reference).
- O sistema deve suportar a vinculação de instâncias de WhatsApp utilizando os métodos de **Pairing Code** (Código de Emparelhamento) e/ou **QR Code**.

# Estrutura Requerida para o PRD
O documento gerado deve conter as seguintes secções detalhadas:
1. **Visão Geral e Objetivos do Produto:** Proposta de valor e objetivos macro.
2. **Arquitetura do Sistema e Fluxo de Dados:** Modelo conceptual, isolamento de tenants, estratégia de concorrência em Golang (Workers/Goroutines) e ciclo de vida do Webhook.
3. **Modelagem de Dados (ERD e Cache):** Estrutura de tabelas do Postgres, índices vetoriais com pgvector e definições exatas das chaves/estruturas de dados no Redis (Listas de debounce, Flags de handover, Sessões de memória).
4. **Requisitos Funcionais Detalhados:** Casos de uso com critérios de aceitação para cada módulo (Canais, Agentes, RAG, Debounce, Handover Automático `fromMe`, Parsing de Saída e Transcrição de Áudio).
5. **Requisitos Não-Funcionais:** Segurança, escalabilidade, latência das respostas de IA e resiliência da ligação à API de mensagens.
6. **Mapeamento de Integrações:** Endpoints, payloads essenciais e mapeamento de webhooks necessários da Evolution API V2 (incluindo `get-media-base64`, `read-messages`, envio de mensagens e instâncias).