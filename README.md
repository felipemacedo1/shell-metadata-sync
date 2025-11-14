# Dev Metadata Sync

Sistema automatizado de coleta, armazenamento e visualização de metadados do GitHub.

## 🎯 Visão Geral

```
GitHub API → Go Collectors → MongoDB Atlas → Export JSON → Next.js Dashboard
```

Coleta dados de repositórios via GitHub API, armazena em MongoDB Atlas, exporta para JSONs estáticos e exibe em dashboard Next.js hospedado no GitHub Pages.

## ✨ Recursos

### Coleta de dados (Go)
- ✅ Rate limit handling e retry exponential backoff
- ✅ Cache com ETag (reduz 90% das requisições)
- ✅ Logs estruturados (JSON/pretty)
- ✅ Validação de dados e detecção de duplicatas
- ✅ Changelog automático de mudanças
- ✅ Testes unitários (9/9 passing)

### Armazenamento
- ✅ MongoDB Atlas (database principal)
- ✅ JSONs estáticos (fallback para GitHub Pages)
- ✅ Sincronização automática a cada 6h

### Dashboard Next.js
- ✅ Gráficos de linguagens e atividade (Recharts)
- ✅ Heatmap de contribuições
- ✅ Listagem de repositórios com filtros
- ✅ Static export para GitHub Pages
- ✅ Modo dual: API routes (dev) + static files (prod)

### Automação (GitHub Actions)
- ✅ Cron schedule a cada 6 horas
- ✅ Manual dispatch com parâmetros
- ✅ Validação automática com jq
- ✅ Commit apenas com mudanças reais

## 📁 Estrutura

```
.
├── data/                    # JSONs estáticos exportados
├── dashboard/               # Next.js app
│   ├── src/
│   │   ├── app/            # Pages e layouts
│   │   ├── components/     # UI components
│   │   └── lib/            # API client e types
│   └── public/data/        # JSONs para static hosting
├── scripts/
│   ├── collectors/         # Go collectors (MongoDB sync)
│   │   ├── user_collector.go
│   │   ├── repos_collector.go
│   │   ├── stats_collector.go
│   │   └── activity_collector.go
│   ├── storage/            # MongoDB client
│   ├── export_from_mongo.go # MongoDB → JSON export
│   ├── update_projects.go   # Legacy JSON-only collector
│   └── test_mongo_connection.go
├── .github/workflows/
│   ├── sync-mongodb.yml     # Sync GitHub → MongoDB → JSON
│   ├── deploy-pages.yml     # Deploy dashboard to Pages
│   └── update-projects.yml  # Legacy workflow
└── bin/                     # Binários compilados
```

## 🚀 Quick Start

### 1. Setup MongoDB Atlas

```bash
./scripts/setup_mongo.sh
```

Ou manualmente:
```bash
cp .env.example .env
# Edite .env com suas credenciais
```

📖 Ver: [MONGODB_SETUP.md](MONGODB_SETUP.md) | [QUICKSTART_MONGO.md](QUICKSTART_MONGO.md)

### 2. Testar conexão

```bash
source .env
go run scripts/test_mongo_connection.go
```

### 3. Sincronizar dados

```bash
./scripts/sync_all.sh
```

Ou manualmente:
```bash
# Build
go build -o bin/user_collector ./scripts/collectors/user_collector.go
go build -o bin/repos_collector ./scripts/collectors/repos_collector.go
go build -o bin/stats_collector ./scripts/collectors/stats_collector.go
go build -o bin/activity_collector ./scripts/collectors/activity_collector.go
go build -o bin/export_from_mongo ./scripts/export_from_mongo.go

# Sync
./bin/user_collector -user=felipemacedo1
./bin/repos_collector -users=felipemacedo1,growthfolio
./bin/stats_collector -user=felipemacedo1
./bin/activity_collector -user=felipemacedo1 -days=90
./bin/export_from_mongo -out=data
```

### 4. Dashboard

```bash
cd dashboard
npm install
npm run dev
```

Acesse: http://localhost:3000

## 📚 Documentação

