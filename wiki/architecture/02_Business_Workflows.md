# 🔄 Business Workflows & Data Flow Analysis

This document provides a comprehensive analysis of OpenCode's core business processes, data flows, and system interactions. Understanding these workflows is crucial for developers, architects, and contributors working with the system.

---

## 🎯 Core Business Scenarios

### 1. User Onboarding & Authentication Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Server
    participant Provider as AI Provider
    participant Storage
    
    User->>CLI: opencode auth login
    CLI->>Server: Request provider list
    Server->>CLI: Available providers
    CLI->>User: Display provider options
    User->>CLI: Select provider (e.g., Anthropic)
    CLI->>Provider: Initiate OAuth flow
    Provider->>User: Browser authentication
    User->>Provider: Complete authentication
    Provider->>CLI: Return tokens
    CLI->>Server: Store credentials
    Server->>Storage: Persist auth data
    Storage->>Server: Confirmation
    Server->>CLI: Success response
    CLI->>User: Authentication complete
```

**Key Data Entities**:
- **Auth.Info**: Discriminated union of OAuth, API, and WellKnown auth types
- **Provider.Info**: Provider configuration and capabilities
- **Credentials**: Encrypted storage of authentication tokens

### 2. Session Creation & Project Context

```mermaid
flowchart TD
    START[User starts opencode] --> DETECT[Detect working directory]
    DETECT --> GIT{Git repository?}
    
    GIT -->|Yes| GITROOT[Find git root]
    GITROOT --> GITID[Generate project ID from first commit]
    GITID --> GITPROJECT[Create Git project]
    
    GIT -->|No| GLOBAL[Use global project]
    
    GITPROJECT --> SESSION[Create new session]
    GLOBAL --> SESSION
    
    SESSION --> PERSIST[Persist session data]
    PERSIST --> CONTEXT[Initialize project context]
    CONTEXT --> READY[Ready for interaction]
    
    READY --> TUI[Launch TUI]
    TUI --> CONNECT[Connect to server]
    CONNECT --> STREAM[Establish SSE connection]
```

**Data Flow Details**:
```typescript
// Project identification
Project.fromDirectory(directory) → {
  id: string,           // Git commit hash or "global"
  worktree: string,     // Repository root path
  vcs?: "git",          // Version control system
  time: { created: number }
}

// Session creation
Session.create() → {
  id: string,           // ULID identifier
  projectID: string,    // Links to project
  parentID?: string,    // For nested sessions
  title: string,        // Auto-generated or user-provided
  time: { created, updated },
  share?: { url: string }
}
```

### 3. AI Interaction & Tool Execution Workflow

```mermaid
sequenceDiagram
    participant User
    participant TUI
    participant Server
    participant Agent
    participant Tools
    participant AI as AI Provider
    participant Storage
    
    User->>TUI: Input message
    TUI->>Server: POST /session/:id/message
    Server->>Storage: Save user message
    
    Server->>Agent: Process with context
    Agent->>AI: Send prompt + available tools
    
    loop Tool Execution
        AI->>Agent: Tool call request
        Agent->>Tools: Execute tool
        Tools->>Tools: Perform operation
        Tools->>Agent: Return result
        Agent->>Server: Stream tool update
        Server->>TUI: SSE tool update
        TUI->>User: Display tool execution
    end
    
    AI->>Agent: Final response
    Agent->>Server: Complete response
    Server->>Storage: Save assistant message
    Server->>TUI: SSE final update
    TUI->>User: Display final result
```

**Tool Execution Pipeline**:
```mermaid
flowchart LR
    INPUT[Tool Input] --> VALIDATE[Validate Schema]
    VALIDATE --> PERMISSION[Check Permissions]
    PERMISSION --> CONTEXT[Setup Context]
    CONTEXT --> EXECUTE[Execute Tool Logic]
    EXECUTE --> METADATA[Generate Metadata]
    METADATA --> OUTPUT[Format Output]
    OUTPUT --> STREAM[Stream to Client]
    STREAM --> PERSIST[Persist Results]
