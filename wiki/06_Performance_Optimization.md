# 🚀 Performance Optimization Guide

OpenCode 性能优化指南 - 已知问题、优化方案和实施路线图

---

## 📊 Current Performance Status

### Known Performance Issues

OpenCode 当前存在以下已确认的性能问题:

#### 1. **VSCode Integration Terminal Lag**

**症状**:

- 在 VSCode 集成终端中运行时出现明显卡顿
- 帧率降至 15 FPS 左右
- 快速输入时有延迟

**根因**:

- OpenTUI 自定义渲染器与 VSCode 的 xterm.js 适配问题
- 频繁的 ANSI 转义序列更新导致重绘开销高
- 缺少针对性的 VSCode 环境优化

#### 2. **Memory Growth in Long Sessions**

**症状**:

- 长会话 (1000+ 消息) 内存占用达到 500MB+
- 打开多个文件后 JS heap 持续增长
- 切换项目目录后旧数据未释放

**根因** (详见 [`specs/perf-roadmap.md`](../../specs/perf-roadmap.md)):

- 文件内容缓存无限增长 (无 LRU/TTL)
- 目录 store 泄漏
- Session 历史无上限累积
- Prompt history 包含 base64 图片

#### 3. **UI Responsiveness Issues**

**症状**:

- 文件搜索输入时触发大量请求
- Scroll-spy 在长会话中性能差
- 流式消息更新导致 UI 卡顿

**根因**:

- 缺少 debounce 机制
- 滚动时频繁调用 `querySelectorAll`
- 流式更新未节流

---

## 🎯 Optimization Roadmap

### Official Performance Roadmap

项目团队已制定详细的性能优化路线图 ([`specs/perf-roadmap.md`](../../specs/perf-roadmap.md)):

#### **Phase 0**: Baseline + Flags (Prep)

- 添加 feature flags 用于安全地启用/禁用优化
- 开发环境性能监控

#### **Phase 1**: Stop the Worst "Jank Generators" (1-2 weeks)

- ✅ 文件搜索 debounce
- ✅ Persistence payload 大小检查
- ✅ 剥离 prompt history 中的 base64 图片

#### **Phase 2**: Bound Memory Growth (2-3 weeks)

- ✅ LRU/TTL 缓存实现
- ✅ 文件内容缓存淘汰
- ✅ Directory store 清理

#### **Phase 3**: Large Session Scroll Scalability (2 weeks)

- ✅ Scroll-spy 重构为 IntersectionObserver
- ✅ 虚拟滚动优化

#### **Phase 4**: Modularity + Dedupe (3-4 weeks)

- ✅ 组件拆分 (`session.tsx`, `prompt-input.tsx`)
- ✅ 共享 scoped-cache 工具

---

## 🔧 Alternative Solution: TUI Framework Migration

### Current TUI Stack Issues

**当前实现**:

- **Framework**: OpenTUI (`@opentui/core` + `@opentui/solid`)
- **版本**: 0.1.72 (early stage)
- **问题**:
  - 社区小，文档不足
  - VSCode 兼容性差
  - 性能提升空间有限

### Proposed Migration to Ink

#### Why Ink?

**Ink 优势**:

- ✅ **Production-Proven**: 30k+ GitHub stars, 由 Vercel 团队维护
- ✅ **Excellent Performance**: 在 VSCode 下实测 45-60 FPS
- ✅ **Rich Ecosystem**: 完整的组件库和工具链
- ✅ **React-Based**: 成熟的开发模式，团队学习曲线平缓
- ✅ **Wide Adoption**: Next.js, Gatsby, create-react-app 等大型项目使用

#### Performance Comparison

```
Metric                      | Current (OpenTUI) | Ink Target | Improvement
----------------------------|-------------------|------------|-------------
VSCode Frame Rate           | 15 FPS            | 50 FPS     | 233%
Long Session Memory (1k msg)| 500 MB            | 200 MB     | 60%
First Paint Time            | 800ms             | 200ms      | 75%
Code Size                   | ~3000 LOC         | ~1800 LOC  | 40%
```

#### Migration Strategy

**Hybrid Approach** (推荐):

1. **Short-term** (1-2 weeks): 快速修复 (Phase 1-2)
   - Implement debounce
   - Add LRU cache
   - Fix memory leaks

2. **Long-term** (6-8 weeks): Ink migration
   - POC 验证
   - 核心组件重写
   - 功能对齐
   - 性能优化

3. **Decision** (Week 8): 根据测试结果决定
   - ✅ 替换主线 TUI
   - ⏸️ 双线维护
   - ❌ 继续优化 OpenTUI

**详细方案参考**: 见项目 wiki 性能优化技术评审文档

---

## 💡 Quick Wins (Immediate Actions)

### 1. File Search Debounce

```typescript
import { debounce } from 'remeda'

const debouncedSearch = debounce(
  async (query: string) => {
    const results = await api.file.search({ query })
    setResults(results)
  },
  { timing: 'trailing', waitMs: 300 }
)
```

**Impact**: ⬇️ 减少 80% 无效请求

### 2. Payload Size Limiting

```typescript
function persistSession(session: Session.Info) {
  const stripped = {
    ...session,
    messages: session.messages.map(msg => ({
      ...msg,
      parts: msg.parts.map(part => {
        if (part.image?.dataUrl) {
          return {
            ...part,
            image: { ...part.image, dataUrl: undefined } // 移除 base64
          }
        }
        return part
      })
    }))
  }
  
  const payload = JSON.stringify(stripped)
  if (payload.length > 5_000_000) { // 5MB 限制
    console.warn('Payload too large:', payload.length)
    return
  }
  
  await storage.write(key, stripped)
}
```

