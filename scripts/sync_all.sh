#!/bin/bash

# Script para sincronizar todos os dados para MongoDB e exportar JSONs
# Uso: ./scripts/sync_all.sh

set -euo pipefail

# Carregar .env se existir
if [ -f .env ]; then
    set -a
    source .env
    set +a
else
    echo "❌ Arquivo .env não encontrado!"
    echo "Execute: ./scripts/setup_mongo.sh"
    exit 1
fi

# Verificar variáveis obrigatórias
if [ -z "${GH_TOKEN:-}" ]; then
    echo "❌ GH_TOKEN não definida no .env"
    exit 1
fi

if [ -z "${MONGODB_URI:-}" ]; then
    echo "❌ MONGODB_URI não definida no .env"
    exit 1
fi

USERS="${GITHUB_USERS:-felipemacedo1,growthfolio}"
IFS=',' read -ra USER_ARRAY <<< "$USERS"

echo "🚀 Sincronização completa para MongoDB Atlas"
echo "============================================="
echo "Users: $USERS"
echo ""

# Build collectors se necessário
if [ ! -f bin/user_collector ]; then
    echo "🔨 Building collectors..."
    go build -o bin/user_collector ./scripts/collectors/user_collector.go
    go build -o bin/repos_collector ./scripts/collectors/repos_collector.go
    go build -o bin/stats_collector ./scripts/collectors/stats_collector.go
    go build -o bin/activity_collector ./scripts/collectors/activity_collector.go
    go build -o bin/export_from_mongo ./scripts/export_from_mongo.go
    echo "✅ Build concluído"
    echo ""
fi

# Sync users
echo "👤 Sincronizando perfis de usuários..."
for user in "${USER_ARRAY[@]}"; do
    echo "   → $user"
    ./bin/user_collector -user="$user" || echo "⚠️  Falha ao sincronizar $user"
done
echo ""

# Sync repos
echo "📚 Sincronizando repositórios..."
./bin/repos_collector -users="$USERS" || echo "⚠️  Falha ao sincronizar repositórios"
echo ""

# Sync languages
echo "💻 Sincronizando linguagens..."
for user in "${USER_ARRAY[@]}"; do
    echo "   → $user"
    ./bin/stats_collector -user="$user" || echo "⚠️  Falha ao sincronizar linguagens de $user"
done
echo ""

# Sync activity
echo "📊 Sincronizando atividade (últimos 90 dias)..."
for user in "${USER_ARRAY[@]}"; do
    echo "   → $user"
    ./bin/activity_collector -user="$user" -days=90 || echo "⚠️  Falha ao sincronizar atividade de $user"
done
echo ""

# Export to JSON
echo "📦 Exportando para JSON..."
./bin/export_from_mongo -out=data || echo "⚠️  Falha ao exportar JSONs"
echo ""

echo "✅ Sincronização completa!"
echo ""
echo "📁 Arquivos gerados em data/:"
ls -lh data/*.json 2>/dev/null || echo "   (nenhum arquivo gerado)"
echo ""
echo "💡 Próximos passos:"
echo "   - Verifique os dados no MongoDB Atlas"
echo "   - Execute o dashboard: cd dashboard && npm run dev"
echo "   - Commit e push: git add data/ && git commit -m 'chore: update data' && git push"
