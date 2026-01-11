# OpenCode SDK API Reference

This document provides a comprehensive reference for the OpenCode SDK v2 API, extracted from the official TUI (opentui) implementation and SDK source code.

## Table of Contents

- [Overview](#overview)
- [Client Initialization](#client-initialization)
- [Event Subscription](#event-subscription)
- [API Categories](#api-categories)
  - [Global](#global)
  - [Project](#project)
  - [PTY (Pseudo-Terminal)](#pty-pseudo-terminal)
  - [Config](#config)
  - [Provider](#provider)
  - [Session](#session)
  - [Message & Parts](#message--parts)
  - [Permission](#permission)
  - [Question](#question)
  - [MCP (Model Context Protocol)](#mcp-model-context-protocol)
  - [App](#app)
  - [Find](#find)
  - [File](#file)
  - [Path](#path)
  - [VCS](#vcs)
  - [LSP](#lsp)
  - [Formatter](#formatter)
  - [Command](#command)
  - [Instance](#instance)
  - [Auth](#auth)
  - [TUI](#tui)
  - [Experimental](#experimental)
    - [Tool](#tool)
    - [Worktree](#worktree)
    - [Resource](#resource)

---

## Overview

The OpenCode SDK provides a TypeScript/JavaScript client for interacting with the OpenCode server. It supports:

- Session management (create, list, delete, fork)
- Message sending and streaming
- Provider and model configuration
- MCP server management
- File operations
- Event subscriptions via SSE

## Client Initialization

```typescript
import { createOpencodeClient } from "@opencode-ai/sdk/v2"

const client = createOpencodeClient({
  baseUrl: "http://localhost:3000",  // OpenCode server URL
  directory: "/path/to/project",      // Optional: project directory
})
```

## Event Subscription

Subscribe to real-time events from the server:

```typescript
// Subscribe to SSE events
const events = await client.event.subscribe()

for await (const event of events.stream) {
  switch (event.type) {
    case "message.updated":
      // Handle message update
      break
    case "message.part.updated":
      // Handle streaming content
      break
    case "session.updated":
      // Handle session changes
      break
    // ... other event types
  }
}
```

### Event Types

| Event Type | Description |
| ---------- | ----------- |
| `installation.updated` | OpenCode version updated |
| `installation.update-available` | New version available |
| `project.updated` | Project info changed |
| `server.instance.disposed` | Server instance disposed |
| `server.connected` | Server connected |
| `global.disposed` | Global disposed |
| `lsp.updated` | LSP status changed |
| `lsp.client.diagnostics` | LSP diagnostics received |
| `message.updated` | Message info updated (role, status, error) |
| `message.removed` | Message was deleted |
| `message.part.updated` | Message part updated (streaming text, tool calls) |
| `message.part.removed` | Message part removed |
| `session.created` | New session created |
| `session.updated` | Session info changed |
| `session.deleted` | Session was deleted |
| `session.status` | Session status changed (idle, busy, retry) |
| `session.idle` | Session became idle |
| `session.diff` | File diff updated |
| `session.compacted` | Session was compacted |
| `session.error` | Session error occurred |
| `permission.asked` | Permission request from AI |
| `permission.replied` | Permission response sent |
| `question.asked` | Question from AI |
| `question.replied` | Question answered |
| `question.rejected` | Question rejected |
| `todo.updated` | Todo list changed |
| `file.edited` | File was edited |
| `file.watcher.updated` | File watcher event (add/change/unlink) |
| `vcs.branch.updated` | Git branch changed |
| `mcp.tools.changed` | MCP tools changed |
| `command.executed` | Command was executed |
| `pty.created` | PTY session created |
| `pty.updated` | PTY session updated |
| `pty.exited` | PTY session exited |
| `pty.deleted` | PTY session deleted |
| `tui.prompt.append` | TUI prompt append event |
| `tui.command.execute` | TUI command execute event |
| `tui.toast.show` | TUI toast show event |
| `tui.session.select` | TUI session select event |

### Event Type Definitions

```typescript
type Event =
  | EventInstallationUpdated
  | EventInstallationUpdateAvailable
  | EventProjectUpdated
  | EventServerInstanceDisposed
  | EventServerConnected
  | EventGlobalDisposed
  | EventLspClientDiagnostics
  | EventLspUpdated
  | EventMessageUpdated
  | EventMessageRemoved
  | EventMessagePartUpdated
  | EventMessagePartRemoved
  | EventPermissionAsked
  | EventPermissionReplied
  | EventSessionStatus
  | EventSessionIdle
  | EventSessionCreated
  | EventSessionUpdated
  | EventSessionDeleted
  | EventSessionDiff
  | EventSessionError
  | EventSessionCompacted
  | EventQuestionAsked
  | EventQuestionReplied
  | EventQuestionRejected
  | EventFileEdited
  | EventFileWatcherUpdated
  | EventTodoUpdated
  | EventVcsBranchUpdated
  | EventMcpToolsChanged
  | EventCommandExecuted
  | EventPtyCreated
  | EventPtyUpdated
  | EventPtyExited
  | EventPtyDeleted
  | EventTuiPromptAppend
  | EventTuiCommandExecute
  | EventTuiToastShow
  | EventTuiSessionSelect

// Global event wrapper
interface GlobalEvent {
  directory: string
  payload: Event
}
```

---

## API Categories

### Global

Server health and lifecycle management.

#### `global.health()`

Get server health information.

```typescript
const response = await client.global.health()
// Returns: { data: { healthy: true, version: string } }
```

#### `global.event()`

Subscribe to global events (SSE endpoint). Prefer using `event.subscribe()` instead.

```typescript
const response = await client.global.event()
// Returns SSE stream of GlobalEvent
```

#### `global.dispose()`

Dispose all server instances.

```typescript
await client.global.dispose()
// Returns: { data: boolean }
```

---

### Project

Project management APIs.

#### `project.list({ directory? })`

Get a list of projects that have been opened with OpenCode.

```typescript
const response = await client.project.list()
// Returns: { data: Project[] }
```

**Response Type:**

```typescript
interface Project {
  id: string
  worktree: string
  vcs?: "git"
  name?: string
  icon?: { url?: string; color?: string }
  time: {
    created: number
    updated: number
    initialized?: number
  }
  sandboxes: string[]
}
```

#### `project.current({ directory? })`

Retrieve the currently active project.

```typescript
const response = await client.project.current()
// Returns: { data: Project }
```

#### `project.update({ projectID, directory?, name?, icon? })`

Update project properties such as name, icon and color.

```typescript
await client.project.update({
  projectID: "project-id",
  name: "My Project",
  icon: { color: "#FF5733" }
})
// Returns: { data: Project }
```

---

### PTY (Pseudo-Terminal)

Manage pseudo-terminal sessions for running shell commands.

#### `pty.list({ directory? })`

Get a list of all active PTY sessions.

```typescript
const response = await client.pty.list()
// Returns: { data: Pty[] }
```

**Response Type:**

```typescript
interface Pty {
  id: string
  title: string
  command: string
  args: string[]
  cwd: string
  status: "running" | "exited"
  pid: number
}
```

#### `pty.create({ directory?, command?, args?, cwd?, title?, env? })`

Create a new PTY session.

```typescript
const response = await client.pty.create({
  command: "bash",
  args: ["-l"],
  cwd: "/path/to/dir",
  title: "My Terminal",
  env: { MY_VAR: "value" }
})
// Returns: { data: Pty }
```

#### `pty.get({ ptyID, directory? })`

Get information about a specific PTY session.

```typescript
const response = await client.pty.get({ ptyID: "pty-id" })
// Returns: { data: Pty }
```

#### `pty.update({ ptyID, directory?, title?, size? })`

Update PTY session properties.

```typescript
await client.pty.update({
  ptyID: "pty-id",
  title: "New Title",
  size: { rows: 24, cols: 80 }
})
// Returns: { data: Pty }
```

#### `pty.remove({ ptyID, directory? })`

Remove and terminate a PTY session.

```typescript
await client.pty.remove({ ptyID: "pty-id" })
// Returns: { data: boolean }
```

#### `pty.connect({ ptyID, directory? })`

Establish a WebSocket connection to interact with a PTY session in real-time.

```typescript
await client.pty.connect({ ptyID: "pty-id" })
// Returns WebSocket connection
```

---

### Config

Configuration management.

#### `config.get({ directory? })`

Get current configuration.

```typescript
const response = await client.config.get()
// Returns: { data: Config }
```

**Response Type:**

```typescript
interface Config {
  $schema?: string
  theme?: string
  keybinds?: KeybindsConfig
  logLevel?: "DEBUG" | "INFO" | "WARN" | "ERROR"
  model?: string              // Default model (e.g., "opencode/glm-4.7-free")
  small_model?: string        // Small model for tasks like title generation
  default_agent?: string      // Default agent name
  username?: string           // Custom username
  share?: "manual" | "auto" | "disabled"
  autoupdate?: boolean | "notify"
  tui?: {
    scroll_speed?: number
    scroll_acceleration?: { enabled: boolean }
    diff_style?: "auto" | "stacked"
  }
  server?: ServerConfig
  agent?: Record<string, AgentConfig>
  provider?: Record<string, ProviderConfig>
  mcp?: Record<string, McpLocalConfig | McpRemoteConfig>
  permission?: PermissionConfig
  experimental?: ExperimentalConfig
}
```

#### `config.update({ directory?, config })`

Update configuration.

```typescript
await client.config.update({
  config: {
    model: "anthropic/claude-3-5-sonnet",
    default_agent: "build"
  }
})
// Returns: { data: Config }
```

#### `config.providers({ directory? })`

Get all configured providers with their models. **This is the primary API for getting available models.**

```typescript
const response = await client.config.providers()
// Returns: { data: { providers: Provider[], default: Record<string, string> } }
```

**Response Type:**

```typescript
interface ConfigProvidersResponse {
  providers: Provider[]
  default: Record<string, string>  // providerID -> default modelID
}

interface Provider {
  id: string           // e.g., "opencode", "anthropic"
  name: string         // Display name
  source: "env" | "config" | "custom" | "api"
  env: string[]
  key?: string
  options: Record<string, unknown>
  models: Record<string, Model>
}

interface Model {
  id: string
  providerID: string
  name: string
  family?: string
  status: "alpha" | "beta" | "deprecated" | "active"
  release_date: string
  capabilities: {
    temperature: boolean
    reasoning: boolean
    attachment: boolean
    toolcall: boolean
    input: { text: boolean; audio: boolean; image: boolean; video: boolean; pdf: boolean }
    output: { text: boolean; audio: boolean; image: boolean; video: boolean; pdf: boolean }
    interleaved: boolean | { field: "reasoning_content" | "reasoning_details" }
  }
  cost: {
    input: number
    output: number
    cache: { read: number; write: number }
  }
  limit: { context: number; output: number }
  options: Record<string, unknown>
  headers: Record<string, string>
  variants?: Record<string, Record<string, unknown>>
}
```

---

### Provider

Provider listing and authentication.

#### `provider.list({ directory? })`

List all available providers (connected and available).

```typescript
const response = await client.provider.list()
// Returns: { data: { all: ProviderInfo[], connected: string[], default: Record<string, string> } }
```

**Response Type:**

```typescript
interface ProviderListResponse {
  all: ProviderInfo[]           // All available providers with models
  connected: string[]           // Array of connected provider IDs
  default: Record<string, string>  // providerID -> default modelID
}
```

#### `provider.auth({ directory? })`

Get authentication methods for all providers.

```typescript
const response = await client.provider.auth()
// Returns: { data: Record<string, ProviderAuthMethod[]> }
```

**Response Type:**

```typescript
interface ProviderAuthMethod {
  type: "oauth" | "api"
  label: string
}
```

#### `provider.oauth.authorize({ providerID, directory?, method })`

Start OAuth flow for a provider.

```typescript
const response = await client.provider.oauth.authorize({
  providerID: "anthropic",
  method: 0  // Index of auth method
})
// Returns: { data: ProviderAuthAuthorization }
```

**Response Type:**

```typescript
interface ProviderAuthAuthorization {
  url: string
  method: "auto" | "code"
  instructions: string
}
```

#### `provider.oauth.callback({ providerID, directory?, method, code? })`

Complete OAuth flow.

```typescript
await client.provider.oauth.callback({
  providerID: "anthropic",
  method: 0,
  code: "authorization-code"
})
// Returns: { data: boolean }
```
```

---

### Session

Session management and messaging.

#### `session.list({ directory?, start?, search?, limit? })`

List all sessions.

```typescript
const response = await client.session.list({
  start: Date.now() - 30 * 24 * 60 * 60 * 1000,  // Last 30 days
  search: "query",  // Optional search
  limit: 30
})
// Returns: { data: Session[] }
```

**Response Type:**

```typescript
interface Session {
  id: string
  projectID: string
  directory: string
  parentID?: string      // For forked sessions
  title: string
  version: string
  summary?: {
    additions: number
    deletions: number
    files: number
    diffs?: FileDiff[]
  }
  share?: { url: string }
  permission?: PermissionRuleset
  revert?: {
    messageID: string
    partID?: string
    snapshot?: string
    diff?: string
  }
  time: {
    created: number
    updated: number
    archived?: number
    compacting?: number
  }
}
```

#### `session.create({ directory?, parentID?, title?, permission? })`

Create a new session.

```typescript
const response = await client.session.create({
  title: "New Session",
  parentID: "parent-session-id",  // Optional: for child sessions
  permission: [{ permission: "bash", pattern: "*", action: "allow" }]
})
// Returns: { data: Session }
```

#### `session.get({ sessionID, directory? })`

Get session details.

```typescript
const response = await client.session.get({ sessionID: "session-id" })
// Returns: { data: Session }
```

#### `session.update({ sessionID, directory?, title?, time? })`

Update session properties.

```typescript
await client.session.update({
  sessionID: "session-id",
  title: "New Title",
  time: { archived: Date.now() }  // Archive the session
})
// Returns: { data: Session }
```

#### `session.delete({ sessionID, directory? })`

Delete a session.

```typescript
await client.session.delete({ sessionID: "session-id" })
// Returns: { data: boolean }
```

#### `session.status({ directory? })`

Get status of all sessions.

```typescript
const response = await client.session.status()
// Returns: { data: Record<string, SessionStatus> }
```

**Response Type:**

```typescript
type SessionStatus = 
  | { type: "idle" }
  | { type: "busy" }
  | { type: "retry"; attempt: number; message: string; next: number }
```

#### `session.messages({ sessionID, directory?, limit? })`

Get all messages in a session.

```typescript
const response = await client.session.messages({
  sessionID: "session-id",
  limit: 100
})
// Returns: { data: Array<{ info: Message, parts: Part[] }> }
```

**Response Type:**

```typescript
type Message = UserMessage | AssistantMessage

interface UserMessage {
  id: string
  sessionID: string
  role: "user"
  agent: string
  model: { providerID: string; modelID: string }
  system?: string
  tools?: Record<string, boolean>
  variant?: string
  time: { created: number }
  summary?: { title?: string; body?: string; diffs: FileDiff[] }
}

interface AssistantMessage {
  id: string
  sessionID: string
  role: "assistant"
  parentID: string
  modelID: string
  providerID: string
  mode: string
  agent: string
  path: { cwd: string; root: string }
  summary?: boolean
  cost: number
  tokens: {
    input: number
    output: number
    reasoning: number
    cache: { read: number; write: number }
  }
  finish?: string
  error?: MessageError
  time: { created: number; completed?: number }
}

type MessageError = 
  | ProviderAuthError 
  | UnknownError 
  | MessageOutputLengthError 
  | MessageAbortedError 
  | ApiError
```

#### `session.prompt({ sessionID, directory?, model?, agent?, parts, ... })`

Send a message to a session. **This is the primary API for chat.**

```typescript
await client.session.prompt({
  sessionID: "session-id",
  model: {
    providerID: "opencode",
    modelID: "glm-4.7-free"
  },
  agent: "build",
  variant: "default",
  messageID: "custom-message-id",  // Optional
  noReply: false,                   // Don't generate AI response
  system: "Custom system prompt",   // Optional
  tools: { bash: true, edit: false }, // Enable/disable tools
  parts: [
    { type: "text", text: "Hello, world!" },
    { type: "file", mime: "image/png", filename: "image.png", url: "data:..." },
    { type: "agent", name: "explore" },
    { type: "subtask", prompt: "...", description: "...", agent: "build" }
  ]
})
// Returns: { data: { info: AssistantMessage, parts: Part[] } }
```

**Part Input Types:**

```typescript
interface TextPartInput {
  id?: string
  type: "text"
  text: string
  synthetic?: boolean
  ignored?: boolean
  time?: { start: number; end?: number }
  metadata?: Record<string, unknown>
}

interface FilePartInput {
  id?: string
  type: "file"
  mime: string
  filename?: string
  url: string
  source?: FilePartSource
}

interface AgentPartInput {
  id?: string
  type: "agent"
  name: string
  source?: { value: string; start: number; end: number }
}

interface SubtaskPartInput {
  id?: string
  type: "subtask"
  prompt: string
  description: string
  agent: string
  command?: string
}
```

#### `session.promptAsync({ sessionID, directory?, ... })`

Send a message asynchronously (returns immediately without waiting for response).

```typescript
await client.session.promptAsync({
  sessionID: "session-id",
  model: { providerID: "opencode", modelID: "glm-4.7-free" },
  parts: [{ type: "text", text: "Hello!" }]
})
// Returns: { data: void } (204 No Content)
```

#### `session.init({ sessionID, directory?, modelID?, providerID?, messageID? })`

Analyze the current application and create an AGENTS.md file.

```typescript
await client.session.init({
  sessionID: "session-id",
  providerID: "opencode",
  modelID: "glm-4.7-free"
})
// Returns: { data: boolean }
```

#### `session.command({ sessionID, directory?, command, arguments, ... })`

Execute a slash command.

```typescript
await client.session.command({
  sessionID: "session-id",
  command: "init",
  arguments: "",
  agent: "build",
  model: "opencode/glm-4.7-free"
})
// Returns: { data: { info: AssistantMessage, parts: Part[] } }
```

#### `session.shell({ sessionID, directory?, command, model?, agent? })`

Execute a shell command (! prefix in TUI).

```typescript
await client.session.shell({
  sessionID: "session-id",
  command: "ls -la",
  model: { providerID: "opencode", modelID: "glm-4.7-free" },
  agent: "build"
})
// Returns: { data: AssistantMessage }
```

#### `session.abort({ sessionID, directory? })`

Abort an active session.

```typescript
await client.session.abort({ sessionID: "session-id" })
// Returns: { data: boolean }
```

#### `session.fork({ sessionID, directory?, messageID? })`

Fork a session at a specific message.

```typescript
const response = await client.session.fork({
  sessionID: "session-id",
  messageID: "message-id"  // Optional: fork point
})
// Returns: { data: Session }
```

#### `session.revert({ sessionID, directory?, messageID?, partID? })`

Revert to a specific message (undo).

```typescript
await client.session.revert({
  sessionID: "session-id",
  messageID: "message-id",
  partID: "part-id"  // Optional: revert specific part
})
// Returns: { data: Session }
```

#### `session.unrevert({ sessionID, directory? })`

Restore reverted messages (redo).

```typescript
await client.session.unrevert({ sessionID: "session-id" })
// Returns: { data: Session }
```

#### `session.share({ sessionID, directory? })`

Create a shareable link.

```typescript
const response = await client.session.share({ sessionID: "session-id" })
// Returns: { data: Session } (with share.url populated)
```

#### `session.unshare({ sessionID, directory? })`

Remove shareable link.

```typescript
await client.session.unshare({ sessionID: "session-id" })
// Returns: { data: Session }
```

#### `session.summarize({ sessionID, directory?, providerID?, modelID?, auto? })`

Compact/summarize a session.

```typescript
await client.session.summarize({
  sessionID: "session-id",
  providerID: "opencode",
  modelID: "glm-4.7-free",
  auto: false
})
// Returns: { data: boolean }
```

#### `session.todo({ sessionID, directory? })`

Get session todos.

```typescript
const response = await client.session.todo({ sessionID: "session-id" })
// Returns: { data: Todo[] }
```

**Response Type:**

```typescript
interface Todo {
  id: string
  content: string
  status: string  // "pending" | "in_progress" | "completed" | "cancelled"
  priority: string  // "high" | "medium" | "low"
}
```

#### `session.diff({ sessionID, directory?, messageID? })`

Get file diffs for a session.

```typescript
const response = await client.session.diff({ sessionID: "session-id" })
// Returns: { data: FileDiff[] }
```

**Response Type:**

```typescript
interface FileDiff {
  file: string
  before: string
  after: string
  additions: number
  deletions: number
}
```

#### `session.children({ sessionID, directory? })`

Get child sessions (forks).

```typescript
const response = await client.session.children({ sessionID: "session-id" })
// Returns: { data: Session[] }
```

#### `session.message({ sessionID, messageID, directory? })`

Get a specific message.

```typescript
const response = await client.session.message({
  sessionID: "session-id",
  messageID: "message-id"
})
// Returns: { data: { info: Message, parts: Part[] } }
```

---

### Message & Parts

**Part Types:**

```typescript
type Part =
  | TextPart
  | ReasoningPart
  | FilePart
  | ToolPart
  | StepStartPart
  | StepFinishPart
  | SnapshotPart
  | PatchPart
  | AgentPart
  | RetryPart
  | CompactionPart
  | SubtaskPart

interface TextPart {
  id: string
  sessionID: string
  messageID: string
  type: "text"
  text: string
  synthetic?: boolean
  ignored?: boolean
  time?: { start: number; end?: number }
  metadata?: Record<string, unknown>
}

interface ReasoningPart {
  id: string
  sessionID: string
  messageID: string
  type: "reasoning"
  text: string
  metadata?: Record<string, unknown>
  time: { start: number; end?: number }
}

interface FilePart {
  id: string
  sessionID: string
  messageID: string
  type: "file"
  mime: string
  filename?: string
  url: string
  source?: FilePartSource
}

interface ToolPart {
  id: string
  sessionID: string
  messageID: string
  type: "tool"
  callID: string
  tool: string
  state: ToolState
  metadata?: Record<string, unknown>
}

type ToolState =
  | { status: "pending"; input: Record<string, unknown>; raw: string }
  | { status: "running"; input: Record<string, unknown>; title?: string; metadata?: Record<string, unknown>; time: { start: number } }
  | { status: "completed"; input: Record<string, unknown>; output: string; title: string; metadata: Record<string, unknown>; time: { start: number; end: number; compacted?: number }; attachments?: FilePart[] }
  | { status: "error"; input: Record<string, unknown>; error: string; metadata?: Record<string, unknown>; time: { start: number; end: number } }

interface StepStartPart {
  id: string
  sessionID: string
  messageID: string
  type: "step-start"
  snapshot?: string
}

interface StepFinishPart {
  id: string
  sessionID: string
  messageID: string
  type: "step-finish"
  reason: string
  snapshot?: string
  cost: number
  tokens: { input: number; output: number; reasoning: number; cache: { read: number; write: number } }
}

interface CompactionPart {
  id: string
  sessionID: string
  messageID: string
  type: "compaction"
  auto: boolean
}
```

#### `part.update({ sessionID, messageID, partID, directory?, part })`

Update a message part.

```typescript
await client.part.update({
  sessionID: "session-id",
  messageID: "message-id",
  partID: "part-id",
  part: { type: "text", text: "Updated text" }
})
// Returns: { data: Part }
```

#### `part.delete({ sessionID, messageID, partID, directory? })`

Delete a message part.

```typescript
await client.part.delete({
  sessionID: "session-id",
  messageID: "message-id",
  partID: "part-id"
})
// Returns: { data: boolean }
```

---

### Permission

Handle permission requests from AI.

#### `permission.list({ directory? })`

List pending permission requests.

```typescript
const response = await client.permission.list()
// Returns: { data: PermissionRequest[] }
```

**Response Type:**

```typescript
interface PermissionRequest {
  id: string
  sessionID: string
  permission: string
  patterns: string[]
  metadata: Record<string, unknown>
  always: string[]
  tool?: {
    messageID: string
    callID: string
  }
}
```

#### `permission.reply({ requestID, directory?, reply, message? })`

Respond to a permission request.

```typescript
await client.permission.reply({
  requestID: "request-id",
  reply: "once",  // "once" | "always" | "reject"
  message: "Optional rejection message"
})
// Returns: { data: boolean }
```

---

### Question

Handle questions from AI.

#### `question.list({ directory? })`

List pending questions.

```typescript
const response = await client.question.list()
// Returns: { data: QuestionRequest[] }
```

**Response Type:**

```typescript
interface QuestionRequest {
  id: string
  sessionID: string
  questions: QuestionInfo[]
  tool?: {
    messageID: string
    callID: string
  }
}

interface QuestionInfo {
  question: string      // Complete question
  header: string        // Very short label (max 12 chars)
  options: QuestionOption[]
  multiple?: boolean    // Allow selecting multiple choices
}

interface QuestionOption {
  label: string         // Display text (1-5 words)
  description: string   // Explanation of choice
}
```

#### `question.reply({ requestID, directory?, answers })`

Answer questions.

```typescript
await client.question.reply({
  requestID: "request-id",
  answers: [["option1", "option2"], ["answer2"]]  // Array of selected labels per question
})
// Returns: { data: boolean }
```

#### `question.reject({ requestID, directory? })`

Reject a question.

```typescript
await client.question.reject({ requestID: "request-id" })
// Returns: { data: boolean }
```

---

### MCP (Model Context Protocol)

Manage MCP servers.

#### `mcp.status({ directory? })`

Get status of all MCP servers.

```typescript
const response = await client.mcp.status()
// Returns: { data: Record<string, McpStatus> }
```

**Response Type:**

```typescript
type McpStatus =
  | { status: "connected" }
  | { status: "disabled" }
  | { status: "failed"; error: string }
  | { status: "needs_auth" }
  | { status: "needs_client_registration"; error: string }
```

#### `mcp.connect({ name, directory? })`

Connect an MCP server.

```typescript
await client.mcp.connect({ name: "mcp-server-name" })
// Returns: { data: boolean }
```

#### `mcp.disconnect({ name, directory? })`

Disconnect an MCP server.

```typescript
await client.mcp.disconnect({ name: "mcp-server-name" })
// Returns: { data: boolean }
```

#### `mcp.add({ directory?, name, config })`

Add a new MCP server.

```typescript
// Local MCP server
await client.mcp.add({
  name: "my-local-mcp",
  config: {
    type: "local",
    command: ["uvx", "my-mcp-server"],
    environment: { MY_VAR: "value" },
    enabled: true,
    timeout: 5000
  }
})

// Remote MCP server
await client.mcp.add({
  name: "my-remote-mcp",
  config: {
    type: "remote",
    url: "https://mcp.example.com",
    headers: { "Authorization": "Bearer token" },
    oauth: { clientId: "...", scope: "read write" },
    enabled: true
  }
})
// Returns: { data: Record<string, McpStatus> }
```

**Config Types:**

```typescript
interface McpLocalConfig {
  type: "local"
  command: string[]
  environment?: Record<string, string>
  enabled?: boolean
  timeout?: number  // ms, default 5000
}

interface McpRemoteConfig {
  type: "remote"
  url: string
  enabled?: boolean
  headers?: Record<string, string>
  oauth?: McpOAuthConfig | false
  timeout?: number
}

interface McpOAuthConfig {
  clientId?: string
  clientSecret?: string
  scope?: string
}
```

#### `mcp.auth.start({ name, directory? })`

Start OAuth authentication flow for an MCP server.

```typescript
const response = await client.mcp.auth.start({ name: "mcp-server" })
// Returns: { data: { authorizationUrl: string } }
```

#### `mcp.auth.callback({ name, directory?, code })`

Complete OAuth authentication with authorization code.

```typescript
await client.mcp.auth.callback({
  name: "mcp-server",
  code: "authorization-code"
})
// Returns: { data: McpStatus }
```

#### `mcp.auth.authenticate({ name, directory? })`

Start OAuth flow and wait for callback (opens browser).

```typescript
await client.mcp.auth.authenticate({ name: "mcp-server" })
// Returns: { data: McpStatus }
```

#### `mcp.auth.remove({ name, directory? })`

Remove OAuth credentials for an MCP server.

```typescript
await client.mcp.auth.remove({ name: "mcp-server" })
// Returns: { data: { success: true } }
```

---

### App

Application-level APIs.

#### `app.agents({ directory? })`

List all available agents.

```typescript
const response = await client.app.agents()
// Returns: { data: Agent[] }
```

**Response Type:**

```typescript
interface Agent {
  name: string
  description?: string
  mode: "subagent" | "primary" | "all"
  native?: boolean
  hidden?: boolean
  topP?: number
  temperature?: number
  color?: string
  permission: PermissionRuleset
  model?: { modelID: string; providerID: string }
  prompt?: string
  options: Record<string, unknown>
  steps?: number
}
```

#### `app.log({ directory?, service, level, message, extra? })`

Write to server logs.

```typescript
await client.app.log({
  service: "tui",
  level: "info",  // "debug" | "info" | "warn" | "error"
  message: "Log message",
  extra: { key: "value" }
})
// Returns: { data: boolean }
```

---

### Find

Search functionality.

#### `find.files({ directory?, query, type?, dirs?, limit? })`

Search for files.

```typescript
const response = await client.find.files({
  query: "package.json",
  type: "file",  // "file" | "directory"
  dirs: "false", // "true" | "false" - include directories
  limit: 20
})
// Returns: { data: string[] }
```

#### `find.text({ directory?, pattern })`

Search for text in files (ripgrep).

```typescript
const response = await client.find.text({ pattern: "TODO" })
// Returns: { data: SearchResult[] }
```

**Response Type:**

```typescript
interface SearchResult {
  path: { text: string }
  lines: { text: string }
  line_number: number
  absolute_offset: number
  submatches: Array<{
    match: { text: string }
    start: number
    end: number
  }>
}
```

#### `find.symbols({ directory?, query })`

Search for code symbols (LSP).

```typescript
const response = await client.find.symbols({ query: "MyClass" })
// Returns: { data: Symbol[] }
```

**Response Type:**

```typescript
interface Symbol {
  name: string
  kind: number
  location: {
    uri: string
    range: Range
  }
}

interface Range {
  start: { line: number; character: number }
  end: { line: number; character: number }
}
```

---

### File

File operations.

#### `file.list({ directory?, path })`

List files in a directory.

```typescript
const response = await client.file.list({ path: "src" })
// Returns: { data: FileNode[] }
```

**Response Type:**

```typescript
interface FileNode {
  name: string
  path: string
  absolute: string
  type: "file" | "directory"
  ignored: boolean
}
```

#### `file.read({ directory?, path })`

Read file content.

```typescript
const response = await client.file.read({ path: "package.json" })
// Returns: { data: FileContent }
```

**Response Type:**

```typescript
interface FileContent {
  type: "text"
  content: string
  diff?: string
  patch?: {
    oldFileName: string
    newFileName: string
    oldHeader?: string
    newHeader?: string
    hunks: Array<{
      oldStart: number
      oldLines: number
      newStart: number
      newLines: number
      lines: string[]
    }>
    index?: string
  }
  encoding?: "base64"
  mimeType?: string
}
```

#### `file.status({ directory? })`

Get git status of all files.

```typescript
const response = await client.file.status()
// Returns: { data: File[] }
```

**Response Type:**

```typescript
interface File {
  path: string
  added: number
  removed: number
  status: "added" | "deleted" | "modified"
}
```

---

### Path

Path information.

#### `path.get({ directory? })`

Get path information.

```typescript
const response = await client.path.get()
// Returns: { data: Path }
```

**Response Type:**

```typescript
interface Path {
  home: string
  state: string
  config: string
  worktree: string
  directory: string
}
```

---

### VCS

Version control information.

#### `vcs.get({ directory? })`

Get VCS (git) information.

```typescript
const response = await client.vcs.get()
// Returns: { data: VcsInfo }
```

**Response Type:**

```typescript
interface VcsInfo {
  branch: string
}
```

---

### LSP

Language Server Protocol status.

#### `lsp.status({ directory? })`

Get LSP server status.

```typescript
const response = await client.lsp.status()
// Returns: { data: LspStatus[] }
```

**Response Type:**

```typescript
interface LspStatus {
  id: string
  name: string
  root: string
  status: "connected" | "error"
}
```

---

### Formatter

Code formatter status.

#### `formatter.status({ directory? })`

Get formatter status.

```typescript
const response = await client.formatter.status()
// Returns: { data: FormatterStatus[] }
```

**Response Type:**

```typescript
interface FormatterStatus {
  name: string
  extensions: string[]
  enabled: boolean
}
```

---

### Command

Available commands.

#### `command.list({ directory? })`

List all available commands.

```typescript
const response = await client.command.list()
// Returns: { data: Command[] }
```

**Response Type:**

```typescript
interface Command {
  name: string
  description?: string
  agent?: string
  model?: string
  mcp?: boolean
  template: string
  subtask?: boolean
  hints: string[]
}
```

---

### Instance

Instance lifecycle.

#### `instance.dispose({ directory? })`

Dispose current instance (triggers reconnection).

```typescript
await client.instance.dispose()
// Returns: { data: boolean }
```

---

### Auth

Authentication management.

#### `auth.set({ providerID, directory?, auth })`

Set authentication credentials.

```typescript
// API key authentication
await client.auth.set({
  providerID: "opencode",
  auth: {
    type: "api",
    key: "your-api-key"
  }
})

// OAuth authentication
await client.auth.set({
  providerID: "anthropic",
  auth: {
    type: "oauth",
    refresh: "refresh-token",
    access: "access-token",
    expires: Date.now() + 3600000
  }
})

// Well-known authentication
await client.auth.set({
  providerID: "copilot",
  auth: {
    type: "wellknown",
    key: "key",
    token: "token"
  }
})
// Returns: { data: boolean }
```

**Auth Types:**

```typescript
type Auth = OAuth | ApiAuth | WellKnownAuth

interface OAuth {
  type: "oauth"
  refresh: string
  access: string
  expires: number
  enterpriseUrl?: string
}

interface ApiAuth {
  type: "api"
  key: string
}

interface WellKnownAuth {
  type: "wellknown"
  key: string
  token: string
}
```

---

### TUI

TUI-specific APIs (for controlling the official TUI).

#### `tui.appendPrompt({ directory?, text })`

Append text to TUI prompt.

```typescript
await client.tui.appendPrompt({ text: "Hello" })
// Returns: { data: boolean }
```

#### `tui.submitPrompt({ directory? })`

Submit the current prompt.

```typescript
await client.tui.submitPrompt()
// Returns: { data: boolean }
```

#### `tui.clearPrompt({ directory? })`

Clear the prompt.

```typescript
await client.tui.clearPrompt()
// Returns: { data: boolean }
```

#### `tui.openHelp({ directory? })`

Open help dialog.

```typescript
await client.tui.openHelp()
// Returns: { data: boolean }
```

#### `tui.openSessions({ directory? })`

Open sessions dialog.

```typescript
await client.tui.openSessions()
// Returns: { data: boolean }
```

#### `tui.openModels({ directory? })`

Open models dialog.

```typescript
await client.tui.openModels()
// Returns: { data: boolean }
```

#### `tui.openThemes({ directory? })`

Open themes dialog.

```typescript
await client.tui.openThemes()
// Returns: { data: boolean }
```

#### `tui.executeCommand({ directory?, command })`

Execute a TUI command.

```typescript
await client.tui.executeCommand({ command: "agent_cycle" })
// Returns: { data: boolean }
```

**Available Commands:**

- `session.list`, `session.new`, `session.share`, `session.interrupt`, `session.compact`
- `session.page.up`, `session.page.down`, `session.half.page.up`, `session.half.page.down`
- `session.first`, `session.last`
- `prompt.clear`, `prompt.submit`
- `agent.cycle`

#### `tui.showToast({ directory?, title?, message, variant, duration? })`

Show a toast notification.

```typescript
await client.tui.showToast({
  title: "Success",
  message: "Operation completed",
  variant: "success",  // "info" | "success" | "warning" | "error"
  duration: 3000       // ms
})
// Returns: { data: boolean }
```

#### `tui.selectSession({ directory?, sessionID })`

Navigate to a session.

```typescript
await client.tui.selectSession({ sessionID: "session-id" })
// Returns: { data: boolean }
```

#### `tui.publish({ directory?, body })`

Publish a TUI event.

```typescript
await client.tui.publish({
  body: {
    type: "tui.toast.show",
    properties: {
      message: "Hello",
      variant: "info"
    }
  }
})
// Returns: { data: boolean }
```

#### `tui.control.next({ directory? })`

Get next TUI request from queue.

```typescript
const response = await client.tui.control.next()
// Returns: { data: { path: string, body: unknown } }
```

#### `tui.control.response({ directory?, body })`

Submit response to TUI request queue.

```typescript
await client.tui.control.response({ body: { result: "ok" } })
// Returns: { data: boolean }
```

---

### Experimental

Experimental APIs (may change).

#### Tool

Tool discovery APIs.

##### `tool.ids({ directory? })`

List all available tool IDs.

```typescript
const response = await client.tool.ids()
// Returns: { data: string[] }
```

##### `tool.list({ directory?, provider, model })`

Get tools with JSON schema parameters for a specific provider/model.

```typescript
const response = await client.tool.list({
  provider: "opencode",
  model: "glm-4.7-free"
})
// Returns: { data: ToolListItem[] }
```

**Response Type:**

```typescript
interface ToolListItem {
  id: string
  description: string
  parameters: unknown  // JSON Schema
}
```

#### Worktree

Git worktree management (sandbox environments).

##### `worktree.list({ directory? })`

List all sandbox worktrees.

```typescript
const response = await client.worktree.list()
// Returns: { data: string[] }
```

##### `worktree.create({ directory?, worktreeCreateInput? })`

Create a new git worktree.

```typescript
const response = await client.worktree.create({
  worktreeCreateInput: {
    name: "feature-branch",
    startCommand: "npm install"
  }
})
// Returns: { data: Worktree }
```

**Response Type:**

```typescript
interface Worktree {
  name: string
  branch: string
  directory: string
}
```

#### Resource

MCP resource discovery.

##### `experimental.resource.list({ directory? })`

List MCP resources.

```typescript
const response = await client.experimental.resource.list()
// Returns: { data: Record<string, McpResource> }
```

**Response Type:**

```typescript
interface McpResource {
  name: string
  uri: string
  description?: string
  mimeType?: string
  client: string
}
```

---

## Usage Examples

### Basic Chat Flow

```typescript
import { createOpencodeClient } from "@opencode-ai/sdk/v2"

async function chat() {
  const client = createOpencodeClient({ baseUrl: "http://localhost:3000" })
  
  // 1. Get available providers and models
  const { data: providersData } = await client.config.providers()
  const provider = providersData.providers.find(p => p.id === "opencode")
  const modelID = providersData.default["opencode"] || Object.keys(provider.models)[0]
  
  // 2. Create a session
  const { data: session } = await client.session.create({})
  
  // 3. Subscribe to events
  const events = await client.event.subscribe({})
  
  // 4. Send a message
  await client.session.prompt({
    sessionID: session.id,
    model: { providerID: "opencode", modelID },
    parts: [{ type: "text", text: "Hello!" }]
  })
  
  // 5. Handle streaming response
  for await (const event of events.stream) {
    if (event.type === "message.part.updated") {
      const part = event.properties.part
      if (part.type === "text") {
        process.stdout.write(part.text)
      }
    }
    if (event.type === "message.updated") {
      const info = event.properties.info
      if (info.role === "assistant" && info.time.completed) {
        break  // Response complete
      }
    }
  }
}
```

### Handling Permissions

```typescript
// In event loop
if (event.type === "permission.asked") {
  const request = event.properties
  console.log(`Permission requested: ${request.tool?.name}`)
  
  // Auto-approve or prompt user
  await client.permission.reply({
    requestID: request.id,
    reply: "once"  // or "always" or "reject"
  })
}
```

---

## Error Handling

All API calls return a response object with either `data` or `error`:

```typescript
const response = await client.session.get({ sessionID: "invalid" })

if (response.error) {
  console.error("Error:", response.error.message)
} else {
  console.log("Session:", response.data)
}
```

For throwing errors on failure:

```typescript
const response = await client.session.get(
  { sessionID: "session-id" },
  { throwOnError: true }
)
// Throws if error occurs
```

---

## Testing

Run the API test script:

```bash
# Start the server in one terminal
cd packages/opencode
bun run src/index.ts serve --port 3000

# In another terminal, run the test script
npx tsx tests/sdk-api-test.ts http://localhost:3000
```

The test script validates all SDK APIs with actual test data and provides a summary of passed, failed, and skipped tests.

---

## Error Handling

All API calls return a response object with either `data` or `error`:

```typescript
const response = await client.session.get({ sessionID: "invalid" })

if (response.error) {
  console.error("Error:", response.error)
} else {
  console.log("Session:", response.data)
}
```

For throwing errors on failure:

```typescript
const response = await client.session.get(
  { sessionID: "session-id" },
  { throwOnError: true }
)
// Throws if error occurs
```

**Error Types:**

```typescript
interface BadRequestError {
  data: unknown
  errors: Array<Record<string, unknown>>
  success: false
}

interface NotFoundError {
  name: "NotFoundError"
  data: { message: string }
}
```

---

## See Also

- [TUI Architecture](./07_TUI_Stack.md)
- [TUI Development Plan](./08_OpenCode_TUI_Development_Plan.md)
- [TUI Implementation Guide](./09_OpenCode_TUI_Implementation_Guide.md)
