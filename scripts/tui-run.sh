#!/bin/bash
# Build and Run Go Bubble TUI
# Usage: ./scripts/tui-run.sh [--build-only] [--port PORT]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TUI_DIR="$ROOT_DIR/packages/tui"
OPENCODE_DIR="$ROOT_DIR/packages/opencode"
TUI_BINARY="$TUI_DIR/dist/tui"

PORT=3456
BUILD_ONLY=false
REBUILD=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --build-only)
            BUILD_ONLY=true
            shift
            ;;
        --rebuild)
            REBUILD=true
            shift
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--build-only] [--rebuild] [--port PORT]"
            exit 1
            ;;
    esac
done

echo "🔧 OpenCode Bubble TUI Runner"
echo "=============================="

# Build TUI binary if needed
if [ "$REBUILD" = true ] || [ ! -f "$TUI_BINARY" ]; then
    echo "📦 Building Go TUI binary..."
    cd "$TUI_DIR"
    mkdir -p dist
    go build -o ./dist/tui ./cmd/opencode/main.go
    echo "✅ TUI binary built: $TUI_BINARY"
else
    echo "✅ TUI binary exists: $TUI_BINARY"
fi

if [ "$BUILD_ONLY" = true ]; then
    echo "Build complete. Exiting."
    exit 0
fi

# Start server and TUI
echo ""

# Check if port is already in use
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "❌ Port $PORT is already in use!"
    echo "   Try a different port with: $0 --port <PORT>"
    echo "   Or kill the process using: kill \$(lsof -t -i:$PORT)"
    exit 1
fi

echo "🚀 Starting OpenCode server on port $PORT..."
cd "$OPENCODE_DIR"

# Start server in background
bun run src/index.ts serve --port $PORT &
SERVER_PID=$!

# Wait for server to be ready
echo "Waiting for server to start..."
MAX_RETRIES=30
for ((i=1; i<=MAX_RETRIES; i++)); do
    if curl -s "http://127.0.0.1:$PORT/api/health" >/dev/null; then
        echo "✅ Server is up!"
        break
    fi
    sleep 0.5
    
    # Check if server process is still alive
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo "❌ Server process died unexpectedly"
        wait $SERVER_PID
        exit 1
    fi
    
    if [ $i -eq $MAX_RETRIES ]; then
        echo "❌ Timed out waiting for server to start"
        kill $SERVER_PID 2>/dev/null || true
        exit 1
    fi
done

echo "✅ Server started (PID: $SERVER_PID)"
echo ""
echo "🖥️  Launching Bubble TUI..."
echo ""

# Run TUI
OPENCODE_SERVER="http://127.0.0.1:$PORT" "$TUI_BINARY"

# Cleanup: stop server when TUI exits
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID 2>/dev/null || true
echo "✅ Done"
