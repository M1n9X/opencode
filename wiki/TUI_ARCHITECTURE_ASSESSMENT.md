# OpenCode TUI 系统架构评估报告

## 📋 执行摘要

OpenCode 采用了**标准的前后端分离架构**，通过 **RESTful API + Server-Sent Events (SSE) + WebSocket** 实现通信。重新构建一套 TUI 系统的工作量为 **中等规模（约 2-4 周全职开发）**，因为：

1. ✅ **后端 API 已标准化** - 完整的 OpenAPI 3.1 规范
2. ✅ **SDK 已生成** - TypeScript SDK 自动生成
3. ✅ **已有 POC** - `packages/opencode-tui-poc` 验证了可行性
4. ⚠️ **需要重新实现** - 现有 Web UI 组件无法直接复用

---

## 🏗️ 当前架构分析

### 1. 前后端分离方案

#### 后端 (Server)
- **位置**: `packages/opencode/src/server/`
- **框架**: Hono (轻量级 Web 框架)
- **运行时**: Bun
- **端口**: 默认 4096
- **协议**:
  - HTTP/REST API (主要通信)
  - Server-Sent Events (实时事件推送)
  - WebSocket (PTY 终端连接)

#### 前端 (Client)
目前有 **3 个客户端实现**:

1. **Web App** (`packages/app/`)
   - 框架: SolidJS + Vite
   - 用途: 浏览器访问
   
2. **Desktop App** (`packages/desktop/`)
   - 框架: Tauri + SolidJS
   - 用途: 桌面应用（macOS/Windows/Linux）
   
3. **TUI POC** (`packages/opencode-tui-poc/`)
   - 框架: Ink (React for CLI)
   - 状态: 概念验证，约 80 行代码
   - 用途: 验证终端 UI 可行性

### 2. API 接口标准化程度

#### ✅ 完全标准化 - OpenAPI 3.1 规范

**API 规范文件**: `packages/sdk/openapi.json` (10,517 行)

**核心 API 端点**:

```
全局管理
├── GET  /global/health          # 健康检查
├── GET  /global/event           # SSE 事件流
└── POST /global/dispose         # 清理实例

项目管理
├── GET    /project              # 列出项目
├── GET    /project/current      # 当前项目
└── PATCH  /project/{id}         # 更新项目

会话管理
├── GET    /session              # 列出会话
├── POST   /session              # 创建会话
├── GET    /session/{id}         # 获取会话详情
├── DELETE /session/{id}         # 删除会话
├── POST   /session/{id}/prompt  # 发送提示词
├── POST   /session/{id}/abort   # 中止执行
└── GET    /session/{id}/todo    # 获取待办事项

消息管理
├── GET    /session/{id}/message           # 列出消息
├── GET    /session/{id}/message/{msgId}   # 获取消息
└── DELETE /session/{id}/message/{msgId}   # 删除消息

终端 (PTY)
├── GET    /pty                  # 列出终端
├── POST   /pty                  # 创建终端
├── GET    /pty/{id}             # 获取终端信息
├── PUT    /pty/{id}             # 更新终端
├── DELETE /pty/{id}             # 删除终端
└── GET    /pty/{id}/connect     # WebSocket 连接

配置管理
├── GET   /config                # 获取配置
└── PATCH /config                # 更新配置

工具管理
├── GET /experimental/tool/ids   # 列出工具 ID
└── GET /experimental/tool       # 列出工具详情

版本控制
└── GET /vcs                     # 获取 VCS 信息

路径信息
└── GET /path                    # 获取路径信息
```

#### SDK 自动生成

**生成脚本**: `packages/sdk/js/script/build.ts`

```typescript
// SDK 自动从 OpenAPI 规范生成
export function createOpencodeClient(config?: Config) {
  const client = createClient(config)
  return new OpencodeClient({ client })
}
```

**使用示例**:
```typescript
import { createOpencodeClient } from "@opencode-ai/sdk"

const client = createOpencodeClient({
  baseUrl: "http://localhost:4096",
  directory: "/path/to/project"
})

// 类型安全的 API 调用
await client.session.create({ title: "New Session" })
await client.session.prompt({ 
  sessionID: "xxx",
  prompt: "Hello AI"
})
```

### 3. 实时通信机制

#### Server-Sent Events (SSE)
用于服务器到客户端的**单向实时推送**:

