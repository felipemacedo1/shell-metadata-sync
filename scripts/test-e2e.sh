#!/bin/bash
# Script de teste end-to-end para todos os collectors
# Simula a execução dos workflows do GitHub Actions localmente

set -euo pipefail

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Funções auxiliares
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_header() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║ $1${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Verificar variáveis de ambiente
check_environment() {
    print_header "🔍 Verificando Ambiente"
    
    if [[ -z "${GITHUB_TOKEN:-}" ]]; then
        log_warning "GITHUB_TOKEN não configurado - rate limit será 60 req/hora"
        log_info "Exporte com: export GITHUB_TOKEN=your_token_here"
    else
        log_success "GITHUB_TOKEN configurado"
    fi
    
    if [[ -z "${MONGO_URI:-}" ]]; then
        log_info "MONGO_URI não configurado - dados não serão salvos no MongoDB"
    else
        log_success "MONGO_URI configurado"
    fi
}

# Compilar todos os collectors
build_collectors() {
    print_header "🔨 Compilando Collectors"
    
    local collectors=("user_collector" "repos_collector" "activity_collector" "stats_collector")
    
    for collector in "${collectors[@]}"; do
        log_info "Building $collector..."
        if go build -o "bin/$collector" "scripts/collectors/$collector.go"; then
            log_success "$collector compilado"
        else
            log_error "Falha ao compilar $collector"
            exit 1
        fi
    done
}

# Executar collector de perfil
run_user_collector() {
    print_header "👤 Coletando Perfil do Usuário"
    
    local user="${1:-felipemacedo1}"
    log_info "Usuário: $user"
    
    if ./bin/user_collector -user="$user"; then
        log_success "Perfil coletado com sucesso"
        
        if [[ -f "data/profile.json" ]]; then
            local name=$(jq -r '.name' data/profile.json)
            local followers=$(jq '.followers' data/profile.json)
            local repos=$(jq '.public_repos' data/profile.json)
            echo "  → Nome: $name"
            echo "  → Followers: $followers"
            echo "  → Repos públicos: $repos"
        fi
    else
        log_error "Falha ao coletar perfil"
        return 1
    fi
}

# Executar collector de repositórios
run_repos_collector() {
    print_header "📦 Coletando Repositórios"
    
    if ./bin/repos_collector; then
        log_success "Repositórios coletados com sucesso"
        
        if [[ -f "data/projects.json" ]] && [[ -f "data/metadata.json" ]]; then
            local repos=$(jq 'length' data/projects.json)
            local version=$(jq -r '.version' data/metadata.json)
            echo "  → Total de repos: $repos"
            echo "  → Versão: $version"
        fi
    else
        log_error "Falha ao coletar repositórios"
        return 1
    fi
}

# Executar collector de atividade
run_activity_collector() {
    print_header "📈 Coletando Atividade"
    
    local user="${1:-felipemacedo1}"
    local days="${2:-30}"
    
    log_info "Usuário: $user | Período: últimos $days dias"
    log_warning "Coletando apenas $days dias para economizar rate limit"
    
    if timeout 300 ./bin/activity_collector -user="$user" -days="$days"; then
        log_success "Atividade coletada com sucesso"
        
        if [[ -f "data/activity-daily.json" ]]; then
            local commits=$(jq '.daily_metrics | to_entries | map(.value.commits) | add' data/activity-daily.json)
            local prs=$(jq '.daily_metrics | to_entries | map(.value.prs) | add' data/activity-daily.json)
            local issues=$(jq '.daily_metrics | to_entries | map(.value.issues) | add' data/activity-daily.json)
            echo "  → Total commits: $commits"
            echo "  → Total PRs: $prs"
            echo "  → Total issues: $issues"
        fi
    else
        log_error "Falha ao coletar atividade"
        return 1
    fi
}

# Executar collector de estatísticas
run_stats_collector() {
    print_header "💻 Coletando Estatísticas de Linguagens"
    
    local user="${1:-felipemacedo1}"
    log_info "Usuário: $user"
    
    if timeout 300 ./bin/stats_collector -user="$user"; then
        log_success "Estatísticas coletadas com sucesso"
        
        if [[ -f "data/languages.json" ]]; then
            local langs=$(jq '.languages | length' data/languages.json)
            local bytes=$(jq '.languages | to_entries | map(.value.bytes) | add' data/languages.json)
            local mb=$((bytes / 1048576))
            echo "  → Total de linguagens: $langs"
            echo "  → Total de código: ~${mb}MB"
            
            log_info "Top 3 linguagens:"
            jq -r '.languages | to_entries | sort_by(.value.percentage) | reverse | limit(3;.[]) | "  → \(.key): \(.value.percentage | floor)%"' data/languages.json
        fi
    else
        log_error "Falha ao coletar estatísticas"
        return 1
    fi
}

# Validar todos os arquivos JSON gerados
validate_json_files() {
    print_header "✅ Validando Arquivos JSON"
    
    local files=("profile.json" "projects.json" "metadata.json" "activity-daily.json" "languages.json")
    local all_valid=true
    
    for file in "${files[@]}"; do
        local path="data/$file"
        if [[ -f "$path" ]]; then
            if jq empty "$path" 2>/dev/null; then
                local size=$(ls -lh "$path" | awk '{print $5}')
                log_success "$file ($size)"
            else
                log_error "$file - JSON inválido!"
                all_valid=false
            fi
        else
            log_warning "$file - não encontrado"
        fi
    done
    
    if [[ "$all_valid" == true ]]; then
        log_success "Todos os arquivos JSON são válidos"
        return 0
    else
        log_error "Alguns arquivos JSON estão inválidos"
        return 1
    fi
}

# Exibir resumo final
show_summary() {
    print_header "📊 Resumo Final"
    
    echo "Arquivos gerados:"
    ls -lh data/*.json | awk '{printf "  %-30s %8s\n", $9, $5}'
    
    echo ""
    echo "Total de dados coletados:"
    du -sh data/ | awk '{print "  → " $1}'
}

# Main
main() {
    local start_time=$(date +%s)
    
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║           🧪 TEST E2E - GitHub Metadata Sync                ║"
    echo "║          Simulação completa dos workflows                   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    check_environment
    build_collectors
    
    local user="${1:-felipemacedo1}"
    local days="${2:-30}"
    
    # Executar todos os collectors em sequência
    run_user_collector "$user" || exit 1
    run_repos_collector || exit 1
    run_activity_collector "$user" "$days" || exit 1
    run_stats_collector "$user" || exit 1
    
    # Validar resultados
    validate_json_files || exit 1
    show_summary
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_header "🎉 Teste E2E Concluído com Sucesso!"
    echo "  Tempo total: ${duration}s"
    echo ""
    echo "Os dados estão prontos para commit:"
    echo "  git add data/*.json"
    echo "  git commit -m \"chore: update metadata (test run) [bot]\""
    echo ""
}

# Executar
main "$@"
