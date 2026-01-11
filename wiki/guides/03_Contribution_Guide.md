# 🤝 Contribution Guide - Developer Onboarding

Welcome to the OpenCode contributor community! This guide provides everything you need to know to start contributing to OpenCode, from setting up your development environment to submitting your first pull request.

---

## 🎯 Getting Started

### Prerequisites

**Required Tools**:
- **Bun**: 1.2.19+ (primary runtime and package manager)
- **Go**: 1.24.x (for TUI development)
- **Git**: Latest version for version control
- **Node.js**: 18+ (for compatibility testing)

**Recommended Tools**:
- **VS Code**: With TypeScript and Go extensions
- **Docker**: For testing deployment scenarios
- **GitHub CLI**: For streamlined PR workflow

### Development Environment Setup

**1. Fork and Clone**:
```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/opencode.git
cd opencode

# Add upstream remote
git remote add upstream https://github.com/sst/opencode.git
```

**2. Install Dependencies**:
```bash
# Install all dependencies
bun install

# Verify installation
bun run typecheck
bun test
```

**3. Development Commands**:
```bash
# Start development server
bun dev

# Run tests
bun test
bun test --watch  # Watch mode

# Type checking
bun run typecheck
bun run typecheck --watch

# Linting
bun run lint
bun run lint --fix

# Build for production
bun run build
```

**4. TUI Development**:
```bash
# Navigate to TUI package
cd packages/tui

# Install Go dependencies
go mod download

# Build TUI
go build -o opencode cmd/opencode/main.go

# Run TUI in development
./opencode --dev
```

---

## 📋 Contribution Types

### 🐛 Bug Fixes

**Process**:
1. Check existing issues to avoid duplicates
2. Create an issue if one doesn't exist
3. Reference the issue in your PR
4. Include tests that verify the fix

**Example Bug Fix PR**:
```markdown
## Bug Fix: Fix session corruption on concurrent access

Fixes #123

### Problem
Sessions could become corrupted when accessed concurrently from multiple clients.

### Solution
- Added file locking mechanism in storage layer
- Implemented retry logic for lock contention
- Added tests for concurrent access scenarios

### Testing
- [x] Unit tests pass
- [x] Integration tests pass
- [x] Manual testing with concurrent clients
```

### ✨ New Features

**Process**:
1. Discuss the feature in GitHub Discussions first
2. Create a detailed feature proposal
3. Get approval from maintainers
4. Implement with comprehensive tests
5. Update documentation

**Feature Development Checklist**:
- [ ] Feature proposal approved
- [ ] Implementation follows architectural patterns
- [ ] Comprehensive test coverage
- [ ] Documentation updated
- [ ] Performance impact assessed
- [ ] Security implications reviewed

### 🔧 New Tools

**Tool Development Template**:
```typescript
import { Tool } from "../tool.js"
import { z } from "zod"

export const MyTool = Tool.define("my-tool", {
  description: "Clear description of what this tool does",
  parameters: z.object({
    input: z.string().describe("Description of the input parameter"),
    options: z.object({
      flag: z.boolean().optional().describe("Optional flag")
    }).optional()
  }),
  execute: async (args, ctx) => {
    // Validate permissions
    const agent = await Agent.get(ctx.agent)
    await Permission.check("my-tool", agent.permission.myTool)
    
    // Implement tool logic
    const result = await performToolOperation(args.input, args.options)
    
    // Return structured result
    return {
      title: `Processed ${args.input}`,
      metadata: {
        input: args.input,
        timestamp: Date.now(),
        // Tool-specific metadata
      },
      output: result
    }
  }
})
```

### 📚 Documentation

**Documentation Standards**:
- Use clear, concise language
- Include code examples
- Add diagrams for complex concepts
- Keep documentation up-to-date with code changes

**Documentation Types**:
- **API Documentation**: JSDoc comments in code
- **User Guides**: Step-by-step instructions
- **Architecture Docs**: High-level system design
- **Tutorials**: Learning-oriented content

---

## 🏗️ Development Workflow

### Branch Strategy

**Branch Naming Convention**:
```bash
# Feature branches
feature/add-new-tool
feature/improve-performance

# Bug fix branches
fix/session-corruption
fix/memory-leak

# Documentation branches
docs/update-api-reference
docs/add-tutorial

# Refactoring branches
refactor/storage-layer
refactor/error-handling
```

### Commit Guidelines

**Commit Message Format**:
```
type(scope): brief description

Detailed explanation of the change, including:
- Why the change was made
- What was changed
- Any breaking changes or migration notes

Closes #123
```

