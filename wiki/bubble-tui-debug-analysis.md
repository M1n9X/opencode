# Bubble TUI Message Sending Issue - Root Cause Analysis

## Problem Report

User confirms Bubble TUI launches successfully but:

1. ❌ Typing message + Enter → No effect
2. ❌ Slash commands like `/models` → Don't work

## Root Cause Analysis

### API Compatibility Comparison

#### opentui (TypeScript) - Working Reference

File: `packages/opencode/src/cli/cmd/tui/component/prompt/index.tsx`

**Line 533-588: Three message types**:

1. **Shell Commands** (line 534-543):

   ```typescript
   sdk.client.session.shell({
     sessionID,
     agent: local.agent.current().name,
     model: {
       providerID: selectedModel.providerID,
       modelID: selectedModel.modelID,
     },
     command: inputText,
   })
   ```

2. **Slash Commands** (line 553-567):

   ```typescript
   sdk.client.session.command({
     sessionID,
     command: command.slice(1),
     arguments: args.join(" "),
     agent: local.agent.current().name,
     model: `${selectedModel.providerID}/${selectedModel.modelID}`,
     messageID,
     variant,
     parts: [...],
   })
   ```

3. **Normal Messages** (line 569-587):

   ```typescript
   sdk.client.session.prompt({
     sessionID,
     ...selectedModel,
     messageID,
     agent: local.agent.current().name,
     model: selectedModel,
     variant,
     parts: [{
       id: Identifier.ascending("part"),
       type: "text",
       text: inputText,
     }, ...],
   })
   ```

#### Bubble TUI (Go) - v0.15.31 SDK

File: `packages/tui/internal/app/app.go`

**Line 771-807: SendPrompt**:

```go
_, err := a.Client.Session.Prompt(ctx, a.Session.ID, opencode.SessionPromptParams{
    Model: opencode.F(opencode.SessionPromptParamsModel{
        ProviderID: opencode.F(a.Provider.ID),
        ModelID:    opencode.F(a.Model.ID),
    }),
    Agent:     opencode.F(a.Agent().Name),
    MessageID: opencode.F(messageID),
    Parts:     opencode.F(message.ToSessionChatParams()),
})
```

**Line 810-844: SendCommand**:

```go
params := opencode.SessionCommandParams{
    Command:   opencode.F(command),
    Arguments: opencode.F(args),
    Agent:     opencode.F(a.Agents[a.AgentIndex].Name),
}
if a.Provider != nil && a.Model != nil {
    params.Model = opencode.F(a.Provider.ID + "/" + a.Model.ID)
}
_, err := a.Client.Session.Command(ctx, a.Session.ID, params)
```

**Line 847-876: SendShell**:

```go
_, err := a.Client.Session.Shell(ctx, a.Session.ID, opencode.SessionShellParams{
    Agent:   opencode.F(a.Agent().Name),
    Command: opencode.F(command),
})
```

---

## Key Findings

### ✅ Good News

1. **All APIs exist in v0.15.31 SDK**:
   - `Session.Prompt()` ✅
   - `Session.Command()` ✅
   - `Session.Shell()` ✅

2. **Bubble TUI has all handlers**:
   - `SendPrompt()` ✅
   - `SendCommand()` ✅
   - `SendShell()` ✅

### ⚠️ Potential Issues

#### 1. Model Parameter Format Difference

**opentui**:

```typescript
model: selectedModel,  // Object with providerID + modelID
```

**Bubble TUI Prompt**:

```go
Model: opencode.F(opencode.SessionPromptParamsModel{
    ProviderID: opencode.F(a.Provider.ID),
    ModelID:    opencode.F(a.Model.ID),
}),
```

**Bubble TUI Command**:

```go
params.Model = opencode.F(a.Provider.ID + "/" + a.Model.ID)  // String format!
```

❗**Inconsistency**: Command uses string `provider/model`, Prompt uses object

#### 2. Missing `variant` Field

opentui sends:

```typescript
variant,  // e.g., "thinking", "fast", etc.
```

Bubble TUI SendPrompt: **No variant field**

#### 3. Shell API Missing Model

- opentui: Sends `model: { providerID, modelID }`
- Bubble TUI: **No model field at all**

---

## Investigation Needed

### 1. Check UI Event Handling

File: `packages/tui/cmd/opencode/main.go`

Need to verify:

- Does Enter key correctly trigger `SendPrompt()`?
- Does `/` prefix correctly trigger `SendCommand()`?
- Are keypresses being captured?

### 2. Check Go SDK Session Methods

File: `packages/sdk/go/session.go` (lines 173-182, 109-118)

Need to compare v0.15.31 vs v1.1.13:

- Are parameter structures compatible?
- Are there new required fields?

### 3. Server-Side API Contract

Need to check `/doc` endpoint for:

- Required vs optional fields
- Model parameter format (object vs string)
- Variant field requirement

---

## Next Steps (Priority Order)

### 🔴 High Priority - Test Input Handling

1. Add debug logging to `SendPrompt()` to verify it's being called
2. Check if Enter key event reaches the handler
3. Verify `/` detection for commands

### 🟡 Medium Priority - API Parameter Alignment

1. Compare v0.15.31 SDK types vs server's OpenAPI schema
2. Add missing `variant` field to Prompt
3. Standardize `model` parameter format

### 🟢 Low Priority - Shell API Enhancement

1. Add model parameter to Shell API call
2. Test shell commands (!ls, etc.)

---

## Hypothesis

**Most Likely**: Input event handling is broken OR Send methods are not being triggered.

**Less Likely**: API parameters are incompatible (but both APIs exist, so partial compatibility).

**Action**: Need to add logging to trace input → handler → API call flow.

---

**Status**: Analysis complete, awaiting user confirmation for debugging approach
**Next Tools**: Add debug logging, test input flow, compare API schemas