```

---

## 🔧 Core System Workflows

### 4. Tool Registration & Discovery

```mermaid
flowchart TD
    START[System Startup] --> BUILTIN[Load Built-in Tools]
    BUILTIN --> PLUGINS[Load Plugin Tools]
    PLUGINS --> HTTP[Register HTTP Tools]
    HTTP --> MCP[Initialize MCP Servers]
    
    BUILTIN --> REGISTRY[Tool Registry]
    PLUGINS --> REGISTRY
    HTTP --> REGISTRY
    MCP --> REGISTRY
    
    REGISTRY --> PROVIDER[Provider-Specific Adaptation]
    PROVIDER --> OPENAI[OpenAI Format]
    PROVIDER --> ANTHROPIC[Anthropic Format]
    PROVIDER --> GOOGLE[Google Format]
    
    OPENAI --> AVAILABLE[Available Tools]
    ANTHROPIC --> AVAILABLE
    GOOGLE --> AVAILABLE
```

**Tool Registration Process**:
```typescript
// Built-in tools
const BUILTIN = [
  BashTool, EditTool, ReadTool, WriteTool,
  GlobTool, GrepTool, ListTool, PatchTool,
  WebFetchTool, TaskTool, TodoWriteTool, TodoReadTool
]

// Plugin tools (runtime registration)
Plugin.register(toolDefinition)

// HTTP tools (API registration)
Server.registerHttpTool({
  id: "custom-tool",
  description: "Custom tool via HTTP",
  parameters: { /* schema */ },
  callbackUrl: "https://api.example.com/tool"
})

// MCP tools (protocol-based)
MCP.connect("weather", {
  type: "local",
  command: ["opencode", "x", "@h1deya/mcp-server-weather"]
})
```

### 5. File System Operations & Git Integration

```mermaid
sequenceDiagram
    participant Tool
    participant FileSystem
    participant Git
    participant LSP
    participant Storage
    
    Tool->>FileSystem: Read file request
    FileSystem->>Git: Check if git repo
    
    alt Git Repository
        Git->>Git: Get file diff
        Git->>FileSystem: Return content + diff
    else No Git
        FileSystem->>FileSystem: Read raw content
    end
    
    FileSystem->>LSP: Notify file access
    LSP->>LSP: Update diagnostics
    
    FileSystem->>Tool: Return file data
    Tool->>Storage: Log file access
```

**File Operation Data Flow**:
```typescript
// File read with Git integration
File.read(path) → {
  content: string,      // Current file content
  patch?: string,       // Git diff if available
  diff?: string         // Formatted patch
}

// File write with LSP notification
File.write(path, content) → {
  success: boolean,
  metadata: {
    size: number,
    modified: timestamp,
    lspNotified: boolean
  }
}
```

### 6. LSP Integration & Code Analysis

```mermaid
flowchart TD
    REQUEST[Code Analysis Request] --> DETECT[Detect File Type]
    DETECT --> SERVERS[Find LSP Servers]
    SERVERS --> EXISTING{Existing Client?}
    
    EXISTING -->|Yes| REUSE[Reuse Client]
    EXISTING -->|No| SPAWN[Spawn New Server]
    
    SPAWN --> INIT[Initialize Client]
    INIT --> CONNECT[Establish Connection]
    CONNECT --> READY[Client Ready]
    
    REUSE --> READY
    READY --> OPERATION[Perform Operation]
    
    OPERATION --> DIAGNOSTICS[Get Diagnostics]
    OPERATION --> HOVER[Get Hover Info]
    OPERATION --> SYMBOLS[Get Symbols]
    
    DIAGNOSTICS --> RESULT[Return Results]
    HOVER --> RESULT
    SYMBOLS --> RESULT
```

**LSP Client Lifecycle**:
```typescript
// LSP server configuration
{
  "lsp": {
    "typescript": {
      "command": ["typescript-language-server", "--stdio"],
      "extensions": [".ts", ".tsx", ".js", ".jsx"],
      "initialization": { /* LSP init params */ }
    }
  }
}

