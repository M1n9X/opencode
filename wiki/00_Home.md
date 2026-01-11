# OpenCode - AI Coding Agent Wiki

<p align="center">
  <img src="../packages/identity/logo-ornate-light.svg" alt="opencode logo" width="300">
</p>

<p align="center">AI coding agent, built for the terminal.</p>

---

## 🚀 Quick Navigation

### Getting Started

- [📦 Installation Guide](guides/01_Installation_Guide.md) - Set up opencode in minutes
- [🏃‍♂️ Quick Start](01_Getting_Started.md) - Your first opencode session

### Architecture & Design

- [🏗️ Architecture Overview](02_Architecture_Overview.md) - System design and philosophy
- [💎 Code Highlights](03_Code_Highlights.md) - Elegant implementations and patterns
- [🧪 Testing Strategy](04_Testing_Strategy.md) - How we ensure quality

### Core Components

- [🔧 Core Modules](architecture/01_Core_Modules/) - Deep dive into system components
- [🖥️ TUI Technology Stack](07_TUI_Stack.md) - Terminal UI implementation details
- [🔄 Business Workflows](architecture/02_Business_Workflows.md) - Key processes and data flows
- [🗄️ Data Model](architecture/03_Data_Model.md) - Database design and entities
- [🚀 Performance Optimization](06_Performance_Optimization.md) - Performance issues and solutions

### Deployment & Operations

- [🚀 Deployment Guide](guides/02_Deployment_Guide.md) - Production deployment
- [🤝 Contribution Guide](guides/03_Contribution_Guide.md) - How to contribute

### Risk Assessment

- [🔒 Security Vulnerabilities](risks/01_Security_Vulnerabilities.md) - Security analysis
- [🐛 Known Issues](risks/02_Known_Bugs_And_Issues.md) - Current limitations
- [♻️ Refactoring Suggestions](risks/03_Refactoring_Suggestions.md) - Improvement opportunities

---

## 📋 Project Overview

**OpenCode** is a sophisticated AI-powered coding assistant designed specifically for terminal environments. Built by the team behind SST, it represents a new paradigm in developer tooling that combines the power of modern AI models with the efficiency of command-line interfaces.

### 🎯 Core Mission

OpenCode aims to revolutionize the developer experience by providing:

- **Terminal-native AI assistance** - Built for developers who live in the terminal
- **Provider-agnostic architecture** - Works with Anthropic, OpenAI, Google, and local models
- **Client-server architecture** - Enables remote control and multiple client interfaces
- **Open source transparency** - 100% open source with MIT license

### ⭐ Key Features

- **🤖 Multi-Provider AI Support** - Anthropic Claude, OpenAI GPT, Google Gemini, local models via Ollama
- **🖥️ Rich Terminal UI** - Built with Go and Bubble Tea for responsive terminal experience
- **🔧 Comprehensive Tool System** - File operations, bash execution, LSP integration, web search
- **📡 Client-Server Architecture** - TypeScript server with Go TUI client
- **🔌 Plugin System** - Extensible via plugins and MCP (Model Context Protocol)
- **🔐 Secure Authentication** - OAuth, API keys, and GitHub Copilot integration
- **📊 Session Management** - Persistent sessions with sharing capabilities
- **🌐 Cloud Integration** - Cloudflare Workers, PlanetScale database
- **🎨 GitHub Actions Integration** - Automated PR reviews and issue handling

### 🏗️ Architecture Highlights

OpenCode follows a sophisticated **client-server architecture** with clear separation of concerns:

- **TypeScript Server** (`packages/opencode`) - Core business logic, AI integration, tool execution
- **TypeScript TUI Client** (`packages/opencode/src/cli/cmd/tui`) - Terminal UI built with OpenTUI/SolidJS
- **Cloud Infrastructure** - Cloudflare Workers for web services, PlanetScale for data persistence
- **Plugin Ecosystem** - Extensible via JavaScript plugins and MCP servers

### 🎯 Target Users

- **Terminal-first developers** who prefer command-line workflows
- **Teams seeking AI assistance** without vendor lock-in
- **Open source enthusiasts** who value transparency and customization
- **DevOps engineers** needing scriptable AI interactions

---

## 🚦 Project Status

- **Version**: 0.6.6 (Active Development)
- **License**: MIT
- **Language**: TypeScript (Server, TUI), JavaScript (Web, Plugins)
- **Runtime**: Bun (Primary), Node.js (Compatible)
- **Database**: PlanetScale (MySQL)
- **Cloud**: Cloudflare Workers
- **CI/CD**: GitHub Actions

---

## 📚 Documentation Structure

This wiki is organized into logical sections for different audiences:

- **Users** → Start with [Getting Started](01_Getting_Started.md) and [Installation Guide](guides/01_Installation_Guide.md)
- **Contributors** → Review [Architecture Overview](02_Architecture_Overview.md) and [Contribution Guide](guides/03_Contribution_Guide.md)
- **Architects** → Explore [Core Modules](architecture/01_Core_Modules/) and [Business Workflows](architecture/02_Business_Workflows.md)
- **Security Reviewers** → Check [Security Vulnerabilities](risks/01_Security_Vulnerabilities.md)

---

*Last updated: January 2025*
