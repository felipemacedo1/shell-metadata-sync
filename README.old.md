# dev-metadata-sync

Coleta e sincroniza metadados de repositórios do GitHub.

## Dados

Arquivos JSON atualizados automaticamente via GitHub Actions:

- `data/profile.json` - Perfil do usuário
- `data/projects.json` - Lista de repositórios
- `data/activity-daily.json` - Atividade diária
- `data/languages.json` - Estatísticas de linguagens

## Uso

```bash
# Compilar
go build -o bin/repos_collector scripts/collectors/repos_collector.go

# Executar
export GITHUB_TOKEN=your_token
./bin/repos_collector
```

## GitHub Pages

Dados públicos disponíveis em:
```
https://felipemacedo1.github.io/dev-metadata-sync/projects.json
```