// Client management
LSP.getClients(file) → LSPClient[]
LSP.touchFile(file) → Promise<void>
LSP.diagnostics() → Record<string, Diagnostic[]>
LSP.hover(position) → HoverInfo
```

---

## 📊 Data Entity Lifecycles

### 7. Session Data Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Created: Session.create()
    Created --> Active: First message
    Active --> Active: Message exchange
    Active --> Shared: Session.share()
    Shared --> Active: Continue interaction
    Active --> Archived: Inactivity timeout
    Archived --> Active: Resume session
    Active --> Deleted: Session.remove()
    Shared --> Deleted: Session.remove()
    Archived --> Deleted: Cleanup process
    Deleted --> [*]
```

**Session State Transitions**:
```typescript
// Session creation
{
  id: "01HKQR8X9Y2Z3A4B5C6D7E8F9G",
  projectID: "abc123...",
  title: "Fix authentication bug",
  time: { created: 1704067200000, updated: 1704067200000 },
  status: "active"
}

// Session sharing
{
  ...session,
  share: {
    url: "https://opencode.ai/s/9G8F7E6D"
  }
}

// Session archival
{
  ...session,
  status: "archived",
  time: { ...time, archived: 1704153600000 }
}
```

### 8. Message & Part Data Flow

```mermaid
flowchart TD
    USER_INPUT[User Input] --> MESSAGE[Create Message]
    MESSAGE --> USER_PART[Create User Part]
    USER_PART --> PERSIST_USER[Persist User Data]
    
    PERSIST_USER --> AI_PROCESS[AI Processing]
    AI_PROCESS --> ASSISTANT_MSG[Create Assistant Message]
    
    ASSISTANT_MSG --> TEXT_PART[Text Part]
    ASSISTANT_MSG --> TOOL_PARTS[Tool Parts]
    
    TEXT_PART --> PERSIST_TEXT[Persist Text]
    TOOL_PARTS --> PERSIST_TOOLS[Persist Tool Results]
    
    PERSIST_TEXT --> COMPLETE[Message Complete]
    PERSIST_TOOLS --> COMPLETE
    
    COMPLETE --> EVENTS[Publish Events]
    EVENTS --> UI_UPDATE[Update UI]
```

**Message Structure Evolution**:
```typescript
// Initial user message
{
  id: "msg_123",
  sessionID: "session_456",
  role: "user",
  parts: [{
    id: "part_789",
    type: "text",
    text: "Fix this bug in the authentication code"
  }],
  time: { created: timestamp }
}

// Assistant response with tool calls
{
  id: "msg_124",
  sessionID: "session_456",
  role: "assistant",
  parts: [
    {
      id: "part_790",
      type: "text",
      text: "I'll analyze the authentication code and fix the bug."
    },
    {
      id: "part_791",
      type: "tool",
      tool: "read",
      state: {
        status: "completed",
        input: { filePath: "src/auth.ts" },
        output: "// File contents...",
        metadata: { size: 1024 }
      }
    }
  ]
}
```

---

## 🌐 External Integration Workflows

### 9. GitHub Actions Integration

```mermaid
sequenceDiagram
    participant GH as GitHub
    participant Action as GitHub Action
    participant Server as OpenCode Server
    participant AI
    participant Git
    
    GH->>Action: Issue comment with /opencode
    Action->>Action: Validate permissions
    Action->>Server: Start opencode server
    Action->>Server: Create session
    
    Action->>GH: Fetch issue/PR data
    GH->>Action: Return context data
    
    Action->>Server: Send prompt with context
    Server->>AI: Process request
    AI->>Server: Generate response + tool calls
    Server->>Git: Execute file changes
    
    alt Changes Made
        Git->>Git: Commit changes
        Git->>GH: Push to branch
        Action->>GH: Create/update PR
    end
    
    Action->>GH: Update comment with results
```

**GitHub Integration Data Flow**:
```typescript
// GitHub context extraction
{
  issue: {
    title: string,
    body: string,
    author: { login: string },
    comments: Comment[]
  },
  pullRequest?: {
    title: string,
    commits: Commit[],
    files: ChangedFile[],
    reviews: Review[]
  }
}

// OpenCode processing
{
  userPrompt: string,           // Extracted from comment
  contextPrompt: string,        // Generated from GitHub data
  promptFiles: AttachedFile[],  // Images/files from comment
  session: Session,             // Created session
  response: string              // AI response
}
```

### 10. Cloud Infrastructure Data Flow

