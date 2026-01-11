# OpenCode TUI 完整实施计划

> **状态**: Phase 1 进行中  
> **最后更新**: 2026-01-10  
> **版本**: v2.0

---

## 📋 目录

1. [项目概述](#项目概述)
2. [技术栈](#技术栈)
3. [已完成工作](#已完成工作)
4. [实施计划](#实施计划)
5. [架构设计](#架构设计)
6. [开发工作流](#开发工作流)

---

## 项目概述

### 目标

创建一个高性能、美观的终端用户界面 (TUI)，作为 OpenCode 的官方交互界面，提供媲美 Gemini CLI 的用户体验。

### 核心优势

- **直接集成**: 作为 monorepo 包，直接调用 OpenCode 内部 API
- **零延迟**: 无 HTTP 开销，本地函数调用
- **流式响应**: 通过 Bus 事件实时获取 AI 响应
- **一致性**: 与 OpenCode 核心共享类型和配置

---

## 技术栈

### 最终确定版本 ✅

```json
{
  "ink": "6.6.0",
  "react": "18.3.1",
  "@types/react": "18.3.12",
  "node": "20.19.5",
  "typescript": "5.8.2"
}
```

**选择原因**:

- Ink 6.6.0: 最新稳定版，功能最全
- React 18.3.1: 稳定且与 Ink 完美兼容
- TypeScript 编译通过，无类型冲突

### 核心依赖

```json
{
  "dependencies": {
    "opencode": "workspace:*",
    "ink": "6.6.0",
    "react": "18.3.1",
    "ink-spinner": "^5.0.0",
    "ink-text-input": "^6.0.0",
    "ink-select-input": "^6.0.0",
    "marked": "catalog:",
    "chalk": "^5.3.0"
  }
}
```

---

## 已完成工作

### ✅ Phase 1.1: Monorepo 集成 (100%)

**时间**: 2026-01-10  
**投入**: 5 小时

#### 成果

**1. 目录结构**

```
opencode/packages/tui/
├── src/
│   ├── App.tsx
│   ├── index.tsx
│   ├── components/
│   ├── contexts/
│   ├── layouts/
│   └── utils/
│       └── opencode-bridge.ts  # 核心集成层
├── bin/
│   └── opencode-tui.js
├── package.json
└── tsconfig.json
```

**2. OpenCode Bridge** 📦

创建了 `opencode-bridge.ts`，实现：

- Session 管理 (create/load/list/delete)
- 消息处理和流式响应
- Bus 事件订阅
- 清晰的 TypeScript 接口

**路径别名优化**:

```typescript
// Before
import { Session } from "../../../opencode/packages/opencode/src/session"

// After  
import { Session } from "@/session"  // ✨
```

**3. 配置优化**

`tsconfig.json`:

```json
{
  "compilerOptions": {
    "paths": {
      "@/*": ["../opencode/src/*"],
      "@tui/*": ["./src/*"]
    },
    "skipLibCheck": true,
    "strict": false
  },
  "exclude": ["../opencode", "../sdk"]
}
```

`package.json`:

```json
{
  "scripts": {
    "dev": "tsx src/index.tsx",
    "build": "tsc",
    "start": "./bin/opencode-tui.js"
  }
}
```

**4. 版本测试矩阵**

| Ink 版本 | React 版本 | 编译 | 运行时 | 选择 |
|---------|-----------|------|--------|------|
| 5.0.1   | 18.3.1    | ✅   | ✅     | -    |
| 6.4.0   | 19.2.3    | ⚠️   | ✅     | -    |
| 6.5.1   | 19.2.3    | ⚠️   | ✅     | -    |
| 6.6.0   | 18.3.1    | ✅   | ✅     | ✅   |

### 🔄 Phase 1.2: OpenCode Processor 集成 (60%)

**时间**: 2026-01-10  
**状态**: 进行中

#### 已完成

**1. API 研究** ✅

深入研究了 OpenCode 内部 API:

- `SessionProcessor.create()` - 处理器创建
- `LLM.buildStreamInput()` - 构建输入
- `MessageV2.create()` - 消息创建
- `Bus.subscribe()` - 事件监听

**2. Bridge 实现** ✅

实现了真实的消息处理流程:

```typescript
async sendMessage(content: string, onChunk?: Function) {
  // 1. 创建用户消息
  const userMessage = await MessageV2.create({
    sessionID: this.sessionId,
    role: "user",
    parts: [{ type: "text", text: content }]
  });

  // 2. 创建助手消息
  const assistantMessage = await MessageV2.create({
    sessionID: this.sessionId,
    role: "assistant",
    agent: currentAgent.id,
    parentID: userMessage.id
  });

  // 3. 创建处理器
  const processor = SessionProcessor.create({
    assistantMessage,
    sessionID: this.sessionId,
    model: currentAgent.model,
    abort: abortController.signal
  });

  // 4. 订阅流式事件
  Bus.subscribe("part.updated", (event) => {
    if (event.data?.delta) {
      onChunk(event.data.delta);
    }
  });

  // 5. 处理消息
  await processor.process(streamInput);
}
```

**3. 流式响应** ✅

通过 Bus 事件实现实时流式输出。

#### 待完成

- [ ] 端到端测试
- [ ] 错误处理优化
- [ ] Agent 切换功能
- [ ] Session 持久化验证

---

## 实施计划

### Phase 1: OpenCode 集成 (当前)

#### 1.1 Monorepo 集成 ✅ (100%)

- [x] 创建 `packages/tui` 结构
- [x] 配置 workspace 依赖
- [x] 设置路径别名
- [x] 创建 OpenCode Bridge
- [x] 移除 POC 冲突包
- [x] 确定稳定版本组合
- [x] 编译系统配置

#### 1.2 Processor 集成 🔄 (60%)

- [x] 研究 OpenCode processor API
- [x] 实现 sendMessage 真实调用
- [x] 连接 Bus 事件流式响应
- [ ] 测试 Session 创建
- [ ] 测试消息发送
- [ ] 验证端到端流程

#### 1.3 错误处理 (待开始)

- [ ] 实现优雅错误处理
- [ ] 权限请求集成
- [ ] Retry 逻辑
- [ ] 中断处理

#### 1.4 Agent 管理 (待开始)

- [ ] Agent 切换功能
- [ ] Agent 配置同步
- [ ] 自定义 Agent 支持

### Phase 2: 核心 UI 组件

#### 2.1 消息显示

- [ ] Markdown 渲染
- [ ] 代码高亮
- [ ] 工具调用卡片
- [ ] 错误消息样式

#### 2.2 交互组件

- [ ] 文本输入增强
- [ ] 多行编辑器
- [ ] 文件选择器
- [ ] 确认对话框

#### 2.3 状态指示

- [ ] Spinner 动画
- [ ] 进度条
- [ ] 连接状态
- [ ] Streaming 指示器

### Phase 3: 高级功能

#### 3.1 Session 管理

- [ ] Session 列表视图
- [ ] Session 切换
- [ ] Session 搜索
- [ ] Session 导出

#### 3.2 历史记录

- [ ] 消息历史浏览
- [ ] 历史搜索
- [ ] 代码片段提取
- [ ] Diff 查看器

#### 3.3 快捷键

- [ ] 全局快捷键
- [ ] Vim 模式支持
- [ ] 自定义绑定
- [ ] 快捷键帮助

### Phase 4: 优化与美化

#### 4.1 性能优化

- [ ] 虚拟滚动
- [ ] 懒加载
- [ ] 内存优化
- [ ] 渲染优化

#### 4.2 UI 美化

- [ ] 主题系统
- [ ] 自定义颜色
- [ ] 动画效果
- [ ] 图标支持

---

## 架构设计

### 核心架构

```
┌─────────────────────────────────────────────┐
│           OpenCode TUI (Ink/React)          │
├─────────────────────────────────────────────┤
│  Components  │  Contexts  │  Layouts        │
├─────────────────────────────────────────────┤
│          OpenCode Bridge Layer              │
├─────────────────────────────────────────────┤
│   Session  │  Message  │  Bus  │  Agent     │
├─────────────────────────────────────────────┤
│        OpenCode Core (workspace:*)          │
└─────────────────────────────────────────────┘
```

### OpenCode Bridge 设计

**职责**:

1. 类型转换: OpenCode ↔ TUI
2. 事件订阅: Bus → React State
3. 会话管理: Session CRUD
4. 消息处理: 流式响应

**接口**:

```typescript
interface OpenCodeBridge {
  // Session
  createSession(root: string): Promise<TUISession>
  loadSession(id: string): Promise<TUISession | null>
  listSessions(): Promise<TUISession[]>
  deleteSession(id: string): Promise<void>
  
  // Message
  sendMessage(
    content: string, 
    onChunk?: (chunk: string) => void
  ): Promise<TUIMessage>
  
  // Agent
  switchAgent(agent: string): Promise<void>
  
  // Cleanup
  cleanup(): void
}
```

### Context 层次

```typescript
<ServerProvider>      // OpenCode Bridge 实例
  <SessionProvider>   // 当前 Session 状态
    <UIStateProvider> // UI 状态 (loading, errors)
      <StreamingProvider> // 流式响应状态
        <App />
```

---

## 开发工作流

### 本地开发

```bash
# 开发模式 (热重载)
cd packages/tui
bun run dev

# 构建  
bun run build

# 类型检查
bun run typecheck

# 启动编译后版本
bun run start
```

### 测试流程

```bash
# 1. 创建 Session
# 输出: Session ID, .opencode/ 目录

# 2. 发送消息
# 输入: "hello"
# 验证: 流式响应, Bus 事件触发

# 3. 检查数据
ls -la .opencode/
# 验证: session 文件, message 文件存在
```

### 调试

```bash
# 查看 Bus 事件
# 在 bridge 中添加日志:
Bus.subscribe("*", (event) => {
  console.log("Bus event:", event);
});

# 检查 Session
const session = await Session.get(sessionId);
console.log(JSON.stringify(session, null, 2));

# 检查 Messages
const messages = await Session.messages({ sessionID: sessionId });
console.log(messages);
```

---

## 关键决策记录

### 1. 为什么选择 Monorepo 集成？

**备选方案**:

- A. HTTP Server + 客户端
- B. Monorepo 包直接集成  ✅
- C. 独立项目 + SDK

**选择 B 的原因**:

- ✅ 零网络延迟
- ✅ 类型安全
- ✅ 代码复用
- ✅ 简化部署
- ❌ 紧耦合 (可接受)

### 2. 为什么选择 Ink？

**对比**:

| 框架 | 优势 | 劣势 |
|------|------|------|
| Ink | React 生态, 组件化 | React 依赖复杂 |
| Blessed | 成熟稳定 | API 老旧 |
| 原生 ANSI | 完全控制 | 开发成本高 |

**选择**: Ink - 开发效率最高，生态最好

### 3. 为什么降级到 React 18？

**问题**: React 19 + Ink 6.6 有类型冲突

**解决方案**:

- A. 等待 Ink 修复  
- B. 降级到 React 18 ✅
- C. 使用 React 19 + skipLibCheck

**选择 B**: 编译干净，运行稳定

---

## 性能指标

### 目标

- Session 创建: < 100ms
- 消息发送: < 50ms (首字节)
- 流式延迟: < 10ms per chunk
- UI 刷新: 60 FPS
- 内存占用: < 100MB

### 当前状态

基于 monorepo 架构，理论上可以达到：

- Session 创建: ~50ms (本地文件IO)
- 消息发送: ~30ms (函数调用)
- Bus 事件: ~1ms (内存通信)

---

## 风险与挑战

### 已解决 ✅

1. **React 版本冲突** - 使用 React 18.3.1
2. **POC 包冲突** - 已移除
3. **TypeScript 编译** - 配置 skipLibCheck
4. **路径别名** - 配置 @/* 别名

### 待解决

1. **端到端测试** - 需要实际运行验证
2. **错误处理** - 权限、中断等场景
3. **性能优化** - 大量消息时的渲染
4. **跨平台** - Windows 终端兼容性

---

## 参考资源

### OpenCode 内部 API

- [Session API](../packages/opencode/src/session/index.ts)
- [MessageV2 API](../packages/opencode/src/session/message-v2.ts)
- [SessionProcessor](../packages/opencode/src/session/processor.ts)
- [Bus 事件](../packages/opencode/src/bus/index.ts)

### Ink 文档

- [Ink 官方文档](https://github.com/vadimdemedes/ink)
- [Ink 组件](https://github.com/vadimdemedes/ink#components)
- [Ink Hooks](https://github.com/vadimdemedes/ink#hooks)

### Gemini CLI 参考

- UI 设计模式
- 组件实现参考
- 交互流程借鉴

---

## 附录

### A. 版本历史

- **v1.0** (2026-01-09): 初始 POC
- **v1.5** (2026-01-10): Monorepo 迁移
- **v2.0** (2026-01-10): Processor 集成

### B. 贡献者

- 架构设计: AI Assistant
- 实现: AI Assistant
- 测试: 待进行

### C. License

MIT (继承 OpenCode)

---

**最后更新**: 2026-01-10  
**下次更新**: Phase 1.2 完成后