- **[MONGODB_SETUP.md](MONGODB_SETUP.md)** - Setup completo do MongoDB Atlas
- **[QUICKSTART_MONGO.md](QUICKSTART_MONGO.md)** - Guia rápido de uso
- **[WORKFLOWS.md](WORKFLOWS.md)** - Documentação dos workflows
- **[STATUS_IMPLEMENTACAO.md](STATUS_IMPLEMENTACAO.md)** - Status do projeto

```json
{
  "metadata": {
    "generated_at": "2025-11-14T01:50:43Z",
    "total_repos": 41,
    "users": ["felipemacedo1"]
  },
  "repositories": [
    {
      "name": "repo-name",
      "owner": "username",
      "description": "Descrição",
      "language": "Go",
      "url": "https://github.com/username/repo",
      "updated_at": "2025-11-14T00:00:00Z"
    }
  ]
}
```

## GitHub Actions

O workflow executa a cada 6 horas ou manualmente:

- Build e execução do script
- Validação do JSON
- Commit apenas se houver mudanças
- Usa cache para otimizar requisições

Secrets opcionais:
- `GH_TOKEN` - Aumenta rate limit da API

## Testes

```bash
go test -v ./scripts/
```

Cobertura:
- Validação de output
- Cache save/load
- Changelog generation
- JSON serialization
   - Detecta duplicatas
   - Retorna erros específicos

7. **Flag `--users` flexível**
   - Aceita lista separada por vírgulas
   - Permite qualquer usuário do GitHub
   - Não mais hardcoded no código

8. **Diff e changelog automático**
   - Compara versão anterior vs nova
   - Gera `CHANGELOG.md` com:
     - Repositórios adicionados
     - Repositórios atualizados
     - Repositórios removidos
   - Formato Markdown limpo

9. **Testes unitários**
   - 9 testes cobrindo funções principais
   - Mock de HTTP server
   - Testes de validação
   - Testes de cache
   - Testes de changelog
   - 100% de testes passando ✅

---

## 🔐 Segurança

- ✅ Secrets não expandidos em blocos `run` do workflow
- ✅ Uso de variáveis de ambiente para tokens
- ✅ Cache em `.cache/` ignorado pelo Git
- ✅ Binários em `bin/` ignorados pelo Git
- ✅ Token opcional (funciona sem, mas com rate limit menor)

---

## 📈 Performance

### Benchmarks

| Métrica | Sem Cache | Com Cache (304) |
|---------|-----------|-----------------|
| Tempo de execução | ~2-3s | ~0.5s |
| Requisições API | N páginas | 1 por usuário |
| Rate limit usado | ~N | 1 |
| Transferência de dados | Full | Mínimo |

### Otimizações Aplicadas

- ✅ Timeout de 30s por requisição
- ✅ Paginação eficiente (100 repos/página)
- ✅ Context cancellation support
- ✅ Escrita atômica com rename
- ✅ Cache em disco persistente

## Troubleshooting

**Rate limit exceeded**
```bash
export GH_TOKEN=ghp_seu_token_aqui
```

**Validação falhou**
```bash
jq . data/projects.json
./bin/update --verbose
```

## Próximos passos

- Integração com MongoDB
- API REST (Next.js)
- Dashboard mais completo
- GraphQL para queries avançadas

## Docs

- [RELATORIO_ANALISE.md](./RELATORIO_ANALISE.md) - Análise técnica
- [CHANGELOG.md](./CHANGELOG.md) - Histórico de mudanças

## Licença

MIT

---

## 👨‍💻 Autor

**Felipe Macedo**
- GitHub: [@felipemacedo1](https://github.com/felipemacedo1)
- Email: felipealexandrej@gmail.com

---

## 🙏 Agradecimentos

- GitHub API documentation
- Go standard library team
- Next.js team
- MongoDB team
- Open source community

---

## 📊 Status do Projeto

![Build Status](https://github.com/felipemacedo1/shell-metadata-sync/workflows/Update%20Projects%20JSON/badge.svg)
![Go Version](https://img.shields.io/badge/Go-1.22-blue)
![Next.js Version](https://img.shields.io/badge/Next.js-16.0-black)
![License](https://img.shields.io/badge/license-MIT-green)

**Última atualização:** 14 de novembro de 2025
