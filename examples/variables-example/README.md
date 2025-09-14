# Variables & Templating Example

This example demonstrates the powerful Variables & Templating feature implemented in Otter CLI.

## Features Demonstrated

### 1. Variable Definition

- Basic variable assignment with `VAR` command
- Variables with different types of values (strings, versions, paths)
- Recursive variable substitution (variables referencing other variables)

### 2. Variable Substitution

- Repository URL substitution using `${VARIABLE_NAME}` syntax
- TARGET path substitution
- Template variable value substitution

### 3. Template Variables

- Passing dynamic values to layers using `TEMPLATE` parameter
- Multiple template variables per layer
- Combining with conditional layers and target directories

### 4. Environment Integration

- Fallback to environment variables with `OTTER_` prefix
- Integration with existing conditional system

## Usage

### Basic Usage

```bash
# Set up environment variables (optional - will use defaults from Otterfile)
export OTTER_REPLICAS=3
export OTTER_ORGANIZATION=mycompany

# Development environment (default)
otter build

# Production environment
export OTTER_ENV=production
otter build
```

### Advanced Usage

```bash
# Override variables using environment variables
export OTTER_PROJECT_NAME=my-custom-service
export OTTER_VERSION=v2.0.0
export OTTER_DATABASE=mysql
export OTTER_LANGUAGE=python

# Use with specific editor
export OTTER_EDITOR=vscode
otter build

# Production deployment with scaling
export OTTER_ENV=production
export OTTER_REPLICAS=5
export OTTER_REGION=eu-west-1
otter build
```

## Variable Resolution Priority

Variables are resolved in this order:

1. **Otterfile variables** (defined with `VAR`) - highest priority
2. **OTTER\_ prefixed environment variables** - medium priority
3. **Direct environment variables** - lowest priority

Example:

```bash
# Otterfile contains: VAR PROJECT_NAME=example-service
# Environment has: OTTER_PROJECT_NAME=override-service
# Result: Uses "example-service" (Otterfile takes precedence)
```

## Example Layer Structure

The example Otterfile would create a structure like:

```
project/
├── services/
│   └── example-microservice/
│       ├── database/
│       ├── cache/
│       └── [language-specific files]
├── config/
│   └── [environment-specific configs]
├── k8s/
│   └── [Kubernetes manifests - production only]
├── .vscode/
│   └── [VS Code settings - if editor=vscode]
└── .cursor/
    └── [Cursor rules - if editor=cursor]
```

## Template Variables Passed to Layers

Different layers receive relevant template variables:

- **project-base**: `project`, `org`
- **go-setup**: `version`
- **postgres-config**: `db_name`
- **dev-config**: `project`, `database`
- **prod-config**: `project`, `image`, `region`
- **dockerfile-go**: `base_image`, `project`, `version`
- **k8s-manifests**: `service`, `image`, `replicas`, `namespace`

This enables layers to be highly customizable and reusable across different projects and environments.