```typescript
// 全局事件流
GET /global/event

// 事件类型
type Event = 
  | { type: "session.created", properties: { info: Session } }
  | { type: "session.updated", properties: { info: Session } }
  | { type: "message.created", properties: { info: Message } }
  | { type: "message.part.updated", properties: { part: Part, delta?: string } }
  | { type: "session.status", properties: { sessionID: string, status: Status } }
  | { type: "lsp.updated", properties: { ... } }
  | { type: "server.heartbeat", properties: {} }
```

**前端订阅示例**:
```typescript
const events = await sdk.global.event()
for await (const event of events.stream) {
  if (event.payload.type === "message.part.updated") {
    // 实时更新 AI 响应
    updateUI(event.payload.properties.part)
  }
}
```

#### WebSocket
仅用于 **PTY 终端的双向通信**:

```typescript
// 连接到终端
GET /pty/{ptyID}/connect (Upgrade: websocket)

// 双向数据流
ws.send(userInput)      // 发送命令
ws.onmessage = (data) => // 接收输出
```

---

## 🎯 重建 TUI 系统工作量评估

### 阶段 1: 核心架构 (3-5 天)

#### 1.1 选择 TUI 框架
**推荐方案**: 继续使用 **Ink** (React for CLI)

**理由**:
- ✅ POC 已验证可行性
- ✅ React 生态成熟，组件丰富
- ✅ 团队熟悉 React/TypeScript
- ✅ 支持 Hooks、状态管理

**替代方案**:
- **Blessed** (Node.js 原生) - 更底层，性能更好，但开发效率低
- **Bubbletea** (Go) - 需要重写，跨语言通信复杂

#### 1.2 SDK 集成 (1 天)
```typescript
// 已有 SDK，直接使用
import { createOpencodeClient } from "@opencode-ai/sdk"

const client = createOpencodeClient({
  baseUrl: process.env.OPENCODE_SERVER || "http://localhost:4096",
  directory: process.cwd()
})
```

#### 1.3 事件订阅系统 (2 天)
```typescript
// 实现 SSE 事件处理
class EventManager {
  async subscribe() {
    const events = await this.client.global.event()
    for await (const event of events.stream) {
      this.emit(event.payload.type, event.payload.properties)
    }
  }
}
```

### 阶段 2: UI 组件开发 (5-8 天)

#### 2.1 核心布局组件 (2 天)
```
┌─────────────────────────────────────┐
│ Header (Agent, Model, Status)       │
├─────────────────────────────────────┤
│                                     │
│ Message List (Scrollable)          │
│ ├─ User Message                    │
│ ├─ AI Response (Streaming)         │
│ └─ Tool Calls (Expandable)         │
│                                     │
├─────────────────────────────────────┤
│ Input Prompt (Multi-line)           │
├─────────────────────────────────────┤
│ Footer (Keybinds, Tips)             │
└─────────────────────────────────────┘
```

**组件清单**:
- `<Header />` - 显示当前 Agent、Model、状态
- `<MessageList />` - 虚拟滚动消息列表
- `<MessageTurn />` - 单个对话回合
- `<ToolCall />` - 工具调用展示
- `<PromptInput />` - 多行输入框
- `<Footer />` - 快捷键提示

#### 2.2 会话管理 UI (2 天)
- 会话列表
- 会话切换
- 会话创建/删除
- 会话历史加载

#### 2.3 文件/终端集成 (2 天)
- 文件选择器
- 终端标签页
- 代码高亮显示

#### 2.4 交互功能 (2 天)
- 键盘快捷键
- 命令面板 (Ctrl+P)
- 上下文菜单
- 权限确认对话框

### 阶段 3: 高级功能 (3-5 天)

#### 3.1 流式响应渲染 (2 天)
```typescript
// 监听 message.part.updated 事件
eventManager.on("message.part.updated", ({ part, delta }) => {
  if (delta) {
    // 增量更新 UI
    appendText(part.id, delta)
  }
})
```

#### 3.2 代码差异显示 (2 天)
- 使用 `diff` 库解析差异
- 终端颜色高亮
- 并排/统一视图切换

#### 3.3 性能优化 (1 天)
- 虚拟滚动 (长消息列表)
- 防抖/节流 (输入、滚动)
- 内存管理 (清理旧消息)

### 阶段 4: 测试与优化 (2-3 天)

#### 4.1 功能测试
- 会话创建/删除
- 消息发送/接收
- 工具调用
- 终端交互

#### 4.2 兼容性测试
- macOS Terminal
- iTerm2
- Windows Terminal
- VSCode 集成终端

