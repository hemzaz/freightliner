# Claude Code Setup Guide for Freightliner

This guide explains the complete Claude Code setup for the Freightliner project.

## 📁 Directory Structure

```
.claude/
├── README.md                          # Main Claude Code documentation
├── CLAUDE_CODE_SETUP.md              # This file
├── settings.json                      # Permissions and settings
├── mcp-config.json                    # MCP server configurations
├── statusline.json                    # Status line configuration
├── tech.md                            # Technical standards
├── structure.md                       # Project structure
├── product.md                         # Product vision
├── commands/                          # Slash commands
│   ├── add-feature.md                 # Feature development workflow
│   ├── replicate-debug.md             # Replication debugging
│   ├── fix-ci.md                      # CI/CD troubleshooting
│   ├── security-audit.md              # Security auditing
│   ├── performance-test.md            # Performance testing
│   ├── deploy.md                      # Kubernetes deployment
│   ├── go-build.md                    # Go build command
│   ├── go-test.md                     # Go test command
│   └── go-lint.md                     # Go lint command
├── agents/                            # Specialized agents
│   ├── COORDINATION.md                # Agent coordination guide
│   ├── WORKFLOWS.md                   # Multi-agent workflows
│   ├── ENHANCED_SYSTEM.md             # Enhanced agent system
│   ├── golang-pro.md                  # Go expert agent
│   ├── architect-reviewer.md          # Architecture review agent
│   ├── test-automator.md              # Testing agent
│   ├── security-auditor.md            # Security review agent
│   ├── code-reviewer.md               # Code review agent
│   ├── performance-engineer.md        # Performance optimization agent
│   ├── deployment-engineer.md         # Deployment agent
│   ├── devops-troubleshooter.md       # Ops troubleshooting agent
│   ├── cloud-architect.md             # Cloud architecture agent
│   ├── api-documenter.md              # Documentation agent
│   ├── prompt-engineer.md             # Prompt optimization agent
│   └── context-manager.md             # Context management agent
├── skills/                            # Domain expertise modules
│   ├── container-registry.md          # Container registry operations
│   ├── go-microservice.md             # Go microservice patterns
│   └── kubernetes-ops.md              # Kubernetes operations
└── workflows/                         # Multi-agent workflows
    ├── feature-development.md         # Complete feature workflow
    └── incident-response.md           # Incident handling workflow
```

## 🎯 Core Components

### 1. Slash Commands

Slash commands provide quick access to common workflows:

| Command | Description | Agent(s) |
|---------|-------------|----------|
| `/add-feature` | Full feature development lifecycle | golang-pro, test-automator, security-auditor |
| `/replicate-debug` | Debug replication issues | devops-troubleshooter |
| `/fix-ci` | Fix CI/CD pipeline failures | deployment-engineer |
| `/security-audit` | Comprehensive security review | security-auditor |
| `/performance-test` | Run performance benchmarks | performance-engineer |
| `/deploy` | Deploy to Kubernetes | deployment-engineer |
| `/go-build` | Build the application | golang-pro |
| `/go-test` | Run tests | test-automator |
| `/go-lint` | Run linting | code-reviewer |

**Creating New Commands:**

1. Create file in `.claude/commands/your-command.md`
2. Use this template:

```markdown
# Your Command Name

Description of what this command does.

## What This Command Does

1. Step 1
2. Step 2
3. Step 3

## Usage

```bash
/your-command [args]
```

## Example

```bash
/your-command example-arg
```

## Tasks

- Task 1
- Task 2
- Task 3
```

### 2. Specialized Agents

Agents are AI specialists for specific domains:

#### Development Agents
- **golang-pro**: Go development expert, knows all Freightliner patterns
- **architect-reviewer**: Reviews architecture and design decisions
- **test-automator**: Creates comprehensive test suites

#### Operations Agents
- **deployment-engineer**: Handles deployments and CI/CD
- **devops-troubleshooter**: Diagnoses production issues
- **cloud-architect**: Cloud infrastructure design

