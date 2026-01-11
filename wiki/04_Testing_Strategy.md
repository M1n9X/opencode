# 🧪 Testing Strategy & Framework

OpenCode employs a comprehensive testing strategy that ensures reliability, performance, and maintainability across its complex architecture. This document outlines the testing philosophy, frameworks, and best practices used throughout the project.

---

## 🎯 Testing Philosophy

### Core Principles

**1. Test Pyramid Approach**
- **Unit Tests**: Fast, isolated tests for individual components
- **Integration Tests**: Component interaction validation
- **End-to-End Tests**: Full workflow validation

**2. Context-Aware Testing**
- Tests run within realistic project contexts
- Proper isolation between test cases
- Cleanup and resource management

**3. Behavior-Driven Testing**
- Tests focus on expected behavior rather than implementation details
- Clear test descriptions that serve as documentation
- Realistic test scenarios that mirror actual usage

---

## 🏗️ Testing Architecture

### Test Organization Structure

```
packages/opencode/test/
├── bun.test.ts              # Test runner configuration
├── fixtures/                # Test data and mock projects
│   └── example/            # Sample project for testing
├── session/                # Session management tests
│   └── fileRegex.test.ts   # File pattern matching tests
└── tool/                   # Tool system tests
    ├── __snapshots__/      # Snapshot test outputs
    ├── bash.test.ts        # Bash tool tests
    ├── edit.test.ts        # Edit tool tests
    ├── register.test.ts    # Tool registration tests
    └── tool.test.ts        # Core tool framework tests
```

### Testing Framework Stack

**Primary Framework**: Bun Test
- **Why Bun**: Native TypeScript support, fast execution, built-in mocking
- **Performance**: Significantly faster than Jest or Mocha
- **Integration**: Seamless integration with Bun runtime

**Supporting Libraries**:
- **Snapshot Testing**: Built-in Bun snapshot support
- **Mocking**: Bun's native mocking capabilities
- **Assertions**: Standard expect-style assertions

---

## 🔧 Tool Testing Framework

### Core Tool Testing Pattern

**Location**: `packages/opencode/test/tool/tool.test.ts`

```typescript
import { describe, expect, test } from "bun:test"
import { GlobTool } from "../../src/tool/glob"
import { ListTool } from "../../src/tool/ls"
import path from "path"
import { Instance } from "../../src/project/instance"

const ctx = {
  sessionID: "test",
  messageID: "",
  toolCallID: "",
  agent: "build",
  abort: AbortSignal.any([]),
  metadata: () => {},
}

const glob = await GlobTool.init()
const list = await ListTool.init()
const projectRoot = path.join(__dirname, "../..")
const fixturePath = path.join(__dirname, "../fixtures/example")
```

**Key Testing Patterns**:

**1. Context Isolation**:
```typescript
test("basic", async () => {
  await Instance.provide(projectRoot, async () => {
    let result = await glob.execute({
      pattern: "*.json",
      path: undefined,
    }, ctx)
    expect(result.metadata).toMatchObject({
      truncated: false,
      count: 2,
    })
  })
})
```

**2. Metadata Validation**:
```typescript
test("truncate", async () => {
  await Instance.provide(projectRoot, async () => {
    let result = await glob.execute({
      pattern: "**/*",
      path: "../../node_modules",
    }, ctx)
    expect(result.metadata.truncated).toBe(true)
  })
})
```

**3. Snapshot Testing**:
```typescript
test("basic", async () => {
  const result = await Instance.provide(projectRoot, async () => {
    return await list.execute({ path: fixturePath, ignore: [".git"] }, ctx)
  })

  // Normalize paths for consistent snapshots
  const normalizedOutput = result.output.replace(fixturePath, "packages/opencode/test/fixtures/example")
  expect(normalizedOutput).toMatchSnapshot()
})
```

---

## 🛡️ Security Testing

### Bash Tool Security Tests

**Location**: `packages/opencode/test/tool/bash.test.ts`

```typescript
describe("tool.bash", () => {
  test("basic", async () => {
    await Instance.provide(projectRoot, async () => {
      const result = await bash.execute({
        command: "echo 'test'",
        description: "Echo test message",
      }, ctx)
      expect(result.metadata.exit).toBe(0)
      expect(result.metadata.output).toContain("test")
    })
  })

  test("cd ../ should fail outside of project root", async () => {
    await Instance.provide(projectRoot, async () => {
      expect(
        bash.execute({
          command: "cd ../",
          description: "Try to cd to parent directory",
        }, ctx)
      ).rejects.toThrow("This command references paths outside of")
    })
  })
})
```

**Security Test Categories**:

**1. Path Traversal Protection**:
- Tests that commands cannot escape project boundaries
- Validation of relative path restrictions
- Symlink attack prevention

**2. Command Injection Prevention**:
- Input sanitization validation
- Shell escape sequence handling
- Environment variable isolation

**3. Resource Limits**:
- Timeout enforcement testing
- Memory usage constraints
- Process isolation validation

---

## 📊 Performance Testing

### Timing and Benchmarking

**Built-in Performance Monitoring**:
```typescript
export async function read(file: string) {
  using _ = log.time("read", { file })  // Automatic timing
  // ... implementation
}
```

**Performance Test Patterns**:
```typescript
test("performance benchmark", async () => {
  const start = performance.now()
  
  await Instance.provide(projectRoot, async () => {
    for (let i = 0; i < 100; i++) {
      await tool.execute(testInput, ctx)
    }
  })
  
  const duration = performance.now() - start
  expect(duration).toBeLessThan(1000) // Should complete in under 1 second
})
```

---

## 🔄 Integration Testing