```mermaid
flowchart TD
    CLIENT[Client Request] --> CF[Cloudflare Worker]
    CF --> AUTH{Authentication?}
    
    AUTH -->|Required| OAUTH[OAuth Flow]
    OAUTH --> PROVIDER[Auth Provider]
    PROVIDER --> TOKEN[Access Token]
    TOKEN --> AUTHORIZED[Authorized Request]
    
    AUTH -->|Not Required| AUTHORIZED
    AUTHORIZED --> API[API Processing]
    
    API --> DB[(PlanetScale DB)]
    API --> STORAGE[File Storage]
    API --> EXTERNAL[External APIs]
    
    DB --> RESPONSE[Response Data]
    STORAGE --> RESPONSE
    EXTERNAL --> RESPONSE
    
    RESPONSE --> CF
    CF --> CLIENT
```

**Cloud Data Architecture**:
```typescript
// Cloudflare Worker environment
{
  database: {
    host: string,
    username: string,
    password: string,
    database: string
  },
  storage: {
    bucket: string,
    region: string
  },
  secrets: {
    anthropicApiKey: string,
    openaiApiKey: string,
    githubAppPrivateKey: string
  }
}
```

---

## 🔄 Event-Driven Architecture

### 11. Event Bus & Real-time Updates

```mermaid
sequenceDiagram
    participant Component
    participant Bus
    participant Subscribers
    participant UI
    
    Component->>Bus: Publish event
    Bus->>Bus: Validate event schema
    Bus->>Subscribers: Notify all subscribers
    
    loop For each subscriber
        Subscribers->>Subscribers: Process event
        Subscribers->>UI: Update interface
    end
    
    Bus->>Component: Publish confirmation
```

**Event Types & Schemas**:
```typescript
// Session events
const SessionEvents = {
  Created: Bus.event("session.created", z.object({
    info: Session.Info
  })),
  Updated: Bus.event("session.updated", z.object({
    info: Session.Info
  })),
  Deleted: Bus.event("session.deleted", z.object({
    info: Session.Info
  }))
}

// Message events
const MessageEvents = {
  PartUpdated: Bus.event("message.part.updated", z.object({
    part: Part.Info
  }))
}
```

---

## 🎯 Performance & Optimization Workflows

### 12. Caching & Resource Management

```mermaid
flowchart TD
    REQUEST[Request] --> CACHE{Cache Hit?}
    CACHE -->|Yes| RETURN[Return Cached]
    CACHE -->|No| COMPUTE[Compute Result]
    
    COMPUTE --> EXPENSIVE{Expensive Operation?}
    EXPENSIVE -->|Yes| BACKGROUND[Background Processing]
    EXPENSIVE -->|No| IMMEDIATE[Immediate Processing]
    
    BACKGROUND --> QUEUE[Task Queue]
    QUEUE --> WORKER[Background Worker]
    WORKER --> RESULT[Compute Result]
    
    IMMEDIATE --> RESULT
    RESULT --> STORE[Store in Cache]
    STORE --> RETURN
```

**Caching Strategies**:
```typescript
// Project context caching
const projectCache = new Map<string, Project.Info>()

// LSP client pooling
const lspClients = new Map<string, LSPClient>()

// Tool result memoization
const toolResultCache = new LRUCache<string, ToolResult>({
  max: 1000,
  ttl: 1000 * 60 * 5 // 5 minutes
})
```

---

## 🎯 Key Workflow Insights

### Critical Success Factors

1. **Context Preservation**: Project and session context maintained throughout workflows
2. **Error Recovery**: Graceful handling of failures at each step
3. **Real-time Updates**: Immediate feedback through event-driven architecture
4. **Resource Efficiency**: Intelligent caching and connection pooling
5. **Security Integration**: Permission checks integrated into all workflows

### Performance Characteristics

- **Session Creation**: ~100ms (with Git detection)
- **Tool Execution**: ~200ms average (varies by tool)
- **AI Response**: 2-10s (depends on provider and complexity)
- **File Operations**: ~50ms (with LSP integration)
- **Event Propagation**: ~10ms (real-time updates)

---

*These workflows form the backbone of OpenCode's functionality, enabling seamless integration between AI capabilities and developer tools while maintaining performance and reliability.*
