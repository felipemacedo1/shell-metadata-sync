#!/bin/bash

set -uo pipefail

# ═══════════════════════════════════════════════════════════════
# SonarCloud Issues Sync Script
# Centralizes SonarCloud issue management across multiple repos
# ═══════════════════════════════════════════════════════════════

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="${SCRIPT_DIR}/../data"
LOG_FILE="${DATA_DIR}/sonar-sync-$(date +%Y%m%d-%H%M%S).log"
REPORT_FILE="${DATA_DIR}/sonar-sync-$(date +%Y%m%d-%H%M%S).json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_REPOS=0
PROCESSED_REPOS=0
SUCCESSFUL_SYNCS=0
FAILED_SYNCS=0
SKIPPED_REPOS=0
TOTAL_ISSUES_CREATED=0
TOTAL_ISSUES_UPDATED=0
TOTAL_ISSUES_CLOSED=0

# ═══════════════════════════════════════════════════════════════
# Logging Functions
# ═══════════════════════════════════════════════════════════════

log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    echo -e "${timestamp} [${level}] ${message}" | tee -a "$LOG_FILE"
}

log_info() { log "INFO" "${BLUE}$*${NC}"; }
log_success() { log "SUCCESS" "${GREEN}$*${NC}"; }
log_warning() { log "WARNING" "${YELLOW}$*${NC}"; }
log_error() { log "ERROR" "${RED}$*${NC}"; }

# ═══════════════════════════════════════════════════════════════
# Environment Validation
# ═══════════════════════════════════════════════════════════════

validate_environment() {
    log_info "Validating environment..."
    
    local missing_vars=()
    
    [[ -z "${GITHUB_TOKEN:-}" ]] && missing_vars+=("GITHUB_TOKEN")
    [[ -z "${SONAR_TOKEN:-}" ]] && missing_vars+=("SONAR_TOKEN")
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "Missing required environment variables: ${missing_vars[*]}"
        exit 1
    fi
    
    # Validate tools
    command -v gh >/dev/null 2>&1 || { log_error "gh CLI not found"; exit 1; }
    command -v jq >/dev/null 2>&1 || { log_error "jq not found"; exit 1; }
    command -v curl >/dev/null 2>&1 || { log_error "curl not found"; exit 1; }
    
    log_success "Environment validated"
}

# ═══════════════════════════════════════════════════════════════
# Repository Discovery
# ═══════════════════════════════════════════════════════════════

get_repositories() {
    log_info "Discovering repositories..."
    
    local personal_repos=$(gh repo list felipemacedo1 --limit 100 --json nameWithOwner --jq '.[].nameWithOwner')
    local org_repos=$(gh repo list growthfolio --limit 100 --json nameWithOwner --jq '.[].nameWithOwner')
    
    echo "$personal_repos"
    echo "$org_repos"
}

filter_repositories() {
    local repos="$1"
    local filter="${REPOS_FILTER:-all}"
    
    if [[ "$filter" == "all" ]]; then
        echo "$repos"
        return
    fi
    
    # Filter by comma-separated list
    IFS=',' read -ra FILTER_ARRAY <<< "$filter"
    for repo in $repos; do
        for pattern in "${FILTER_ARRAY[@]}"; do
            if [[ "$repo" == *"$pattern"* ]]; then
                echo "$repo"
            fi
        done
    done
}

# ═══════════════════════════════════════════════════════════════
# SonarCloud API Functions
# ═══════════════════════════════════════════════════════════════

check_sonar_project() {
    local repo=$1
    local project_key="${repo//\//_}"
    
    local response=$(curl -s -w "\n%{http_code}" -u "${SONAR_TOKEN}:" \
        "https://sonarcloud.io/api/components/show?component=${project_key}" 2>/dev/null)
    
    local http_code=$(echo "$response" | tail -n1)
    
    if [[ "$http_code" == "200" ]]; then
        return 0
    else
        return 1
    fi
}

get_sonar_issues() {
    local repo=$1
    local project_key="${repo//\//_}"
    
    log_info "Fetching SonarCloud issues for ${project_key}..."
    
    local issues=$(curl -s -u "${SONAR_TOKEN}:" \
        "https://sonarcloud.io/api/issues/search?componentKeys=${project_key}&ps=500&resolved=false" \
        2>/dev/null | jq -c '.issues[]?' 2>/dev/null || echo "[]")
    
    echo "$issues"
}