#### Quality Agents
- **security-auditor**: Security reviews and vulnerability scanning
- **code-reviewer**: Code quality and style reviews
- **performance-engineer**: Performance optimization

#### Support Agents
- **api-documenter**: Documentation creation
- **context-manager**: Multi-agent coordination
- **prompt-engineer**: Prompt optimization

**Using Agents:**

```bash
# Invoke specific agent
/add-feature "New feature" --agent=golang-pro

# Multiple agents in parallel
/add-feature "New feature" --agents="golang-pro,security-auditor,test-automator"
```

**Agent Configuration:**

Each agent has a markdown file in `.claude/agents/` defining:
- Expertise areas
- Responsibilities
- Tools available
- Coordination patterns

### 3. Skills

Skills are reusable knowledge modules:

#### container-registry
- ECR and GCR authentication
- Image replication patterns
- Registry debugging
- Multi-architecture support

#### go-microservice
- Interface-driven design
- Error handling patterns
- Context usage
- Testing strategies
- HTTP server patterns

#### kubernetes-ops
- Deployment strategies
- High availability patterns
- Security best practices
- Monitoring setup

**Using Skills:**

Skills are automatically available to agents. Reference them in commands:

```markdown
See .claude/skills/go-microservice.md for implementation patterns.
```

### 4. Workflows

Workflows coordinate multiple agents for complex tasks:

#### feature-development
- Complete feature lifecycle
- Multi-agent coordination
- Quality gates
- Documentation updates

#### incident-response
- Severity assessment
- Root cause analysis
- Fix implementation
- Post-mortem review

**Using Workflows:**

```bash
/add-feature "Description"
# Automatically uses feature-development workflow
```

### 5. MCP Servers

Model Context Protocol servers extend capabilities:

#### Configured MCPs
- **filesystem**: File system operations
- **github**: GitHub API access (requires GITHUB_TOKEN)
- **aws**: AWS services (ECR, KMS, Secrets Manager)
- **google-cloud**: GCP services (GCR, Artifact Registry)
- **kubernetes**: Kubernetes cluster operations
- **docker**: Docker daemon access
- **prometheus**: Metrics access (disabled by default)

**Enabling/Configuring MCPs:**

Edit `.claude/mcp-config.json`:

```json
{
  "mcpServers": {
    "your-mcp": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-your-mcp"],
      "env": {
        "API_KEY": "${YOUR_API_KEY}"
      },
      "enabled": true
    }
  }
}
```

**Environment Variables Required:**
- `GITHUB_TOKEN`: For GitHub MCP
- `GOOGLE_APPLICATION_CREDENTIALS`: For Google Cloud MCP
- `KUBECONFIG`: For Kubernetes MCP (defaults to ~/.kube/config)
- `PROMETHEUS_URL`: For Prometheus MCP (optional)

## 🚀 Getting Started

### Initial Setup

1. **Clone the repository:**
```bash
git clone https://github.com/hemzaz/freightliner.git
cd freightliner
```

2. **Install dependencies:**
```bash
make setup
```

3. **Configure environment:**
```bash
# AWS credentials
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# GCP credentials
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json

# GitHub token for MCP
export GITHUB_TOKEN=your-github-token
```

4. **Verify Claude Code setup:**
```bash
# Check that .claude directory exists
ls .claude/

# Verify slash commands
ls .claude/commands/

# Verify agents
ls .claude/agents/

# Verify skills
ls .claude/skills/
```

### Daily Development Workflow

1. **Start your work:**
```bash
# Check project status
make help

# Pull latest changes
git pull origin main
```

2. **Develop a feature:**
```bash
/add-feature "Your feature description"
```

This will:
- Create feature spec
- Design interfaces
- Implement code
- Add tests
- Review security
- Review code quality
- Update documentation

3. **Fix issues:**
```bash
# CI failures
/fix-ci

# Replication issues
/replicate-debug ecr/repo gcr/repo

# Performance problems
/performance-test
```

4. **Deploy:**
```bash
# To staging
/deploy staging v1.2.0

# To production
/deploy production v1.2.0
```

## 🎓 Advanced Usage

### Multi-Agent Coordination

