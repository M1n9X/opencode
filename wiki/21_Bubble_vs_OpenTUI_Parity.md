# Bubble Tea TUI vs OpenTUI Parity Checklist

**目的**：让 `packages/tui`（Bubble Tea 方案）在功能与细节上完全对齐 `packages/opencode/src/cli/cmd/tui`（OpenTUI 方案），并给出实现优先级。

## 现状对照（功能 / 状态）
- 入口与渲染：两端各自渲染栈；Bubble 缺少 console `onCopySelection`/OSC52 钩子。
- 路由/页面：OpenTUI 有 RouteProvider；Bubble 以单页 Chat 手工切换 Home/Chat，功能等价。
- 键盘体系：OpenTUI `useKeyboard`；Bubble `commands`+leader+防抖，行为近似。
- 输入与附件：双方支持文件/符号/代理附件、历史导航、bash 模式；Bubble 有 StashDialog，但无 frecency UI；OpenTUI 有 stash/frecency provider。
- 自动完成：`/` 命令、`@` 代理/文件/符号均已对齐。
- 消息视图：两端均支持思考块、工具详情、代码折叠；Bubble 代码/工具高亮与折叠细节弱于 OpenTUI。
- 侧边栏：会话/MCP/LSP/Todo/Diff 均已实现，Bubble 支持折叠。
- 权限弹窗：OpenTUI 展示 diff/路径/模式，支持 once/always/reject 多阶段；Bubble 仅按键 enter/esc/a，缺少 UI 与 diff 预览。
- 问卷（question）：双方完整实现。
- Session 列表/重命名/分享：双方具备。
- 时间线 & 分叉：OpenTUI 时间线可直接 fork；Bubble 仅跳转/还原，无 fork 入口。
- 子会话/子代理：双方具备。
- 主题/模型/代理/MCP Dialog：双方具备。
- 导出：双方支持 4 个选项保存/打开。
- 标签/文件搜索：OpenTUI 有 Tag Dialog；Bubble 仅 FileDialog，无 tag 过滤。
- Diff 视图：OpenTUI `<diff>` 组件用于权限/消息内嵌；Bubble 仅侧边栏统计，缺少嵌入展示。
- 终端特性：OpenTUI 启用 kitty 键盘、鼠标选择复制到剪贴板；Bubble 仅鼠标选区复制，缺少 console copy hook。

## 缺口列表（需补齐）
1) 权限弹窗完整交互与 diff 预览（once/always/reject，多阶段 UI）。  
2) 时间线中一键 fork（选中消息 -> fork 子会话）。  
3) 终端复制/kitty 对齐：支持 OSC52/onCopySelection，kitty 键盘选项。  
4) 消息渲染细节：工具调用层级展开、代码/工具高亮与 OpenTUI 一致；在权限/工具输出中嵌入 diff。  
5) Tag 补全/搜索：新增 Tag Dialog 或扩展 FileDialog。  
6) frecency/stash UI parity（可选）：在输入区增加入口或继续对话框方式。  

## 优先级与行动计划
P1 权限弹窗：在 Bubble 增加 Modal（复用 `components/diff`），实现 once/always/reject 与 diff/路径/模式展示；与服务器交互调用 `Session.Permissions.Respond`。  
P2 时间线 fork：在 `dialog/timeline.go` 添加 fork 操作（绑定按键/回车），调用现有 `ForkDialog` 或直接 `Session.Fork`。  
P3 终端复制 & kitty：在 util 中增加 OSC52 复制，复制/选择统一到系统剪贴板；后续可启用 kitty keyboard 选项。  
P4 渲染细节：增强工具块（JSON pretty、长输出可展开/折叠），嵌入 diff 支持，匹配主题高亮。  
P5 Tag 搜索：新增 `dialog/tag.go`（备用），保持 FileDialog 为默认文件选择；后续可接入真实 tags API 时启用。  
P6 Frecency UI：提供快捷入口显示 stash/frecency（低风险，按需）。  

## 完成判定
- 功能行为与 OpenTUI 一致：相同快捷键/交互路径能完成同结果。  
- 视觉与信息：权限/时间线/工具输出呈现的信息不少于 OpenTUI。  
- 终端交互：kitty/OSC52 下复制与键盘输入体验匹配。  
- 回归用例：权限请求、时间线 fork、tag 选择、工具调用展示、长文复制、侧边栏数据等均可复现。  
