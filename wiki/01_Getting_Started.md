# 🏃‍♂️ Getting Started with OpenCode

Welcome to OpenCode! This guide will get you up and running with your first AI coding session in under 5 minutes.

---

## 📋 Prerequisites

Before installing OpenCode, ensure you have:

- **Operating System**: macOS, Linux, or Windows
- **Terminal**: Any modern terminal (Terminal.app, iTerm2, Windows Terminal, etc.)
- **Internet Connection**: For AI model access and installation
- **Git** (recommended): For project management features

---

## 🚀 Quick Installation

### Option 1: One-Line Install (Recommended)

```bash
# YOLO install - fastest way to get started
curl -fsSL https://opencode.ai/install | bash
```

### Option 2: Package Managers

```bash
# npm/bun/pnpm/yarn
npm i -g opencode-ai@latest

# Homebrew (macOS and Linux)
brew install sst/tap/opencode

# Arch Linux
paru -S opencode-bin
```

### Option 3: Custom Installation Directory

```bash
# Install to custom directory
OPENCODE_INSTALL_DIR=/usr/local/bin curl -fsSL https://opencode.ai/install | bash

# XDG compliant installation
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://opencode.ai/install | bash
```

> **💡 Tip**: Remove versions older than 0.1.x before installing the latest version.

---

## 🔑 Authentication Setup

OpenCode supports multiple AI providers. Choose your preferred option:

### Anthropic Claude (Recommended)

1. **Sign up for Claude Pro/Max** at [anthropic.com](https://www.anthropic.com/news/claude-pro)
2. **Authenticate with OpenCode**:
   ```bash
   opencode auth login
   ```
3. **Select Anthropic** and choose "Claude Pro/Max"
4. **Complete browser authentication** when prompted

### OpenAI

1. **Get API key** from [platform.openai.com](https://platform.openai.com/api-keys)
2. **Add to OpenCode**:
   ```bash
   opencode auth login
   ```
3. **Select OpenAI** and enter your API key

### Local Models (Ollama)

1. **Install Ollama** from [ollama.ai](https://ollama.ai)
2. **Configure in opencode.json**:
   ```json
   {
     "$schema": "https://opencode.ai/config.json",
     "provider": {
       "ollama": {
         "npm": "@ai-sdk/openai-compatible",
         "name": "Ollama (local)",
         "options": {
           "baseURL": "http://localhost:11434/v1"
         },
         "models": {
           "llama2": {
             "name": "Llama 2"
           }
         }
       }
     }
   }
   ```

---

## 🎯 Your First Session

### 1. Start OpenCode

```bash
opencode
```

This launches the Terminal User Interface (TUI) with a rich, interactive experience.

### 2. Select Your Model

Use the `/models` command to choose your AI model:
```
/models
```

### 3. Start Coding!

Try these example prompts:

**File Operations:**
```
Create a simple Python web server using Flask
```

**Code Analysis:**
```
Analyze the security of this authentication function
```

**Bug Fixing:**
```
Fix the memory leak in this C++ code
```

**Documentation:**
```
Generate comprehensive documentation for this API
```

### 4. Explore Tools

OpenCode provides powerful built-in tools:

- **File Operations**: `read`, `write`, `edit`, `list`
- **Search**: `grep`, `glob` for finding files and content
- **Execution**: `bash` for running commands
- **Web**: `webfetch` for accessing online resources
- **LSP Integration**: Real-time code analysis and diagnostics

---

## 🔧 Configuration

### Basic Configuration

Create `opencode.json` in your project root:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-3-5-sonnet-20241022",
  "share": "auto",
  "lsp": {
    "typescript": {
      "command": ["typescript-language-server", "--stdio"],
      "extensions": [".ts", ".tsx", ".js", ".jsx"]
    }
  }
}
```

### Advanced Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-3-5-sonnet-20241022",
  "share": "disabled",
  "permission": {
    "edit": "allow",
    "bash": {
      "npm": "allow",
      "git": "allow",
      "rm": "deny"
    }
  },
  "mcp": {
    "weather": {
      "type": "local",
      "command": ["opencode", "x", "@h1deya/mcp-server-weather"]
    }
  }
}
```

---

## 🎨 TUI Navigation

### Key Bindings

- **Ctrl+C**: Exit OpenCode
- **Enter**: Send message
- **↑/↓**: Navigate message history
- **Tab**: Auto-complete commands
- **Ctrl+L**: Clear screen
- **Ctrl+R**: Refresh session

### Commands

- `/models` - Select AI model
- `/agents` - Switch between agents (general, build, plan)
- `/share` - Share current session
- `/export` - Export session data
- `/help` - Show help information

---

## 🌟 Pro Tips

### 1. Project Context
OpenCode automatically detects Git repositories and provides project-aware assistance:
```bash
cd your-project
opencode
```

### 2. Session Persistence
Sessions are automatically saved and can be resumed:
```bash
opencode --session <session-id>
```

### 3. Agent Specialization
Use specialized agents for different tasks:
- **General**: Research and multi-step tasks
- **Build**: Code generation and modification
- **Plan**: High-level planning and architecture

### 4. Tool Combinations
Combine tools for powerful workflows:
```
First, use `grep` to find all TODO comments, then use `edit` to address each one
```

### 5. Sharing Sessions
Share interesting sessions with your team:
```
/share
```

---

## 🆘 Troubleshooting

### Common Issues

**OpenCode not found after installation:**
```bash
# Reload your shell configuration
source ~/.bashrc  # or ~/.zshrc
```

**Permission denied errors:**
```bash
# Check installation directory permissions
ls -la ~/.opencode/bin/
chmod +x ~/.opencode/bin/opencode
```

**Model authentication failures:**
```bash
# Re-authenticate
opencode auth login
```

**TUI display issues:**
```bash
# Check terminal compatibility
echo $TERM
# Try with different terminal if needed
```

### Getting Help

- **Documentation**: [opencode.ai/docs](https://opencode.ai/docs)
- **Discord Community**: [opencode.ai/discord](https://opencode.ai/discord)
- **GitHub Issues**: [github.com/sst/opencode/issues](https://github.com/sst/opencode/issues)

---

## 🎯 Next Steps

Now that you have OpenCode running:

1. **Explore the [Architecture Overview](02_Architecture_Overview.md)** to understand how it works
2. **Check out [Code Highlights](03_Code_Highlights.md)** for advanced usage patterns
3. **Read the [Contribution Guide](guides/03_Contribution_Guide.md)** if you want to contribute
4. **Review [Security Guidelines](risks/01_Security_Vulnerabilities.md)** for production use

---

*Happy coding with OpenCode! 🚀*
