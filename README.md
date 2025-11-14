# Dev Metadata Sync

Sistema de coleta e visualização de metadados do GitHub.

## Sobre

Coleta dados de repositórios públicos via GitHub API, armazena em MongoDB e JSONs, e disponibiliza um dashboard Next.js para visualização.

## Recursos

**Coleta de dados (Go)**
- Rate limit handling e retry automático
- Cache com ETag para otimizar requisições
- Logs estruturados
- Validação de dados
- Changelog de mudanças

**API REST (Next.js)**
```
GET /api/profile
GET /api/projects
GET /api/languages
GET /api/activity
GET /api/metadata
```

**Dashboard**
- Gráficos de linguagens e atividade
- Heatmap de contribuições
- Lista de repositórios

**Automação (GitHub Actions)**
- Execução a cada 6h
- Validação automática
- Commit apenas se houver mudanças

## Estrutura

```
data/                  # JSONs gerados
dashboard/             # Next.js app
  app/api/            # API endpoints
  components/         # Componentes React
scripts/              # Collectors Go
  update_projects.go  # Script principal
  collectors/         # Outros coletores
.github/workflows/    # Automação CI/CD
```

## Uso

### Build

```bash
go build -o bin/update ./scripts/update_projects.go
```

### Executar

```bash
# Básico
./bin/update

# Com opções
./bin/update --users=user1,user2 --verbose
```

### Flags disponíveis

```
--users        Usuários separados por vírgula (padrão: felipemacedo1,growthfolio)
--out          Arquivo de saída (padrão: data/projects.json)
--cache-dir    Diretório de cache (padrão: .cache)
--changelog    Arquivo de changelog (padrão: CHANGELOG.md)
--verbose      Logs detalhados
--json-logs    Logs em formato JSON
--token        GitHub token (ou use GH_TOKEN env)
```

## Formato do JSON

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
