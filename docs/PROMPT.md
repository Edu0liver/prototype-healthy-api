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

## 3. Mensagem para contato de suporte da loja dizendo que nossos serviços estão indisponiveis
