#!/bin/bash

# Script para copiar JSONs para pasta pública do dashboard
# Permite que o site funcione estaticamente no GitHub Pages

set -euo pipefail

# Detectar diretório raiz do projeto
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "📦 Copiando JSONs para dashboard/public/data/"

# Navegar para diretório raiz
cd "$ROOT_DIR"

# Criar diretório se não existir
mkdir -p dashboard/public/data

# Remover symlinks existentes para evitar conflitos
rm -f dashboard/public/data/projects.json

# Copiar todos os JSONs (usar -L para seguir symlinks)
cp -L data/projects.json dashboard/public/data/projects.json
cp data/profile.json dashboard/public/data/profile.json 2>/dev/null || echo "⚠️  profile.json não encontrado"
cp data/languages.json dashboard/public/data/languages.json 2>/dev/null || echo "⚠️  languages.json não encontrado"
cp data/activity-daily.json dashboard/public/data/activity.json 2>/dev/null || echo "⚠️  activity-daily.json não encontrado"
cp data/metadata.json dashboard/public/data/metadata.json 2>/dev/null || echo "⚠️  metadata.json não encontrado"

echo "✅ JSONs copiados com sucesso!"
echo ""
echo "Arquivos disponíveis em:"
echo "  - dashboard/public/data/projects.json"
echo "  - dashboard/public/data/profile.json"
echo "  - dashboard/public/data/languages.json"
echo "  - dashboard/public/data/activity.json"
echo "  - dashboard/public/data/metadata.json"
