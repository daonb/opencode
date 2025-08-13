# Custom Commands Guide

Custom commands let you create your own slash commands in opencode using simple markdown files. They're perfect for automating repetitive tasks, creating project-specific workflows, and sharing common operations with your team.

## Quick Start

Create your first custom command in 3 steps:

```bash
# 1. Create the commands directory
mkdir -p .opencode/commands

# 2. Create a command file
cat > .opencode/commands/hello.md << 'EOF'
---
name: "hello-world"
description: "Say hello with style"
trigger: ["hello", "hi"]
---

# Hello Command

Hello, $ARGUMENTS! 👋

!echo "Current time: $(date)"
!echo "You are in: $(pwd)"
EOF

# 3. Use it in opencode
# Type: /hello world
```

## File Format

Custom commands use markdown files with YAML frontmatter:

```markdown
---
name: "command-name" # Required: unique identifier
description: "What it does" # Required: shown in completions
trigger: ["cmd", "alias"] # Required: trigger words (array)
---

# Your Command Content

Regular markdown text and special features:

- Use `!command` to run bash commands
- Use `@file.txt` to include file contents
- Use `$ARGUMENTS` for user input
```

### Required Fields

| Field         | Description                       | Example               |
| ------------- | --------------------------------- | --------------------- |
| `name`        | Unique command identifier         | `"deploy-app"`        |
| `description` | Brief description for completions | `"Deploy to staging"` |
| `trigger`     | Array of trigger words            | `["deploy", "d"]`     |

### Optional Fields

| Field    | Description                                      | Example                                  |
| -------- | ------------------------------------------------ | ---------------------------------------- |
| `model`  | Specific model to use (defaults to system model) | `"anthropic/claude-3-5-sonnet-20241022"` |
| `prompt` | Prompt to send with command output to AI         | `"Analyze the following output"`         |

## Directory Structure

Commands are loaded from two locations:

```
~/.opencode/commands/          # User-level (all projects)
├── git-shortcuts.md
├── docker-utils.md
└── personal-tools.md

.opencode/commands/            # Project-level (current project only)
├── deploy.md
├── test.md
└── build.md
```

**User-level commands** are available everywhere. **Project-level commands** are only available in that specific project.

## AI Model Integration

Custom commands can optionally integrate with AI models to process their output. This is controlled by two optional fields in the command's frontmatter:

- **`prompt`**: If provided, the command output will be sent to the AI along with this prompt
- **`model`**: Optionally specify which model to use (defaults to system model if not specified)

### Behavior Rules

1. **No `prompt` field**: Command runs and shows output only, no AI involvement
2. **`prompt` only**: Command output is sent to the default AI model with the specified prompt
3. **Both `prompt` and `model`**: Command output is sent to the specified model with the prompt

### AI Integration Example

```markdown
---
name: "analyze-logs"
description: "Analyze application logs for errors"
trigger: ["logs", "analyze-logs"]
prompt: "Analyze these logs and summarize any errors or issues found"
model: "anthropic/claude-3-5-sonnet-20241022"
---

# Log Analysis

!tail -n 100 /var/log/app.log
```

This command will:

1. Execute `tail -n 100 /var/log/app.log`
2. Send the log output to Claude 3.5 Sonnet with the analysis prompt
3. Display the AI's analysis to the user

## Core Features

### 1. Bash Command Execution

Lines starting with `!` execute as bash commands:

```markdown
---
name: "system-info"
description: "Show system information"
trigger: ["sysinfo", "sys"]
---

# System Information

## Git Status

!git status --short

## Disk Usage

!df -h

## Memory Usage

!free -h
```

### 2. File References

Lines starting with `@` include file contents:

```markdown
---
name: "show-config"
description: "Display project configuration files"
trigger: ["config", "cfg"]
---

# Project Configuration

## Package.json

@package.json

## TypeScript Config

@tsconfig.json

## Environment Example

@.env.example
```

**Path Resolution**: Relative paths are resolved from the command file's directory.

### 3. Argument Substitution

Use `$ARGUMENTS` to include user input:

