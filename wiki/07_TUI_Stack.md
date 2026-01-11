# 🏗️ TUI Technology Stack

OpenCode Terminal UI 技术栈详解

---

## 📋 Overview

OpenCode 的 TUI (Terminal User Interface) 已从早期的 Go + Bubble Tea 架构迁移到 **TypeScript + OpenTUI/SolidJS** 架构，实现了完全的 TypeScript 技术栈统一。

---

## 🔧 Current Stack

### Core Technologies

```yaml
Runtime: Bun 1.3.5
Language: TypeScript 5.8.2
Framework: OpenTUI + SolidJS
Location: packages/opencode/src/cli/cmd/tui/
```

### Key Dependencies

```json
{
  "@opentui/core": "0.1.72",
  "@opentui/solid": "0.1.72",
  "solid-js": "1.9.10",
  "opentui-spinner": "0.0.6"
}
```

---

## 🎨 Architecture

### Component Structure

```
packages/opencode/src/cli/cmd/tui/
├── app.tsx                    # 主应用入口
├── routes/
│   └── session/
│       ├── index.tsx          # 会话主界面
│       ├── permission.tsx     # 权限对话框
│       └── question.tsx       # 问题提示
├── component/
│   ├── dialog-session-list.tsx
│   ├── dialog-status.tsx
│   ├── dialog-command.tsx
│   ├── dialog-mcp.tsx
│   ├── dialog-provider.tsx
│   ├── logo.tsx
│   ├── prompt/
│   │   ├── index.tsx          # 输入框
│   │   └── autocomplete.tsx   # 自动完成
│   └── textarea-keybindings.ts
├── ui/
│   ├── dialog.tsx
│   ├── toast.tsx
│   ├── link.tsx
│   ├── spinner.ts
│   ├── dialog-*.tsx           # 各种对话框
│   └── ...
├── context/
│   ├── local.tsx              # 本地状态
│   ├── theme.tsx              # 主题管理
│   ├── keybind.tsx            # 键盘绑定
│   └── exit.tsx               # 退出处理
└── util/
    ├── editor.ts              # 编辑器集成
    └── terminal.ts            # 终端工具
```

### Core Concepts

#### 1. **Reactive Rendering (SolidJS)**

```typescript
import { createSignal, createEffect } from 'solid-js'

function ChatView() {
  const [messages, setMessages] = createSignal<Message[]>([])
  
  // 细粒度响应式更新
  createEffect(() => {
    console.log('Messages updated:', messages().length)
  })
  
  return <MessageList messages={messages()} />
}
```

**优势**:

- 细粒度响应式 (无虚拟 DOM 开销)
- 自动依赖追踪
- 高效的更新机制

#### 2. **Custom Terminal Renderer**

```typescript
import { CliRenderer } from '@opentui/core'

const renderer = new CliRenderer({
  width: terminalWidth,
  height: terminalHeight,
})

renderer.render(vnode)
```

**特点**:

- 直接操作终端缓冲区
- 智能 diff 减少重绘
- ANSI  转义序列优化

#### 3. **Keyboard Event Handling**

```typescript
import { useKeyboard } from '@opentui/solid'

function Dialog({ onClose }) {
  const keyboard = useKeyboard()
  
  createEffect(() => {
    const handler = keyboard()
    if (handler.key === 'escape') {
      onClose()
    }
  })
}
```

#### 4. **Terminal Dimensions Adaptation**

```typescript
import { useTerminalDimensions } from '@opentui/solid'

function ResponsiveLayout() {
  const { width, height } = useTerminalDimensions()
  
  return (
    <Box
      width={width()}
      height={height()}
      flexDirection="column"
    >
      {/* Content */}
    </Box>
  )
}
```

---

## 🆚 Comparison: OpenTUI vs Alternatives

### OpenTUI

**优点**:

- ✅ 基于 SolidJS,细粒度响应式
- ✅ TypeScript 原生支持
- ✅ 与服务器层技术栈统一

**缺点**:

- ⚠️ 版本 0.1.x,成熟度较低
- ⚠️ 社区小,文档少
- ⚠️ VSCode 集成终端性能差
- ⚠️ 长期维护前景不明

### Ink (React-based)

**优点**:

- ✅ 成熟稳定 (30k+ stars)
- ✅ VSCode 兼容性优秀
- ✅ 丰富的组件生态
- ✅ React 主流框架

**缺点**:

- ⚠️ 虚拟 DOM 开销 (实际影响可忽略)
- ⚠️ 需要从 SolidJS 迁移

