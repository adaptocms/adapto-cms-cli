#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Generating client from OpenAPI spec..."
cd "$PROJECT_ROOT/api"
oapi-codegen --config oapi-codegen.yaml openapi.json
echo "Generated internal/client/generated.go"