```markdown
---
name: "docker-run"
description: "Run docker container with arguments"
trigger: ["docker", "run"]
---

# Docker Run

Running container with: $ARGUMENTS

!docker run $ARGUMENTS
!echo "Container started with args: $ARGUMENTS"
```

**Usage**: `/docker -it ubuntu:latest bash`

## Practical Examples

### Example 1: Silent Command (No AI Integration)

```markdown
---
name: "git-status"
description: "Show git status without AI analysis"
trigger: ["status", "gs"]
---

# Git Status

!git status --short
!git log --oneline -5
```

This command runs and shows output directly without any AI involvement.

### Example 2: AI Analysis with Default Model

```markdown
---
name: "test-failures"
description: "Run tests and analyze failures"
trigger: ["test-analyze", "ta"]
prompt: "Analyze these test results and suggest fixes for any failures"
---

# Test Analysis

!npm test 2>&1
```

This command will run tests and ask the default AI model to analyze failures.

### Example 3: Specific Model with Custom Prompt

```markdown
---
name: "code-review"
description: "Review recent code changes"
trigger: ["review", "cr"]
model: "anthropic/claude-3-5-sonnet-20241022"
prompt: "Review these git changes for potential issues, bugs, and improvements"
---

# Code Review

!git diff HEAD~1 HEAD
!git log --stat -1
```

This command will show recent changes and have Claude 3.5 Sonnet perform a code review.

### Git Workflow Command

```markdown
---
name: "git-workflow"
description: "Complete git workflow with branch and commit"
trigger: ["workflow", "wf"]
---

# Git Workflow

Creating branch and committing changes: $ARGUMENTS

## Create and switch to branch

!git checkout -b feature/$ARGUMENTS

## Show current status

!git status

## Stage all changes

!git add .

## Commit with message

!git commit -m "feat: $ARGUMENTS"

## Push to remote

!git push -u origin feature/$ARGUMENTS
```

### Project Setup Command

```markdown
---
name: "project-setup"
description: "Initialize new project with dependencies"
trigger: ["setup", "init"]
---

# Project Setup

Setting up project: $ARGUMENTS

## Create directory structure

!mkdir -p src tests docs

## Initialize package.json

!npm init -y

## Install common dependencies

!npm install $ARGUMENTS

## Create basic files

!touch src/index.js tests/index.test.js README.md

## Initialize git

!git init
!git add .
!git commit -m "Initial commit"

## Show project structure

!tree -L 2 || ls -la
```

### Environment Management

```markdown
---
name: "env-manager"
description: "Manage environment configurations"
trigger: ["env", "environment"]
---

# Environment Manager

Managing environment: $ARGUMENTS

## Current Environment Files

@.env.example
@.env.local

## Environment Variables

!printenv | grep -E "(NODE_ENV|PORT|DATABASE)" || echo "No environment variables found"

## Docker Environment

!docker-compose config 2>/dev/null || echo "No docker-compose.yml found"
```

### Testing Suite

```markdown
---
name: "test-suite"
description: "Run comprehensive test suite"
trigger: ["test", "t"]
---

# Test Suite

Running tests with options: $ARGUMENTS

## Pre-test checks

!echo "Checking test environment..."
!npm list --depth=0 | grep -E "(jest|mocha|vitest)" || echo "No test framework detected"

## Run tests

!npm test $ARGUMENTS

## Coverage report

!npm run coverage 2>/dev/null || echo "No coverage script found"

## Test results summary

!echo "Test run completed at $(date)"
```

## Advanced Patterns

### Conditional Logic with Bash

```markdown
---
name: "smart-deploy"
description: "Deploy with environment detection"
trigger: ["deploy"]
---

# Smart Deploy

Deploying to: $ARGUMENTS

## Environment Detection

!if [ "$ARGUMENTS" = "prod" ]; then echo "🚨 PRODUCTION DEPLOYMENT"; else echo "📦 Development deployment"; fi

## Pre-deployment checks

!npm run lint
!npm run test

## Build and deploy

!npm run build
!if [ "$ARGUMENTS" = "prod" ]; then npm run deploy:prod; else npm run deploy:dev; fi
```

### Multi-file References

