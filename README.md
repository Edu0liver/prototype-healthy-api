# API

SaaS multi-tenant white-label em Go com atendimento automatizado por IA.

## Pré-requisitos

- Go 1.26+
- Docker + Docker Compose
- `goose` — `go install github.com/pressly/goose/v3/cmd/goose@latest`

## Rodando localmente

```bash
# 1. Copie e configure as variáveis de ambiente
cp .env.example .env
# edite .env com suas credenciais

# 2. Suba Postgres + Redis
make up

# 3. Instale dependências
make deps

# 4. Rode as migrations
make migrate-up

# 5. Inicie o servidor
make run
```

O servidor sobe em `http://localhost:8080`.
