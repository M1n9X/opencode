# 📦 Installation Guide - Complete Setup Instructions

This comprehensive guide covers all installation methods, configuration options, and troubleshooting steps for OpenCode across different platforms and environments.

---

## 🎯 Installation Overview

### System Requirements

**Minimum Requirements**:
- **OS**: macOS 10.15+, Linux (Ubuntu 18.04+, CentOS 7+), Windows 10+
- **Memory**: 512MB RAM available
- **Storage**: 100MB free disk space
- **Network**: Internet connection for AI provider access

**Recommended Requirements**:
- **OS**: Latest stable versions
- **Memory**: 2GB+ RAM for optimal performance
- **Storage**: 1GB+ for sessions and cache
- **Terminal**: Modern terminal with Unicode support

**Development Requirements** (for contributors):
- **Bun**: 1.2.19+ (primary runtime)
- **Go**: 1.24.x (for TUI development)
- **Node.js**: 18+ (compatibility)
- **Git**: For version control integration

---

## 🚀 Quick Installation Methods

### Method 1: One-Line Install (Recommended)

**Universal Install Script**:
```bash
curl -fsSL https://opencode.ai/install | bash
```

**What it does**:
- Detects your OS and architecture automatically
- Downloads the latest release binary
- Installs to `~/.opencode/bin/` by default
- Adds to PATH in your shell configuration
- Sets appropriate permissions

**Custom Installation Directory**:
```bash
# Install to custom directory
OPENCODE_INSTALL_DIR=/usr/local/bin curl -fsSL https://opencode.ai/install | bash

# XDG Base Directory compliant
XDG_BIN_DIR=$HOME/.local/bin curl -fsSL https://opencode.ai/install | bash

# User bin directory
mkdir -p $HOME/bin
OPENCODE_INSTALL_DIR=$HOME/bin curl -fsSL https://opencode.ai/install | bash
```

### Method 2: Package Managers

**npm/yarn/pnpm/bun**:
```bash
# npm
npm install -g opencode-ai@latest

# yarn
yarn global add opencode-ai@latest

# pnpm
pnpm add -g opencode-ai@latest

# bun
bun add -g opencode-ai@latest
```

**Homebrew (macOS and Linux)**:
```bash
# Add the tap
brew tap sst/tap

# Install opencode
brew install opencode

# Update
brew upgrade opencode
```

**Arch Linux**:
```bash
# Using paru
paru -S opencode-bin

# Using yay
yay -S opencode-bin

# Manual AUR installation
git clone https://aur.archlinux.org/opencode-bin.git
cd opencode-bin
makepkg -si
```

### Method 3: Manual Installation

**Download and Install**:
```bash
# Determine your platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
  x86_64) ARCH="x64" ;;
  aarch64) ARCH="arm64" ;;
  arm64) ARCH="arm64" ;;
esac

# Download
curl -L -o opencode.zip \
  "https://github.com/sst/opencode/releases/latest/download/opencode-${OS}-${ARCH}.zip"

# Extract and install
unzip opencode.zip
chmod +x opencode
mv opencode ~/.local/bin/  # or your preferred bin directory

# Cleanup
rm opencode.zip
```

---

## 🔧 Platform-Specific Instructions

### macOS Installation

**Homebrew (Recommended)**:
```bash
brew install sst/tap/opencode
```

**Manual Installation**:
```bash
# Download for macOS
curl -L -o opencode-macos.zip \
  "https://github.com/sst/opencode/releases/latest/download/opencode-darwin-$(uname -m).zip"

# Install
unzip opencode-macos.zip
chmod +x opencode
mv opencode /usr/local/bin/

# Add to PATH (if not already)
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**macOS Security Note**:
If you get a security warning, run:
```bash
xattr -d com.apple.quarantine /usr/local/bin/opencode
```

### Linux Installation

**Ubuntu/Debian**:
```bash
# Using the install script
curl -fsSL https://opencode.ai/install | bash

# Manual installation
wget https://github.com/sst/opencode/releases/latest/download/opencode-linux-x64.zip
unzip opencode-linux-x64.zip
chmod +x opencode
sudo mv opencode /usr/local/bin/
```

**CentOS/RHEL/Fedora**:
```bash
# Using curl
curl -fsSL https://opencode.ai/install | bash

