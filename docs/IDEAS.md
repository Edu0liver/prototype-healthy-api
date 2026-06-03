# Ideias para Implementações Futuras

---

## 1. Infraestrutura e Engenharia de Prompts (Meta-Agentes)

### Agente Gerador de Prompts (Meta-Agente)

**Descrição:**
Uma funcionalidade focada em aperfeiçoar e otimizar prompts para outros agentes do sistema.

**Como funciona:**

* Utiliza um fluxo interativo de perguntas predefinidas.
* Coleta o contexto exato e os requisitos do usuário.
* Estrutura o prompt ideal com base nas respostas fornecidas.

**Objetivo:**
Ajudar a entender com precisão o que o agente final deve fazer, reduzindo a tentativa e erro e aumentando a eficiência da IA.

 Ideias para Implementações Futuras — Prototype Healthy

Este documento reúne propostas de melhorias, novas funcionalidades e arquiteturas de agentes de Inteligência Artificial para o projeto **Prototype Healthy**. As ideias estão divididas por categorias para facilitar o planeamento e o desenvolvimento do backlog

---

## 2. Funcionalidades Centradas no Utilizador (Pacientes)

### Agente de Triagem e Pré-Avaliação de Sintomas

* **Descrição:** Um assistente virtual que realiza uma triagem preliminar com base em sintomas relatados pelo utilizador.
* **Como funciona:** Segue protocolos internacionais validados (como o Protocolo de Manchester) para fazer perguntas direcionadas e avaliar o nível de urgência.
* **Objetivo:** Orientar o utilizador sobre a necessidade de procurar um serviço de urgência, uma consulta de rotina ou cuidados caseiros, sem emitir diagnósticos definitivos.

### Integração com Dispositivos Wearables

* **Descrição:** Ligação do ecossistema a smartwatches e smartbands (via APIs como Apple HealthKit ou Google Fit).
* **Como funciona:** Um agente analisa continuamente métricas como frequência cardíaca, qualidade do sono, níveis de oxigénio e atividade física.
* **Objetivo:** Fornecer alertas preditivos e relatórios semanais automatizados sobre a evolução da saúde física do utilizador.

### Gestor Inteligente de Medicação e Adesão

* **Descrição:** Um assistente interativo para controlo de tratamentos médicos.
* **Como funciona:** Além de emitir alertas, o agente conversa com o utilizador para registar se a medicação foi tomada, mapear efeitos secundários e notificar familiares ou médicos em caso de esquecimentos recorrentes.

---

## 3. Funcionalidades para Profissionais de Saúde

### Agente Resumidor de Histórico Clínico (Prontuários)

* **Descrição:** Ferramenta de apoio à decisão clínica que processa históricos médicos longos e complexos.
* **Como funciona:** Analisa documentos anteriores, exames e notas de consultas, gerando um resumo executivo estruturado para o médico ler antes de iniciar a consulta.
* **Objetivo:** Destacar de imediato alergias, patologias crónicas, tratamentos em curso e os últimos exames relevantes, otimizando o tempo da consulta.

### Assistente de Diagnóstico por RAG (Retrieval-Augmented Generation)

* **Descrição:** Um agente de consulta para médicos baseado em evidência científica de confiança.
* **Como funciona:** Utiliza uma base de conhecimento atualizada com artigos do PubMed, diretrizes da OMS e manuais de medicina. O médico introduz os sintomas e o agente sugere diagnósticos diferenciais fundamentados e com links para as fontes.

---

## 4. Segurança, Moderação e Conformidade

### Agente de Auditoria Clínica (Fact-Checker Interno)

* **Descrição:** Uma camada de segurança de "segunda opinião" automatizada na arquitetura do sistema.
* **Como funciona:** Monitoriza as respostas geradas pelos agentes de IA destinados aos pacientes e valida se há alguma recomendação clinicamente perigosa, incorreta ou ambígua antes de a exibir no ecrã.

### Agente de Anonimização e Privacidade (RGPD / LGPD)

* **Descrição:** Um gateway de segurança para proteção de Dados Pessoais de Saúde (dados sensíveis).
* **Como funciona:** Intercepta as mensagens do utilizador antes de serem enviadas para as APIs de LLM externas e mascara automaticamente dados de identificação pessoal (como nomes, cartões de cidadão, moradas e números de telefone).
