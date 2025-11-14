#!/bin/bash

# Script para configurar MongoDB Atlas connection
# Ajuda o usuário a criar o arquivo .env corretamente

set -euo pipefail

echo "🔧 Configuração MongoDB Atlas"
echo "=============================="
echo ""

# Verificar se .env já existe
if [ -f .env ]; then
    echo "⚠️  Arquivo .env já existe!"
    read -p "Deseja sobrescrever? (s/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        echo "❌ Operação cancelada"
        exit 0
    fi
fi

# Obter GitHub Token
echo "1️⃣  GitHub Token"
echo "   Crie um token em: https://github.com/settings/tokens"
echo "   Permissões: repo, read:user, read:org"
echo ""
read -p "Digite seu GitHub Token: " GH_TOKEN

# Obter MongoDB Connection String
echo ""
echo "2️⃣  MongoDB Atlas Connection String"
echo "   Formato: mongodb+srv://usuario:senha@cluster.mongodb.net/?appName=..."
echo ""
read -p "Digite sua Connection String: " MONGODB_URI

# Obter Database Name
echo ""
echo "3️⃣  Database Name (padrão: dev_metadata)"
read -p "Digite o nome do database [dev_metadata]: " MONGODB_DATABASE
MONGODB_DATABASE=${MONGODB_DATABASE:-dev_metadata}

# Obter GitHub Users
echo ""
echo "4️⃣  GitHub Users para coletar (separados por vírgula)"
read -p "Digite os usuários [felipemacedo1,growthfolio]: " GITHUB_USERS
GITHUB_USERS=${GITHUB_USERS:-felipemacedo1,growthfolio}

# Criar arquivo .env
cat > .env << EOF
# GitHub API Token
GH_TOKEN=$GH_TOKEN

# MongoDB Atlas Connection String
MONGODB_URI=$MONGODB_URI

# MongoDB Database Name
MONGODB_DATABASE=$MONGODB_DATABASE

# GitHub Users
GITHUB_USERS=$GITHUB_USERS
EOF

echo ""
echo "✅ Arquivo .env criado com sucesso!"
echo ""
echo "🧪 Testar conexão:"
echo "   source .env && go run scripts/test_mongo_connection.go"
echo ""
echo "📦 Build collectors:"
echo "   go build -o bin/user_collector ./scripts/collectors/user_collector.go"
echo "   go build -o bin/repos_collector ./scripts/collectors/repos_collector.go"
echo "   go build -o bin/stats_collector ./scripts/collectors/stats_collector.go"
echo "   go build -o bin/activity_collector ./scripts/collectors/activity_collector.go"
echo ""
echo "🚀 Executar sincronização:"
echo "   ./bin/user_collector -user=felipemacedo1"
echo "   ./bin/repos_collector -users=felipemacedo1,growthfolio"
echo ""

# Perguntar se quer testar conexão agora
read -p "Testar conexão agora? (s/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Ss]$ ]]; then
    echo ""
    echo "🧪 Testando conexão..."
    export MONGODB_URI="$MONGODB_URI"
    export MONGODB_DATABASE="$MONGODB_DATABASE"
    go run scripts/test_mongo_connection.go
fi
