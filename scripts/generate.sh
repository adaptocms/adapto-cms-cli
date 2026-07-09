#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SPEC_URL="${SPEC_URL:-https://api.adaptocms.com/openapi.json}"

echo "Fetching OpenAPI spec from ${SPEC_URL}..."
curl -fsSL "$SPEC_URL" -o "$PROJECT_ROOT/api/openapi.json"

echo "Generating client from OpenAPI spec..."
cd "$PROJECT_ROOT/api"
oapi-codegen --config oapi-codegen.yaml openapi.json
echo "Generated internal/client/generated.go"
