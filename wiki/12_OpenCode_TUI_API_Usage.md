# OpenCode TUI API 使用指南

本文档是 [OpenCode SDK API Reference](./10_OpenCode_SDK_API_Reference.md) 的补充，专注于 TUI (Terminal User Interface) 如何使用这些 API。

## 概述

OpenCode TUI 通过 SDK 客户端与后端服务器通信。本文档记录了 TUI 源代码中实际使用的所有 API 调用位置和使用场景。

## TUI 架构

### 通信方式

TUI 支持两种通信方式：

#### 1. HTTP 模式（外部访问）

```typescript
// thread.ts
const server = await client.call("server", networkOpts)
url = server.url  // http://localhost:3000
```

#### 2. RPC 模式（内部优化）

```typescript
// thread.ts
url = "http://opencode.internal"
customFetch = createWorkerFetch(client)  // RPC wrapper
events = createEventSource(client, cwd)  // RPC events
```

### 数据同步

TUI 使用 `context/sync.tsx` 管理所有状态同步。

---

## API 使用映射

### 启动流程

TUI 启动时按以下顺序调用 API：

#### 第一阶段：阻塞加载（必须完成）

| API | 文件位置 | 用途 |
|-----|---------|------|
| `config.providers()` | `context/sync.tsx:322` | 加载所有提供商和模型 |
| `provider.list()` | `context/sync.tsx:328` | 加载提供商连接状态 |
| `app.agents()` | `context/sync.tsx:333` | 加载代理列表 |
| `config.get()` | `context/sync.tsx:334` | 加载用户配置 |
| `session.list()` | `context/sync.tsx:316` | 加载最近 30 天会话（如果使用 --continue） |

#### 第二阶段：非阻塞加载

| API | 文件位置 | 用途 |
|-----|---------|------|
| `session.list()` | `context/sync.tsx:343` | 加载会话列表（如果未使用 --continue） |
| `command.list()` | `context/sync.tsx:344` | 加载可用命令 |
| `lsp.status()` | `context/sync.tsx:345` | 加载 LSP 服务器状态 |
| `mcp.status()` | `context/sync.tsx:346` | 加载 MCP 服务器状态 |
| `experimental.resource.list()` | `context/sync.tsx:347` | 加载实验性资源 |
| `formatter.status()` | `context/sync.tsx:348` | 加载格式化器状态 |
| `session.status()` | `context/sync.tsx:349` | 加载所有会话状态 |
| `provider.auth()` | `context/sync.tsx:352` | 加载认证方法 |
| `vcs.get()` | `context/sync.tsx:353` | 加载 VCS 信息 |
| `path.get()` | `context/sync.tsx:354` | 加载路径信息 |

**代码片段:**

```typescript
// context/sync.tsx:313-367
async function bootstrap() {
  const blockingRequests: Promise<unknown>[] = [
    sdk.client.config.providers({}, { throwOnError: true }),
    sdk.client.provider.list({}, { throwOnError: true }),
    sdk.client.app.agents({}, { throwOnError: true }),
    sdk.client.config.get({}, { throwOnError: true }),
    ...(args.continue ? [sessionListPromise] : []),
  ]
  
  await Promise.all(blockingRequests)
  
  // Non-blocking
  Promise.all([
    sdk.client.command.list(),
    sdk.client.lsp.status(),
    sdk.client.mcp.status(),
    // ... 其他
  ])
}
```

---

### 会话管理

#### 创建和加载会话

| 操作 | API | 文件位置 | 触发条件 |
|------|-----|---------|---------|
| 创建会话 | `session.create()` | `component/prompt/index.tsx:505` | 提交提示词时无当前会话 |
| 同步会话 | `session.get()` | `context/sync.tsx:402` | 打开会话页面 |
| 加载消息 | `session.messages()` | `context/sync.tsx:403` | 同步会话（限制 100 条） |
| 加载待办 | `session.todo()` | `context/sync.tsx:404` | 同步会话 |
| 加载差异 | `session.diff()` | `context/sync.tsx:405` | 同步会话 |

