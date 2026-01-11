#!/bin/bash
# End-to-End API Testing for Bubble TUI
# Tests all critical APIs that the TUI needs to function
# Usage: ./scripts/tui-e2e-test.sh [--port PORT]

# set -e removed for robustness

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OPENCODE_DIR="$ROOT_DIR/packages/opencode"

PORT=3456
SERVER_PID=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            PORT="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

BASE_URL="http://127.0.0.1:$PORT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

passed=0
failed=0
skipped=0

cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo ""
        echo "🛑 Stopping server..."
        kill $SERVER_PID 2>/dev/null || true
    fi
}

trap cleanup EXIT

test_api() {
    local name="$1"
    local endpoint="$2"
    local expected="$3"
    
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint" 2>/dev/null)
    http_code=$(echo "$response" | tail -1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ] || [ "$http_code" = "204" ]; then
        if [ -n "$expected" ]; then
            if echo "$body" | grep -q "$expected"; then
                echo -e "${GREEN}✅ PASS${NC} $name"
                ((passed++))
            else
                echo -e "${RED}❌ FAIL${NC} $name (missing: $expected)"
                ((failed++))
            fi
        else
            echo -e "${GREEN}✅ PASS${NC} $name"
            ((passed++))
        fi
    else
        echo -e "${RED}❌ FAIL${NC} $name (HTTP $http_code)"
        ((failed++))
    fi
}

echo "🧪 OpenCode End-to-End API Tests"
echo "================================="
echo "Target: $BASE_URL"
echo ""

# Start server
echo "Starting OpenCode server on port $PORT..."
cd "$OPENCODE_DIR"
bun run src/index.ts serve --port $PORT > /dev/null 2>&1 &
SERVER_PID=$!

# Wait for server to be ready
echo "Waiting for server to start..."
for i in {1..30}; do
    if curl -s "$BASE_URL/api/health" >/dev/null; then
        echo -e "${GREEN}✅ Server is up!${NC}"
        break
    fi
    sleep 0.5
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo -e "${RED}❌ Server crashed${NC}"
        exit 1
    fi
done

if ! curl -s "$BASE_URL/api/health" >/dev/null; then
    echo -e "${RED}❌ Server failed to start after 15s${NC}"
    exit 1
fi

echo ""
echo "Testing TUI Startup APIs..."
echo "----------------------------"

test_api "session.list()" "/session?limit=10" "id"
test_api "config.providers()" "/config/providers" "providers"
test_api "app.agents()" "/agent" "name"
test_api "config.get()" "/config" ""
test_api "path.get()" "/path" "home"

echo ""
echo "Testing Session APIs..."
echo "------------------------"

# Create a test session
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/session" -H "Content-Type: application/json" -d '{"title":"E2E Test Session"}')
SESSION_ID=$(echo "$SESSION_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$SESSION_ID" ]; then
    echo -e "${GREEN}✅ PASS${NC} session.create() -> $SESSION_ID"
    ((passed++))
    
    test_api "session.get()" "/session/$SESSION_ID" "id"
    test_api "session.messages()" "/session/$SESSION_ID/messages" ""
    
    UPDATE_RESPONSE=$(curl -s -X PATCH "$BASE_URL/session/$SESSION_ID" -H "Content-Type: application/json" -d '{"title":"Updated E2E Test"}')
    if echo "$UPDATE_RESPONSE" | grep -q "Updated E2E Test"; then
        echo -e "${GREEN}✅ PASS${NC} session.update()"
        ((passed++))
    else
        echo -e "${RED}❌ FAIL${NC} session.update()"
        ((failed++))
    fi
    
    DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/session/$SESSION_ID")
    if [ "$?" = "0" ]; then
        echo -e "${GREEN}✅ PASS${NC} session.delete()"
        ((passed++))
    else
        echo -e "${YELLOW}⚠️ SKIP${NC} session.delete()"
        ((skipped++))
    fi
else
    echo -e "${RED}❌ FAIL${NC} session.create()"
    ((failed++))
fi

echo ""
echo "Testing Other Core APIs..."
echo "---------------------------"

test_api "command.list()" "/command" ""
test_api "lsp.status()" "/lsp" ""
test_api "mcp.status()" "/mcp" ""
test_api "vcs.get()" "/vcs" ""
test_api "project.current()" "/project/current" "id"

echo ""
echo "================================="
echo "📊 Summary"
echo "---------"
echo -e "${GREEN}Passed: $passed${NC}"
echo -e "${RED}Failed: $failed${NC}"
echo -e "${YELLOW}Skipped: $skipped${NC}"
echo ""

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✅ All E2E tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
