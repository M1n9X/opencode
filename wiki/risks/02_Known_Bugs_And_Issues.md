# 🐛 Known Bugs & Issues - Current Limitations

This document tracks known bugs, limitations, and issues in OpenCode that users and contributors should be aware of. Issues are categorized by severity and component.

---

## 🎯 Issue Classification

### Severity Levels

- 🔴 **Critical**: System crashes, data loss, or complete feature failure
- 🟠 **High**: Major functionality impaired, significant user impact
- 🟡 **Medium**: Minor functionality issues, workarounds available
- 🟢 **Low**: Cosmetic issues, edge cases, minor inconveniences
- 📋 **Enhancement**: Feature requests and improvements

### Component Categories

- **🖥️ TUI**: Terminal User Interface (Go-based)
- **⚙️ Server**: TypeScript server and API
- **🔧 Tools**: Individual tool implementations
- **🔗 Integration**: External service integrations
- **📦 Installation**: Setup and deployment issues
- **🌐 Cloud**: Cloud infrastructure and services

---

## 🔴 Critical Issues

### 1. TUI Crashes on Large Output - 🔴 Critical

**Component**: 🖥️ TUI  
**Status**: Open  
**Reported**: Multiple users  

**Description**:
The TUI crashes when tools return very large outputs (>10MB), causing the entire session to terminate.

**Reproduction Steps**:
1. Run a command that generates large output: `bash "find / -name '*.log' 2>/dev/null"`
2. TUI becomes unresponsive
3. Application crashes with memory allocation error

**Error Message**:
```
runtime: out of memory: cannot allocate 134217728-byte block (536870912 in use)
fatal error: out of memory
```

**Workaround**:
- Use output redirection: `bash "find / -name '*.log' 2>/dev/null | head -1000"`
- Limit command scope to smaller directories

**Root Cause**:
TUI attempts to load entire output into memory before rendering, causing OOM on large responses.

**Proposed Fix**:
Implement streaming output with pagination in the TUI component.

### 2. Session Data Corruption on Concurrent Access - 🔴 Critical

**Component**: ⚙️ Server  
**Status**: Open  
**Reported**: Development team  

**Description**:
Concurrent access to the same session from multiple clients can cause data corruption in session files.

**Reproduction Steps**:
1. Open same session in multiple terminal windows
2. Send messages simultaneously from both clients
3. Session data becomes corrupted or messages are lost

**Error Symptoms**:
- Duplicate message IDs
- Missing message parts
- JSON parsing errors when loading sessions

**Workaround**:
- Use only one client per session
- Avoid concurrent operations on the same session

**Root Cause**:
Lack of proper file locking mechanism in storage layer.

---

## 🟠 High Priority Issues

### 3. LSP Diagnostics Not Updating - 🟠 High

**Component**: 🔧 Tools  
**Status**: Open  
**Reported**: Multiple users  

**Description**:
Language Server Protocol diagnostics don't update after file modifications, showing stale error information.

**Affected Languages**:
- TypeScript/JavaScript
- Python
- Go
- Rust

**Reproduction Steps**:
1. Open a file with syntax errors
2. Fix the errors using the edit tool
3. Run `lsp_diagnostics` - still shows old errors

**Workaround**:
Restart the LSP server: `bash "pkill -f typescript-language-server"`

**Root Cause**:
LSP client doesn't properly notify server of file changes made outside the editor.

### 4. Authentication Token Expiry Not Handled - 🟠 High

**Component**: 🔗 Integration  
**Status**: Open  
**Reported**: Production users  

**Description**:
When AI provider tokens expire, the system doesn't gracefully handle re-authentication, causing all requests to fail.

**Error Message**:
```
Authentication failed: Token expired
Please re-authenticate with: opencode auth login
```

**Impact**:
- All AI interactions fail
- No automatic token refresh
- User must manually re-authenticate