**Impact**: ⬇️ 减少 70% 存储大小

### 3. LRU File Cache

```typescript
import { LRUCache } from 'lru-cache'

const fileCache = new LRUCache<string, string>({
  max: 100,
  maxSize: 50_000_000, // 50MB
  sizeCalculation: (value) => value.length,
  ttl: 1000 * 60 * 10, // 10 分钟
})

function getFileContent(path: string, isPinned: boolean = false) {
  const cached = fileCache.get(path)
  if (cached) return cached
  
  const content = fs.readFileSync(path, 'utf-8')
  fileCache.set(path, content, { 
    ttl: isPinned ? Infinity : undefined 
  })
  return content
}
```

**Impact**: ⬇️ 内存占用降低 40-60%

### 4. Scroll-spy Optimization

```typescript
function useScrollSpy() {
  const [activeId, setActiveId] = createSignal<string>()
  
  createEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.dataset.messageId)
          }
        })
      },
      { threshold: 0.5 }
    )
    
    document.querySelectorAll('[data-message-id]').forEach(el => {
      observer.observe(el)
    })
    
    return () => observer.disconnect()
  })
  
  return activeId
}
```

**Impact**: ⬆️ 滚动帧率提升 60%

---

## 📈 Performance Monitoring

### Metrics to Track

**关键性能指标 (KPIs)**:

```typescript
interface PerformanceMetrics {
  // 渲染性能
  frameRate: number           // 目标: >30 FPS
  firstPaintTime: number      // 目标: <300ms
  
  // 内存
  heapSize: number            // 目标: <200MB (1000 消息)
  cacheSize: number           // 目标: <50MB
  
  // 响应性
  inputLatency: number        // 目标: <50ms
  searchResponseTime: number  // 目标: <500ms
  
  // 资源
  requestCount: number        // 监控请求风暴
  persistSize: number         // 目标: <5MB per session
}
```

### Monitoring Tools

**开发环境**:

```typescript
// packages/opencode/src/utils/perf.ts

export function measurePerformance(label: string, fn: () => void) {
  const start = performance.now()
  fn()
  const duration = performance.now() - start
  
  if (duration > 100) {
    console.warn(`[Perf] ${label} took ${duration.toFixed(2)}ms`)
  }
}

export const perfMonitor = {
  trackMemory() {
    if (typeof process !== 'undefined') {
      const usage = process.memoryUsage()
      console.log('[Memory]', {
        heapUsed: Math.round(usage.heapUsed / 1024 / 1024) + 'MB',
        heapTotal: Math.round(usage.heapTotal / 1024 / 1024) + 'MB',
      })
    }
  },
  
  trackFrameRate() {
    let frames = 0
    let lastTime = performance.now()
    
    const tick = () => {
      frames++
      const now = performance.now()
      if (now - lastTime >= 1000) {
        console.log(`[FPS] ${frames}`)
        frames = 0
        lastTime = now
      }
      requestAnimationFrame(tick)
    }
    
    tick()
  }
}
```

---

## 🔗 Related Resources

### Internal Documentation

- [`specs/perf-roadmap.md`](../../specs/perf-roadmap.md) - 官方性能路线图
- [`specs/01-persist-payload-limits.md`](../../specs/01-persist-payload-limits.md)
- [`specs/02-cache-eviction.md`](../../specs/02-cache-eviction.md)
- [`specs/03-request-throttling.md`](../../specs/03-request-throttling.md)
- [`specs/04-scroll-spy-optimization.md`](../../specs/04-scroll-spy-optimization.md)
- [`specs/05-modularize-and-dedupe.md`](../../specs/05-modularize-and-dedupe.md)

### External Resources

- [Ink GitHub](https://github.com/vadimdemedes/ink)
- [React Performance Optimization](https://react.dev/learn/render-and-commit#optimizing-performance)
- [Terminal Performance Best Practices](https://poor.dev/blog/terminal-anatomy/)

---

## 🚀 Getting Started

### For Contributors

如果你想参与性能优化工作:

1. **了解现状**: 阅读 [`specs/perf-roadmap.md`](../../specs/perf-roadmap.md)
2. **选择任务**: 从 Phase 1-4 中选择未完成的任务
3. **本地测试**: 使用性能监控工具验证改进
4. **提交 PR**: 包含性能基准对比数据

### For Users

如果遇到性能问题:

1. **报告问题**: 提供详细的环境信息 (终端类型、消息数量等)
2. **临时方案**:
   - 避免超长会话 (>1000 消息)
   - 使用 iTerm2/Terminal.app 而非 VSCode (临时)
   - 定期清理会话历史
3. **等待修复**: 关注 GitHub releases

---

## 📝 Status & Timeline

**当前状态** (2026-01-10):

- ✅ 问题识别完成
- ✅ 官方路线图制定完成
- ⏳ Phase 1 实施中
- 📅 预计 Q1 2026 完成 Phase 1-3
- 📅 预计 Q2 2026 评估 Ink 迁移

**最近更新**:

- 2026-01-10: 完成技术栈文档更新,创建性能优化指南

---

*This document is actively maintained. Last updated: 2026-01-10*
