#!/bin/bash

# Script to copy data files from /data to dashboard/public/data
# This ensures the dashboard always has the latest data

set -e

# Get script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

SOURCE_DIR="$PROJECT_ROOT/data"
DEST_DIR="$PROJECT_ROOT/dashboard/public/data"

echo "🔄 Syncing data files to dashboard..."

# Create destination directory if it doesn't exist
mkdir -p "$DEST_DIR"

# Copy all JSON files
if [ -d "$SOURCE_DIR" ]; then
    cp -v "$SOURCE_DIR"/*.json "$DEST_DIR/" 2>/dev/null || echo "⚠️  No JSON files found in $SOURCE_DIR"
    echo "✅ Data sync complete!"
    echo "📊 Files synced to $DEST_DIR"
else
    echo "❌ Source directory $SOURCE_DIR not found!"
    exit 1
fi

# List copied files
echo ""
echo "📁 Available data files:"
ls -lh "$DEST_DIR"
