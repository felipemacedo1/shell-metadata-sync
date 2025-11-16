#!/bin/bash
set -e

ORG="${1:-growthfolio}"
AUTHOR="${2:-felipemacedo1}"
DAYS="${3:-90}"
OUTPUT="${4:-data/activity-daily-secondary.json}"

echo "📊 Collecting activity for $AUTHOR in org $ORG (last $DAYS days)"

# Calculate date range
END_DATE=$(date -u +%Y-%m-%d)
START_DATE=$(date -u -d "$DAYS days ago" +%Y-%m-%d)

echo "   Period: $START_DATE to $END_DATE"

# Build search query: commits by author in org repos
QUERY="author:$AUTHOR org:$ORG committer-date:$START_DATE..$END_DATE"

echo "🔍 Query: $QUERY"
echo ""

# Use GitHub CLI to search commits
gh api search/commits \
  -X GET \
  -F q="$QUERY" \
  -F per_page=100 \
  --paginate \
  --jq '.items[] | {
    date: (.commit.author.date[:10]),
    repo: .repository.full_name,
    sha: .sha[0:7]
  }' > /tmp/commits.jsonl 2>/dev/null || true

# Count commits per day
if [ -s /tmp/commits.jsonl ]; then
  cat /tmp/commits.jsonl | jq -r '.date' | sort | uniq -c | awk '{print $2, $1}'
  TOTAL=$(wc -l < /tmp/commits.jsonl)
  echo ""
  echo "✅ Found $TOTAL commits"
else
  echo "⚠️  No commits found"
fi

