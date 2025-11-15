#!/bin/bash

# Script to create secondary profile data files for aggregation
# This takes the current data (which is from growthfolio) and saves it as secondary files
# Then the primary files should be regenerated for felipemacedo1

set -e

DATA_DIR="./data"
DASHBOARD_DATA_DIR="./dashboard/public/data"

echo "📊 Creating secondary profile data files for aggregation..."

# Check if data directory exists
if [ ! -d "$DATA_DIR" ]; then
    echo "❌ Data directory not found: $DATA_DIR"
    exit 1
fi

# Copy current data as secondary (growthfolio)
echo "📋 Copying growthfolio data as secondary..."

if [ -f "$DATA_DIR/profile.json" ]; then
    cp "$DATA_DIR/profile.json" "$DATA_DIR/profile-secondary.json"
    echo "✓ Created profile-secondary.json"
fi

if [ -f "$DATA_DIR/activity-daily.json" ]; then
    cp "$DATA_DIR/activity-daily.json" "$DATA_DIR/activity-daily-secondary.json"
    echo "✓ Created activity-daily-secondary.json"
fi

if [ -f "$DATA_DIR/languages.json" ]; then
    cp "$DATA_DIR/languages.json" "$DATA_DIR/languages-secondary.json"
    echo "✓ Created languages-secondary.json"
fi

echo ""
echo "🔄 Now you should run the collectors for felipemacedo1 to generate primary data:"
echo ""
echo "  cd bin"
echo "  ./user_collector -user=felipemacedo1 -token=\$GH_TOKEN"
echo "  ./activity_collector -user=felipemacedo1 -token=\$GH_TOKEN -days=90"
echo "  ./stats_collector -user=felipemacedo1 -token=\$GH_TOKEN"
echo ""
echo "✅ Secondary files created successfully!"
echo ""
echo "📁 Files in $DATA_DIR:"
ls -lh "$DATA_DIR"/*.json 2>/dev/null || echo "No JSON files found"
