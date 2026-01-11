#!/bin/bash
# Build Go TUI binary only
# Usage: ./scripts/tui-build.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TUI_DIR="$ROOT_DIR/packages/tui"

echo "📦 Building Go Bubble TUI"
echo "========================="
echo ""

cd "$TUI_DIR"

# Clean and build
mkdir -p dist
echo "🔨 Compiling..."
go build -o ./dist/tui ./cmd/opencode/main.go

echo ""
echo "✅ Build successful!"
echo "   Binary: $TUI_DIR/dist/tui"
echo ""
echo "To run the TUI:"
echo "   OPENCODE_SERVER=http://localhost:3000 ./packages/tui/dist/tui"
