# 🔒 Security Vulnerabilities & Risk Assessment

This document provides a comprehensive security analysis of OpenCode, identifying potential vulnerabilities, attack vectors, and mitigation strategies based on OWASP Top 10 and industry security standards.

---

## 🎯 Security Assessment Overview

### Risk Classification

**Risk Levels**:
- 🔴 **Critical**: Immediate security threat requiring urgent attention
- 🟠 **High**: Significant security risk that should be addressed soon
- 🟡 **Medium**: Moderate risk that should be planned for remediation
- 🟢 **Low**: Minor risk with acceptable current mitigation
- ✅ **Mitigated**: Risk properly addressed with current controls

### Assessment Methodology

This analysis follows:
- **OWASP Top 10 2021** security risks
- **NIST Cybersecurity Framework** guidelines
- **Static code analysis** of the codebase
- **Architecture review** of system components
- **Threat modeling** of attack vectors

---

## 🚨 Critical Security Risks

### 1. Command Injection (Bash Tool) - 🔴 Critical

**Location**: `packages/opencode/src/tool/bash.ts`

**Vulnerability Description**:
The Bash tool executes user-provided commands with limited input validation, potentially allowing command injection attacks.

**Current Code**:
```typescript
const proc = Bun.spawn({
  cmd: ["bash", "-c", args.command],  // Direct command execution
  cwd: Instance.directory,
  stdout: "pipe",
  stderr: "pipe",
  env: { ...process.env, HOME: Instance.directory, PWD: Instance.directory }
})
```

**Attack Vectors**:
- **Command Chaining**: `ls; rm -rf /`
- **Command Substitution**: `ls $(rm important.txt)`
- **Environment Manipulation**: Commands that modify PATH or other env vars
- **Privilege Escalation**: Commands that attempt to gain higher privileges

**Risk Impact**:
- Complete system compromise
- Data destruction or theft
- Unauthorized access to sensitive files
- Lateral movement within the system

**Mitigation Strategies**:
```typescript
// Enhanced command validation
const validateCommand = (command: string): void => {
  const dangerous = [
    /;\s*rm\s+-rf/,           // Destructive rm commands
    /\$\([^)]*\)/,            // Command substitution
    /`[^`]*`/,                // Backtick execution
    /\|\s*sh/,                // Pipe to shell
    />\s*\/dev\/null\s*2>&1/, // Output redirection
    /sudo|su\s/,              // Privilege escalation
    /curl.*\|\s*bash/,        // Download and execute
  ]
  
  for (const pattern of dangerous) {
    if (pattern.test(command)) {
      throw new Error(`Potentially dangerous command pattern detected: ${pattern}`)
    }
  }
}

// Whitelist approach for allowed commands
const allowedCommands = new Set([
  'ls', 'cat', 'grep', 'find', 'git', 'npm', 'yarn', 'bun',
  'node', 'python', 'go', 'cargo', 'make', 'docker'
])

const validateCommandWhitelist = (command: string): void => {
  const baseCommand = command.split(' ')[0]
  if (!allowedCommands.has(baseCommand)) {
    throw new Error(`Command not in whitelist: ${baseCommand}`)
  }
}
```

### 2. Path Traversal Vulnerabilities - 🔴 Critical

**Location**: Multiple file operation tools

**Vulnerability Description**:
Insufficient path validation could allow access to files outside the project directory.

**Current Validation**:
```typescript
// Insufficient validation in some tools
if (args.command.includes('..')) {
  throw new Error("Command references paths outside of project")
}
```

**Attack Vectors**:
- **Directory Traversal**: `../../../etc/passwd`
- **Symlink Attacks**: Creating symlinks to sensitive files
- **Encoded Paths**: URL-encoded or Unicode-encoded path traversal
- **Windows Path Separators**: Using `\` instead of `/` on Unix systems

**Enhanced Mitigation**:
```typescript
import path from 'path'

const validatePath = (filePath: string, baseDir: string): string => {
  // Resolve and normalize the path
  const resolved = path.resolve(baseDir, filePath)
  const normalized = path.normalize(resolved)
  
  // Ensure the path is within the base directory
  if (!normalized.startsWith(path.resolve(baseDir))) {
    throw new Error(`Path traversal attempt detected: ${filePath}`)
  }
  
  // Check for symlink attacks
  const stats = fs.lstatSync(normalized, { throwIfNoEntry: false })
  if (stats?.isSymbolicLink()) {
    const target = fs.readlinkSync(normalized)
    const resolvedTarget = path.resolve(path.dirname(normalized), target)
    if (!resolvedTarget.startsWith(path.resolve(baseDir))) {
      throw new Error(`Symlink points outside base directory: ${filePath}`)
    }
  }
  
  return normalized
}
```

---

## 🟠 High Security Risks

### 3. Authentication Token Exposure - 🟠 High

**Location**: `packages/opencode/src/auth/index.ts`

**Vulnerability Description**:
Authentication tokens are stored in plaintext files with basic file permissions.

**Current Implementation**:
```typescript
export async function set(key: string, info: Info) {
  const file = Bun.file(filepath)
  const data = await all()
  await Bun.write(file, JSON.stringify({ ...data, [key]: info }, null, 2))
  await fs.chmod(file.name!, 0o600) // Only owner read/write
}
```

**Risks**:
- Tokens readable by other processes running as same user
- No encryption at rest
- Potential token leakage in logs or error messages
- No token rotation mechanism

**Enhanced Security**:
```typescript
import crypto from 'crypto'

