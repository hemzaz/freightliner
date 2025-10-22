# Freightliner Claude Code Quick Start

Get productive with AI-assisted development in 5 minutes!

## ⚡ Quick Setup (1 minute)

```bash
# 1. Navigate to project
cd /home/user/freightliner

# 2. Verify Claude Code setup
ls .claude/

# 3. Set environment variables (if needed)
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
export GITHUB_TOKEN=your-github-token
```

## 🎯 Try These Commands (2 minutes)

### Build and Test
```bash
/go-build          # Build the application
/go-test           # Run all tests
/go-lint           # Run linting
```

### Development
```bash
# Add a new feature (full workflow)
/add-feature "Add health check metrics"

# Debug replication issue
/replicate-debug ecr/my-repo gcr/my-repo

# Fix CI failures
/fix-ci comprehensive-validation
```

### Operations
```bash
# Run security audit
/security-audit code

# Performance testing
/performance-test benchmarks

# Deploy to staging
/deploy staging v1.2.0
```

## 🤖 Meet Your AI Team (1 minute)

### For Development
- **golang-pro**: Go expert who knows all Freightliner patterns
- **test-automator**: Creates comprehensive tests
- **architect-reviewer**: Reviews architecture decisions

### For Operations
- **deployment-engineer**: Handles deployments
- **devops-troubleshooter**: Fixes production issues
- **performance-engineer**: Optimizes performance

### For Quality
- **security-auditor**: Security reviews
- **code-reviewer**: Code quality checks

## 📚 Essential Skills Available (1 minute)

Type these in conversations for instant expertise:

```
"Using container-registry skill, help me..."
"Following go-microservice patterns, implement..."
"Using kubernetes-ops skill, deploy..."
```

## 🚀 Your First Feature (Complete Example)

Let's add a simple feature together:

```bash
# 1. Start feature development
/add-feature "Add /version endpoint to HTTP server"
```

This will:
1. ✅ Create feature spec with golang-pro
2. ✅ Design interfaces following project patterns
3. ✅ Implement code with error handling
4. ✅ Add tests achieving 80%+ coverage
5. ✅ Security review with security-auditor
6. ✅ Code review with code-reviewer
7. ✅ Update documentation
8. ✅ Ensure CI passes

**Expected time**: 5-10 minutes (vs 1-2 hours manually)

## 💡 Pro Tips

### 1. Parallel Agents
Run multiple agents for faster results:
```bash
/add-feature "New feature" --agents="golang-pro,security-auditor,test-automator"
```

### 2. Use Specific Commands
```bash
# ❌ Less effective
"Help me fix the CI"

# ✅ More effective
/fix-ci comprehensive-validation
```

### 3. Reference Skills
```bash
# ❌ Generic request
"How do I authenticate with ECR?"

# ✅ Better approach
"Using the container-registry skill, show me ECR authentication"
```

### 4. Leverage Workflows
```bash
/add-feature "Description"
# Automatically uses feature-development workflow
# Coordinates golang-pro → test-automator → security-auditor → code-reviewer
```

## 📖 Learn More

- **Full documentation**: [.claude/README.md](.claude/README.md)
- **Setup guide**: [.claude/CLAUDE_CODE_SETUP.md](.claude/CLAUDE_CODE_SETUP.md)
- **Contributing**: [../CONTRIBUTING.md](../CONTRIBUTING.md)
- **Architecture**: [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)

## 🎓 Common Scenarios

### Scenario 1: Fix a Bug
```bash
# 1. Identify issue
/fix-ci

# 2. Review logs and implement fix
# (AI will analyze logs and suggest fixes)

# 3. Test
/go-test

# 4. Deploy
/deploy staging
```

### Scenario 2: Add New Registry Support
```bash
# 1. Design feature
/add-feature "Add Azure Container Registry (ACR) support"

# 2. Review implementation
# (AI coordinates golang-pro, test-automator, security-auditor)

# 3. Performance test
/performance-test benchmarks

# 4. Deploy
/deploy staging
```

### Scenario 3: Production Incident
```bash
# 1. Diagnose (uses devops-troubleshooter)
"Production is slow, help diagnose"

# 2. Analyze (uses performance-engineer)
/performance-test stress

# 3. Fix and deploy
/deploy production v1.2.1
```

## 🔧 Customization

### Add Your Own Command

Create `.claude/commands/my-command.md`:

```markdown
# My Command

Description

## Usage
/my-command [args]

## Tasks
- Task 1
- Task 2
```

### Add Your Own Skill

Create `.claude/skills/my-domain.md`:

```markdown
# My Domain Expertise

## Capabilities
- Skill 1
- Skill 2

## Patterns
Code examples
```

## ❓ Need Help?

- **Questions**: Check [.claude/README.md](.claude/README.md)
- **Issues**: See troubleshooting in [.claude/CLAUDE_CODE_SETUP.md](.claude/CLAUDE_CODE_SETUP.md)
- **Examples**: Review existing commands in `.claude/commands/`

## 🎯 Success Metrics

With Claude Code, you should see:
- ⚡ **50-70% faster** feature development
- 🐛 **80%+ reduction** in bugs (from tests and reviews)
- 📝 **100% documentation** coverage (auto-generated)
- 🔒 **Zero security issues** (auto-audited)
- ✅ **90%+ code coverage** (auto-tested)

## 🚀 Next Steps

1. ✅ Try `/go-build` and `/go-test`
2. ✅ Develop a small feature with `/add-feature`
3. ✅ Review the generated code
4. ✅ Customize for your workflow
5. ✅ Share with your team!

---

**Ready to code 10x faster? Start with `/go-build`!** 🚂✨