**Workaround**:
Run `opencode auth login` to refresh tokens.

**Proposed Fix**:
Implement automatic token refresh with fallback to user prompt.

### 5. File Watcher Memory Leak - 🟠 High

**Component**: ⚙️ Server  
**Status**: Open  
**Reported**: Long-running sessions  

**Description**:
File watchers for LSP integration accumulate over time, causing memory usage to grow continuously.

**Symptoms**:
- Memory usage increases over time
- System becomes slow after extended use
- Eventually leads to OOM errors

**Reproduction**:
- Run OpenCode for several hours with active file editing
- Monitor memory usage - shows continuous growth

**Workaround**:
Restart OpenCode periodically to clear accumulated watchers.

---

## 🟡 Medium Priority Issues

### 6. Unicode Handling in File Paths - 🟡 Medium

**Component**: 🔧 Tools  
**Status**: Open  
**Reported**: International users  

**Description**:
File operations fail on paths containing Unicode characters (Chinese, Japanese, Arabic, etc.).

**Affected Tools**:
- `read`, `write`, `edit`
- `bash` (when dealing with Unicode filenames)

**Error Example**:
```
Error: ENOENT: no such file or directory, open '/path/to/文件.txt'
```

**Workaround**:
Use ASCII-only file names when possible.

### 7. Git Integration Issues with Submodules - 🟡 Medium

**Component**: 🔧 Tools  
**Status**: Open  
**Reported**: Enterprise users  

**Description**:
Git-related operations don't properly handle repositories with submodules.

**Issues**:
- Project detection fails in submodule directories
- Git diff doesn't show submodule changes
- File operations may access files outside main repository

**Workaround**:
Run OpenCode from the main repository root.

### 8. Terminal Resize Handling - 🟡 Medium

**Component**: 🖥️ TUI  
**Status**: Open  
**Reported**: Multiple users  

**Description**:
TUI doesn't properly handle terminal window resizing, causing display corruption.

**Symptoms**:
- Text wrapping issues after resize
- UI elements positioned incorrectly
- Scrolling behavior becomes erratic

**Workaround**:
Restart OpenCode after significant terminal size changes.

### 9. Slow Performance on Large Repositories - 🟡 Medium

**Component**: ⚙️ Server  
**Status**: Open  
**Reported**: Enterprise users  

**Description**:
Initial project loading and file operations are slow in repositories with >100k files.

**Performance Issues**:
- Project initialization takes >30 seconds
- File globbing operations timeout
- LSP startup is delayed

**Workaround**:
- Use `.opencodeignore` to exclude large directories
- Work in smaller subdirectories when possible

---

## 🟢 Low Priority Issues

### 10. Inconsistent Error Messages - 🟢 Low

**Component**: Multiple  
**Status**: Open  
**Reported**: UX feedback  

**Description**:
Error messages vary in format and helpfulness across different components.

**Examples**:
- Some errors show stack traces, others don't
- Inconsistent error codes
- Missing suggestions for resolution

**Impact**: Minor UX degradation

### 11. Tab Completion Edge Cases - 🟢 Low

**Component**: 🖥️ TUI  
**Status**: Open  
**Reported**: Power users  

**Description**:
Tab completion doesn't work correctly in some edge cases:
- Files with spaces in names
- Hidden files (starting with .)
- Symlinked directories

**Workaround**: Type full paths manually

### 12. Log File Rotation Missing - 🟢 Low

**Component**: ⚙️ Server  
**Status**: Open  
**Reported**: Long-running deployments  

**Description**:
Log files grow indefinitely without rotation, potentially filling disk space.

**Impact**: Disk space usage over time

**Workaround**: Manually clean log files periodically

---

## 📋 Enhancement Requests

### 13. Multi-Language Support in TUI - 📋 Enhancement

**Component**: 🖥️ TUI  
**Status**: Planned  
**Requested by**: International users  

