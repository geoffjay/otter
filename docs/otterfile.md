# Otterfile Documentation

The Otterfile is the core configuration file for Otter, defining which layers to apply to your development environment. This document covers the complete syntax and features available.

## Basic Syntax

An Otterfile uses a Dockerfile-like syntax with commands written in uppercase. Each command operates on a single line (multi-line commands are not currently supported).

### Comments

Lines starting with `#` are treated as comments and are ignored during parsing.

```dockerfile
# This is a comment
LAYER git@github.com:example/base.git
```

## LAYER Command

The `LAYER` command is the primary command for defining layers to be applied to your project.

### Basic Syntax

```dockerfile
LAYER <repository-url> [TARGET <target-path>] [IF <condition>]
```

### Parameters

- **`<repository-url>`** (required): The git repository URL containing the layer files
- **`TARGET <target-path>`** (optional): The directory where layer files should be copied (default: current directory)
- **`IF <condition>`** (optional): A condition that must be met for the layer to be applied

### Examples

```dockerfile
# Basic layer - applies to current directory
LAYER git@github.com:otter-layers/go-base.git

# Layer with custom target directory
LAYER git@github.com:otter-layers/vscode-config.git TARGET .vscode

# Layer with both target and condition
LAYER git@github.com:otter-layers/prod-config.git TARGET config IF env=production
```

## Conditional Layers

Conditional layers allow you to apply different configurations based on your environment, operating system, editor, or custom variables.

### Condition Syntax

Conditions use a simple `key=value` format:

```dockerfile
IF key=value
```

### Built-in Condition Variables

#### Environment (`env` or `environment`)

Controls layer application based on the current environment.

**Environment Variable Priority:**

1. `OTTER_ENV`
2. `ENV`
3. `NODE_ENV`
4. Default: `development`

**Examples:**

```dockerfile
LAYER git@github.com:otter-layers/dev-tools.git IF env=development
LAYER git@github.com:otter-layers/prod-config.git IF env=production
LAYER git@github.com:otter-layers/test-setup.git IF environment=test
```

**Setting Environment:**

```bash
export OTTER_ENV=production
otter build
```

#### Operating System (`os`)

Applies layers based on the current operating system.

**Possible Values:**

- `darwin` (macOS)
- `linux`
- `windows`
- Other values from Go's `runtime.GOOS`

**Examples:**

```dockerfile
LAYER git@github.com:otter-layers/macos-dev.git IF os=darwin
LAYER git@github.com:otter-layers/linux-dev.git IF os=linux
LAYER git@github.com:otter-layers/windows-dev.git IF os=windows
```

#### Architecture (`arch`)

Applies layers based on system architecture.

**Possible Values:**

- `amd64`
- `arm64`
- Other values from Go's `runtime.GOARCH`

**Examples:**

```dockerfile
LAYER git@github.com:otter-layers/arm64-tools.git IF arch=arm64
LAYER git@github.com:otter-layers/x86-tools.git IF arch=amd64
```

#### Editor (`editor`)

Applies layers based on the detected or configured editor.

**Editor Detection Priority:**

1. `OTTER_EDITOR` environment variable
2. `EDITOR` environment variable
3. Auto-detection (looks for `.vscode` or `.cursor` directories)

**Examples:**

```dockerfile
LAYER git@github.com:otter-layers/vscode-settings.git IF editor=vscode
LAYER git@github.com:otter-layers/cursor-rules.git IF editor=cursor
LAYER git@github.com:otter-layers/vim-config.git IF editor=vim
```

**Setting Editor:**

```bash
export OTTER_EDITOR=vscode
otter build
```

### Custom Variables

You can define custom conditions using environment variables prefixed with `OTTER_`.

**Example:**

```dockerfile
LAYER git@github.com:otter-layers/react-setup.git IF framework=react
LAYER git@github.com:otter-layers/vue-setup.git IF framework=vue
```

**Usage:**

```bash
export OTTER_FRAMEWORK=react
otter build
```

## Complete Examples

### Full-Stack Development Environment