#### 4.3 性能测试
- 长会话 (1000+ 消息)
- 大文件显示
- 并发请求

---

## 📊 工作量总结

| 阶段 | 任务 | 预估时间 | 难度 |
|------|------|----------|------|
| 1 | 核心架构 | 3-5 天 | 🟢 低 |
| 2 | UI 组件开发 | 5-8 天 | 🟡 中 |
| 3 | 高级功能 | 3-5 天 | 🟡 中 |
| 4 | 测试与优化 | 2-3 天 | 🟢 低 |
| **总计** | | **13-21 天** | |

**全职开发**: 2-4 周  
**兼职开发**: 4-8 周

---

## 🚀 实施建议

### 方案 A: 渐进式开发 (推荐)

**第 1 周**: 最小可用产品 (MVP)
- ✅ 基础会话创建/发送消息
- ✅ 流式 AI 响应显示
- ✅ 简单的消息列表

**第 2 周**: 核心功能
- ✅ 工具调用展示
- ✅ 文件/终端集成
- ✅ 键盘快捷键

**第 3 周**: 高级功能
- ✅ 代码差异显示
- ✅ 会话历史管理
- ✅ 性能优化

**第 4 周**: 打磨与测试
- ✅ Bug 修复
- ✅ 用户体验优化
- ✅ 文档编写

### 方案 B: 复用现有 POC

**优势**:
- 已验证 Ink 可行性
- 基础布局已完成
- 可快速迭代

**需要扩展**:
```typescript
// 当前 POC (80 行)
packages/opencode-tui-poc/src/index.tsx

// 扩展为完整 TUI (预估 2000-3000 行)
packages/opencode-tui/
├── src/
│   ├── components/
│   │   ├── Header.tsx
│   │   ├── MessageList.tsx
│   │   ├── MessageTurn.tsx
│   │   ├── ToolCall.tsx
│   │   ├── PromptInput.tsx
│   │   └── Footer.tsx
│   ├── hooks/
│   │   ├── useSession.ts
│   │   ├── useMessages.ts
│   │   └── useEvents.ts
│   ├── utils/
│   │   ├── sdk.ts
│   │   └── formatting.ts
│   └── index.tsx
└── package.json
```

---

## ⚠️ 技术风险与挑战

### 1. 终端兼容性 (中等风险)
**问题**: 不同终端对 ANSI 转义序列支持不一致

**解决方案**:
- 使用 `chalk` 库统一颜色处理
- 检测终端能力 (`supports-color`)
- 提供降级方案 (纯文本模式)

### 2. 性能问题 (低风险)
**问题**: 长会话可能导致渲染卡顿

**解决方案**:
- 虚拟滚动 (`react-window`)
- 消息分页加载
- 定期清理旧消息

### 3. 复杂交互 (中等风险)
**问题**: 终端 UI 交互不如 GUI 直观

**解决方案**:
- 丰富的键盘快捷键
- 命令面板 (类似 VSCode)
- 清晰的视觉反馈

---

## 🎨 UI 设计参考

### 参考项目
1. **Gemini CLI** - Google 的终端 AI 助手
2. **GitHub Copilot CLI** - 命令行代码助手
3. **Warp Terminal** - 现代化终端
4. **Cursor** - AI 代码编辑器

### 设计原则
- **简洁**: 避免过度装饰
- **高效**: 键盘优先操作
- **清晰**: 明确的状态指示
- **响应式**: 适配不同终端尺寸

---

## 📝 结论

### ✅ 可行性: 高

1. **后端 API 完全标准化** - OpenAPI 3.1 规范，无需修改
2. **SDK 已自动生成** - TypeScript SDK 开箱即用
3. **POC 已验证** - Ink 框架可行
4. **架构清晰** - 前后端分离，易于扩展

### 📊 工作量: 中等

- **最小可用产品**: 1 周
- **完整功能**: 2-4 周
- **打磨优化**: 额外 1-2 周

### 🚀 建议

1. **立即开始**: 基于现有 POC 扩展
2. **渐进式开发**: 先 MVP，后完善
3. **复用 SDK**: 无需重新实现 API 客户端
4. **参考 Gemini CLI**: 学习成熟的 TUI 设计

### 🎯 下一步行动

1. 确认 TUI 功能范围
2. 设计详细的 UI 原型
3. 搭建开发环境
4. 实现 MVP (第 1 周目标)

---

**评估日期**: 2026-01-10  
**评估人**: AI Assistant  
**项目**: OpenCode TUI System
