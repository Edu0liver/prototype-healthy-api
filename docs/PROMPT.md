# PROMPT: Criação de Frontend Next.js baseado no Backend Local

Atue como um Desenvolvedor Frontend Sênior especialista em React, Next.js (App Router), TypeScript e Tailwind CSS.

Seu objetivo é criar um projeto de frontend do zero em Next.js dentro de uma nova pasta chamada `frontend` no mesmo diretório atual. Este frontend irá consumir a API do projeto backend existente neste repositório.

## 1. Stack Tecnológico Base

- **Framework:** Next.js (App Router)
- **Linguagem:** TypeScript
- **Estilização:** Tailwind CSS + Material UI (MUI)
  *(Nota: Configure o MUI corretamente para funcionar com o App Router do Next.js usando 'use client' onde necessário e garanta que os estilos do Tailwind não conflitem com os do MUI).*
- **Consumo de API:** Axios (com instância configurada e interceptors para injetar o Token JWT).
- **Gerenciamento de Estado/Autenticação:** Context API.

## 2. Descoberta de Contexto (Análise do Backend)

Antes de escrever o código do frontend, analise os arquivos do projeto backend atual para identificar:

1. Como funciona o fluxo de autenticação (mecanismo de login, geração e validação do token JWT).
2. Quais são as principais entidades e rotas disponíveis (especialmente as rotas de autenticação e as rotas da tela principal/dashboard).
3. A estrutura dos dados (tipagens/interfaces) que a API retorna.

## 3. Estrutura de Páginas e Fluxos a Criar

Crie as seguintes telas no Next.js adaptadas ao que você encontrou no backend:

1. **Página de Login (`/login`):** Tela pública com formulário estruturado com os campos necessários identificados na API. Ao logar com sucesso, salve o JWT com segurança (ex: cookies ou localStorage) e redirecione.
2. **Dashboard/Home (`/`):** Tela privada. Busque os dados da rota principal da sua API (ex: listagem de dados) e exiba-os em uma tabela ou cards usando componentes do Material UI.
3. **Layout Global:** Crie um layout limpo com uma barra de navegação/sidebar (com botão de Logout) para as páginas privadas.

## 4. Regras de Negócio e Requisitos Técnicos

- **Proteção de Rotas:** Implemente um `middleware.ts` no Next.js para impedir que usuários não autenticados acessem a Dashboard, redirecionando-os para `/login`.
- **Gerenciamento de Estado:** O Context API deve gerenciar o estado global de autenticação (`user`, `isAuthenticated`, função de `login` e `logout`).
- **Tratamento de Erros:** Configure o Axios para capturar erros (400, 401, 500) e implemente um sistema de feedbacks visuais (como Toast Notifications ou Alerts do MUI).
- **Tipagem Estrita:** Crie interfaces TypeScript claras para todas as respostas da API baseando-se no backend.

## 5. Próximos Passos (Plano de Ação)

Por favor, execute nesta ordem:

1. Crie a pasta para o projeto no diretório anterior a este projeto e inicialize o projeto Next.js com as configurações solicitadas.
2. Apresente a árvore de arquivos recomendada.
3. Crie os arquivos de configuração (Axios, Context de Autenticação, Middleware).
4. Desenvolva as páginas e componentes de UI.

Me avise se encontrar alguma ambiguidade no backend antes de começar a codificar o frontend.
