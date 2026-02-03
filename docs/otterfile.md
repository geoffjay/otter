# Otterfile Documentation

The Otterfile is the core configuration file for Otter, defining which layers to apply to your development environment.
This document covers the complete syntax and features available.

## Basic Syntax

An Otterfile uses a Dockerfile-like syntax with commands written in uppercase. Each command operates on a single line
(multi-line commands are not currently supported).

### Comments

Lines starting with `#` are treated as comments and are ignored during parsing.

```dockerfile
# This is a comment
LAYER git@github.com:example/base.git
```

## VAR Command

The `VAR` command allows you to define reusable variables that can be used throughout your Otterfile for dynamic configuration.

### Basic Syntax

```dockerfile
VAR <variable-name>=<value>
```

### Parameters

- **`<variable-name>`** (required): The name of the variable (case-sensitive)
- **`<value>`** (required): The value to assign to the variable

### Examples

```dockerfile
# Define project variables
VAR PROJECT_NAME=my-api
VAR GO_VERSION=1.21
VAR DATABASE=postgres
VAR DESCRIPTION=My awesome API project

# Variables can reference other variables
VAR BASE_PATH=src/${PROJECT_NAME}
```

## LAYER Command

The `LAYER` command is the primary command for defining layers to be applied to your project.

### Basic Syntax

```dockerfile
LAYER <repository-url> [TARGET <target-path>] [IF <condition>] [TEMPLATE <key=value>...]
```

### Parameters

- **`<repository-url>`** (required): The layer source - can be:
  - Git repository URL (e.g., `git@github.com:user/repo.git`)
  - Local directory path (e.g., `./layers/my-layer`)
  - Absolute path (e.g., `/path/to/layer`)
  - File URI (e.g., `file:///absolute/path/to/layer`)
- **`TARGET <target-path>`** (optional): The directory where layer files should be copied (default: current directory)
- **`IF <condition>`** (optional): A condition that must be met for the layer to be applied
- **`TEMPLATE <key=value>...`** (optional): Template variables to pass to the layer

### Examples

```dockerfile
# Basic layer - applies to current directory
LAYER git@github.com:otter-layers/go-base.git

# Layer with custom target directory
LAYER git@github.com:otter-layers/vscode-config.git TARGET .vscode

# Layer with both target and condition
LAYER git@github.com:otter-layers/prod-config.git TARGET config IF env=production

# Layer with template variables
LAYER git@github.com:otter-layers/dockerfile.git TEMPLATE go_version=1.21 project_name=my-api

# Layer with variable substitution
LAYER git@github.com:otter-layers/${DATABASE}-setup.git TARGET database

# Complex layer with all parameters
LAYER git@github.com:otter-layers/service-config.git TARGET services/${SERVICE_NAME} IF env=production TEMPLATE version=${VERSION} database=${DATABASE}

# Local directory layers
LAYER ./layers/base-config TARGET config
LAYER ../shared/common-layer TARGET shared

# Absolute path layer
LAYER /company/shared/security-baseline

# File URI layer
LAYER file:///path/to/shared/layer TARGET shared
```

## Local Layers

Local layers allow you to use directories on your local filesystem as layer sources instead of remote Git repositories.
This is particularly useful for development, testing, and rapid iteration.

### Local Layer Types

#### Relative Path Layers

```dockerfile
# Relative to current directory
LAYER ./layers/dev-config
LAYER ./layers/database-setup TARGET database

# Relative to parent directory
LAYER ../shared/common-config TARGET config
```

#### Absolute Path Layers

```dockerfile
# Unix/Linux/macOS absolute path
LAYER /company/shared/base-layer

# Windows absolute path
LAYER C:/shared/layers/windows-config

# Network path (if supported by OS)
LAYER //server/shared/layer TARGET network
```

#### File URI Layers

```dockerfile
# File URI with absolute path
LAYER file:///absolute/path/to/layer

# File URI useful for cross-platform paths
LAYER file://C:/shared/layer TARGET windows IF os=windows
LAYER file:///shared/layer TARGET unix IF os!=windows
```

### Local Layer Benefits

#### 1. **Rapid Development and Testing**

```dockerfile
# Modify ./layers/my-config locally and test immediately
LAYER ./layers/my-config TARGET config

# No need to commit and push to remote repository
```

#### 2. **Mixed Local and Remote Layers**

```dockerfile
# Use local layers for development
LAYER ./layers/dev-tools IF env=development

# Use remote layers for production
LAYER git@github.com:company/prod-config.git IF env=production
```

#### 3. **Shared Team Resources**