**代码片段:**

```typescript
// context/sync.tsx:399-421
async sync(sessionID: string) {
  if (fullSyncedSessions.has(sessionID)) return
  
  const [session, messages, todo, diff] = await Promise.all([
    sdk.client.session.get({ sessionID }, { throwOnError: true }),
    sdk.client.session.messages({ sessionID, limit: 100 }),
    sdk.client.session.todo({ sessionID }),
    sdk.client.session.diff({ sessionID }),
  ])
  
  // 更新本地 store
  fullSyncedSessions.add(sessionID)
}
```

#### 会话操作

| 操作 | API | 文件位置 | 用户触发 |
|------|-----|---------|---------|
| 重命名 | `session.update()` | `component/dialog-session-rename.tsx:22` | 会话重命名对话框 |
| 删除 | `session.delete()` | `component/dialog-session-list.tsx:96` | 会话列表删除按钮 |
| 分享 | `session.share()` | `routes/session/index.tsx:290-291` | Ctrl+X → "Share session" |
| 取消分享 | `session.unshare()` | `routes/session/index.tsx:384-385` | Ctrl+X → "Unshare session" |
| 压缩 | `session.summarize()` | `routes/session/index.tsx:369` | Ctrl+X → "Compact session" |
| 分支 | `session.fork()` | `routes/session/dialog-fork-from-timeline.tsx:35` | Ctrl+X → "Fork from message" |
| 撤销 | `session.revert()` | `routes/session/index.tsx:404-409` | Ctrl+X → "Undo previous message" |
| 重做 | `session.unrevert()` | `routes/session/index.tsx:440` | Ctrl+X → "Redo" |
| 中止 | `session.abort()` | `routes/session/index.tsx:400` | 撤销前中止会话 |

**代码片段:**

```typescript
// routes/session/index.tsx:290-300
{
  title: "Share session",
  keybind: "session_share",
  onSelect: async (dialog) => {
    await sdk.client.session
      .share({ sessionID: route.sessionID })
      .then((res) => Clipboard.copy(res.data!.share!.url))
      .then(() => toast.show({ message: "Share URL copied!", variant: "success" }))
  }
}
```

#### 消息发送

| 操作 | API | 文件位置 | 说明 |
|------|-----|---------|------|
| 发送提示词 | `session.prompt()` | `component/prompt/index.tsx:570` | 用户提交消息 |
| Shell 命令 | `session.shell()` | `component/prompt/index.tsx:535` | `!` 前缀命令 |
| 斜杠命令 | `session.command()` | `component/prompt/index.tsx:554` | `/` 前缀命令 |

**代码片段:**

```typescript
// component/prompt/index.tsx:570
sdk.client.session.prompt({
  sessionID: sessionID,
  model: { providerID, modelID },
  agent: agentName,
  variant: "default",
  parts: [
    { type: "text", text: input },
    ...parts.map(p => ({ type: "file", ... }))
  ]
})
```

---

### 文件操作

| 操作 | API | 文件位置 | 用途 |
|------|-----|---------|------|
| 文件搜索 | `find.files()` | `component/prompt/autocomplete.tsx:189` | @ 自动完成 |
| 文件标签 | `find.files()` | `component/dialog-tag.tsx:18` | 文件选择对话框 |

**代码片段:**

```typescript
// component/prompt/autocomplete.tsx:189
const result = await sdk.client.find.files({
  query: match[1],
  limit: 10
})
```

---

### 权限和问题

#### 权限管理

| 操作 | API | 文件位置 | 触发时机 |
|------|-----|---------|---------|
| 列出权限 | `permission.list()` | 事件驱动 | `permission.asked` 事件 |
| 允许一次 | `permission.reply()` | `routes/session/permission.tsx:159` | 用户点击"Allow once" |
| 拒绝 | `permission.reply()` | `routes/session/permission.tsx:169` | 用户点击"Reject" |
| 始终允许 | `permission.reply()` | `routes/session/permission.tsx:244` | 用户点击"Always allow" |

