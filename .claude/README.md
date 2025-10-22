# Claude Code Development Environment for Freightliner

Welcome to the Claude Code-enhanced Freightliner development environment! This directory contains configurations, workflows, and tools to supercharge development with AI assistance.

## 🚀 Quick Start

### Available Slash Commands

Run these commands from the Claude Code interface:

```bash
/add-feature [description]       # Add a new feature with full workflow
/replicate-debug [src] [dst]     # Debug replication issues
/fix-ci [workflow]               # Fix CI/CD pipeline failures
/security-audit [focus]          # Run security audit
/performance-test [scenario]     # Run performance tests
/deploy [env] [version]          # Deploy to Kubernetes
/go-build                        # Build the application
/go-test                         # Run tests
/go-lint                         # Run linting
```

### Development Workflows

#### 1. Feature Development
```bash
/add-feature "Add support for Azure Container Registry (ACR)"
```

This will:
- ✅ Create feature specification
- ✅ Design interfaces following project patterns
- ✅ Implement with proper error handling
- ✅ Add comprehensive tests (80%+ coverage)
- ✅ Security review
- ✅ Code review
- ✅ Performance review
- ✅ Update documentation
- ✅ Ensure CI/CD passes

#### 2. Bug Fixes
```bash
/fix-ci comprehensive-validation
```

Diagnoses and fixes CI failures:
- Analyzes GitHub Actions logs
- Identifies failure patterns
- Implements fixes
- Validates locally before pushing

#### 3. Production Deployment
```bash
/deploy production v1.2.0
```

Handles full deployment:
- Pre-deployment validation
- Helm deployment
- Health checks
- Smoke tests
- Monitoring setup

## 🤖 Specialized Agents

The project is configured with specialized agents for different tasks:

### Development Agents

#### **golang-pro**
Expert in Go development following Freightliner patterns
- Interface-driven design
- Error handling patterns
- Context usage
- Worker pools
- Go best practices

#### **architect-reviewer**
Reviews code for architectural consistency
- SOLID principles
- Interface design
- Package organization
- Dependency management

#### **test-automator**
Creates comprehensive test suites
- Table-driven tests
- Integration tests
- 80%+ coverage
- Mock usage
- Test fixtures

### Operations Agents

#### **deployment-engineer**
Handles deployments and CI/CD
- Kubernetes deployments
- Helm chart management
- CI/CD pipeline fixes
- Docker operations

#### **devops-troubleshooter**
Diagnoses production issues
- Log analysis
- Metrics review
- Incident response
- Root cause analysis

#### **performance-engineer**
Optimizes performance
- Profiling and benchmarking
- Bottleneck identification
- Load testing
- Resource optimization

### Security & Quality Agents

#### **security-auditor**
Reviews security aspects
- Vulnerability scanning
- Authentication/authorization review
- Secrets management
- Container security

#### **code-reviewer**
Reviews code quality
- Code style
- Best practices
- Error handling
- Documentation

## 📚 Skills

Domain-specific knowledge modules:

### **container-registry**
Expert in container registry operations
- ECR and GCR authentication
- Image replication
- Registry debugging
- Multi-architecture images

### **go-microservice**
Go microservice development patterns
- Interface-driven design
- Error handling
- Context usage
- Testing patterns
- HTTP servers

### **kubernetes-ops**
Kubernetes operations
- Deployment strategies
- High availability
- Monitoring setup
- Security best practices

## 🔧 MCP Servers

Model Context Protocol servers for extended capabilities:

### Enabled by Default
- **filesystem**: File system access
- **github**: GitHub API access
- **aws**: AWS services (ECR, KMS, Secrets Manager)
- **google-cloud**: GCP services (GCR, Artifact Registry)
- **kubernetes**: K8s cluster operations
- **docker**: Docker daemon access

### Configure Additional MCPs
Edit `.claude/mcp-config.json` to enable/configure:
- **prometheus**: Metrics access (disabled by default)

## 📊 Status Line

The status line shows real-time project information:
```
🚂 Freightliner | main | ✓ | Go 1.24.5 | ✓ Built
```

Configure in `.claude/statusline.json`

## 🔄 Workflows

Pre-defined multi-agent workflows:

### Feature Development
`.claude/workflows/feature-development.md`
- Complete feature development lifecycle
- Multi-agent coordination
- Quality gates
- Documentation updates

### Incident Response
`.claude/workflows/incident-response.md`
- Severity assessment
- Root cause analysis
- Fix implementation
- Post-incident review

## 📝 Project Documentation

Key files for AI context:

- **CLAUDE.md**: Project guide for Claude Code (root)
- **tech.md**: Technical standards and architecture
- **structure.md**: Project structure overview
- **product.md**: Product vision and goals

## 🎯 Common Tasks

### Start Development
```bash
# Check status
make help

# Build and test
make dev

# Run quality checks
make quality
```

### Debug Issues
```bash
# Debug replication
/replicate-debug ecr/my-repo gcr/my-repo

# Check CI failures
/fix-ci

# Performance issues
/performance-test
```

### Deploy
```bash
# Deploy to staging
/deploy staging v1.2.0

# Deploy to production
/deploy production v1.2.0
```

## 🔐 Security

### Credentials Setup
Required environment variables:
- `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`
- `GOOGLE_APPLICATION_CREDENTIALS`
- `GITHUB_TOKEN` (for MCP)

### Secrets Management
- AWS Secrets Manager integration
- Google Secret Manager integration
- Kubernetes secrets for deployments

## 📖 Documentation

### Code Documentation
- All exported functions have GoDoc comments
- Package-level documentation in README files
- Architecture docs in `docs/ARCHITECTURE.md`

### Operational Documentation
- Deployment guides in `docs/`
- Troubleshooting runbooks
- Security implementation guides
- Performance optimization reports

## 🤝 Multi-Agent Coordination

Agents can work in parallel for efficiency:

```bash
# Run multiple agents in parallel
# This will invoke golang-pro, security-auditor, and test-automator concurrently
/add-feature "New feature" --agents="golang-pro,security-auditor,test-automator"
```

The **context-manager** agent automatically coordinates:
- Context preservation across agents
- Result aggregation
- Conflict resolution
- Final report generation

## 💡 Tips

1. **Use specific commands**: `/add-feature` is better than generic requests
2. **Parallel agents**: Run multiple agents for faster results
3. **Review outputs**: Always review AI-generated code and tests
4. **Incremental changes**: Make small, testable changes
5. **Quality gates**: Ensure all checks pass before merging

## 📚 Additional Resources

- [Go Best Practices](.claude/skills/go-microservice.md)
- [Container Registry Ops](.claude/skills/container-registry.md)
- [Kubernetes Ops](.claude/skills/kubernetes-ops.md)
- [Agent Coordination](.claude/agents/COORDINATION.md)
- [Enhanced System](.claude/agents/ENHANCED_SYSTEM.md)

## 🆘 Getting Help

- Check existing commands: `ls .claude/commands/`
- Review workflows: `ls .claude/workflows/`
- Read skills: `ls .claude/skills/`
- Check agent docs: `ls .claude/agents/`

## 🔄 Updates

This environment is continuously improved. To update:
```bash
git pull origin main
# Review .claude/ directory for new features
```

---

**Happy Coding with Claude!** 🚂✨
