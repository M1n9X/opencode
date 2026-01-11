# Bubble TUI Revival Project

**Project Status**: 🚧 In Progress  
**Start Date**: 2026-01-11  
**Target Branch**: `refactor`  
**Base Version**: v1.1.13 (origin/dev)  
**Goal**: Restore Go Bubble TUI while maintaining latest backend features

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Project Background](#project-background)
- [Architecture Evolution](#architecture-evolution)
- [Strategy Decision](#strategy-decision)
- [API Evolution Analysis](#api-evolution-analysis)
- [Migration Phases](#migration-phases)
- [Implementation Checklist](#implementation-checklist)
- [Testing Plan](#testing-plan)
- [Risks and Mitigation](#risks-and-mitigation)
- [References](#references)

---

## Executive Summary

### Objective

Migrate the `refactor` branch to use the **Go Bubble TUI** from v0.15.31 while syncing all backend updates from the latest `dev` branch (v1.1.13), effectively replacing the current TypeScript/Ink TUI implementation.

### Key Metrics

| Metric                | Value                      |
| :-------------------- | :------------------------- |
| **Commits to Bridge** | 3,419 commits              |
| **Timeline**          | Oct 31, 2025 → Jan 11, 2026|
| **Server API Growth** | +82% (1,589 → 2,894 lines) |
| **Estimated Effort**  | 3-5 days                   |
| **Risk Level**        | Medium                     |

### Chosen Strategy

**Strategy B: Backward Adapt**  
Start from latest `dev`, remove TypeScript TUI, port back Bubble TUI with API adaptations.

**Rationale**: Far more practical than cherry-picking 3,419 commits. Clear scope, testable phases, future-proof.

---

## Project Background

### Timeline

1. **v0.15.31 (Oct 31, 2025)**: Original release with Go Bubble TUI
   - Backend: TypeScript/Hono (1,589 lines)
   - Frontend: Go Bubble TUI (`packages/tui/`)
   - API: HTTP/REST + SSE

2. **Commit 96bdeb3c7 (Oct 31, 2025)**: "OpenTUI is here"
   - Introduced TypeScript/Ink TUI in `packages/opencode/src/cli/cmd/tui/`
   - Both TUIs coexisted briefly

3. **Commit f68374ad2 (Nov 2, 2025)**: "DELETE GO BUBBLETEA CRAP HOORAY"
   - Go Bubble TUI completely removed
   - Only TypeScript TUI remains

4. **v1.1.13 (Current dev)**: Latest version
   - Backend: TypeScript/Hono (2,894 lines, +82%)
   - Frontend: TypeScript/Ink TUI only
   - API: HTTP/REST + SSE + WebSockets

### Current State

- `refactor` branch: Based on v0.15.31 (still has Bubble TUI)
- `dev` branch: At v1.1.13 (TypeScript TUI only)
- Goal: `refactor` → latest dev backend + Bubble TUI frontend

---

## Architecture Evolution

### Component Comparison

| Component            | v0.15.31           | v1.1.13 (Current)                  |
| :------------------- | :----------------- | :--------------------------------- |
| **Backend Core**     | packages/opencode  | packages/opencode                  |
| **Server API**       | 1,589 lines        | 2,894 lines (+82%)                 |
| **TUI Frontend**     | Go (Bubble Tea)    | TypeScript (Ink)                   |
| **TUI Location**     | packages/tui/      | packages/opencode/src/cli/cmd/tui/ |
| **Communication**    | HTTP + SSE         | HTTP + SSE + WebSocket             |
| **SDK Client**       | Go SDK (Stainless) | TypeScript SDK                     |
| **Event Streaming**  | SSE only           | SSE + Custom events                |

### API Evolution Summary

Based on reviewing `wiki/10_OpenCode_SDK_API_Reference.md`:

**New Endpoints** (likely added post-v0.15.31):

- `pty.*` - PTY session management (6 endpoints)
- `question.*` - AI question handling (3 endpoints)
- `permission.*` refactored to `PermissionNext`
- `tui.*` - TUI control APIs (13 endpoints, TypeScript TUI specific)
- `experimental.worktree.*` - Git worktree support
- `experimental.resource.*` - MCP resources

**Enhanced Endpoints**:

- `session.prompt()` - More part types (subtask, agent)
- `config.providers()` - Enhanced model capabilities
- WebSocket support for PTY connections

**Breaking Changes** (potential):

- Permission model: `Permission` → `PermissionNext`
- Event types: New event categories (PTY, TUI control)
- Error handling: More detailed error types

---

## Strategy Decision

### Strategy A vs Strategy B

| Aspect           | Strategy A (Forward Port)           | Strategy B (Backward Adapt)              |
| :--------------- | :---------------------------------- | :--------------------------------------- |
| **Approach**     | v0.15.31 + cherry-pick 3419 commits | Latest dev - TypeScript TUI + Bubble TUI |
| **Complexity**   | Very High                           | Medium                                   |
| **Workload**     | 2-4 weeks                           | 3-5 days                                 |
| **Risk**         | High (merge conflicts, dependencies)| Medium (API adaptation)                  |
| **Future Sync**  | Difficult                           | Easy                                     |
| **Choice**       | ❌ Rejected                         | ✅ Selected                              |

### Why Strategy B Wins

1. **Quantifiable Scope**: ~50 Go files to adapt vs reviewing 3,419 commits
2. **Clean Base**: Latest backend is tested and working
3. **Clear Contract**: OpenAPI spec defines exact API shape
4. **Forward Compatible**: Future dev merges are straightforward
5. **Lower Risk**: Working with known-good components

---

## API Evolution Analysis

### Critical API Changes

Based on `wiki/11_OpenCode_API_Testing.md` (97.4% core API success):

#### 1. TUI Startup APIs (14 APIs - All Required)

| API                            | Status        | Notes                      |
| :----------------------------- | :------------ | :------------------------- |
| `session.list()`               | ✅ Compatible | Load recent sessions       |
| `config.providers()`           | ✅ Compatible | Load providers/models      |
| `provider.list()`              | ✅ Compatible | Connection status          |
| `app.agents()`                 | ✅ Compatible | Agent list                 |
| `config.get()`                 | ✅ Compatible | User config                |
| `command.list()`               | ✅ Compatible | Available commands         |
| `lsp.status()`                 | ✅ Compatible | LSP servers                |
| `mcp.status()`                 | ✅ Compatible | MCP servers                |
| `experimental.resource.list()` | ⚠️ New        | May not exist in v0.15.31  |
| `formatter.status()`           | ⚠️ New        | May not exist in v0.15.31  |
| `session.status()`             | ✅ Compatible | Session statuses           |
| `provider.auth()`              | ✅ Compatible | Auth methods               |
| `vcs.get()`                    | ✅ Compatible | VCS info                   |
| `path.get()`                   | ✅ Compatible | Path info                  |

**Action**: Bubble TUI must handle missing experimental APIs gracefully.

#### 2. Session Management APIs (9 core APIs)

All core session APIs remain compatible:

- `session.create/get/update/delete` - No changes
- `session.messages/todo/diff/children` - Compatible
- `session.prompt()` - Enhanced with new part types (backward compatible)

**Action**: Bubble TUI can ignore new optional part types (`subtask`, `agent`).

#### 3. New PTY APIs (6 endpoints)

```text
pty.list/create/get/update/remove/connect
```

**Decision**: Skip PTY support in Bubble TUI initially (not critical).

#### 4. Permission Model Changes

- v0.15.31: `Permission` interface
- v1.1.13: `PermissionNext` interface in `server.ts`

**Action**: **CRITICAL** - Update Go Bubble TUI permission handling code to match `PermissionNext` schema.

#### 5. SDK Generation Infrastructure

- v0.15.31: `packages/sdk/stainless/stainless.yml` existed
- v1.1.13: `packages/sdk/stainless` directory **DELETED**

**Action**: Must recreate Stainless configuration from scratch or using v0.15.31 backup to enable Go SDK generation. `openapi-ts` is now used for JS SDK only.

#### 6. Event Streaming Enhancements

**New Event Types**:

```text
pty.created/updated/exited/deleted
question.asked/replied/rejected
tui.prompt.append/command.execute/toast.show/session.select
formatter.* events
experimental.* events
```

**Action**: Bubble TUI should ignore unknown event types (already standard practice).

### API Compatibility Matrix

| Category             | Compatibility    | Action                |
| :------------------- | :--------------- | :-------------------- |
| **Session Core**     | ✅ Full          | No changes needed     |
| **Config/Provider**  | ✅ Full          | No changes needed     |
| **File Operations**  | ✅ Full          | No changes needed     |
| **MCP**              | ✅ Full          | No changes needed     |
| **Events (Core)**    | ✅ Full          | Ignore new types      |
| **Permission**       | ⚠️ Schema Change | Update struct         |
| **Experimental APIs**| ⚠️ New           | Add graceful fallback |
| **PTY**              | ➕ New           | Skip (non-essential)  |
| **TUI Control**      | ➕ New           | Not applicable        |

---

## Migration Phases

### Phase 1: Preparation (2 hours)

**Goal**: Set up environment and gather intelligence

#### Phase 1 Tasks

1. **Document Current API** (30 min)

   ```bash
   # On latest dev branch
   cd packages/opencode
   bun run src/index.ts serve --port 3000
   
   # In another terminal
   curl http://localhost:3000/doc > ~/openapi-v1.1.13.json
   ```

2. **Extract v0.15.31 Bubble TUI** (30 min)

   ```bash
   git checkout v0.15.31
   cp -r packages/tui ~/bubble-tui-backup
   git checkout refactor
   ```

3. **API Diff Analysis** (1 hour)
   - Compare OpenAPI schemas
   - Document breaking changes
   - Create adaptation checklist

### Phase 2: Backend Sync (30 min)

**Goal**: Ensure refactor has latest dev backend

#### Phase 2 Tasks

1. **Verify refactor is up-to-date**

   ```bash
   git checkout refactor
   git log --oneline -1
   # Should match origin/dev HEAD
   ```

2. **Confirm server.ts version**

   ```bash
   wc -l packages/opencode/src/server/server.ts
   # Should be ~2894 lines
   ```

### Phase 3: Remove TypeScript TUI (1 hour)

**Goal**: Clean slate for Bubble TUI

#### Phase 3 Tasks

1. **Delete TypeScript TUI code**

   ```bash
   rm -rf packages/opencode/src/cli/cmd/tui/
   rm -rf packages/opencode/src/cli/cmd/opentui/
   ```

2. **Update package.json**
   - Remove Ink dependencies
   - Remove React dependencies (if only used by TUI)

3. **Update index.ts command registration**

   ```typescript
   // packages/opencode/src/index.ts
   // Remove: .command(TuiCommand)
   // Will re-add after Bubble TUI restore
   ```

### Phase 4: Restore Bubble TUI Structure (2 hours)

**Goal**: Bring back Go Bubble TUI skeleton

#### Phase 4 Tasks

1. **Copy Bubble TUI code**

   ```bash
   cp -r ~/bubble-tui-backup packages/tui
   ```

2. **Restore TUI launcher**

   ```bash
   # Copy from v0.15.31
   git show v0.15.31:packages/opencode/src/cli/cmd/tui.ts > \
     packages/opencode/src/cli/cmd/tui.ts
   ```

3. **Update go.mod**

   ```bash
   cd packages/tui
   go get -u  # Update all dependencies
   go mod tidy
   ```

4. **Test basic compilation**

   ```bash
   cd cmd/opencode
   go build -o ../../dist/tui ./main.go
   ```

### Phase 5: SDK Regeneration (3 hours)

**Goal**: Generate Go SDK from latest OpenAPI spec

#### Phase 5 Tasks

1. **Extract OpenAPI spec**

   ```bash
   # Save from running server
   curl http://localhost:3000/doc > /tmp/openapi.json
   ```

2. **Update Stainless config**

   ```yaml
   # packages/sdk/stainless/stainless.yml
   # Verify Go SDK generation settings
   targets:
     go:
       package_name: opencode
       production_repo: sst/opencode-sdk-go
   ```

3. **Generate new Go SDK**

   ```bash
   # This requires Stainless CLI (OpenCode team tool)
   # Alternatively, use the existing SDK generation workflow
   # OR manually update SDK types based on OpenAPI diffs
   ```

4. **Update Bubble TUI imports**

   ```go
   // Update to use new SDK package
   import "github.com/sst/opencode-sdk-go"
   ```

### Phase 6: API Adaptation (8-12 hours)

**Goal**: Make Bubble TUI work with new APIs

#### Focus Areas

##### 6.1 Permission Handling (2 hours)

**File**: `packages/tui/internal/components/dialog/permission.go`

**Changes Needed**:

- Update `Permission` struct to match `PermissionNext`
- Verify permission reply format

##### 6.2 Event Handling (2 hours)

**File**: `packages/tui/internal/api/events.go`

**Changes Needed**:

- Skip unknown event types gracefully
- Add debug logging for new events
- Test SSE connection stability

##### 6.3 Session Prompt (2 hours)

**File**: `packages/tui/internal/api/session.go`

**Changes Needed**:

- Verify `session.prompt()` payload structure
- Test with new optional fields
- Ensure backward compatibility

##### 6.4 Config/Provider APIs (1 hour)

**File**: `packages/tui/internal/api/config.go`

**Changes Needed**:

- Update `Provider` struct for new model capabilities
- Handle new provider auth methods

##### 6.5 New Experimental APIs (1 hour)

**Strategy**: Add graceful fallbacks

```go
// Example pattern
result, err := client.Experimental.Resource.List(ctx, opts)
if err != nil {
    // Log and continue - not critical for TUI startup
    log.Debug("Experimental resource API not available: %v", err)
    return
}
```

##### 6.6 TUI Launch Logic (2 hours)

**File**: `packages/opencode/src/cli/cmd/tui.ts`

**Changes Needed**:

- Verify server URL passing (OPENCODE_SERVER env var)
- Test Go binary compilation and execution
- Handle exit codes properly

### Phase 7: Integration Testing (8-12 hours)

**Goal**: Validate end-to-end functionality

#### Test Scenarios (ordered by priority)

1. **TUI Startup** (2 hours)
   - [ ] TUI launches without errors
   - [ ] All startup APIs succeed
   - [ ] Session list loads
   - [ ] Models/providers display

2. **Session Management** (2 hours)
   - [ ] Create new session
   - [ ] List sessions
   - [ ] Open existing session
   - [ ] Delete session

3. **Message Flow** (3 hours)
   - [ ] Send text message
   - [ ] Receive streaming response
   - [ ] Display message history
   - [ ] Handle errors gracefully

4. **Permissions & Questions** (1 hour)
   - [ ] Permission dialog appears
   - [ ] Allow/reject works
   - [ ] Question prompts work

5. **File Operations** (1 hour)
   - [ ] File search works
   - [ ] File attachment works

6. **Advanced Features** (2 hours)
   - [ ] MCP servers display
   - [ ] Agent selection works
   - [ ] Model switching works
   - [ ] Session fork (if supported)

### Phase 8: Documentation (2 hours)

**Goal**: Document changes and usage

#### Deliverables

1. **Update this wiki page** with final results
2. **Create migration notes** in `CHANGES.md`
3. **Update README** if TUI usage changed
4. **Add troubleshooting guide** for common issues

---

## Implementation Checklist

### Pre-Implementation

- [x] Strategy decision made
- [x] API documentation reviewed
- [x] Wiki document created
- [ ] Team approval obtained

### Phase 1: Preparation

- [ ] Extract v0.15.31 Bubble TUI to backup
- [ ] Document current OpenAPI spec
- [ ] Create API diff analysis
- [ ] Set up development environment

### Phase 2: Backend Sync

- [ ] Verify refactor branch is on latest dev
- [ ] Confirm server.ts version (2894 lines)
- [ ] Test backend server runs

### Phase 3: Remove TypeScript TUI

- [ ] Delete `packages/opencode/src/cli/cmd/tui/`
- [ ] Delete `packages/opencode/src/cli/cmd/opentui/`
- [ ] Remove Ink dependencies from package.json
- [ ] Update index.ts command registration
- [ ] Commit: "chore: remove TypeScript/Ink TUI"

### Phase 4: Restore Bubble TUI

- [ ] Copy packages/tui from v0.15.31
- [ ] Restore tui.ts launcher
- [ ] Update go.mod dependencies
- [ ] Test Go compilation
- [ ] Commit: "feat: restore Go Bubble TUI skeleton"

### Phase 5: SDK Regeneration

- [ ] Generate latest OpenAPI spec
- [ ] Update Stainless config (if needed)
- [ ] Generate new Go SDK
- [ ] Update Bubble TUI imports
- [ ] Commit: "chore: update Go SDK to v1.1.13"

### Phase 6: API Adaptation

- [ ] Update permission handling
- [ ] Update event handling
- [ ] Update session prompt
- [ ] Update config/provider structs
- [ ] Add experimental API fallbacks
- [ ] Fix TUI launch logic
- [ ] Commit: "feat: adapt Bubble TUI to v1.1.13 APIs"

### Phase 7: Testing

- [ ] TUI startup test
- [ ] Session management test
- [ ] Message flow test
- [ ] Permissions test
- [ ] File operations test
- [ ] Advanced features test
- [ ] Fix bugs found during testing
- [ ] Commit: "test: validate Bubble TUI integration"

### Phase 8: Documentation

- [ ] Update this wiki page
- [ ] Update CHANGES.md
- [ ] Update README if needed
- [ ] Create troubleshooting guide
- [ ] Commit: "docs: Bubble TUI revival completion"

### Post-Implementation

- [ ] Create PR for review
- [ ] Run CI/CD tests
- [ ] User acceptance testing
- [ ] Merge to refactor branch

---

## Testing Plan

### Unit Tests

**Scope**: Individual API calls

```bash
# Test each Bubble TUI component
cd packages/tui
go test ./internal/api/...
go test ./internal/components/...
```

### Integration Tests

**Scope**: TUI ↔ Server interaction

| Test Case         | Expected Behavior                          |
| :---------------- | :----------------------------------------- |
| Server connection | TUI connects to `OPENCODE_SERVER`          |
| Session list      | Displays all sessions                      |
| Message send      | HTTP request succeeds, SSE events received |
| Event streaming   | All event types handled or ignored         |
| Permission flow   | Dialog appears, user can reply             |

### Regression Tests

**Scope**: Features that existed in v0.15.31

- [ ] All v0.15.31 features still work
- [ ] No UI regressions
- [ ] Performance not degraded

### Acceptance Criteria

- ✅ TUI launches in <2 seconds
- ✅ All 14 startup APIs succeed
- ✅ Can send message and receive response
- ✅ Zero crashes during 30-minute session
- ✅ Memory usage <100MB idle

---

## Risks and Mitigation

### Risk 1: OpenAPI Spec Unavailable

**Impact**: High  
**Probability**: Low  
**Mitigation**:

- OpenAPI spec is auto-generated by `hono-openapi`
- Can manually extract from live server
- Worst case: manual API inspection via API testing

### Risk 2: Breaking API Changes

**Impact**: High  
**Probability**: Medium  
**Mitigation**:

- API is well-documented (wiki docs)
- 97.4% test compatibility confirmed
- Most changes are additive, not breaking

### Risk 3: Go SDK Generation Issues

**Impact**: High  
**Probability**: Medium  
**Mitigation**:

- Stainless SDK generation is automated
- Can manually port SDK if tooling fails
- v0.15.31 SDK can serve as base

### Risk 4: Missing Features in Bubble TUI

**Impact**: Medium  
**Probability**: High  
**Mitigation**:

- Document missing features
- Add TODO markers for future implementation
- Focus on core functionality first

### Risk 5: Performance Regression

**Impact**: Low  
**Probability**: Low  
**Mitigation**:

- Bubble Tea is inherently performant
- API calls are same HTTP as before
- Monitor memory/CPU in testing

### Risk 6: Time Overrun

**Impact**: Medium  
**Probability**: Medium  
**Mitigation**:

- Phased approach allows partial delivery
- Core functionality prioritized
- Can skip experimental features if needed

---

## References

### Internal Documentation

- [OpenCode SDK API Reference](./10_OpenCode_SDK_API_Reference.md) - Complete API docs
- [OpenCode API Testing](./11_OpenCode_API_Testing.md) - Test results
- [OpenCode TUI API Usage](./12_OpenCode_TUI_API_Usage.md) - TypeScript TUI usage patterns

### Key Commits

- `81c6177` - v0.15.31 release (Go Bubble TUI)
- `96bdeb3c7` - "OpenTUI is here" (TypeScript TUI introduced)
- `f68374ad2` - "DELETE GO BUBBLETEA CRAP HOORAY" (Go TUI removed)
- `efbab087d` - v1.1.13 release (current dev)

### External Resources

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Hono OpenAPI](https://github.com/honojs/hono-openapi) - API documentation
- [Stainless](https://www.stainlessapi.com/) - SDK generator

---

## Change Log

| Date       | Author   | Change                    |
| :--------- | :------- | :------------------------ |
| 2026-01-11 | AI Agent | Initial document creation |

---

**Next Steps**: Proceed with Phase 1 (Preparation) after approval.
