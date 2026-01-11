# OpenCode TUI 完整开发计划

> **项目**: OpenCode TUI (基于 Ink + React)  
> **策略**: Server First → Core UI → Enhancement  
> **预计时间**: 6 周  
> **更新日期**: 2026-01-10

---

## 📋 目录

1. [项目背景](#项目背景)
2. [架构设计](#架构设计)
3. [OpenCode Server API 分析](#opencode-server-api-分析)
4. [实施计划](#实施计划)
5. [技术栈](#技术栈)
6. [验收标准](#验收标准)

---

## 项目背景

### 现状

- **当前 TUI**: OpenTUI (Go-based) - 性能问题严重
  - VSCode 集成终端 FPS: 15
  - 内存占用高
  - 渲染慢

- **MVP TUI**: 已完成基础框架
  - ✅ Ink + React 架构
  - ✅ Context 层次结构
  - ✅ Session/Agent 管理
  - ✅ Mock Server
  - ❌ 未连接真实 Server

### 目标

对标 **Gemini CLI** 的 TUI 实现：

- 流畅的终端体验 (50+ FPS)
- 完整的 Markdown 和代码高亮
- 丰富的 UI 组件
- 完善的错误处理

---

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                   OpenCode TUI (Ink)                    │
├─────────────────────────────────────────────────────────┤
│  Contexts (4层)                                         │
│  ├── ServerContext      # Server 连接管理               │
│  ├── SessionContext     # Session 状态                  │
│  ├── UIStateContext     # UI 对话框状态                 │
│  └── StreamingContext   # 流式消息                      │
├─────────────────────────────────────────────────────────┤
│  Components                                             │
│  ├── Header, Footer, Prompt                             │
│  ├── MessageList (虚拟滚动)                             │
│  ├── Dialogs (Session/Agent/Help)                       │
│  └── Messages (User/AI/Tool/Error)                      │
├─────────────────────────────────────────────────────────┤
│  Utils                                                  │
│  ├── api.ts          # HTTP Client                      │
│  ├── storage.ts      # 本地持久化                       │
│  └── stream-client.ts # 流式响应                        │
└─────────────────────────────────────────────────────────┘
                          ↓
                    (HTTP/WebSocket)
                          ↓
┌─────────────────────────────────────────────────────────┐
│               OpenCode Server (内部)                     │
├─────────────────────────────────────────────────────────┤
│  Session 模块 (packages/opencode/src/session/)         │
│  ├── create/fork/list/get                              │
│  ├── messages/updateMessage/removeMessage              │
│  ├── diff/summary/revert                               │
│  └── share (sync via function API)                     │
└─────────────────────────────────────────────────────────┘
```

### 数据流

```
用户输入 → Prompt Component
    ↓
App.handleSubmit()
    ↓
API Client (api.ts)
    ↓
OpenCode Internal Session API
    ↓
Stream Response (SSE/WebSocket)
    ↓
StreamingContext (实时更新)
    ↓
MessageList Component (显示)
```

---

## OpenCode Server API 分析

### Server 架构

OpenCode 使用**内部模块调用**，不是传统的 HTTP REST API：

```typescript
// packages/opencode/src/session/index.ts

export namespace Session {
  // Session 管理
  create()                    // 创建新 Session
  fork()                      // Fork Session
  list()                      // 列出所有 Session
  get(id)                     // 获取 Session 详情
  update(id, editor)          // 更新 Session
  
  // 消息管理
  messages(sessionID, limit)  // 获取消息列表
  updateMessage(msg)          // 更新消息
  removeMessage(sessionID, messageID)
  
  // Part 管理
  updatePart(part)            // 更新 Part (流式)
  
  // 其他功能
  diff(sessionID)              // 获取代码变更
  summary(sessionID)           // 生成摘要
  share/unshare                // 分享功能
}
```

### 对接策略

由于 OpenCode Server 是内部模块，TUI 有两种对接方式：

#### 方案 1: 直接调用 (推荐)

TUI 作为 OpenCode 的一个入口点，直接调用内部模块：

```typescript
// opencode-tui/src/utils/opencode-bridge.ts

import { Session } from '@opencode/session'
import { MessageV2 } from '@opencode/session/message-v2'

export class OpenCodeBridge {
  async createSession(projectRoot: string) {
    return await Session.create({
      directory: projectRoot
    })
  }
  
  async sendMessage(sessionID: string, content: string) {
    // 使用内部 Session 流程
    const msg = await MessageV2.create(...)
    // 触发 AI 处理
    await Session.Processor.process(...)
  }
}
```

**优点**:

- 无需额外 HTTP 层
- 性能最优
- 直接复用现有逻辑

**缺点**:

- TUI 必须在 OpenCode 项目内
- 耦合度高

#### 方案 2: HTTP Wrapper (备选)

创建一个轻量级 HTTP Server 包装内部 API：

```typescript
// packages/opencode/src/server/tui-api.ts

import { Hono } from 'hono'
import { Session } from '../session'

export const tuiAPI = new Hono()
  .post('/sessions', async (c) => {
    const { projectRoot } = await c.req.json()
    const session = await Session.create({ directory: projectRoot })
    return c.json(session)
  })
  .get('/sessions/:id/messages', async (c) => {
    const sessionID = c.req.param('id')
    const messages = await Session.messages({ sessionID })
    return c.json(messages)
  })
  .post('/sessions/:id/messages', async (c) => {
    const sessionID = c.req.param('id')
    const { content } = await c.req.json()
    // 处理消息...
    return c.json({ message })
  })
  // ... 其他端点
```

**优点**:

- TUI 可独立部署
- 解耦合

**缺点**:

- 额外的 HTTP 开销
- 需要维护 API 层

### 选定方案

**使用方案 1 (直接调用)**，原因：

1. TUI 本身就是 OpenCode 的一部分
2. 性能最优
3. 简化架构

---

## 实施计划

### Phase 1: OpenCode Server 集成 (Week 1-2) ⭐⭐⭐

#### Week 1: 基础集成

**Day 1-2: 创建 OpenCode Bridge**

```bash
# 创建文件
touch opencode-tui/src/utils/opencode-bridge.ts

# 实现内容
- [ ] Session 创建/列表/获取
- [ ] 消息发送/获取
- [ ] Agent 切换
- [ ] 项目初始化
```

**Day 3-4: 更新 API Client**

```typescript
// api.ts 改为使用 OpenCodeBridge

import { OpenCodeBridge } from './opencode-bridge'

export class OpenCodeClient {
  private bridge: OpenCodeBridge
  
  constructor(config: ServerConfig) {
    if (config.useMock) {
      this.bridge = new MockBridge()
    } else {
      this.bridge = new OpenCodeBridge()
    }
  }
  
  async createSession(projectRoot: string) {
    return await this.bridge.createSession(projectRoot)
  }
  
  // ... 其他方法
}
```

**Day 5-7: 流式响应**

OpenCode 使用 `MessageV2.Part` 的增量更新方式：

```typescript
// stream-client.ts

export class OpenCodeStreamClient {
  async streamMessage(sessionID: string, content: string, onChunk: Callback) {
    const bus = Bus.subscribe(MessageV2.Event.PartUpdated, (event) => {
      onChunk({
        type: 'content',
        data: event.delta
      })
    })
    
    // 触发处理
    await Session.Processor.process(...)
    
    // 清理
    bus.unsubscribe()
  }
}
```

#### Week 2: 测试和优化

**测试清单**:

- [ ] 创建 Session 并看到本地 `.opencode/` 目录
- [ ] 发送消息触发 AI 响应
- [ ] 流式响应实时显示每个 Part
- [ ] Agent 切换生效
- [ ] Session 列表正确
- [ ] 退出后重启恢复 Session

**验收标准**:

- 所有基础功能通过 `useMock: false` 测试
- Session 数据正确持久化
- 无明显性能问题

---

### Phase 2: 核心 UI 组件移植 (Week 3-4) ⭐⭐

参考 **Gemini CLI** 实现：

#### Week 3: Markdown & 代码高亮

**Day 1-3: Markdown 渲染**

```bash
# 安装依赖
npm install marked marked-terminal

# 创建组件
src/components/shared/Markdown.tsx
src/utils/markdown-formatter.ts
```

实现要点：

- 将 Markdown 转换为终端友好格式
- 支持标题、列表、代码块、引用
- 集成到 AIMessage 组件

**Day 4-7: 代码语法高亮**

```bash
# 安装依赖
npm install shiki

# 创建组件
src/components/shared/CodeBlock.tsx
```

支持语言：

- TypeScript/JavaScript
- Python, Go, Rust
- JSON, YAML, Shell

#### Week 4: 消息组件美化

**创建文件**:

```
src/components/messages/UserMessage.tsx
src/components/messages/AIMessage.tsx
src/components/messages/ToolCallMessage.tsx
src/components/messages/ErrorMessage.tsx
```

**样式参考** (Gemini CLI):

```
┌─────────────────────────────────────┐
│ 👤 You                    10:30 AM  │
│ ┌─────────────────────────────────┐ │
│ │ create a login component        │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

---

### Phase 3: 高级功能 (Week 5-6) ⭐

#### Week 5: 虚拟滚动

```bash
npm install @tanstack/react-virtual

# 创建组件
src/components/shared/VirtualMessageList.tsx
```

目标：支持 10,000+ 消息流畅滚动

#### Week 6: 搜索 & 美化

- 消息搜索 (Ctrl+F)
- Diff 显示
- 主题系统
- 动画优化

---

## 技术栈

### 核心框架

- **Ink**: 6.6.0 (React for CLI)
- **React**: 19.2.3
- **TypeScript**: 5.x

### OpenCode 依赖

```json
{
  "dependencies": {
    "@opencode/session": "workspace:*",
    "@opencode/bus": "workspace:*",
    "@opencode/storage": "workspace:*",
    "@opencode/util": "workspace:*"
  }
}
```

### UI 增强

- **marked** + **marked-terminal**: Markdown 渲染
- **shiki**: 代码语法高亮
- **@tanstack/react-virtual**: 虚拟滚动
- **diff**: Diff 计算

---

## 验收标准

### Phase 1 完成标准

**功能性**:

- [ ] 连接真实 OpenCode Session
- [ ] 创建/列表/切换 Session
- [ ] 发送消息并接收 AI 响应
- [ ] 流式响应实时显示
- [ ] Agent 切换 (build/plan)
- [ ] 数据正确持久化到 `.opencode/`

**性能**:

- [ ] 消息发送延迟 < 100ms
- [ ] 流式响应首字节 < 500ms
- [ ] 无明显卡顿

**代码质量**:

- [ ] TypeScript 类型完整
- [ ] 错误处理完善
- [ ] 代码注释清晰

### Phase 2 完成标准

**UI 质量**:

- [ ] Markdown 正确渲染
- [ ] 代码块语法高亮
- [ ] 消息样式丰富
- [ ] 对齐 Gemini CLI 视觉风格

### Phase 3 完成标准

**用户体验**:

- [ ] 10,000+ 消息流畅滚动
- [ ] 搜索功能正常
- [ ] Diff 显示清晰
- [ ] 整体体验优秀

---

## 项目结构

```
opencode-tui/
├── src/
│   ├── index.tsx              # 入口
│   ├── App.tsx                # 主应用
│   ├── components/
│   │   ├── Header.tsx
│   │   ├── MessageList.tsx
│   │   ├── Prompt.tsx
│   │   ├── Footer.tsx
│   │   ├── shared/
│   │   │   ├── Dialog.tsx
│   │   │   ├── Spinner.tsx
│   │   │   ├── Markdown.tsx
│   │   │   ├── CodeBlock.tsx
│   │   │   └── VirtualMessageList.tsx
│   │   ├── dialogs/
│   │   │   ├── HelpDialog.tsx
│   │   │   ├── SessionBrowser.tsx
│   │   │   └── AgentDialog.tsx
│   │   └── messages/
│   │       ├── UserMessage.tsx
│   │       ├── AIMessage.tsx
│   │       ├── ToolCallMessage.tsx
│   │       └── ErrorMessage.tsx
│   ├── contexts/
│   │   ├── ServerContext.tsx
│   │   ├── SessionContext.tsx
│   │   ├── UIStateContext.tsx
│   │   └── StreamingContext.tsx
│   ├── layouts/
│   │   └── DefaultLayout.tsx
│   └── utils/
│       ├── api.ts              # API Client (facade)
│       ├── opencode-bridge.ts  # ⭐ OpenCode 内部调用
│       ├── storage.ts          # 本地存储
│       ├── stream-client.ts    # 流式响应
│       └── markdown-formatter.ts
├── package.json
├── tsconfig.json
└── README.md
```

---

## 开发工作流

### 开发模式

```bash
cd /Users/mxue/GitRepos/Codebreeze/opencode-tui

# 启动 TUI
npm run dev

# 类型检查
npm run typecheck

# 构建
npm run build
```

### 测试流程

1. **集成测试** (Phase 1)

   ```bash
   # 启动 TUI
   npm run dev
   
   # 测试步骤
   1. 创建新 Session
   2. 输入 "hello" 测试响应
   3. 输入 "create a component" 测试流式
   4. 按 Tab 切换 Agent
   5. Ctrl+L 查看 Session 列表
   ```

2. **UI 测试** (Phase 2)

   ```bash
   # 测试 Markdown
   输入: "Generate a **bold** text and `code`"
   预期: 看到加粗和代码格式
   
   # 测试代码高亮
   输入: "Show me a TypeScript function"
   预期: 看到语法高亮的代码块
   ```

3. **性能测试** (Phase 3)

   ```bash
   # 创建长会话 (1000+ 消息)
   # 测试滚动流畅度
   # 测试搜索性能
   ```

---

## 风险和挑战

### 技术风险

1. **OpenCode 内部 API 稳定性**
   - **风险**: Session API 可能变更
   - **缓解**: 创建 Bridge 层抽象，隔离依赖

2. **流式响应实现**
   - **风险**: Bus Event 机制可能不适合 TUI
   - **缓解**: 实现 Adapter 模式转换

3. **性能问题**
   - **风险**: 虚拟滚动可能不够流畅
   - **缓解**: 使用成熟库 @tanstack/react-virtual

### 依赖风险

- TUI 必须在 OpenCode 项目内运行
- 与 OpenCode 版本强绑定

**缓解方案**: 未来可选支持方案 2 (HTTP API)

---

## 附录

### Gemini CLI 组件参考

| Gemini CLI 组件 | OpenCode TUI 对应 | 优先级 |
|----------------|-----------------|--------|
| HistoryItemDisplay | MessageList + 消息组件 | P0 |
| MarkdownDisplay | Markdown.tsx | P0 |
| CodeColorizer | CodeBlock.tsx | P0 |
| Table | Table.tsx | P1 |
| Scrollable | VirtualMessageList.tsx | P1 |
| Dialog系列 | Dialogs | P0 |

### 时间线总览

```
Week 1-2:  ████████ Phase 1 - Server 集成
Week 3-4:  ████████ Phase 2 - UI 组件
Week 5-6:  ████████ Phase 3 - 高级功能
```

**总计**: 6 周完整实现

---

**下一步行动**: 开始 Phase 1.1 - 创建 OpenCode Bridge
