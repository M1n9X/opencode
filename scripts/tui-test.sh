#!/bin/bash
# Run Go TUI tests
# Usage: ./scripts/tui-test.sh [--verbose]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TUI_DIR="$ROOT_DIR/packages/tui"

VERBOSE=""
if [[ "$1" == "--verbose" ]] || [[ "$1" == "-v" ]]; then
    VERBOSE="-v"
fi

echo "🧪 OpenCode Bubble TUI Test Runner"
echo "==================================="
echo ""

cd "$TUI_DIR"

echo "📋 Running Go tests..."
echo ""

go test $VERBOSE ./...

echo ""
echo "✅ All tests passed!"