### Blessed / Blessed-contrib

**优点**:

- ✅ 功能完整
- ✅ 社区成熟

**缺点**:

- ⚠️ 不支持 JSX
- ⚠️ 性能一般
- ⚠️ 维护不活跃

### Comparison Table

| Feature | OpenTUI | Ink | Blessed |
|---------|---------|-----|---------|
| Framework | SolidJS | React | Vanilla JS |
| JSX Support | ✅ | ✅ | ❌ |
| TypeScript | ✅ | ✅ | Partial |
| Performance | Good | Excellent | Moderate |
| VSCode Compat | Poor | Excellent | Good |
| Ecosystem | Small | Large | Moderate |
| Maturity | 0.1.x | Stable | Stable |
| Maintenance | Unknown | Active | Low |

---

## 🚀 Migration Considerations

### Why Consider Ink?

根据性能分析 (详见 [`06_Performance_Optimization.md`](./06_Performance_Optimization.md)):

1. **VScode 卡顿问题**: Ink 在 VSCode 集成终端中实测 45-60 FPS
2. **性能提升**: 预期 50-70% 性能提升
3. **生态成熟**: 大量现成组件,开发效率高
4. **长期可维护**: React 主流框架,人才充足

### Migration Effort

**预计工作量**: 1.5-2 人月

**核心任务**:

- 组件迁移 (Dialog, Prompt, MessageList, etc.)
- 虚拟滚动实现
- 快捷键系统
- Diff 显示
- 集成测试

**详细方案**: 见性能优化技术评审文档

---

## 📚 Development Guide

### Setting Up Development Environment

```bash
# 安装依赖
cd packages/opencode
bun install

# 开发模式运行 TUI
bun run --conditions=browser src/index.ts tui

# TypeScript 类型检查
bun run typecheck
```

### Creating a New Component

```typescript
// packages/opencode/src/cli/cmd/tui/component/my-component.tsx

import { TextAttributes } from '@opentui/core'
import { useKeyboard } from '@opentui/solid'
import { createSignal } from 'solid-js'

export function MyComponent() {
  const [state, setState] = createSignal('initial')
  const keyboard = useKeyboard()
  
  createEffect(() => {
    const k = keyboard()
    // Handle keyboard events
  })
  
  return (
    <box flexDirection="column">
      <text style={ TextAttributes.bold }>
        My Component: {state()}
      </text>
    </box>
  )
}
```

### Styling with TextAttributes

```typescript
import { TextAttributes, fg } from '@opentui/core'

// 文字样式
<text style={TextAttributes.bold}>Bold Text</text>
<text style={TextAttributes.dim}>Dim Text</text>
<text style={TextAttributes.italic}>Italic Text</text>

// 颜色
<text style={fg('#00ff00')}>Green Text</text>
<text style={[TextAttributes.bold, fg('#ff0000')]}>Bold Red</text>
```

### Keyboard Shortcuts

```typescript
import { useKeyboard } from '@opentui/solid'

function App() {
  const keyboard = useKeyboard()
  
  createEffect(() => {
    const k = keyboard()
    
    switch (k.ctrl && k.name) {
      case 'c':
        // Ctrl+C
        process.exit(0)
        break
      case 'l':
        // Ctrl+L - Clear screen
        clearScreen()
        break
    }
    
    switch (k.name) {
      case 'tab':
        // Tab - Switch agent
        switchAgent()
        break
      case 'escape':
        // Esc - Close dialog
        closeDialog()
        break
    }
  })
}
```

---

## 🔗 Related Documentation

### Internal

- [Architecture Overview](./02_Architecture_Overview.md)
- [Performance Optimization](./06_Performance_Optimization.md)
- [OpenCode vs Crush](./05_Opencode_vs_Crush.md)

### External

- [OpenTUI Documentation](https://github.com/wobsoriano/opentui) (Note: Limited)
- [SolidJS Guide](https://www.solidjs.com/guides/getting-started)
- [Ink Alternative](https://github.com/vadimdemedes/ink)

---

## 📝 Status

**当前状态**:

- ✅ TypeScript 技术栈统一
- ✅ 基础功能完整
- ⚠️ 性能问题待优化
- 🔄 考虑迁移到 Ink

**技术债务**:

- OpenTUI 成熟度低
- VSCode 兼容性差
- 性能瓶颈

**改进计划**:

- 短期: 渐进式性能优化
- 长期: 评估 Ink 迁移

---

*Last updated: 2026-01-10*