# Manual with yum/dnf
curl -L -o opencode.zip \
  "https://github.com/sst/opencode/releases/latest/download/opencode-linux-x64.zip"
unzip opencode.zip
chmod +x opencode
sudo mv opencode /usr/local/bin/
```

**Alpine Linux**:
```bash
# Install dependencies
apk add --no-cache curl unzip

# Install opencode
curl -fsSL https://opencode.ai/install | bash
```

### Windows Installation

**PowerShell Installation**:
```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/sst/opencode/releases/latest/download/opencode-windows-x64.zip" -OutFile "opencode.zip"

# Extract
Expand-Archive -Path "opencode.zip" -DestinationPath "."

# Move to a directory in PATH
Move-Item "opencode.exe" "$env:USERPROFILE\bin\"

# Add to PATH if needed
$env:PATH += ";$env:USERPROFILE\bin"
```

**Windows Subsystem for Linux (WSL)**:
```bash
# Use the Linux installation method
curl -fsSL https://opencode.ai/install | bash
```

---

## 🔑 Authentication Setup

### Provider Configuration

**Anthropic Claude (Recommended)**:
```bash
# Start authentication flow
opencode auth login

# Select Anthropic from the list
# Choose "Claude Pro/Max" for browser authentication
# Or "Create API Key" for API key setup
```

**OpenAI Setup**:
```bash
# Get API key from https://platform.openai.com/api-keys
opencode auth login

# Select OpenAI
# Enter your API key when prompted
```

**Google Gemini Setup**:
```bash
# Get API key from https://makersuite.google.com/app/apikey
opencode auth login

# Select Google
# Enter your API key when prompted
```

**Local Models (Ollama)**:
```bash
# Install Ollama first
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama2

# Configure opencode
cat > opencode.json << EOF
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
EOF
```

### Authentication Verification

**Test Authentication**:
```bash
# Check available models
opencode models

# Test with a simple prompt
opencode run "Hello, can you help me with coding?"
```

---

## ⚙️ Configuration

### Basic Configuration

**Create `opencode.json`** in your project root:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-3-5-sonnet-20241022",
  "share": "auto",
  "permission": {
    "edit": "allow",
    "bash": {
      "*": "ask",
      "git": "allow",
      "npm": "allow",
      "yarn": "allow"
    },
    "webfetch": "ask"
  }
}
```

### Advanced Configuration

**Full Configuration Example**:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-3-5-sonnet-20241022",
  "share": "disabled",
  "permission": {
    "edit": "allow",
    "bash": {
      "*": "deny",
      "git": "allow",
      "npm": "allow",
      "yarn": "allow",
      "bun": "allow",
      "ls": "allow",
      "cat": "allow",
      "grep": "allow",
      "find": "allow"
    },
    "webfetch": "ask"
  },
  "lsp": {
    "typescript": {
      "command": ["typescript-language-server", "--stdio"],
      "extensions": [".ts", ".tsx", ".js", ".jsx"],
      "initialization": {
        "preferences": {
          "includeInlayParameterNameHints": "all"
        }
      }
    },
    "python": {
      "command": ["pylsp"],
      "extensions": [".py"],
      "env": {
        "PYTHONPATH": "./src"
      }
    }
  },
  "mcp": {
    "weather": {
      "type": "local",
      "command": ["opencode", "x", "@h1deya/mcp-server-weather"]
    },
    "database": {
      "type": "sse",
      "url": "https://mcp-server.example.com/sse",
      "headers": {
        "Authorization": "Bearer your-token"
      }
    }
  }
}
```

### Environment Variables

**Configuration via Environment**:
```bash
# Set default model
export OPENCODE_MODEL="anthropic/claude-3-5-sonnet-20241022"

# Set data directory
export OPENCODE_DATA_DIR="$HOME/.config/opencode"

# Enable debug logging
export OPENCODE_LOG_LEVEL="DEBUG"

# Disable auto-sharing
export OPENCODE_SHARE="false"
```

---

## 🔧 Development Setup

### For Contributors

**Prerequisites**:
```bash
# Install Bun
curl -fsSL https://bun.sh/install | bash

# Install Go (1.24.x)
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**Clone and Setup**:
```bash
# Clone repository
git clone https://github.com/sst/opencode.git
cd opencode

# Install dependencies
bun install

# Run in development mode
bun dev

# Run tests
bun test

# Type checking
bun run typecheck
```