For complex tasks, use multiple agents in parallel:

```bash
# Feature development with parallel security and performance review
/add-feature "New ACR support" --agents="golang-pro,security-auditor,performance-engineer"
```

**The context-manager agent automatically:**
- Preserves context across agents
- Aggregates results
- Resolves conflicts
- Generates final report

### Custom Workflows

Create custom workflows in `.claude/workflows/`:

```markdown
# My Custom Workflow

## Workflow Steps

### 1. Step Name (agent-name)
Description and tasks

### 2. Next Step (other-agent)
Description and tasks

## Parallel Execution
- Step A (agent-1)
- Step B (agent-2)
```

### Extending Skills

Add new skills in `.claude/skills/`:

```markdown
# My Skill Name

Description of the skill domain.

## Capabilities
- Capability 1
- Capability 2

## Common Patterns
Code examples and patterns

## Best Practices
Guidelines and recommendations
```

## 🔧 Customization

### Permissions

Edit `.claude/settings.json` to control tool access:

```json
{
  "permissions": {
    "allow": [
      "Bash(go:*)",
      "Bash(make:*)",
      "Bash(your-command:*)"
    ],
    "deny": []
  }
}
```

### Status Line

Customize `.claude/statusline.json`:

```json
{
  "format": "🚂 Freightliner | {component1} | {component2}",
  "components": {
    "component1": {
      "command": "your-command",
      "refresh": 5000,
      "color": "blue"
    }
  }
}
```

### Agent Behavior

Modify agent files in `.claude/agents/` to:
- Adjust expertise areas
- Change tools used
- Modify coordination patterns
- Add new responsibilities

## 📊 Monitoring & Metrics

### Command Usage

Track which commands are most useful:
```bash
# Review command files
ls -lh .claude/commands/

# See agent activity
# (logged automatically by Claude Code)
```

### Agent Performance

Monitor agent effectiveness:
- Success rate of automated fixes
- Time saved vs manual work
- Code quality improvements
- Test coverage increases

### Workflow Efficiency

Measure workflow improvements:
- Feature development time
- Bug fix turnaround
- Deployment frequency
- Incident resolution time

## 🐛 Troubleshooting

### Commands Not Working

1. Check file exists: `ls .claude/commands/your-command.md`
2. Verify markdown format is correct
3. Check permissions in `.claude/settings.json`

### MCPs Not Loading

1. Verify environment variables set
2. Check MCP configuration: `cat .claude/mcp-config.json`
3. Ensure npx is available: `which npx`
4. Check MCP logs in Claude Code output

### Agents Not Coordinating

1. Verify context-manager agent is configured
2. Check workflow file format
3. Review agent coordination docs: `.claude/agents/COORDINATION.md`

## 📚 Resources

### Internal Documentation
- [Main README](.claude/README.md)
- [Agent Coordination](.claude/agents/COORDINATION.md)
- [Enhanced System](.claude/agents/ENHANCED_SYSTEM.md)
- [Workflows](.claude/agents/WORKFLOWS.md)

### Project Documentation
- [Architecture](../docs/ARCHITECTURE.md)
- [Contributing](../CONTRIBUTING.md)
- [Security](../docs/SECURITY.md)

### External Resources
- [Claude Code Documentation](https://docs.claude.com/claude-code)
- [MCP Specification](https://modelcontextprotocol.io)
- [Go Best Practices](https://golang.org/doc/effective_go)

## 🎯 Best Practices

1. **Use specific commands** instead of generic requests
2. **Leverage parallel agents** for faster development
3. **Review AI output** before committing
4. **Extend skills** as you learn domain knowledge
5. **Document workflows** that work well
6. **Share commands** with the team
7. **Monitor effectiveness** and adjust configurations

## 🚀 Next Steps

1. Try basic commands: `/go-build`, `/go-test`
2. Develop a small feature: `/add-feature "simple feature"`
3. Review generated code and tests
4. Customize agent behavior for your needs
5. Create custom commands for frequent tasks
6. Share improvements with the team

---

**Happy Coding with Claude!** 🚂✨