```dockerfile
# Reference shared company layers
LAYER file:///company/shared/security-policies TARGET security
LAYER file:///company/shared/docker-configs TARGET docker
```

### Development Workflow

#### 1. **Create Local Layer**

```bash
# Create a local layer directory
mkdir -p ./layers/my-layer
echo "config: value" > ./layers/my-layer/config.yaml
```

#### 2. **Use in Otterfile**

```dockerfile
# Reference the local layer
LAYER ./layers/my-layer TARGET config
```

#### 3. **Test and Iterate**

```bash
# Test the layer
otter build

# Modify layer files
echo "new_config: new_value" >> ./layers/my-layer/config.yaml

# Test again immediately
otter build
```

#### 4. **Graduate to Remote Repository**

```bash
# When ready, create a remote repository
git init ./layers/my-layer
cd ./layers/my-layer
git remote add origin git@github.com:company/my-layer.git
git push -u origin main

# Update Otterfile
# LAYER ./layers/my-layer TARGET config              # Local
# LAYER git@github.com:company/my-layer.git TARGET config  # Remote
```

### Local Layer Validation

Local layers are validated when processed:

- **Directory exists**: The path must point to an existing directory
- **Directory accessible**: Must have read permissions
- **Path resolution**: Relative paths are resolved to absolute paths

### Local Layer Examples

#### Multi-Environment Setup

```dockerfile
# Base configuration (always applied)
LAYER ./layers/base-config

# Environment-specific local layers
LAYER ./layers/dev-environment IF env=development
LAYER ./layers/staging-environment IF env=staging

# Remote production layer (more controlled)
LAYER git@github.com:company/prod-config.git IF env=production
```

#### Team Development Setup

```dockerfile
# Individual developer layers
LAYER ./layers/personal-dev-config TARGET dev

# Shared team layers from network storage
LAYER file:///team/shared/coding-standards TARGET standards
LAYER file:///team/shared/docker-setup TARGET docker

# Project-specific layers
LAYER ./layers/project-config TARGET config
```

#### Platform-Specific Layers

```dockerfile
# Local platform-specific configurations
LAYER ./layers/macos-dev IF os=darwin
LAYER ./layers/linux-dev IF os=linux
LAYER ./layers/windows-dev IF os=windows

# Shared platform tools
LAYER file:///company/platform/macos TARGET platform IF os=darwin
LAYER file:///company/platform/linux TARGET platform IF os=linux
```

#### Template Variables with Local Layers

```dockerfile
VAR PROJECT_NAME=my-app
VAR TEAM=backend

# Local layer with template variables
LAYER ./layers/app-config TARGET config TEMPLATE project=${PROJECT_NAME} team=${TEAM}

# The config files in ./layers/app-config can use ${project} and ${team}
```

### Best Practices

#### 1. **Layer Organization**

```
project/
├── Otterfile
├── layers/                    # Local layers
│   ├── base-config/
│   ├── dev-tools/
│   ├── test-setup/
│   └── app-specific/
└── src/                      # Application code
```

#### 2. **Naming Conventions**

```dockerfile
# Use descriptive names
LAYER ./layers/database-config TARGET database
LAYER ./layers/api-gateway-config TARGET gateway

# Avoid generic names
# LAYER ./layers/config1          # Bad
# LAYER ./layers/stuff            # Bad
```

#### 3. **Environment Separation**

```dockerfile
# Keep environment-specific layers separate
LAYER ./layers/base-config                          # Common
LAYER ./layers/dev-overrides IF env=development     # Dev-specific
LAYER ./layers/prod-overrides IF env=production     # Prod-specific
```

#### 4. **Documentation**

```dockerfile
# Comment your local layers
# Base application configuration
LAYER ./layers/app-config TARGET config

# Development tools and utilities
LAYER ./layers/dev-tools IF env=development TARGET tools

# Platform-specific development setup
LAYER ./layers/macos-setup IF os=darwin TARGET platform
```

### Migration Strategies

#### From Remote to Local (for development)

```dockerfile
# Step 1: Clone remote layer locally
# git clone git@github.com:company/layer.git ./layers/layer

# Step 2: Update Otterfile
# LAYER git@github.com:company/layer.git      # Old
LAYER ./layers/layer                          # New

# Step 3: Develop and test locally
# Step 4: Push changes back to remote when ready
```

#### From Local to Remote (for sharing)

```dockerfile
# Step 1: Create remote repository for local layer
# Step 2: Update Otterfile
# LAYER ./layers/shared-config                        # Local only
LAYER git@github.com:company/shared-config.git       # Now remote
```

