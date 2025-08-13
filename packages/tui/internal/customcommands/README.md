# Custom Commands Technical Documentation

This package implements a comprehensive custom slash commands system for OpenCode, allowing users to create their own commands using markdown files with YAML frontmatter.

## Overview

The custom commands system provides:

1. **Markdown-based command definitions** with YAML frontmatter
2. **Multi-location command discovery** (user-level and project-level)
3. **Bash command execution** with security validation
4. **File content inclusion** and argument substitution
5. **Lazy loading and caching** for optimal performance
6. **Thread-safe operations** for concurrent access
7. **Integration with existing command system** and completion

## Architecture

### Core Components

1. **`CustomCommand`** - Command structure with metadata and execution logic
2. **`Scanner`** - Directory scanning and command discovery
3. **`Registry`** - Command management, caching, and integration
4. **`CommandCache`** - Thread-safe caching with modification tracking

### Directory Structure

Commands are loaded from multiple locations in priority order:

- **User-level**: `~/.opencode/commands/` - Available across all projects
- **Project-level**: `.opencode/commands/` - Available only in the current project

## Command File Format

Custom commands are defined in markdown files (`.md`) with YAML frontmatter:

```markdown
---
name: "command-name"
description: "Brief description of what the command does"
trigger: ["trigger1", "trigger2"]
---

# Command Content

This is the command content that will be processed when executed.

You can include:

- Regular text
- Bash commands with !
- File references with @
- Argument substitution with $ARGUMENTS
```

### Required Fields

- `name`: Unique identifier for the command
- `description`: Brief description shown in completions
- `trigger`: Array of trigger words that activate the command

### Example Commands

#### Simple Git Status Command

```markdown
---
name: "git-status"
description: "Show git status with formatting"
trigger: ["status", "st"]
---

# Git Status

!git status --porcelain
!echo "Files changed: $(git status --porcelain | wc -l)"
```

#### Build Command with Arguments

```markdown
---
name: "build-project"
description: "Build project with optional target"
trigger: ["build", "b"]
---

# Build Project

Building with arguments: $ARGUMENTS

!make $ARGUMENTS
```

#### File Reference Command

```markdown
---
name: "show-config"
description: "Display project configuration"
trigger: ["config", "cfg"]
---

# Project Configuration

@package.json
@tsconfig.json
```

## Features

### 1. Bash Command Execution

Lines starting with `!` are executed as bash commands:

```markdown
!echo "Hello World"
!git status
!npm install $ARGUMENTS
```

### 2. File References

Lines starting with `@` include file content:

```markdown
@README.md
@src/config.ts
@.env.example
```

Relative paths are resolved relative to the command file's directory.

### 3. Argument Substitution

Use `$ARGUMENTS` to substitute user-provided arguments:

```markdown
Building target: $ARGUMENTS
!make $ARGUMENTS
```

### 4. Integration with Completion System

Custom commands automatically appear in the slash command completion dialog, alongside built-in commands.

## Implementation Details

### Core Components

1. **CustomCommand**: Represents a single custom command
2. **Scanner**: Scans directories for command files
3. **Registry**: Manages loaded commands and integrates with the app
4. **CompletionProvider**: Provides completion suggestions

### Key Files

- `command.go`: Core command structure and parsing
- `scanner.go`: Directory scanning and command loading
- `registry.go`: Command management and integration
- `completion.go`: Completion provider implementation

### Integration Points

The custom commands system integrates with:

- **App initialization**: Commands are loaded when the app starts
- **Completion system**: Custom commands appear in completions
- **Command execution**: Custom commands are executed alongside built-in commands

## Usage Examples

### Creating a Custom Command

1. Create the commands directory:

   ```bash
   mkdir -p .opencode/commands
   ```

2. Create a command file (e.g., `deploy.md`):

   ```markdown
   ---
   name: "deploy-app"
   description: "Deploy application to specified environment"
   trigger: ["deploy", "d"]
   ---

   # Deploy Application

   Deploying to environment: $ARGUMENTS

   !npm run build
   !docker build -t myapp .
   !docker push myapp:$ARGUMENTS
   !kubectl apply -f k8s/$ARGUMENTS/
   ```

3. Use the command in OpenCode:
   ```
   /deploy production
   ```

### Command with File References

```markdown
---
name: "project-info"
description: "Show project information"
trigger: ["info", "project"]
---

# Project Information

## Package Configuration

@package.json

## TypeScript Configuration

@tsconfig.json

## Git Status

!git status --short
```

### Command with Complex Logic

```markdown
---
name: "test-runner"
description: "Run tests with coverage"
trigger: ["test", "t"]
---

# Test Runner

Running tests with arguments: $ARGUMENTS

!echo "Starting test run..."
!npm test $ARGUMENTS
!echo "Generating coverage report..."
!npm run coverage
!echo "Test run complete!"
```

## Error Handling

The system handles various error conditions gracefully:

- **Invalid YAML frontmatter**: Commands with invalid frontmatter are skipped
- **Missing required fields**: Commands without name or triggers are ignored
- **File not found**: Missing referenced files result in error messages
- **Command execution failures**: Bash command failures are reported to the user

## Performance Considerations

- Commands are loaded once at startup and cached
- Directory scanning is performed only when explicitly refreshed
- Command execution is performed asynchronously
- Large file references are handled efficiently

## Security Considerations

- Bash commands are executed in the current working directory
- File references are restricted to readable files
- No automatic privilege escalation
- Commands run with the same permissions as the OpenCode process

## Future Enhancements

Potential future improvements include:

1. **Command templates**: Reusable command templates
2. **Environment variables**: Access to environment variables in commands
3. **Conditional execution**: If/then logic in commands
4. **Command chaining**: Ability to chain multiple commands
5. **Interactive prompts**: Support for user input during execution
6. **Command validation**: Syntax checking and validation
7. **Hot reloading**: Automatic reloading when command files change

## Testing

The package includes comprehensive tests covering:

- Command parsing and validation
- Directory scanning
- Argument substitution
- Bash command execution
- File reference resolution
- Error handling scenarios

Run tests with:

```bash
go test ./internal/customcommands/...
```
