Faça um documento, descrevendo exatamente como é essa aplicação, e como deve ser, como se fosse um PRD, ele será usado como referencia durante todo o desenvolvimento, então pode faze-lo bem completo, longo e detalhado.

Descrevendo um sistema multi-tenant/white-label, para empresas, com esse módulo de atendimento automatizado, como o JSON colado sugere, só que implementado em Golang, utilizando junto Evolution API, Redis, Postgres e OpenAI para os agentes.

Detalhes abaixo que o sistema deve possuir:

- Uma "company" pode ter várias automações, varios agentes, de Whatsapp ou Instagram.
- Ao usuário de uma empresa vincular o número de celular (Whatsapp) ou conta do Instagram, ele poderá vincular somente um agente para nesse número/conta.
- O usuário, nesse sistema, será capaz de ter acesso aos logs de conversas, configurar bases de dados dos agentes e configurar prompt de agentes.
- Uma base de dados (RAG), é onde usuários das empresas podem colocar informações e arquivos para que os agentes possam se basear por eles (acredito que pgvector seja o ideal).
- Uma "company" pode ter várias base de dados (RAG), que pode ser referências por vários agentes.
- Um agente pode usar várias bases de dados (RAG).
- Um agente pode alternar para o modo "atendimento com humano" (redis acredito que resolve), então ele não interage mais na conversação.
- Verificar se a integração com a Evolution API está correta de acordo com a documentação (V2 https://doc.evolution-api.com/v2/api-reference).
- É possivel vincular o Whatsapp via pairing code e/ou qrcode (https://doc.evolution-api.com/v2/api-reference/instance-controller/instance-connect).