**代码片段:**

```typescript
// routes/session/permission.tsx:159
sdk.client.permission.reply({
  requestID: request.id,
  reply: "once"
})
```

#### 问题回答

| 操作 | API | 文件位置 | 触发时机 |
|------|-----|---------|---------|
| 列出问题 | `question.list()` | 事件驱动 | `question.asked` 事件 |
| 回答 | `question.reply()` | `routes/session/question.tsx:46` | 用户提交答案 |
| 拒绝 | `question.reject()` | `routes/session/question.tsx:53` | 用户点击"Reject" |

---

### MCP 管理

| 操作 | API | 文件位置 | 用途 |
|------|-----|---------|------|
| 查看状态 | `mcp.status()` | `component/dialog-mcp.tsx:60` | MCP 对话框 |
| 连接 | `mcp.connect()` | `context/local.tsx:363` | 启用 MCP 服务器 |
| 断开 | `mcp.disconnect()` | `context/local.tsx:360` | 禁用 MCP 服务器 |

**代码片段:**

```typescript
// context/local.tsx:360-363
if (enabled) {
  await sdk.client.mcp.disconnect({ name })
} else {
  await sdk.client.mcp.connect({ name })
}
```

---

### 提供商认证

#### OAuth 流程

| 步骤 | API | 文件位置 | 说明 |
|------|-----|---------|------|
| 1. 启动 | `provider.oauth.authorize()` | `component/dialog-provider.tsx:70` | 获取授权 URL |
| 2. 回调 | `provider.oauth.callback()` | `component/dialog-provider.tsx:124` | 完成授权 |
| 3. 刷新 | `instance.dispose()` | `component/dialog-provider.tsx:132` | 重新加载配置 |

#### API 密钥流程

| 步骤 | API | 文件位置 | 说明 |
|------|-----|---------|------|
| 1. 设置 | `auth.set()` | `component/dialog-provider.tsx:229` | 保存 API 密钥 |
| 2. 刷新 | `instance.dispose()` | `component/dialog-provider.tsx:236` | 重新加载配置 |

**代码片段:**

```typescript
// component/dialog-provider.tsx:70
const result = await sdk.client.provider.oauth.authorize({
  providerID: provider.id,
  method: methodIndex
})

// 打开授权 URL
if (result.data!.method === "auto") {
  openUrl(result.data!.url)
}
```

---

### 事件处理

TUI 通过 SSE 接收实时事件：

```typescript
// context/sync.tsx:107-308
sdk.event.listen((e) => {
  const event = e.details
  switch (event.type) {
    case "message.updated":
      // 更新消息到 store
      break
    case "message.part.updated":
      // 更新消息部分（流式响应）
      break
    case "session.status":
      // 更新会话状态
      break
    case "permission.asked":
      // 添加权限请求到 store
      break
    case "question.asked":
      // 添加问题到 store
      break
    case "lsp.updated":
      // 重新加载 LSP 状态
      sdk.client.lsp.status()
      break
    // ... 其他事件
  }
})
```

**处理的事件类型:**

| 事件 | 处理位置 | 操作 |
|------|---------|------|
| `server.instance.disposed` | `sync.tsx:110` | 重新启动 bootstrap |
| `permission.asked` | `sync.tsx:128` | 添加到权限列表 |
| `permission.replied` | `sync.tsx:113` | 从权限列表移除 |
| `question.asked` | `sync.tsx:166` | 添加到问题列表 |
| `question.replied` | `sync.tsx:150` | 从问题列表移除 |
| `session.updated` | `sync.tsx:208` | 更新会话信息 |
| `session.deleted` | `sync.tsx:196` | 从列表移除 |
| `session.status` | `sync.tsx:223` | 更新会话状态 |
| `message.updated` | `sync.tsx:228` | 更新/添加消息 |
| `message.removed` | `sync.tsx:249` | 移除消息 |
| `message.part.updated` | `sync.tsx:263` | 更新消息部分 |
| `message.part.removed` | `sync.tsx:284` | 移除消息部分 |
| `lsp.updated` | `sync.tsx:298` | 重新加载 LSP 状态 |
| `vcs.branch.updated` | `sync.tsx:303` | 更新分支信息 |
| `todo.updated` | `sync.tsx:188` | 更新待办列表 |
| `session.diff` | `sync.tsx:192` | 更新文件差异 |