**Description**:
Add support for non-English languages in the TUI interface.

**Requirements**:
- Internationalization (i18n) framework
- Translation files for major languages
- RTL language support

### 14. Plugin System for Custom Tools - 📋 Enhancement

**Component**: 🔧 Tools  
**Status**: Under consideration  
**Requested by**: Enterprise users  

**Description**:
Allow users to create and install custom tools without modifying core code.

**Requirements**:
- Plugin API specification
- Sandboxed execution environment
- Plugin marketplace/registry

### 15. Visual Diff Tool - 📋 Enhancement

**Component**: 🔧 Tools  
**Status**: Planned  
**Requested by**: Multiple users  

**Description**:
Add a tool for visual diff comparison of files with syntax highlighting.

**Features**:
- Side-by-side diff view
- Syntax highlighting
- Word-level diff highlighting
- Integration with git diff

---

## 🔧 Platform-Specific Issues

### macOS Issues

**16. Gatekeeper Warnings - 🟡 Medium**
- Unsigned binary triggers security warnings
- Users must manually allow execution
- **Workaround**: Code signing in future releases

**17. Terminal.app Compatibility - 🟢 Low**
- Some Unicode characters don't render correctly
- **Workaround**: Use iTerm2 or other modern terminals

### Linux Issues

**18. Wayland Clipboard Access - 🟡 Medium**
- Clipboard operations fail on Wayland
- **Workaround**: Use X11 session or manual copy/paste

**19. Alpine Linux Compatibility - 🟢 Low**
- Binary doesn't work on Alpine Linux (musl libc)
- **Workaround**: Use Docker or different distribution

### Windows Issues

**20. PowerShell Integration - 🟡 Medium**
- Bash tool doesn't work with PowerShell commands
- **Workaround**: Use WSL or Git Bash

**21. Path Separator Issues - 🟢 Low**
- Mixed use of `/` and `\` in file paths
- **Workaround**: Use forward slashes consistently

---

## 🔍 Debugging Information

### Collecting Debug Information

**For Bug Reports, Please Include**:

```bash
# System information
opencode debug system

# Configuration
opencode debug config

# Recent logs
opencode debug logs --lines 100

# Session information (if applicable)
opencode debug session <session-id>
```

### Common Debug Commands

```bash
# Check OpenCode version
opencode --version

# Verify installation
opencode doctor

# Test authentication
opencode auth status

# Check available models
opencode models

# Validate configuration
opencode config validate
```

### Log Locations

**Default Log Paths**:
- **macOS**: `~/Library/Logs/opencode/`
- **Linux**: `~/.local/share/opencode/logs/`
- **Windows**: `%APPDATA%/opencode/logs/`

**Log Levels**:
- `ERROR`: Critical errors and failures
- `WARN`: Warnings and recoverable issues
- `INFO`: General information and status
- `DEBUG`: Detailed debugging information

---

## 📊 Issue Tracking

### Reporting New Issues

**Before Reporting**:
1. Check this document for known issues
2. Search existing GitHub issues
3. Try the suggested workarounds
4. Collect debug information

**Report Issues At**:
- **GitHub Issues**: [github.com/sst/opencode/issues](https://github.com/sst/opencode/issues)
- **Discord Community**: [opencode.ai/discord](https://opencode.ai/discord)

**Issue Template**:
```markdown
## Bug Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [macOS/Linux/Windows]
- OpenCode Version: [version]
- Terminal: [terminal name and version]

## Debug Information
[Paste output of `opencode debug system`]

## Additional Context
Any other relevant information
```

### Issue Prioritization

**Priority Factors**:
- **User Impact**: How many users are affected
- **Severity**: How badly it affects functionality
- **Workaround Availability**: Whether there's a viable workaround
- **Implementation Complexity**: How difficult it is to fix

---

*This document is regularly updated as new issues are discovered and existing issues are resolved. Check back frequently for the latest information.*
