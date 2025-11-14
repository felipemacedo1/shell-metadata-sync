# Quick Start - MongoDB Sync

## 1. Configure MongoDB Atlas

```bash
# Opção 1: Script interativo
./scripts/setup_mongo.sh

# Opção 2: Manual
cp .env.example .env
# Edite .env com suas credenciais
```

**Importante**: No MongoDB Atlas, adicione seu IP em **Network Access**:
- Security → Network Access → Add IP Address
- Para dev: `0.0.0.0/0` (permite qualquer IP)
- Para prod: adicione IPs específicos

## 2. Teste a conexão

```bash
source .env
go run scripts/test_mongo_connection.go
```

Saída esperada:
```
✅ Conexão estabelecida com sucesso!
✅ Documento de teste inserido
✅ Documento de teste removido
```

## 3. Build dos collectors

```bash
go build -o bin/user_collector ./scripts/collectors/user_collector.go
go build -o bin/repos_collector ./scripts/collectors/repos_collector.go
go build -o bin/stats_collector ./scripts/collectors/stats_collector.go
go build -o bin/activity_collector ./scripts/collectors/activity_collector.go
go build -o bin/export_from_mongo ./scripts/export_from_mongo.go
```

## 4. Sincronizar dados

### Coletar dados de usuário
```bash
./bin/user_collector -user=felipemacedo1
./bin/user_collector -user=growthfolio
```

### Coletar repositórios
```bash
./bin/repos_collector -users=felipemacedo1,growthfolio
```

### Coletar estatísticas de linguagens
```bash
./bin/stats_collector -user=felipemacedo1
./bin/stats_collector -user=growthfolio
```

### Coletar atividade (últimos 90 dias)
```bash
./bin/activity_collector -user=felipemacedo1 -days=90
./bin/activity_collector -user=growthfolio -days=90
```

## 5. Exportar para JSON

```bash
./bin/export_from_mongo -out=data
```

Arquivos gerados:
- `data/profile.json`
- `data/projects.json`
- `data/languages.json`
- `data/activity-daily.json`
- `data/metadata.json`

## 6. Verificar no MongoDB Atlas

1. Acesse [cloud.mongodb.com](https://cloud.mongodb.com)
2. Database → Browse Collections
3. Database: `dev_metadata`
4. Collections:
   - `users` - Perfis de usuários
   - `repositories` - Repositórios
   - `languages` - Linguagens por usuário
   - `activity` - Atividade diária

## 7. Configurar GitHub Actions

### Adicionar secrets no repositório

Settings → Secrets and variables → Actions → New repository secret:

- **Name**: `GH_TOKEN`  
  **Value**: Seu GitHub Personal Access Token

- **Name**: `MONGODB_URI`  
  **Value**: `mongodb+srv://usuario:senha@cluster.mongodb.net/?appName=...`

### Workflow automático

O workflow `.github/workflows/sync-mongodb.yml` executa:
- ✅ A cada 6 horas
- ✅ Manual via "Run workflow"
- ✅ Coleta dados → MongoDB
- ✅ Exporta MongoDB → JSON
- ✅ Commit dos JSONs

## 8. Dashboard Next.js

O dashboard consome os JSONs estáticos:

```bash
cd dashboard
npm install
npm run dev
```

Acesse: http://localhost:3000

## Troubleshooting

### ❌ "authentication failed"
- Verifique senha no `.env`
- URL-encode caracteres especiais: `@` → `%40`, `#` → `%23`

### ❌ "connection timeout"
- MongoDB Atlas → Security → Network Access
- Adicione seu IP ou `0.0.0.0/0`

### ❌ "database not found"
- Database é criado automaticamente na primeira inserção
- Certifique-se que `MONGODB_DATABASE=dev_metadata`

### 📊 Ver logs detalhados
```bash
./bin/user_collector -user=felipemacedo1 -v
```

## Fluxo completo

```
GitHub API
    ↓
Go Collectors (sync a cada 6h)
    ↓
MongoDB Atlas (dev_metadata)
    ↓
Export to JSON (data/)
    ↓
Next.js Dashboard (GitHub Pages)
```

## Comandos úteis

```bash
# Sincronização completa
./scripts/sync_all.sh

# Apenas exportar JSONs
./bin/export_from_mongo -out=data

# Rebuild all
go build -o bin/user_collector ./scripts/collectors/user_collector.go && \
go build -o bin/repos_collector ./scripts/collectors/repos_collector.go && \
go build -o bin/stats_collector ./scripts/collectors/stats_collector.go && \
go build -o bin/activity_collector ./scripts/collectors/activity_collector.go && \
go build -o bin/export_from_mongo ./scripts/export_from_mongo.go

# Dashboard build
cd dashboard && npm run build
```