```markdown
---
name: "project-overview"
description: "Complete project overview"
trigger: ["overview", "info"]
---

# Project Overview

## Configuration Files

@package.json
@tsconfig.json
@.gitignore

## Documentation

@README.md
@CHANGELOG.md

## Recent Activity

!git log --oneline -10
!git status --short
```

## Best Practices

### 1. **Use Descriptive Names**

```yaml
# Good
name: "deploy-to-staging"
trigger: ["deploy-staging", "stage"]

# Avoid
name: "d"
trigger: ["d"]
```

### 2. **Provide Clear Descriptions**

```yaml
# Good
description: "Deploy application to staging environment with health checks"

# Avoid
description: "Deploy stuff"
```

### 3. **Handle Errors Gracefully**

```markdown
!npm test || echo "❌ Tests failed - deployment aborted"
!docker build . || (echo "❌ Build failed" && exit 1)
```

### 4. **Use Multiple Triggers**

```yaml
trigger: ["deploy", "d", "ship"] # Multiple ways to invoke
```

### 5. **Add Context and Feedback**

```markdown
!echo "🚀 Starting deployment to $ARGUMENTS..."
!deploy-script $ARGUMENTS
!echo "✅ Deployment completed successfully!"
```

### 6. **Organize by Purpose**

```
.opencode/commands/
├── git/
│   ├── workflow.md
│   └── cleanup.md
├── docker/
│   ├── build.md
│   └── deploy.md
└── testing/
    ├── unit.md
    └── e2e.md
```

## Troubleshooting

### Command Not Found

- Check file is in correct directory (`.opencode/commands/` or `~/.opencode/commands/`)
- Verify file has `.md` extension
- Ensure YAML frontmatter is valid

### YAML Parsing Errors

```yaml
# Correct format
---
name: "my-command"
description: "Does something useful"
trigger: ["cmd"]
---
# Common mistakes to avoid:
# - Missing quotes around strings with special characters
# - Incorrect indentation
# - Missing required fields
```

### Bash Command Failures

- Commands run in current working directory
- Use full paths for executables if needed: `!/usr/bin/git status`
- Check command exists: `!which docker || echo "Docker not installed"`

### File Reference Issues

- Relative paths are resolved from command file location
- Use absolute paths if needed: `@/etc/hosts`
- Check file permissions and existence

## Security Considerations

- Commands run with same permissions as opencode
- Bash commands execute in current directory
- No automatic privilege escalation
- Be cautious with user input in `$ARGUMENTS`
- Avoid hardcoding sensitive information

## Tips and Tricks

### 1. **Create Command Templates**

Keep a template file for new commands:

```markdown
---
name: "template-command"
description: "Template for new commands"
trigger: ["template"]
---

# Command Template

Purpose: $ARGUMENTS

## Pre-checks

!echo "Starting command..."

## Main logic

# Add your logic here

## Cleanup

!echo "Command completed!"
```

### 2. **Use Environment Variables**

```markdown
!echo "Building for environment: ${NODE_ENV:-development}"
!docker build -t myapp:${VERSION:-latest} .
```

### 3. **Combine with Git Hooks**

Create commands that work with git hooks for consistent workflows.

### 4. **Share Team Commands**

Commit `.opencode/commands/` to version control to share with your team.

### 5. **Debug Commands**

Add debug output to troubleshoot:

```markdown
!echo "Debug: Arguments received: '$ARGUMENTS'"
!echo "Debug: Current directory: $(pwd)"
!echo "Debug: Environment: $NODE_ENV"
```

## Troubleshooting

### Command Not Found

- Check file has `.md` extension
- Verify YAML frontmatter starts with `---`
- Ensure command file is in `.opencode/commands/` directory
- Try `/help` to see if command appears in completion list

### Command Fails to Execute

- Check bash command syntax with `!echo "test"`
- Verify file paths exist before using `@filename`
- Use debug output: `!echo "Debug: $ARGUMENTS"`
- Check logs for validation errors

### Permission Issues

- Ensure command directories are readable
- Check file permissions on `.md` files
- For bash commands, verify executables are in PATH

---

**Need help?** Join our [Discord community](https://opencode.ai/discord) or check the [main documentation](https://opencode.ai/docs).