## Conditional Layers

Conditional layers allow you to apply different configurations based on your environment, operating system, editor, or
custom variables.

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

## Variables & Templating

Variables and templating provide powerful ways to make your Otterfiles dynamic and reusable across different
environments and projects.

### Variable Definition

Variables are defined using the `VAR` command and can be used throughout your Otterfile:

```dockerfile
# Define variables at the top of your Otterfile
VAR PROJECT_NAME=my-awesome-api
VAR GO_VERSION=1.21
VAR DATABASE=postgres
VAR TEAM=backend
```

### Variable Substitution

Variables can be referenced using `${VARIABLE_NAME}` syntax in:

- Layer repository URLs
- Target directory paths
- Template variable values

```dockerfile
# Use variables in repository URLs
LAYER git@github.com:otter-layers/${DATABASE}-setup.git

# Use variables in target paths
LAYER git@github.com:otter-layers/go-project.git TARGET src/${PROJECT_NAME}

# Variables can reference other variables
VAR BASE_PATH=services/${PROJECT_NAME}
LAYER git@github.com:templates/config.git TARGET ${BASE_PATH}/config
```

### Template Variables

Template variables allow you to pass dynamic values to layers using the `TEMPLATE` parameter:

```dockerfile
# Pass template variables to layers
LAYER git@github.com:otter-layers/dockerfile.git TEMPLATE go_version=${GO_VERSION} project=${PROJECT_NAME}

# Multiple template variables
LAYER git@github.com:otter-layers/k8s-config.git TEMPLATE service=${PROJECT_NAME} version=v1.0 replicas=3
```

### Variable Priority

Variables are resolved in the following order (highest to lowest priority):

1. **Otterfile variables** - Variables defined with `VAR` command
2. **OTTER\_ environment variables** - Environment variables prefixed with `OTTER_`
3. **Direct environment variables** - Regular environment variables

```bash
# Environment variables can be used as fallbacks
export OTTER_TEAM=frontend
export DATABASE_URL=postgres://localhost/mydb

# These will be available as ${TEAM} and ${DATABASE_URL} in your Otterfile
```

### Advanced Examples

#### Multi-Service Project

```dockerfile
# Project-wide variables
VAR ORG_NAME=mycompany
VAR PROJECT_NAME=ecommerce
VAR VERSION=v2.1.0

# Service-specific variables
VAR AUTH_SERVICE=auth-service
VAR PAYMENT_SERVICE=payment-service
VAR INVENTORY_SERVICE=inventory-service

# Base layers for all services
LAYER git@github.com:${ORG_NAME}/microservice-base.git
LAYER git@github.com:${ORG_NAME}/monitoring.git

# Individual service layers
LAYER git@github.com:${ORG_NAME}/${AUTH_SERVICE}.git@${VERSION} TARGET services/${AUTH_SERVICE}
LAYER git@github.com:${ORG_NAME}/${PAYMENT_SERVICE}.git@${VERSION} TARGET services/${PAYMENT_SERVICE}
LAYER git@github.com:${ORG_NAME}/${INVENTORY_SERVICE}.git@${VERSION} TARGET services/${INVENTORY_SERVICE}

# Environment-specific configurations
LAYER git@github.com:${ORG_NAME}/dev-config.git IF env=development TEMPLATE project=${PROJECT_NAME}
LAYER git@github.com:${ORG_NAME}/prod-config.git IF env=production TEMPLATE project=${PROJECT_NAME} version=${VERSION}
```

#### Framework-Agnostic Setup

```dockerfile
# Framework and language variables
VAR FRAMEWORK=react
VAR LANGUAGE=typescript
VAR BUILD_TOOL=vite

# Dynamic layer selection based on framework
LAYER git@github.com:otter-layers/${FRAMEWORK}-base.git TARGET frontend
LAYER git@github.com:otter-layers/${LANGUAGE}-config.git TARGET frontend
LAYER git@github.com:otter-layers/${BUILD_TOOL}-setup.git TARGET frontend TEMPLATE framework=${FRAMEWORK}

# Database layer based on environment variable
LAYER git@github.com:otter-layers/${DATABASE}-setup.git TEMPLATE db_name=${PROJECT_NAME}_${ENVIRONMENT}
```

#### Environment-Specific Configurations

