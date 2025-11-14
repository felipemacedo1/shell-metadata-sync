# Configuração MongoDB Atlas

## 1. Criar Database e Collections no Atlas

Acesse o [MongoDB Atlas](https://cloud.mongodb.com/) e:

### 1.1 Criar Database
- Database Name: `dev_metadata`

### 1.2 Criar Collections
Crie as seguintes collections:

```
dev_metadata
  ├── users           # Perfis de usuários
  ├── repositories    # Repositórios
  ├── languages       # Linguagens por usuário
  ├── activity        # Atividade diária
  └── metadata        # Metadados de sync
```

## 2. Configurar Connection String

### 2.1 Obter senha do usuário
No Atlas:
- Security → Database Access
- Usuário: `felipemacedo1`
- Copie ou redefina a senha

### 2.2 Criar arquivo `.env`

```bash
cp .env.example .env
```

Edite `.env` com seus dados:

```env
GH_TOKEN=ghp_seu_token_github_aqui
MONGODB_URI=mongodb+srv://felipemacedo1:SUA_SENHA_AQUI@terminal-cluster.1m4vtj1.mongodb.net/?appName=terminal-cluster
MONGODB_DATABASE=dev_metadata
GITHUB_USERS=felipemacedo1,growthfolio
```

⚠️ **Importante**: 
- Substitua `<db_password>` pela senha real
- A senha deve ser URL-encoded (caracteres especiais como `@`, `#`, `%` devem ser escapados)

## 3. Instalar dependências Go

```bash
cd scripts/storage
go mod tidy
cd ../..
go mod download
```

## 4. Testar conexão

```bash
# Testar collector de usuário
go run scripts/collectors/user_collector.go -user=felipemacedo1

# Testar collector de repositórios
go run scripts/collectors/repos_collector.go -users=felipemacedo1,growthfolio

# Testar collector de linguagens
go run scripts/collectors/stats_collector.go -user=felipemacedo1

# Testar collector de atividade
go run scripts/collectors/activity_collector.go -user=felipemacedo1 -days=30
```

## 5. Build dos collectors

```bash
# Build todos os collectors
go build -o bin/user_collector ./scripts/collectors/user_collector.go
go build -o bin/repos_collector ./scripts/collectors/repos_collector.go
go build -o bin/stats_collector ./scripts/collectors/stats_collector.go
go build -o bin/activity_collector ./scripts/collectors/activity_collector.go
```

## 6. Executar sincronização completa

```bash
# Sincronizar todos os dados
./bin/user_collector -user=felipemacedo1
./bin/user_collector -user=growthfolio
./bin/repos_collector -users=felipemacedo1,growthfolio
./bin/stats_collector -user=felipemacedo1
./bin/stats_collector -user=growthfolio
./bin/activity_collector -user=felipemacedo1 -days=90
./bin/activity_collector -user=growthfolio -days=90
```

## 7. Configurar GitHub Actions

### 7.1 Adicionar secrets no GitHub

Repository → Settings → Secrets and variables → Actions:

- `GH_TOKEN`: Seu token do GitHub
- `MONGODB_URI`: Connection string completo

### 7.2 Workflow já configurado

O workflow `.github/workflows/update-projects.yml` irá:
- Executar a cada 6 horas
- Sincronizar dados para MongoDB
- Gerar JSONs estáticos para o dashboard

## 8. Verificar dados no MongoDB Atlas

No Atlas:
- Browse Collections
- Selecione database `dev_metadata`
- Verifique as collections: `users`, `repositories`, `languages`, `activity`

## Schema das Collections

### users
```json
{
  "_id": "felipemacedo1",
  "name": "Felipe Macedo",
  "bio": "...",
  "avatar_url": "...",
  "followers": 100,
  "following": 50,
  "public_repos": 25,
  "total_stars_received": 500,
  "total_forks_received": 50,
  "organizations": ["growthfolio"],
  "generated_at": "2025-11-14T..."
}
```

### repositories
```json
{
  "_id": "felipemacedo1/repo-name",
  "name": "repo-name",
  "owner": "felipemacedo1",
  "description": "...",
  "language": "Go",
  "url": "https://github.com/...",
  "stars": 10,
  "forks": 2,
  "updated_at": "2025-11-14T..."
}
```

### languages
```json
{
  "_id": "felipemacedo1",
  "user": "felipemacedo1",
  "languages": {
    "Go": {
      "bytes": 50000,
      "repos": 5,
      "percentage": 45.5
    }
  },
  "top_languages": ["Go", "TypeScript", "Python"],
  "generated_at": "2025-11-14T..."
}
```

### activity
```json
{
  "_id": "felipemacedo1_2025-11-14",
  "user": "felipemacedo1",
  "date": "2025-11-14",
  "commits": 5,
  "prs": 2,
  "issues": 1,
  "reviews": 0
}
```

## Troubleshooting

### Erro: "authentication failed"
- Verifique se a senha está correta
- Certifique-se que a senha foi URL-encoded
- Verifique se o usuário `felipemacedo1` tem permissões

### Erro: "connection timeout"
- Verifique Network Access no Atlas
- Adicione seu IP: Security → Network Access → Add IP Address
- Ou permita acesso de qualquer lugar: `0.0.0.0/0` (não recomendado para produção)

### Erro: "database not found"
- O database é criado automaticamente na primeira inserção
- Certifique-se que `MONGODB_DATABASE=dev_metadata` está no `.env`

### Verificar logs
```bash
# Com verbose
go run scripts/collectors/user_collector.go -user=felipemacedo1 -v

# Ver collections
mongosh "mongodb+srv://terminal-cluster.1m4vtj1.mongodb.net" --apiVersion 1 --username felipemacedo1
```
