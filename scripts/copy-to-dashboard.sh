#!/bin/bash
set -e

echo "📦 Copying data files to dashboard..."

mkdir -p dashboard/public/data

cp -v data/*.json dashboard/public/data/ 2>/dev/null || echo "⚠️ No JSON files found in data/"

echo "✅ Data copied successfully"
ls -lh dashboard/public/data/
