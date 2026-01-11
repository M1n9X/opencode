# 🚀 Deployment Guide - Production Infrastructure

This guide covers deploying OpenCode's cloud infrastructure, setting up production environments, and managing scalable deployments for teams and organizations.

---

## 🎯 Deployment Overview

### Architecture Components

OpenCode's production deployment consists of several interconnected services:

```mermaid
graph TB
    subgraph "Edge Layer"
        CF[Cloudflare Workers]
        CDN[Cloudflare CDN]
    end
    
    subgraph "Application Layer"
        API[API Worker]
        AUTH[Auth Worker]
        WEB[Web Interface]
        CONSOLE[Admin Console]
    end
    
    subgraph "Data Layer"
        DB[(PlanetScale MySQL)]
        KV[Cloudflare KV]
        R2[Cloudflare R2]
    end
    
    subgraph "External Services"
        STRIPE[Stripe Billing]
        GITHUB[GitHub App]
        PROVIDERS[AI Providers]
    end
    
    CF --> API
    CF --> AUTH
    CF --> WEB
    CF --> CONSOLE
    
    API --> DB
    AUTH --> KV
    WEB --> R2
    
    API --> STRIPE
    AUTH --> GITHUB
    API --> PROVIDERS
```

### Deployment Strategies

**1. Serverless-First Architecture**
- Cloudflare Workers for compute
- PlanetScale for database
- Automatic scaling and global distribution

**2. Multi-Environment Support**
- Development, staging, and production environments
- Branch-based database deployments
- Environment-specific configurations

**3. Infrastructure as Code**
- SST (Serverless Stack) for infrastructure management
- TypeScript-based configuration
- Automated deployments via GitHub Actions

---

## 🏗️ Infrastructure Setup

### Prerequisites

