#!/bin/bash
set -e

echo "📦 Copying data files from ./data to ./dashboard/public/data"

# Create target directory
mkdir -p dashboard/public/data

# Copy all JSON files
echo "📂 Source files:"
ls -lh data/*.json

echo ""
echo "📋 Copying files..."
cp -v data/*.json dashboard/public/data/

echo ""
echo "✅ Done! Files in dashboard/public/data:"
ls -lh dashboard/public/data/
