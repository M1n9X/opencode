# Bubble TUI Debug Testing Guide

## 🎯 根本原因分析

### 发现的机制

**Enter 键应该的工作流程：**

1. 用户按 Enter 键
2. TUI的 Update() 方法接收 `tea.KeyPressMsg`（line 104）
3. 检查命令匹配（line 327-335）
4. **`InputSubmitCommand`** 绑定到 **"enter"** 键 (`command.go`:330-333)
5. 匹配成功后执行 `executeCommand` (line 1511-1514)
6. 调用 `a.editor.Submit()`
7. Submit() 发送 `app.SendPrompt` 消息
8. TUI接收并调用 `a.app.SendPrompt()` (line 398-416)

### 关键代码位置

**命令定义** (`packages/tui/internal/commands/command.go:330-333`):

```go
{
    Name:        InputSubmitCommand,
    Description: "submit message",
    Keybindings: parseBindings("enter"),
},
```

**命令执行** (`packages/tui/internal/tui/tui.go:1511-1514`):

```go
case commands.InputSubmitCommand:
    updated, cmd := a.editor.Submit()
    a.editor = updated.(chat.EditorComponent)
    cmds = append(cmds, cmd)
```

**消息发送** (`packages/tui/internal/components/chat/editor.go:479-544`):

```go
func (m *editorComponent) Submit() (tea.Model, tea.Cmd) {
    // ... 处理逻辑
    cmds = append(cmds, util.CmdHandler(app.SendPrompt(prompt)))
    return m, tea.Batch(cmds...)
}
```

## 🔍 调试版本

### 新增日志

编译了带调试日志的版本 `tui-debug`，会输出：

1. **每个按键**: `[TUI/KeyPress] key=enter`
2. **命令匹配前**: `[TUI/Commands] Matching key=enter leader=false`
3. **匹配结果**: `[TUI/Commands] Matched count=1`
4. **执行命令**: `[TUI/Commands] Executing commands=1 first=input_submit`

## 📋 测试步骤

### Terminal 1 - 启动服务器

```bash
cd /Users/mxue/GitRepos/Codebreeze/opencode/packages/opencode
bun run src/index.ts serve --port 3001
```

### Terminal 2 - 运行调试版 TUI

```bash
cd /Users/mxue/GitRepos/Codebreeze/opencode
OPENCODE_SERVER=http://localhost:3001 packages/tui/dist/tui-debug .
```

### 操作测试

1. **测试 1: 普通消息**
   - 输入: `hello world`
   - 按: Enter
   - **预期日志**:

     ```
     [TUI/KeyPress] key=enter
     [TUI/Commands] Matching key=enter
     [TUI/Commands] Matched count=1
     [TUI/Commands] Executing first=input_submit
     [SendPrompt] Called text="hello world"
     [SendPrompt] Message sent successfully
     ```

2. **测试 2: Slash 命令**
   - 输入: `/models`
   - 按: Enter
   - **预期日志**:

     ```
     [TUI/KeyPress] key=enter
     [TUI/Commands] Matched count=1
     [SendCommand] Called command="models"
     [SendCommand] Command sent successfully
     ```

## 🐛 可能的问题

### 问题 1: 根本看不到 `[TUI/KeyPress]` 日志

→ **原因**: 键盘事件没有到达TUI Update()
→ **解决**: 检查 Bubble Tea 初始化

### 问题 2: 看到 `[TUI/KeyPress]` 但 `Matched count=0`

→ **原因**: 命令匹配逻辑有问题
→ **解决**: 检查 Commands.Matches() 实现

### 问题 3: `Matched count=1` 但没执行

→ **原因**: ExecuteCommandsMsg 处理有问题
→ **解决**: 检查 executeCommand() 方法

### 问题 4: 执行了但没发送

→ **原因**: Submit() → SendPrompt 流程断了
→ **解决**: 已添加日志，能看到具体哪步失败

## 📊 日志分析指南

**完全没有日志** = 进程可能崩溃或日志被重定向
**只有 KeyPress** = Enter没被识别为命令
**有 Matched 但 count=0** = 命令绑定失败
**有 Executing** = 命令系统工作，问题在Submit/SendPrompt

---

**编译时间**: 2026-01-11 18:27
**调试二进制**: packages/tui/dist/tui-debug (25MB)