class SecureTokenStorage {
  private encryptionKey: Buffer
  
  constructor() {
    // Derive key from system-specific data
    this.encryptionKey = crypto.scryptSync(
      process.env.USER + process.platform,
      'opencode-salt',
      32
    )
  }
  
  encrypt(data: string): string {
    const iv = crypto.randomBytes(16)
    const cipher = crypto.createCipher('aes-256-gcm', this.encryptionKey)
    cipher.setAAD(Buffer.from('opencode-auth'))
    
    let encrypted = cipher.update(data, 'utf8', 'hex')
    encrypted += cipher.final('hex')
    
    const authTag = cipher.getAuthTag()
    return iv.toString('hex') + ':' + authTag.toString('hex') + ':' + encrypted
  }
  
  decrypt(encryptedData: string): string {
    const [ivHex, authTagHex, encrypted] = encryptedData.split(':')
    const iv = Buffer.from(ivHex, 'hex')
    const authTag = Buffer.from(authTagHex, 'hex')
    
    const decipher = crypto.createDecipher('aes-256-gcm', this.encryptionKey)
    decipher.setAAD(Buffer.from('opencode-auth'))
    decipher.setAuthTag(authTag)
    
    let decrypted = decipher.update(encrypted, 'hex', 'utf8')
    decrypted += decipher.final('utf8')
    
    return decrypted
  }
}
```

### 4. Server-Side Request Forgery (SSRF) - 🟠 High

**Location**: `packages/opencode/src/tool/webfetch.ts`

**Vulnerability Description**:
The WebFetch tool can make requests to arbitrary URLs, potentially accessing internal services.

**Attack Vectors**:
- **Internal Service Access**: `http://localhost:8080/admin`
- **Cloud Metadata Access**: `http://169.254.169.254/metadata`
- **File System Access**: `file:///etc/passwd`
- **Port Scanning**: Probing internal network services

**Mitigation Strategy**:
```typescript
const validateUrl = (url: string): void => {
  const parsed = new URL(url)
  
  // Block dangerous protocols
  const allowedProtocols = ['http:', 'https:']
  if (!allowedProtocols.includes(parsed.protocol)) {
    throw new Error(`Protocol not allowed: ${parsed.protocol}`)
  }
  
  // Block private IP ranges
  const hostname = parsed.hostname
  const privateRanges = [
    /^127\./,                    // Loopback
    /^10\./,                     // Private Class A
    /^172\.(1[6-9]|2[0-9]|3[01])\./, // Private Class B
    /^192\.168\./,               // Private Class C
    /^169\.254\./,               // Link-local
    /^::1$/,                     // IPv6 loopback
    /^fc00:/,                    // IPv6 private
  ]
  
  for (const range of privateRanges) {
    if (range.test(hostname)) {
      throw new Error(`Private IP address not allowed: ${hostname}`)
    }
  }
  
  // Block cloud metadata endpoints
  const blockedHosts = [
    '169.254.169.254',           // AWS/GCP metadata
    'metadata.google.internal',  // GCP metadata
    '100.100.100.200',          // Alibaba Cloud
  ]
  
  if (blockedHosts.includes(hostname)) {
    throw new Error(`Blocked hostname: ${hostname}`)
  }
}
```

---

## 🟡 Medium Security Risks

### 5. Information Disclosure - 🟡 Medium

**Location**: Error handling and logging throughout the application

**Vulnerability Description**:
Detailed error messages and stack traces may leak sensitive information.

**Current Issues**:
```typescript
// Potentially leaking sensitive information
.onError((err, c) => {
  log.error("failed", { error: err }) // Full error object logged
  if (err instanceof NamedError) {
    return c.json(err.toObject(), { status: 400 })
  }
  return c.json(new NamedError.Unknown({ message: err.toString() }).toObject(), {
    status: 400,
  })
})
```

**Secure Error Handling**:
```typescript
const sanitizeError = (error: Error, isDevelopment: boolean) => {
  if (isDevelopment) {
    return {
      message: error.message,
      stack: error.stack,
      name: error.name
    }
  }
  
  // Production: Generic error messages
  const safeErrors = {
    'ValidationError': 'Invalid input provided',
    'AuthenticationError': 'Authentication failed',
    'AuthorizationError': 'Access denied',
    'NotFoundError': 'Resource not found'
  }
  
  return {
    message: safeErrors[error.name] || 'An error occurred',
    code: 'GENERIC_ERROR'
  }
}
```

### 6. Session Management Issues - 🟡 Medium

**Location**: Session handling in server and storage

**Vulnerability Description**:
Sessions may not have proper timeout, invalidation, or secure storage.

**Current Implementation Gaps**:
- No automatic session expiration
- No session invalidation on security events
- Session data stored without integrity checks