```dockerfile
# Base configuration
VAR SERVICE_NAME=user-api
VAR BASE_IMAGE=alpine:3.18

# Environment-specific variables via environment variables
# OTTER_REPLICAS, OTTER_MEMORY_LIMIT, OTTER_CPU_LIMIT should be set

# Docker configuration with environment-specific resources
LAYER git@github.com:otter-layers/dockerfile.git TEMPLATE \
  base_image=${BASE_IMAGE} \
  service=${SERVICE_NAME}

# Kubernetes manifests with environment-specific scaling
LAYER git@github.com:otter-layers/k8s-manifests.git TARGET k8s IF env=production TEMPLATE \
  service=${SERVICE_NAME} \
  replicas=${REPLICAS} \
  memory_limit=${MEMORY_LIMIT} \
  cpu_limit=${CPU_LIMIT}
```

### Best Practices

#### 1. Group Variables Logically

```dockerfile
# Project metadata
VAR PROJECT_NAME=my-api
VAR VERSION=1.0.0
VAR DESCRIPTION=My awesome API

# Technology stack
VAR LANGUAGE=go
VAR DATABASE=postgres
VAR CACHE=redis

# Infrastructure
VAR CLOUD_PROVIDER=aws
VAR REGION=us-west-2
```

#### 2. Use Descriptive Variable Names

```dockerfile
# Good - descriptive names
VAR GO_VERSION=1.21
VAR POSTGRES_VERSION=15
VAR REDIS_VERSION=7-alpine

# Avoid - generic names
VAR VER=1.21
VAR DB=postgres
VAR V1=15
```

#### 3. Document Variable Usage

```dockerfile
# Database configuration variables
VAR DATABASE=postgres          # Database type: postgres, mysql, sqlite
VAR DB_VERSION=15              # Database version
VAR DB_NAME=${PROJECT_NAME}_db # Generated database name

# Use the variables
LAYER git@github.com:otter-layers/${DATABASE}-${DB_VERSION}.git TEMPLATE db_name=${DB_NAME}
```

#### 4. Validate Critical Variables

Ensure critical environment variables are set before running `otter build`:

```bash
# Example setup script
#!/bin/bash
set -e

# Check required environment variables
: "${OTTER_PROJECT_NAME:?Environment variable OTTER_PROJECT_NAME is required}"
: "${OTTER_DATABASE:?Environment variable OTTER_DATABASE is required}"

# Run otter build
otter build
```

## Hooks & Lifecycle Events

Hooks allow you to execute shell commands at various points during the build process. This is useful for setup scripts,
validation, cleanup, and post-processing.

### Global Hooks

Global hooks execute once per build, regardless of how many layers are applied.

#### ON_BEFORE_BUILD

Runs before any layers are processed. Use this for pre-build setup, validation, or environment preparation.

```dockerfile
ON_BEFORE_BUILD: ["echo 'Starting build...'"]
ON_BEFORE_BUILD: ["./scripts/validate-env.sh", "mkdir -p tmp"]
```

#### ON_AFTER_BUILD

Runs after all layers have been successfully applied. Use this for post-build tasks like running tests, installing
dependencies, or generating documentation.

```dockerfile
ON_AFTER_BUILD: ["go mod tidy"]
ON_AFTER_BUILD: ["npm install", "npm run build"]
```

#### ON_ERROR

Runs when any error occurs during the build process (including hook failures). Use this for cleanup, notifications, or
error handling.

```dockerfile
ON_ERROR: ["echo 'Build failed, cleaning up...'"]
ON_ERROR: ["./scripts/cleanup.sh", "rm -rf tmp"]
```

### Per-Layer Hooks

Per-layer hooks execute for each layer individually, allowing layer-specific setup and post-processing.

#### BEFORE

Runs immediately before a layer's files are copied. Use this for layer-specific preparation.

```dockerfile
LAYER git@github.com:otter-layers/database.git BEFORE ["chmod +x scripts/*.sh"]
LAYER git@github.com:otter-layers/config.git BEFORE ["mkdir -p config", "backup-existing.sh"]
```

#### AFTER

Runs immediately after a layer's files are copied successfully. Use this for layer-specific post-processing.

```dockerfile
LAYER git@github.com:otter-layers/go-project.git AFTER ["go mod tidy"]
LAYER git@github.com:otter-layers/npm-setup.git AFTER ["npm install", "npm run setup"]
```

### Hook Syntax

Hooks use JSON array syntax for specifying commands:

```dockerfile
# Single command
ON_BEFORE_BUILD: ["echo 'hello'"]

# Multiple commands (executed sequentially)
ON_AFTER_BUILD: ["go mod tidy", "go build ./...", "go test ./..."]

# Layer hooks can be combined with other parameters
LAYER git@github.com:example/layer.git TARGET config IF env=production BEFORE ["validate.sh"] AFTER ["post-setup.sh"]
```