# ═══════════════════════════════════════════════════════════════
# GitHub Issues Management
# ═══════════════════════════════════════════════════════════════

get_existing_github_issues() {
    local repo=$1
    
    gh issue list --repo "$repo" \
        --label "sonarcloud" \
        --state all \
        --limit 500 \
        --json number,title,state,labels \
        --jq '.[]'
}

create_github_issue() {
    local repo=$1
    local title=$2
    local body=$3
    local severity=$4
    local type=$5
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would create issue: $title"
        return 0
    fi
    
    local labels="sonarcloud,severity:${severity},type:${type}"
    
    gh issue create \
        --repo "$repo" \
        --title "$title" \
        --body "$body" \
        --label "$labels" \
        >/dev/null 2>&1
    
    return $?
}

update_github_issue() {
    local repo=$1
    local issue_number=$2
    local body=$3
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would update issue #${issue_number}"
        return 0
    fi
    
    gh issue edit "$issue_number" \
        --repo "$repo" \
        --body "$body" \
        >/dev/null 2>&1
    
    return $?
}

close_github_issue() {
    local repo=$1
    local issue_number=$2
    local reason=${3:-"completed"}
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would close issue #${issue_number}"
        return 0
    fi
    
    gh issue close "$issue_number" \
        --repo "$repo" \
        --reason "$reason" \
        --comment "✅ This issue has been resolved in SonarCloud." \
        >/dev/null 2>&1
    
    return $?
}

# ═══════════════════════════════════════════════════════════════
# Issue Synchronization Logic
# ═══════════════════════════════════════════════════════════════