**TUI Development**:
```bash
# Navigate to TUI package
cd packages/tui

# Install Go dependencies
go mod download

# Build TUI
go build -o opencode cmd/opencode/main.go

# Run TUI in development
./opencode
```

---

## 🔍 Troubleshooting

### Common Issues

**1. Command Not Found**:
```bash
# Check if opencode is in PATH
which opencode

# If not found, add to PATH
echo 'export PATH="$HOME/.opencode/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Or for zsh
echo 'export PATH="$HOME/.opencode/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**2. Permission Denied**:
```bash
# Make executable
chmod +x ~/.opencode/bin/opencode

# Check file permissions
ls -la ~/.opencode/bin/opencode
```

**3. SSL/TLS Errors**:
```bash
# Update certificates (Linux)
sudo apt-get update && sudo apt-get install ca-certificates

# macOS
brew install ca-certificates

# Set certificate bundle (if needed)
export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
```

**4. Authentication Issues**:
```bash
# Clear auth cache
rm ~/.opencode/data/auth.json

# Re-authenticate
opencode auth login

# Check auth status
opencode auth status
```

**5. TUI Display Issues**:
```bash
# Check terminal compatibility
echo $TERM

# Try with different TERM
TERM=xterm-256color opencode

# For tmux users
TERM=screen-256color opencode
```

### Platform-Specific Issues

**macOS Gatekeeper**:
```bash
# If blocked by Gatekeeper
sudo spctl --master-disable  # Temporarily disable
# Or
xattr -d com.apple.quarantine /path/to/opencode
```

**Linux Missing Dependencies**:
```bash
# Ubuntu/Debian
sudo apt-get install libc6 libgcc-s1

# CentOS/RHEL
sudo yum install glibc libgcc
```

**Windows Antivirus**:
- Add opencode.exe to antivirus exclusions
- Run as administrator if needed
- Check Windows Defender SmartScreen settings

### Performance Issues

**Slow Startup**:
```bash
# Check for large session history
ls -la ~/.opencode/data/session/

# Clean old sessions (optional)
find ~/.opencode/data/session/ -mtime +30 -delete

# Disable auto-share for faster startup
export OPENCODE_SHARE=false
```

**High Memory Usage**:
```bash
# Monitor memory usage
opencode debug memory

# Limit concurrent sessions
export OPENCODE_MAX_SESSIONS=5

# Clear cache
rm -rf ~/.opencode/cache/
```

---

## 🔄 Updates and Maintenance

### Updating OpenCode

**Automatic Update Check**:
```bash
# Check for updates
opencode upgrade

# Force update
opencode upgrade --force
```

**Manual Update**:
```bash
# Using package managers
brew upgrade opencode          # Homebrew
npm update -g opencode-ai      # npm
paru -Syu opencode-bin        # Arch Linux

# Using install script
curl -fsSL https://opencode.ai/install | bash
```

### Maintenance Tasks

**Clean Up Old Data**:
```bash
# Remove old sessions (older than 30 days)
find ~/.opencode/data/session/ -mtime +30 -delete

# Clear cache
rm -rf ~/.opencode/cache/

# Compact database (if applicable)
opencode maintenance compact
```

**Backup Configuration**:
```bash
# Backup important data
tar -czf opencode-backup.tar.gz ~/.opencode/data/auth.json opencode.json

# Restore from backup
tar -xzf opencode-backup.tar.gz -C /
```

---

## 🆘 Getting Help

### Support Channels

- **Documentation**: [opencode.ai/docs](https://opencode.ai/docs)
- **Discord Community**: [opencode.ai/discord](https://opencode.ai/discord)
- **GitHub Issues**: [github.com/sst/opencode/issues](https://github.com/sst/opencode/issues)
- **GitHub Discussions**: [github.com/sst/opencode/discussions](https://github.com/sst/opencode/discussions)

### Diagnostic Information

**Generate Debug Report**:
```bash
# System information
opencode debug system

# Configuration check
opencode debug config

# Connection test
opencode debug connection

# Full diagnostic
opencode debug all > opencode-debug.txt
```

---

*This installation guide ensures you can get OpenCode running smoothly on any supported platform with proper configuration and troubleshooting support.*