---

## API 使用模式

### 1. 阻塞加载模式

用于 TUI 启动时必须完成的初始化：

```typescript
const blockingRequests = [
  sdk.client.config.providers({}, { throwOnError: true }),
  sdk.client.provider.list({}, { throwOnError: true }),
  // ...
]

await Promise.all(blockingRequests)
```

### 2. 非阻塞加载模式

用于可以后台加载的数据：

```typescript
Promise.all([
  sdk.client.command.list(),
  sdk.client.lsp.status(),
  // ...
]).then(() => setStore("status", "complete"))
```

### 3. 懒加载模式

用于按需加载的数据：

```typescript
async sync(sessionID: string) {
  if (fullSyncedSessions.has(sessionID)) return
  
  const [session, messages, todo, diff] = await Promise.all([...])
  fullSyncedSessions.add(sessionID)
}
```

### 4. 事件驱动模式

用于实时更新：

```typescript
sdk.event.listen((e) => {
  const event = e.details
  // 根据事件类型更新 store
})
```

### 5. 乐观更新模式

用于提升用户体验：

```typescript
// 先更新 UI
setStore("session", sessionID, { title: newTitle })

// 后台保存
sdk.client.session.update({ sessionID, title: newTitle })
```

---

## 错误处理

### 关键 API 错误处理

```typescript
// context/sync.tsx:359-366
await Promise.all(blockingRequests)
  .catch(async (e) => {
    Log.Default.error("tui bootstrap failed", {
      error: e instanceof Error ? e.message : String(e),
    })
    await exit(e)
  })
```

### 可选 API 错误处理

```typescript
// 静默失败，不阻塞 UI
sdk.client.lsp.status().then((x) => setStore("lsp", x.data!))
```

---

## 性能优化

### 1. 批量请求

```typescript
// 一次性加载所有会话数据
const [session, messages, todo, diff] = await Promise.all([
  sdk.client.session.get({ sessionID }),
  sdk.client.session.messages({ sessionID, limit: 100 }),
  sdk.client.session.todo({ sessionID }),
  sdk.client.session.diff({ sessionID }),
])
```

### 2. 限制数据量

```typescript
// 限制消息数量
session.messages({ sessionID, limit: 100 })

// 限制时间范围
session.list({ start: Date.now() - 30 * 24 * 60 * 60 * 1000 })
```

### 3. 事件批处理

```typescript
// context/sdk.tsx:29-55
const flush = () => {
  const events = queue
  queue = []
  batch(() => {
    for (const event of events) {
      emitter.emit(event.type, event)
    }
  })
}
```

### 4. 缓存已加载数据

```typescript
// 避免重复加载
const fullSyncedSessions = new Set<string>()
if (fullSyncedSessions.has(sessionID)) return
```

---

## 测试覆盖

参见 [OpenCode API Testing](./11_OpenCode_API_Testing.md) 查看完整的测试结果。

**TUI 关键路径测试覆盖:**

- ✅ TUI 启动: 14/14 API (100%)
- ✅ 会话管理: 9/9 核心 API (100%)
- ✅ 文件操作: 4/4 API (100%)
- ✅ 整体核心: 38/39 API (97.4%)

---

## 参考资料

- [OpenCode SDK API Reference](./10_OpenCode_SDK_API_Reference.md) - 完整 API 文档
- [OpenCode API Testing](./11_OpenCode_API_Testing.md) - API 测试报告
- [TUI Development Guide](./09_OpenCode_TUI_Implementation_Guide.md) - TUI 开发指南

---

**文档版本:** 1.0  
**最后更新:** 2026-01-10  
**基于代码:** OpenCode TUI latest