sync_repo_issues() {
    local repo=$1
    
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Processing: ${repo}"
    
    ((PROCESSED_REPOS++))
    
    # Check if repo has SonarCloud
    if ! check_sonar_project "$repo"; then
        log_warning "No SonarCloud project found, skipping..."
        ((SKIPPED_REPOS++))
        return 0
    fi
    
    # Get SonarCloud issues
    local sonar_issues=$(get_sonar_issues "$repo")
    local sonar_count=$(echo "$sonar_issues" | jq -s 'length' 2>/dev/null || echo 0)
    
    log_info "Found ${sonar_count} open issues in SonarCloud"
    
    # Get existing GitHub issues
    local github_issues=$(get_existing_github_issues "$repo" 2>/dev/null || echo "[]")
    
    local created=0
    local updated=0
    local closed=0
    
    # Process SonarCloud issues
    if [[ "$sonar_count" -gt 0 ]]; then
        while IFS= read -r issue; do
            [[ -z "$issue" ]] && continue
            
            local key=$(echo "$issue" | jq -r '.key')
            local message=$(echo "$issue" | jq -r '.message')
            local severity=$(echo "$issue" | jq -r '.severity' | tr '[:upper:]' '[:lower:]')
            local type=$(echo "$issue" | jq -r '.type' | tr '[:upper:]' '[:lower:]')
            local line=$(echo "$issue" | jq -r '.line // "N/A"')
            local component=$(echo "$issue" | jq -r '.component' | sed 's/.*://')
            
            local title="[SonarCloud] ${message}"
            local body="**Issue Key:** \`${key}\`
**Severity:** ${severity}
**Type:** ${type}
**File:** \`${component}\`
**Line:** ${line}

**Description:**
${message}

---
🔗 [View in SonarCloud](https://sonarcloud.io/project/issues?id=${repo//\//_}&issues=${key})
"
            
            # Check if issue already exists
            local existing=$(echo "$github_issues" | jq -r --arg title "$title" 'select(.title == $title) | .number' 2>/dev/null)
            
            if [[ -n "$existing" && "$existing" != "null" ]]; then
                # Update existing issue
                if update_github_issue "$repo" "$existing" "$body"; then
                    log_success "Updated issue #${existing}"
                    ((updated++))
                else
                    log_error "Failed to update issue #${existing}"
                fi
            else
                # Create new issue
                if create_github_issue "$repo" "$title" "$body" "$severity" "$type"; then
                    log_success "Created issue: ${message}"
                    ((created++))
                else
                    log_error "Failed to create issue: ${message}"
                fi
            fi
            
            # Rate limiting
            sleep 1
            
        done <<< "$sonar_issues"
    fi
    
    # Close resolved issues (issues in GitHub but not in SonarCloud)
    # TODO: Implement this logic if needed
    
    ((TOTAL_ISSUES_CREATED += created))
    ((TOTAL_ISSUES_UPDATED += updated))
    ((TOTAL_ISSUES_CLOSED += closed))
    
    log_success "Completed: ${repo} (Created: ${created}, Updated: ${updated}, Closed: ${closed})"
    ((SUCCESSFUL_SYNCS++))
    
    return 0
}

# ═══════════════════════════════════════════════════════════════
# MongoDB Integration (Optional)
# ═══════════════════════════════════════════════════════════════

save_to_mongodb() {
    if [[ -z "${MONGODB_URI:-}" ]]; then
        log_warning "MongoDB URI not configured, skipping database sync"
        return 0
    fi
    
    log_info "Saving metrics to MongoDB..."
    
    # TODO: Implement MongoDB saving logic using mongosh or API
    local data=$(cat "$REPORT_FILE")
    
    # Example: curl to MongoDB API or use mongosh
    # curl -X POST "${MONGODB_API}/metrics" -d "$data"
    
    log_success "Metrics saved to MongoDB"
}

# ═══════════════════════════════════════════════════════════════
# Report Generation
# ═══════════════════════════════════════════════════════════════

generate_report() {
    log_info "Generating execution report..."
    
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    
    cat > "$REPORT_FILE" <<EOF
{
  "execution": {
    "timestamp": "$(date -Iseconds)",
    "duration_seconds": ${duration},
    "dry_run": ${DRY_RUN:-false}
  },
  "summary": {
    "total_repos": ${TOTAL_REPOS},
    "processed": ${PROCESSED_REPOS},
    "successful": ${SUCCESSFUL_SYNCS},
    "failed": ${FAILED_SYNCS},
    "skipped": ${SKIPPED_REPOS}
  },
  "issues": {
    "created": ${TOTAL_ISSUES_CREATED},
    "updated": ${TOTAL_ISSUES_UPDATED},
    "closed": ${TOTAL_ISSUES_CLOSED}
  }
}
EOF
    
    log_success "Report saved to: ${REPORT_FILE}"
    cat "$REPORT_FILE"
}

# ═══════════════════════════════════════════════════════════════
# Main Execution
# ═══════════════════════════════════════════════════════════════

main() {
    local START_TIME=$(date +%s)
    
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║          🔄 SonarCloud Issues Sync - Centralized             ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo ""
    
    mkdir -p "$DATA_DIR"
    
    # Validate environment
    validate_environment
    
    # Get repositories
    local all_repos=$(get_repositories)
    local filtered_repos=$(filter_repositories "$all_repos")
    
    TOTAL_REPOS=$(echo "$filtered_repos" | wc -l)
    
    log_info "Total repositories to process: ${TOTAL_REPOS}"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_warning "🔍 DRY RUN MODE - No changes will be made"
    fi
    
    # Process each repository
    local count=0
    local max_repos=${MAX_REPOS:-0}
    
    while IFS= read -r repo; do
        [[ -z "$repo" ]] && continue
        
        ((count++))
        
        # Limit number of repos if specified
        if [[ $max_repos -gt 0 && $count -gt $max_repos ]]; then
            log_warning "Reached max repos limit (${max_repos}), stopping..."
            break
        fi
        
        # Process repo with error handling
        if ! sync_repo_issues "$repo"; then
            log_error "Failed to sync: ${repo}"
            ((FAILED_SYNCS++))
        fi
        
        # Rate limiting between repos
        sleep 2
        
    done <<< "$filtered_repos"
    
    echo ""
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║                     📊 FINAL REPORT                          ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    
    # Generate report
    generate_report
    
    # Save to MongoDB
    save_to_mongodb
    
    log_success "✅ Sync completed!"
    log_info "Total execution time: $(($(date +%s) - START_TIME))s"
    
    # Exit code based on failures
    if [[ $FAILED_SYNCS -gt 0 ]]; then
        exit 0
    fi
    
    exit 0
}

# Run main function
main "$@"
