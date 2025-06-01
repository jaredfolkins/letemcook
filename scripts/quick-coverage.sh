#!/bin/bash
# Quick coverage script - can be run from any directory within the project

# Find the project root (directory containing go.mod)
PROJECT_ROOT=""
current_dir="$(pwd)"

while [[ "$current_dir" != "/" ]]; do
    if [[ -f "$current_dir/go.mod" ]]; then
        PROJECT_ROOT="$current_dir"
        break
    fi
    current_dir="$(dirname "$current_dir")"
done

if [[ -z "$PROJECT_ROOT" ]]; then
    echo "❌ Error: Could not find go.mod file. Please run from within the project directory."
    exit 1
fi

echo "📁 Project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"
exec ./scripts/coverage.sh 