**Enhanced Session Security**:
```typescript
interface SecureSession {
  id: string
  userId?: string
  createdAt: number
  lastAccessedAt: number
  expiresAt: number
  ipAddress: string
  userAgent: string
  isActive: boolean
  securityFlags: {
    requiresReauth: boolean
    suspiciousActivity: boolean
  }
}

const validateSession = (session: SecureSession): boolean => {
  // Check expiration
  if (Date.now() > session.expiresAt) {
    return false
  }
  
  // Check for suspicious activity
  if (session.securityFlags.suspiciousActivity) {
    return false
  }
  
  // Update last accessed time
  session.lastAccessedAt = Date.now()
  
  return true
}
```

---

## 🟢 Low Security Risks

### 7. Dependency Vulnerabilities - 🟢 Low

**Current Status**: Regular dependency updates and security scanning

**Mitigation**:
- Automated dependency scanning via GitHub Dependabot
- Regular updates of critical dependencies
- Use of `npm audit` and `bun audit` for vulnerability detection

**Recommendations**:
```bash
# Regular security audits
bun audit
npm audit --audit-level high

# Update dependencies
bun update
npm update

# Check for outdated packages
bun outdated
npm outdated
```

### 8. Rate Limiting - 🟢 Low

**Current Status**: Basic rate limiting via Cloudflare

**Enhancement Opportunities**:
```typescript
// Application-level rate limiting
const rateLimiter = new Map<string, { count: number; resetTime: number }>()

const checkRateLimit = (identifier: string, limit: number, windowMs: number): boolean => {
  const now = Date.now()
  const record = rateLimiter.get(identifier)
  
  if (!record || now > record.resetTime) {
    rateLimiter.set(identifier, { count: 1, resetTime: now + windowMs })
    return true
  }
  
  if (record.count >= limit) {
    return false
  }
  
  record.count++
  return true
}
```

---

## ✅ Well-Mitigated Risks

### 9. SQL Injection - ✅ Mitigated

**Status**: Well protected through ORM usage and parameterized queries

**Current Protection**:
- Use of PlanetScale with prepared statements
- No direct SQL construction from user input
- Proper input validation and sanitization

### 10. Cross-Site Scripting (XSS) - ✅ Mitigated

**Status**: Limited risk due to terminal-based interface

**Current Protection**:
- Terminal UI doesn't render HTML
- Server-side rendering with proper escaping
- Content Security Policy headers

---

## 🛡️ Security Recommendations

### Immediate Actions (Critical/High Risks)

1. **Implement Command Whitelist**: Restrict bash tool to approved commands only
2. **Enhanced Path Validation**: Implement comprehensive path traversal protection
3. **Token Encryption**: Encrypt authentication tokens at rest
4. **SSRF Protection**: Add URL validation and IP filtering to WebFetch tool

### Short-term Improvements (Medium Risks)

1. **Error Handling**: Implement secure error handling with sanitized messages
2. **Session Security**: Add session timeout and security monitoring
3. **Input Validation**: Comprehensive input validation across all tools
4. **Audit Logging**: Enhanced security event logging

### Long-term Enhancements (Low Risks)

1. **Security Headers**: Implement comprehensive security headers
2. **Content Security Policy**: Strict CSP for web interfaces
3. **Penetration Testing**: Regular security assessments
4. **Security Training**: Developer security awareness programs

---

## 🔍 Security Monitoring

### Security Metrics

**Key Indicators**:
- Failed authentication attempts
- Command execution patterns
- File access outside project boundaries
- Unusual network requests
- Error rate spikes

**Monitoring Implementation**:
```typescript
const securityMetrics = {
  failedAuth: 0,
  suspiciousCommands: 0,
  pathTraversalAttempts: 0,
  blockedRequests: 0
}

const logSecurityEvent = (event: string, details: any) => {
  log.warn('security.event', {
    event,
    timestamp: Date.now(),
    sessionId: details.sessionId,
    userId: details.userId,
    details: sanitizeForLogging(details)
  })
  
  securityMetrics[event] = (securityMetrics[event] || 0) + 1
}
```

### Incident Response

**Response Procedures**:
1. **Detection**: Automated monitoring and alerting
2. **Assessment**: Rapid security impact analysis
3. **Containment**: Immediate threat isolation
4. **Eradication**: Remove security threats
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Post-incident analysis

---

## 📋 Security Checklist

### Development Security

- [ ] Input validation on all user inputs
- [ ] Output encoding for all dynamic content
- [ ] Secure authentication and session management
- [ ] Proper error handling without information leakage
- [ ] Secure communication (HTTPS/TLS)
- [ ] Access control and authorization checks
- [ ] Security logging and monitoring

### Deployment Security

- [ ] Secure configuration management
- [ ] Regular security updates
- [ ] Network security controls
- [ ] Backup and recovery procedures
- [ ] Incident response plan
- [ ] Security monitoring and alerting
- [ ] Regular security assessments

---

*This security assessment provides a foundation for maintaining and improving OpenCode's security posture. Regular reviews and updates of this analysis are recommended as the system evolves.*