**Required Accounts**:
- [Cloudflare](https://cloudflare.com) - Workers, KV, R2, DNS
- [PlanetScale](https://planetscale.com) - MySQL database
- [Stripe](https://stripe.com) - Payment processing (optional)
- [GitHub](https://github.com) - Source code and CI/CD

**Required Tools**:
```bash
# Install SST CLI
npm install -g sst

# Install dependencies
bun install

# Verify setup
sst version
```

### Environment Configuration

**1. Environment Variables**:
```bash
# Create .env file
cat > .env << EOF
# Cloudflare
CLOUDFLARE_API_TOKEN=your_cloudflare_token

# PlanetScale
PLANETSCALE_SERVICE_TOKEN=your_planetscale_token
PLANETSCALE_SERVICE_TOKEN_ID=your_token_id

# Stripe (optional)
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# AI Providers
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
XAI_API_KEY=xai-...

# GitHub App
GITHUB_APP_ID=123456
GITHUB_APP_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----..."

# Auth Providers
GITHUB_CLIENT_ID_CONSOLE=your_github_client_id
GITHUB_CLIENT_SECRET_CONSOLE=your_github_client_secret
GOOGLE_CLIENT_ID=your_google_client_id
EOF
```

**2. SST Configuration** (`sst.config.ts`):
```typescript
export default $config({
  app(input) {
    return {
      name: "opencode",
      removal: input?.stage === "production" ? "retain" : "remove",
      protect: ["production"].includes(input?.stage),
      home: "cloudflare",
      providers: {
        stripe: {
          apiKey: process.env.STRIPE_SECRET_KEY,
        },
        planetscale: "0.4.1",
      },
    }
  },
  async run() {
    const { api } = await import("./infra/app.js")
    const { auth } = await import("./infra/cloud.js")
    return {
      api: api.url,
      auth: auth.url,
    }
  },
})
```

### Database Setup

**1. PlanetScale Database Creation**:
```bash
# Create database
pscale database create opencode --region us-east

# Create production branch
pscale branch create opencode production

# Create development branch
pscale branch create opencode development

# Deploy schema
pscale deploy-request create opencode development production
pscale deploy-request deploy opencode <deploy-request-number>
```

**2. Database Schema**:
```sql
-- Users table
CREATE TABLE users (
  id VARCHAR(255) PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255),
  avatar_url TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  subscription_status ENUM('free', 'pro', 'team') DEFAULT 'free',
  subscription_expires_at TIMESTAMP NULL,
  monthly_tokens_used BIGINT DEFAULT 0,
  monthly_tokens_limit BIGINT DEFAULT 100000,
  INDEX idx_email (email),
  INDEX idx_subscription (subscription_status, subscription_expires_at)
);

-- Shared sessions table
CREATE TABLE shared_sessions (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255),
  session_id VARCHAR(255) NOT NULL,
  title VARCHAR(500),
  description TEXT,
  is_public BOOLEAN DEFAULT FALSE,
  view_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_sessions (user_id, created_at),
  INDEX idx_public_sessions (is_public, created_at)
);

-- Auth tokens table
CREATE TABLE auth_tokens (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  provider ENUM('anthropic', 'openai', 'google', 'github-copilot') NOT NULL,
  token_type ENUM('oauth', 'api', 'wellknown') NOT NULL,
  encrypted_token TEXT NOT NULL,
  expires_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY unique_user_provider (user_id, provider)
);
```

---

## 🌐 Cloudflare Workers Deployment

### API Worker Configuration

**Worker Script** (`infra/app.ts`):
```typescript
const GITHUB_APP_ID = new sst.Secret("GITHUB_APP_ID")
const GITHUB_APP_PRIVATE_KEY = new sst.Secret("GITHUB_APP_PRIVATE_KEY")
const bucket = new sst.cloudflare.Bucket("Bucket")

export const api = new sst.cloudflare.Worker("Api", {
  domain: `api.${domain}`,
  handler: "packages/function/src/api.ts",
  environment: {
    WEB_DOMAIN: domain,
  },
  url: true,
  link: [bucket, GITHUB_APP_ID, GITHUB_APP_PRIVATE_KEY],
  transform: {
    worker: (args) => {
      args.logpush = true
      args.bindings = $resolve(args.bindings).apply((bindings) => [
        ...bindings,
        {
          name: "SYNC_SERVER",
          type: "durable_object_namespace",
          className: "SyncServer",
        },
      ])
      args.migrations = {
        oldTag: $app.stage === "production" ? "" : "v1",
        newTag: $app.stage === "production" ? "" : "v1",
      }
    },
  },
})
```

### Auth Worker Configuration

**Auth Service** (`infra/cloud.ts`):
```typescript
const GITHUB_CLIENT_ID_CONSOLE = new sst.Secret("GITHUB_CLIENT_ID_CONSOLE")
const GITHUB_CLIENT_SECRET_CONSOLE = new sst.Secret("GITHUB_CLIENT_SECRET_CONSOLE")
const GOOGLE_CLIENT_ID = new sst.Secret("GOOGLE_CLIENT_ID")
const authStorage = new sst.cloudflare.Kv("AuthStorage")

export const auth = new sst.cloudflare.Worker("AuthApi", {
  domain: `auth.${domain}`,
  handler: "cloud/function/src/auth.ts",
  url: true,
  link: [database, authStorage, GITHUB_CLIENT_ID_CONSOLE, GITHUB_CLIENT_SECRET_CONSOLE, GOOGLE_CLIENT_ID],
})
```

### Web Interface Deployment

**SolidStart Application**:
```typescript
new sst.cloudflare.x.SolidStart("Console", {
  domain,
  path: "cloud/app",
  link: [
    database,
    AUTH_API_URL,
    STRIPE_WEBHOOK_SECRET,
    STRIPE_SECRET_KEY,
    ANTHROPIC_API_KEY,
    XAI_API_KEY,
    BASETEN_API_KEY,
  ],
  environment: {
    VITE_AUTH_URL: auth.url.apply((url) => url!),
  },
  transform: {
    server: {
      transform: {
        worker: {
          placement: { mode: "smart" },
          tailConsumers: logProcessor ? [{ service: logProcessor.nodes.worker.scriptName }] : [],
        },
      },
    },
  },
})
```

---

## 🔧 Deployment Commands

### Development Deployment

**Deploy to Development**:
```bash
# Deploy development environment
sst deploy --stage dev

# View deployment status
sst status --stage dev

# View logs
sst logs --stage dev
```

### Staging Deployment

**Deploy to Staging**:
```bash
# Deploy staging environment
sst deploy --stage staging

# Run integration tests
npm run test:integration --stage staging

# Smoke tests
curl https://api-staging.opencode.ai/health
```

### Production Deployment

**Deploy to Production**:
```bash
# Deploy production (with protection)
sst deploy --stage production

# Verify deployment
sst status --stage production

# Monitor deployment
sst logs --stage production --tail
```

### Rollback Procedures

**Emergency Rollback**:
```bash
# List recent deployments
sst history --stage production

# Rollback to previous version
sst rollback --stage production --version <version-id>

# Verify rollback
curl https://api.opencode.ai/health
```

---

## 🔐 Security Configuration

### SSL/TLS Setup

**Cloudflare SSL Configuration**:
```typescript
// Automatic SSL via Cloudflare
const domain = {
  production: "opencode.ai",
  staging: "staging.opencode.ai",
  dev: "dev.opencode.ai"
}[$app.stage] || `${$app.stage}.opencode.ai`

// SSL settings in Cloudflare dashboard:
// - SSL/TLS encryption mode: Full (strict)
// - Always Use HTTPS: On
// - HTTP Strict Transport Security (HSTS): Enabled
```

### Secrets Management

**Managing Secrets**:
```bash
# Set secrets for production
sst secret set ANTHROPIC_API_KEY "sk-ant-..." --stage production
sst secret set STRIPE_SECRET_KEY "sk_live_..." --stage production
sst secret set GITHUB_APP_PRIVATE_KEY "$(cat private-key.pem)" --stage production

# List secrets
sst secret list --stage production

# Remove secrets
sst secret remove OLD_SECRET --stage production
```

### Access Control

**IAM Configuration**:
```typescript
// Cloudflare API token permissions:
// - Zone:Zone Settings:Edit
// - Zone:Zone:Read
// - Account:Cloudflare Workers:Edit
// - Account:Account Settings:Read

// PlanetScale permissions:
// - Database administrator
// - Branch administrator
```

---

## 📊 Monitoring & Observability

### Logging Configuration

**Structured Logging**:
```typescript
// Log processor for production
let logProcessor
if ($app.stage === "production" || $app.stage === "frank") {
  const HONEYCOMB_API_KEY = new sst.Secret("HONEYCOMB_API_KEY")
  logProcessor = new sst.cloudflare.Worker("LogProcessor", {
    handler: "cloud/function/src/log-processor.ts",
    link: [HONEYCOMB_API_KEY],
  })
}
```

### Health Checks

**Health Monitoring Endpoints**:
```typescript
// API health check
app.get("/health", async (c) => {
  const health = {
    status: "healthy",
    timestamp: new Date().toISOString(),
    version: process.env.VERSION,
    environment: process.env.STAGE,
    database: await checkDatabaseHealth(),
    external: await checkExternalServices(),
  }
  return c.json(health)
})
```

### Performance Monitoring

**Metrics Collection**:
```bash
# Cloudflare Analytics
# - Request volume and latency
# - Error rates and status codes
# - Geographic distribution

# Custom metrics via Workers Analytics Engine
# - Tool execution times
# - AI provider response times
# - User session metrics
```

---

## 🔄 CI/CD Pipeline

### GitHub Actions Workflow

**Deployment Pipeline** (`.github/workflows/deploy.yml`):
```yaml
name: Deploy

on:
  push:
    branches: [main, dev]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - run: bun install
      - run: bun test
      - run: bun run typecheck

  deploy-dev:
    if: github.ref == 'refs/heads/dev'
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - run: bun install
      - run: bunx sst deploy --stage dev
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          PLANETSCALE_SERVICE_TOKEN: ${{ secrets.PLANETSCALE_SERVICE_TOKEN }}

  deploy-production:
    if: github.ref == 'refs/heads/main'
    needs: test
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - run: bun install
      - run: bunx sst deploy --stage production
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          PLANETSCALE_SERVICE_TOKEN: ${{ secrets.PLANETSCALE_SERVICE_TOKEN }}
          STRIPE_SECRET_KEY: ${{ secrets.STRIPE_SECRET_KEY }}
```

### Automated Testing

**Integration Tests**:
```bash
# API endpoint tests
npm run test:api --stage staging

# Database migration tests
npm run test:db --stage staging

# End-to-end tests
npm run test:e2e --stage staging
```

---

## 📈 Scaling Considerations

### Performance Optimization

**Cloudflare Workers Optimization**:
- Use Durable Objects for stateful operations
- Implement caching strategies with KV
- Optimize bundle size and cold start times
- Use Smart Placement for reduced latency

**Database Scaling**:
- PlanetScale automatic scaling
- Read replicas for global distribution
- Connection pooling and query optimization
- Database branching for schema changes

### Cost Management

**Resource Optimization**:
```typescript
// Cost-effective resource allocation
const workerConfig = {
  // CPU time limits
  cpuMs: $app.stage === "production" ? 50000 : 10000,
  
  // Memory limits
  memoryMB: $app.stage === "production" ? 128 : 64,
  
  // Request limits
  requestsPerMinute: $app.stage === "production" ? 1000 : 100,
}
```

---

## 🔧 Maintenance & Operations

### Regular Maintenance Tasks

**Database Maintenance**:
```bash
# Weekly database cleanup
pscale shell opencode production << EOF
DELETE FROM shared_sessions WHERE expires_at < NOW() - INTERVAL 7 DAY;
DELETE FROM auth_tokens WHERE expires_at < NOW();
OPTIMIZE TABLE users, shared_sessions, auth_tokens;
EOF
```

**Log Rotation**:
```bash
# Cloudflare Workers logs are automatically rotated
# Custom log retention via Logpush
```

### Backup Procedures

**Database Backups**:
```bash
# PlanetScale automatic backups (daily)
# Manual backup creation
pscale backup create opencode production

# List backups
pscale backup list opencode production

# Restore from backup
pscale backup restore opencode production <backup-id>
```

### Disaster Recovery

**Recovery Procedures**:
1. **Database Recovery**: Restore from PlanetScale backup
2. **Worker Recovery**: Redeploy from Git repository
3. **DNS Recovery**: Update Cloudflare DNS settings
4. **Certificate Recovery**: Automatic via Cloudflare

---

## 🎯 Best Practices

### Deployment Best Practices

1. **Environment Parity**: Keep dev/staging/prod as similar as possible
2. **Blue-Green Deployments**: Use staging slots for zero-downtime deployments
3. **Feature Flags**: Control feature rollouts with environment variables
4. **Monitoring**: Comprehensive monitoring and alerting
5. **Documentation**: Keep deployment docs up to date

### Security Best Practices

1. **Secrets Rotation**: Regular rotation of API keys and tokens
2. **Access Control**: Principle of least privilege
3. **Audit Logging**: Comprehensive audit trails
4. **Vulnerability Scanning**: Regular security scans
5. **Incident Response**: Clear incident response procedures

---

*This deployment guide ensures reliable, scalable, and secure production deployments of OpenCode's cloud infrastructure.*