**Commit Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Example Commits**:
```bash
feat(tools): add file compression tool

Add new tool for compressing and decompressing files using gzip.
Supports both single files and directories with recursive compression.

- Add compression tool implementation
- Add comprehensive tests
- Update tool registry
- Add documentation

Closes #456

fix(storage): prevent session corruption on concurrent access

Add file locking mechanism to prevent race conditions when multiple
clients access the same session simultaneously.

- Implement file locking in storage layer
- Add retry logic for lock contention
- Add tests for concurrent scenarios

Fixes #123
```

### Pull Request Process

**1. Before Creating PR**:
```bash
# Sync with upstream
git fetch upstream
git rebase upstream/main

# Run all checks
bun run typecheck
bun test
bun run lint

# Test your changes
bun dev  # Manual testing
```

**2. PR Template**:
```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Performance impact assessed

## Checklist
- [ ] Code follows the project's style guidelines
- [ ] Self-review of code completed
- [ ] Code is commented, particularly in hard-to-understand areas
- [ ] Corresponding changes to documentation made
- [ ] No new warnings introduced
```

**3. PR Review Process**:
1. Automated checks must pass
2. At least one maintainer review required
3. Address all review feedback
4. Squash commits before merge (if requested)

---

## 🧪 Testing Guidelines

### Test Structure

**Test Organization**:
```
tests/
├── unit/           # Unit tests for individual components
│   ├── tools/     # Tool-specific tests
│   ├── storage/   # Storage layer tests
│   └── utils/     # Utility function tests
├── integration/   # Integration tests
│   ├── api/       # API endpoint tests
│   ├── session/   # Session management tests
│   └── tools/     # Tool integration tests
└── e2e/           # End-to-end tests
    ├── cli/       # CLI workflow tests
    └── tui/       # TUI interaction tests
```

### Writing Tests

**Unit Test Example**:
```typescript
import { describe, it, expect, beforeEach } from 'bun:test'
import { BashTool } from '../src/tool/bash.js'
import { createMockContext } from './helpers/mock-context.js'

describe('BashTool', () => {
  let tool: BashTool
  let mockContext: ToolContext
  
  beforeEach(() => {
    tool = new BashTool()
    mockContext = createMockContext({
      sessionID: 'test-session',
      agent: 'test-agent'
    })
  })
  
  it('should execute simple commands successfully', async () => {
    const result = await tool.execute({
      command: 'echo "hello world"',
      description: 'Test echo command'
    }, mockContext)
    
    expect(result.output.trim()).toBe('hello world')
    expect(result.metadata.exitCode).toBe(0)
    expect(result.status).toBe('success')
  })
  
  it('should handle command failures gracefully', async () => {
    const result = await tool.execute({
      command: 'exit 1',
      description: 'Test failing command'
    }, mockContext)
    
    expect(result.status).toBe('error')
    expect(result.metadata.exitCode).toBe(1)
  })
  
  it('should validate dangerous commands', async () => {
    await expect(tool.execute({
      command: 'rm -rf /',
      description: 'Dangerous command'
    }, mockContext)).rejects.toThrow('Potentially dangerous command')
  })
})
```

**Integration Test Example**:
```typescript
import { describe, it, expect, beforeAll, afterAll } from 'bun:test'
import { TestServer } from './helpers/test-server.js'
import { TestClient } from './helpers/test-client.js'

describe('Session API', () => {
  let server: TestServer
  let client: TestClient
  
  beforeAll(async () => {
    server = new TestServer()
    await server.start()
    client = new TestClient(server.url)
  })
  
  afterAll(async () => {
    await server.stop()
  })
  
  it('should create and retrieve sessions', async () => {
    // Create session
    const createResponse = await client.post('/session', {
      title: 'Test Session'
    })
    
    expect(createResponse.status).toBe(200)
    const session = createResponse.data
    expect(session.title).toBe('Test Session')
    
    // Retrieve session
    const getResponse = await client.get(`/session/${session.id}`)
    expect(getResponse.status).toBe(200)
    expect(getResponse.data).toEqual(session)
  })
})
```

### Test Helpers

**Mock Context Helper**:
```typescript
export function createMockContext(overrides: Partial<ToolContext> = {}): ToolContext {
  return {
    sessionID: 'mock-session',
    messageID: 'mock-message',
    agent: 'mock-agent',
    abort: new AbortController().signal,
    metadata: jest.fn(),
    ...overrides
  }
}
```

