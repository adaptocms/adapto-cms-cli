#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SPEC_URL="${SPEC_URL:-https://api.adaptocms.com/openapi.json}"

echo "Fetching OpenAPI spec from ${SPEC_URL}..."
curl -fsSL "$SPEC_URL" -o "$PROJECT_ROOT/api/openapi.json"

echo "Downgrading spec 3.1 -> 3.0 for oapi-codegen..."
python3 "$SCRIPT_DIR/downgrade_openapi.py" "$PROJECT_ROOT/api/openapi.json" "$PROJECT_ROOT/api/openapi.json"

echo "Generating client from OpenAPI spec..."
cd "$PROJECT_ROOT/api"
oapi-codegen --config oapi-codegen.yaml openapi.json
echo "Generated internal/client/generated.go"