```dockerfile
# Base project setup - always applied
LAYER git@github.com:otter-layers/base-project.git

# Environment-specific configurations
LAYER git@github.com:otter-layers/dev-tools.git IF env=development
LAYER git@github.com:otter-layers/prod-config.git IF env=production
LAYER git@github.com:otter-layers/test-setup.git IF env=test

# Operating system specific tools
LAYER git@github.com:otter-layers/macos-dev.git IF os=darwin
LAYER git@github.com:otter-layers/linux-dev.git IF os=linux

# Editor configurations
LAYER git@github.com:otter-layers/vscode-settings.git IF editor=vscode TARGET .vscode
LAYER git@github.com:otter-layers/cursor-rules.git IF editor=cursor TARGET .cursor

# Framework-specific setup
LAYER git@github.com:otter-layers/react-frontend.git IF framework=react TARGET frontend
LAYER git@github.com:otter-layers/go-backend.git IF language=go TARGET backend

# Database setup
LAYER git@github.com:otter-layers/postgres-config.git IF database=postgres
LAYER git@github.com:otter-layers/mysql-config.git IF database=mysql
```

### Multi-Environment Setup

```dockerfile
# Development environment
LAYER git@github.com:otter-layers/dev-docker.git IF env=development
LAYER git@github.com:otter-layers/hot-reload.git IF env=development
LAYER git@github.com:otter-layers/debug-tools.git IF env=development

# Staging environment
LAYER git@github.com:otter-layers/staging-config.git IF env=staging
LAYER git@github.com:otter-layers/monitoring.git IF env=staging

# Production environment
LAYER git@github.com:otter-layers/prod-docker.git IF env=production
LAYER git@github.com:otter-layers/prod-monitoring.git IF env=production
LAYER git@github.com:otter-layers/security-config.git IF env=production
```

## Usage Examples

### Setting Up Environment

```bash
# Development setup (default)
otter build

# Production setup
export OTTER_ENV=production
otter build

# Custom framework and database
export OTTER_FRAMEWORK=react
export OTTER_DATABASE=postgres
otter build

# Editor-specific setup
export OTTER_EDITOR=vscode
otter build
```

### Project Initialization

```bash
# Initialize project
otter init

# Edit Otterfile with your layer definitions
nano Otterfile

# Apply layers
otter build
```

## Best Practices

### 1. Layer Organization

```dockerfile
# Group related layers together with comments
# Base setup
LAYER git@github.com:otter-layers/project-base.git

# Environment-specific
LAYER git@github.com:otter-layers/dev-tools.git IF env=development
LAYER git@github.com:otter-layers/prod-config.git IF env=production

# Platform-specific
LAYER git@github.com:otter-layers/macos-setup.git IF os=darwin
```

### 2. Use Descriptive Repository Names

```dockerfile
# Good - descriptive names
LAYER git@github.com:company-layers/go-microservice-base.git
LAYER git@github.com:company-layers/kubernetes-deployment.git

# Avoid - generic names
LAYER git@github.com:company-layers/template1.git
LAYER git@github.com:company-layers/config.git
```

### 3. Combine Conditions Strategically

```dockerfile
# Target specific combinations
LAYER git@github.com:otter-layers/docker-dev.git IF env=development
LAYER git@github.com:otter-layers/docker-prod.git IF env=production
```

### 4. Use Consistent Environment Variables

Create a `.env` file or documentation for your team:

```bash
# Common environment variables for this project
export OTTER_ENV=development        # development, staging, production
export OTTER_FRAMEWORK=react        # react, vue, angular
export OTTER_DATABASE=postgres      # postgres, mysql, sqlite
export OTTER_EDITOR=vscode          # vscode, cursor, vim
```

## Troubleshooting

### Debug Layer Application

```bash
# See which layers would be applied
otter build --dry-run

# Verbose output showing condition evaluation
otter build --verbose
```

### Common Issues

1. **Layer not applied**: Check that your environment variables match the condition exactly
2. **Wrong layers applied**: Verify environment variable values and condition syntax
3. **Syntax errors**: Ensure proper spacing around `=` in conditions

### Checking Environment

```bash
# Check current environment variables
env | grep OTTER_

# Check OS and architecture
go env GOOS GOARCH
```

## Migration Guide

### From Basic Layers to Conditional Layers

**Before:**

```dockerfile
LAYER git@github.com:otter-layers/dev-config.git
```

**After:**

```dockerfile
LAYER git@github.com:otter-layers/dev-config.git IF env=development
LAYER git@github.com:otter-layers/prod-config.git IF env=production
```

This ensures environment-appropriate configurations are applied automatically.