### Execution Order

The build process executes hooks in this order:

1. **ON_BEFORE_BUILD** hooks (once at start)
2. For each layer:
   - **BEFORE** hooks for the layer
   - Clone/update and copy layer files
   - **AFTER** hooks for the layer
3. **ON_AFTER_BUILD** hooks (once at end)

If any step fails:
- The build process stops immediately
- **ON_ERROR** hooks are executed (if defined)
- The build command exits with an error

### Error Handling

- Each command in a hook array runs sequentially
- If any command fails (non-zero exit code), the build stops
- ON_ERROR hooks are always attempted when an error occurs
- Hook commands inherit the current working directory (project root)

### Examples

#### Basic Setup with Hooks

```dockerfile
# Validate environment before starting
ON_BEFORE_BUILD: ["./scripts/check-deps.sh"]

# Base layer
LAYER git@github.com:otter-layers/go-base.git

# Database layer with setup script
LAYER git@github.com:otter-layers/postgres-config.git AFTER ["./scripts/db-init.sh"]

# Run tests after build
ON_AFTER_BUILD: ["go test ./..."]

# Cleanup on failure
ON_ERROR: ["echo 'Build failed'", "./scripts/cleanup.sh"]
```

#### Environment-Specific Hooks

```dockerfile
VAR PROJECT_NAME=my-api

# Development setup
LAYER git@github.com:otter-layers/dev-tools.git IF env=development AFTER ["npm install", "npm run dev:setup"]

# Production setup with validation
LAYER git@github.com:otter-layers/prod-config.git IF env=production BEFORE ["./scripts/prod-validate.sh"] AFTER ["./scripts/prod-verify.sh"]

ON_AFTER_BUILD: ["echo 'Build complete for ${PROJECT_NAME}'"]
```

#### Complex Multi-Layer Setup

```dockerfile
ON_BEFORE_BUILD: ["echo 'Starting build...'", "mkdir -p .otter/tmp"]

# Base configuration
LAYER git@github.com:otter-layers/base-config.git

# API layer with dependency installation
LAYER git@github.com:otter-layers/go-api.git TARGET api BEFORE ["mkdir -p api"] AFTER ["cd api && go mod download"]

# Frontend layer with build step
LAYER git@github.com:otter-layers/react-app.git TARGET web AFTER ["cd web && npm install && npm run build"]

# Docker configuration
LAYER git@github.com:otter-layers/docker-compose.git AFTER ["docker-compose config --quiet"]

ON_AFTER_BUILD: ["echo 'All layers applied successfully'", "./scripts/final-setup.sh"]
ON_ERROR: ["echo 'Build failed, see logs for details'", "rm -rf .otter/tmp"]
```

## Complete Examples

### Full-Stack Development Environment with Variables

```dockerfile
# Project variables
VAR PROJECT_NAME=my-fullstack-app
VAR FRONTEND_FRAMEWORK=react
VAR BACKEND_LANGUAGE=go
VAR DATABASE_TYPE=postgres

# Base project setup - always applied
LAYER git@github.com:otter-layers/base-project.git TEMPLATE project=${PROJECT_NAME}

# Environment-specific configurations
LAYER git@github.com:otter-layers/dev-tools.git IF env=development TEMPLATE project=${PROJECT_NAME}
LAYER git@github.com:otter-layers/prod-config.git IF env=production TEMPLATE project=${PROJECT_NAME}
LAYER git@github.com:otter-layers/test-setup.git IF env=test

# Operating system specific tools
LAYER git@github.com:otter-layers/macos-dev.git IF os=darwin
LAYER git@github.com:otter-layers/linux-dev.git IF os=linux

# Editor configurations
LAYER git@github.com:otter-layers/vscode-settings.git IF editor=vscode TARGET .vscode
LAYER git@github.com:otter-layers/cursor-rules.git IF editor=cursor TARGET .cursor

# Framework-specific setup using variables
LAYER git@github.com:otter-layers/${FRONTEND_FRAMEWORK}-frontend.git TARGET frontend TEMPLATE project=${PROJECT_NAME}
LAYER git@github.com:otter-layers/${BACKEND_LANGUAGE}-backend.git TARGET backend TEMPLATE project=${PROJECT_NAME}

# Database setup using variables
LAYER git@github.com:otter-layers/${DATABASE_TYPE}-config.git TEMPLATE db_name=${PROJECT_NAME}_${ENVIRONMENT}
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
