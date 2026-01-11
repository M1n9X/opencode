# Bubble TUI Testing - Message Sending Debug

## ✅ Progress

- **/models command works** - Command matching system is functional
- **Enter key reaches TUI layer** - textarea no longer intercepts it
- **Message sending still fails** - Need to debug Submit() flow

## 🧪 Test with Debug Logs

### Terminal 1 - Server

```bash
cd packages/opencode
bun run src/index.ts serve --port 3001
```

### Terminal 2 - TUI with Debug Output

```bash
cd /Users/mxue/GitRepos/Codebreeze/opencode
OPENCODE_SERVER=http://localhost:3001 packages/tui/dist/tui . 2>&1 | tee /tmp/tui-test.log
```

### Test Steps

1. Type `hello world`
2. Press **Enter**
3. Check the output for debug messages

### Expected Debug Output

**If InputSubmitCommand is triggered:**

```
[DEBUG] KeyPress: "enter"
[DEBUG] Matching commands for key="enter"
[DEBUG] Matched 1 commands
[DEBUG] Executing command: input_submit
[DEBUG] InputSubmitCommand triggered!
[DEBUG] Submit() called, value="hello world"
```

**If Submit() returns early:**

```
[DEBUG] Submit() called, value=""
[DEBUG] Submit() empty value, returning
```

### Possible Issues

**Case 1: No "InputSubmitCommand triggered" message**
→ Command isn't matching or being executed
→ Check if Enter is being handled elsewhere

**Case 2: "Submit() called, value=''"**  
→ Editor value is empty when Submit() is called
→ Editor.Value() might not be returning the text

**Case 3: Submit() called but no message sent**
→ Issue is in Submit() logic or SendPrompt flow

## Next Steps

Based on the debug output, we'll know exactly where the flow breaks and can fix it.

---
**Status**: Awaiting test results with debug output
**Binary**: packages/tui/dist/tui (with comprehensive logging)