---

## 📖 Code Style Guide

### TypeScript Guidelines

**Naming Conventions**:
```typescript
// Use PascalCase for classes and interfaces
class SessionManager {}
interface ToolContext {}

// Use camelCase for functions and variables
const sessionId = 'abc123'
function executeCommand() {}

// Use SCREAMING_SNAKE_CASE for constants
const MAX_RETRY_ATTEMPTS = 3
const DEFAULT_TIMEOUT = 30000

// Use kebab-case for file names
// session-manager.ts
// tool-registry.ts
```

**Type Definitions**:
```typescript
// Prefer interfaces over types for object shapes
interface User {
  id: string
  name: string
  email: string
}

// Use types for unions and computed types
type Status = 'pending' | 'completed' | 'failed'
type UserKeys = keyof User

// Use generics for reusable components
interface Repository<T> {
  find(id: string): Promise<T | null>
  save(entity: T): Promise<void>
}
```

**Function Guidelines**:
```typescript
// Prefer async/await over Promises
async function fetchUser(id: string): Promise<User> {
  const response = await fetch(`/users/${id}`)
  return response.json()
}

// Use explicit return types for public APIs
export async function createSession(
  options: CreateSessionOptions
): Promise<Session> {
  // Implementation
}

// Use function declarations for top-level functions
function validateInput(input: unknown): boolean {
  // Implementation
}

// Use arrow functions for callbacks and short functions
const users = await Promise.all(
  userIds.map(id => fetchUser(id))
)
```

### Go Guidelines (TUI)

**Naming Conventions**:
```go
// Use PascalCase for exported functions and types
func CreateSession() *Session {}
type SessionManager struct {}

// Use camelCase for unexported functions and variables
func validateInput() bool {}
var sessionCache map[string]*Session

// Use ALL_CAPS for constants
const MAX_RETRY_ATTEMPTS = 3
const DEFAULT_TIMEOUT = 30 * time.Second
```

**Error Handling**:
```go
// Always handle errors explicitly
result, err := performOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use custom error types for specific errors
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}
```

---

## 🔍 Code Review Guidelines

### For Authors

**Before Requesting Review**:
- [ ] Code is self-reviewed
- [ ] All tests pass
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] Commit messages are clear

**Responding to Feedback**:
- Address all comments
- Ask for clarification if needed
- Update code based on suggestions
- Re-request review after changes

### For Reviewers

**Review Checklist**:
- [ ] Code correctness and logic
- [ ] Test coverage and quality
- [ ] Performance implications
- [ ] Security considerations
- [ ] Documentation completeness
- [ ] Code style consistency

**Review Guidelines**:
- Be constructive and specific
- Explain the "why" behind suggestions
- Acknowledge good practices
- Focus on the code, not the person
- Use GitHub's suggestion feature for small changes

---

## 🚀 Release Process

### Version Management

**Semantic Versioning**:
- **Major** (1.0.0): Breaking changes
- **Minor** (0.1.0): New features, backward compatible
- **Patch** (0.0.1): Bug fixes, backward compatible

### Release Checklist

**Pre-Release**:
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version bumped
- [ ] Security review completed

**Release**:
- [ ] Create release branch
- [ ] Final testing in staging
- [ ] Create GitHub release
- [ ] Deploy to production
- [ ] Monitor for issues

---

## 🤝 Community Guidelines

### Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Discord**: Real-time community chat
- **Email**: security@opencode.ai for security issues

### Getting Help

**For Contributors**:
- Check existing documentation first
- Search GitHub issues and discussions
- Ask in Discord #contributors channel
- Tag maintainers in GitHub for urgent issues

**For Maintainers**:
- Respond to issues within 48 hours
- Provide constructive feedback on PRs
- Help onboard new contributors
- Maintain project roadmap and priorities

---

## 🎯 Contribution Recognition

### Contributor Levels

**First-time Contributors**:
- Welcome package and mentorship
- Good first issue labels
- Detailed feedback on first PR

**Regular Contributors**:
- Recognition in release notes
- Invitation to contributor Discord channels
- Input on project direction

**Core Contributors**:
- Commit access to repository
- Participation in architectural decisions
- Mentorship responsibilities

### Recognition Programs

- **Contributor of the Month**: Featured in newsletter
- **Annual Contributors**: Special recognition and swag
- **Conference Speakers**: Support for speaking at events

---

*Thank you for contributing to OpenCode! Your efforts help make AI-powered development accessible to everyone.*