### Session Management Tests

**Location**: `packages/opencode/test/session/fileRegex.test.ts`

```typescript
describe("session file regex", () => {
  test("matches expected patterns", async () => {
    const patterns = [
      "**/*.ts",
      "src/**/*.js",
      "!node_modules/**"
    ]
    
    for (const pattern of patterns) {
      const result = await Session.findFiles(pattern)
      expect(result).toBeDefined()
      expect(Array.isArray(result)).toBe(true)
    }
  })
})
```

**Integration Test Scope**:
- **Session Lifecycle**: Creation, updates, persistence, cleanup
- **Tool Coordination**: Multiple tools working together
- **Event System**: Event publishing and subscription
- **Storage Operations**: File system and database interactions

---

## 🎭 Mocking and Fixtures

### Test Fixtures

**Project Structure**: `packages/opencode/test/fixtures/example/`
```
example/
├── package.json
├── src/
│   ├── index.ts
│   └── utils.ts
├── tests/
│   └── example.test.ts
└── README.md
```

**Fixture Usage**:
```typescript
const fixturePath = path.join(__dirname, "../fixtures/example")

test("list directory contents", async () => {
  const result = await Instance.provide(fixturePath, async () => {
    return await listTool.execute({ path: "." }, ctx)
  })
  
  expect(result.output).toContain("package.json")
  expect(result.output).toContain("src/")
})
```

### Mocking Strategies

**AI Provider Mocking**:
```typescript
// Mock AI responses for predictable testing
const mockProvider = {
  chat: jest.fn().mockResolvedValue({
    content: "Mocked AI response",
    usage: { tokens: 100 }
  })
}
```

**File System Mocking**:
```typescript
// Mock file operations for isolated testing
const mockFS = {
  readFile: jest.fn(),
  writeFile: jest.fn(),
  exists: jest.fn()
}
```

---

## 🚀 Continuous Integration Testing

### GitHub Actions Integration

**Test Workflow**:
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - run: bun install
      - run: bun test
      - run: bun run typecheck
```

**Test Categories in CI**:
- **Unit Tests**: All tool and component tests
- **Type Checking**: TypeScript compilation validation
- **Linting**: Code style and quality checks
- **Security Scans**: Dependency vulnerability checks

---

## 📈 Test Coverage Strategy

### Coverage Goals

**Target Coverage Levels**:
- **Critical Paths**: 95%+ coverage (auth, security, data persistence)
- **Business Logic**: 85%+ coverage (tools, sessions, agents)
- **UI Components**: 70%+ coverage (TUI interactions)
- **Integration Points**: 90%+ coverage (API endpoints, external services)

**Coverage Measurement**:
```bash
# Generate coverage report
bun test --coverage

# Coverage thresholds in package.json
{
  "scripts": {
    "test:coverage": "bun test --coverage --coverage-threshold=85"
  }
}
```

---

## 🔍 Test Data Management

### Test Database Strategy

**Isolated Test Environments**:
- Each test suite uses isolated database instances
- Automatic cleanup after test completion
- Seed data for consistent test scenarios

**Test Data Patterns**:
```typescript
beforeEach(async () => {
  await setupTestDatabase()
  await seedTestData()
})

afterEach(async () => {
  await cleanupTestDatabase()
})
```

---

## 🎯 Testing Best Practices

### Code Quality Standards

**1. Test Naming Conventions**:
```typescript
describe("Tool.BashTool", () => {
  describe("when executing safe commands", () => {
    test("should return successful exit code", async () => {
      // Test implementation
    })
  })
  
  describe("when attempting path traversal", () => {
    test("should throw security error", async () => {
      // Security test
    })
  })
})
```

**2. Assertion Patterns**:
```typescript
// Prefer specific assertions
expect(result.metadata.exit).toBe(0)
expect(result.output).toContain("expected content")

// Avoid generic assertions
expect(result).toBeTruthy() // Too vague
```

**3. Test Independence**:
```typescript
// Each test should be independent
test("should work independently", async () => {
  await Instance.provide(testContext, async () => {
    // Test logic that doesn't depend on other tests
  })
})
```

---

## 🔧 Running Tests

### Local Development

```bash
# Run all tests
bun test

# Run specific test file
bun test test/tool/bash.test.ts

# Run tests with coverage
bun test --coverage

# Run tests in watch mode
bun test --watch

# Type checking
bun run typecheck
```

### Test Configuration

**bun.test.ts Configuration**:
```typescript
import { beforeAll, afterAll } from "bun:test"
import { Log } from "./src/util/log"

beforeAll(async () => {
  // Initialize test environment
  await Log.init({ print: false, level: "ERROR" })
})

afterAll(async () => {
  // Cleanup test environment
})
```

---

## 📋 Test Maintenance

### Regular Maintenance Tasks

**1. Snapshot Updates**:
```bash
# Update snapshots when output format changes
bun test --update-snapshots
```

**2. Dependency Updates**:
```bash
# Update test dependencies
bun update
bun test # Verify tests still pass
```

**3. Performance Monitoring**:
```bash
# Monitor test execution time
bun test --reporter=verbose
```

---

## 🎯 Future Testing Enhancements

### Planned Improvements

**1. Visual Regression Testing**:
- TUI screenshot comparison
- Layout consistency validation
- Cross-platform rendering tests

**2. Load Testing**:
- Concurrent session handling
- Memory usage under load
- Performance degradation thresholds

**3. Chaos Engineering**:
- Network failure simulation
- Resource exhaustion testing
- Recovery mechanism validation

---

*This comprehensive testing strategy ensures OpenCode maintains high quality and reliability as it evolves and scales.*
