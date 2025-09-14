# Local Layers Example

This example demonstrates the **Local Layers & Layer Discovery** feature, which allows you to use local directories as layers instead of having to push changes to Git repositories every time you want to test modifications.

## Features Demonstrated

### 1. **Local Directory Layers**

- Relative path layers: `./layers/base-config`
- Conditional local layers: `./layers/dev-tools IF env=development`
- Local layers with template variables

### 2. **File URI Support**

- Absolute path layers using `file://` scheme
- Useful for shared layers across multiple projects

### 3. **Mixed Local and Remote Layers**

- You can combine local layers with remote Git repositories
- Perfect for development workflow: local testing → remote sharing

## Directory Structure

```
local-layers-example/
├── Otterfile                   # Main configuration
├── README.md                   # This file
└── layers/                     # Local layer definitions
    ├── base-config/            # Base application configuration
    │   ├── app.conf
    │   └── .gitignore
    ├── dev-tools/              # Development environment setup
    │   ├── docker-compose.yml
    │   └── scripts/
    │       └── start-dev.sh
    ├── app-config/             # Application-specific config with templating
    │   └── config.json
    ├── editor-config/          # Editor settings (VS Code)
    │   └── settings.json
    └── macos-specific/         # Platform-specific setup
        └── setup-macos.sh
```

## Usage

### Basic Usage

```bash
# Initialize the project
cd examples/local-layers-example
otter init

# Apply all layers (development environment by default)
otter build
```

### Environment-Specific Usage

```bash
# Development environment (includes dev-tools layer)
export OTTER_ENV=development
otter build

# Production environment (excludes dev-tools layer)
export OTTER_ENV=production
otter build
```

### Editor-Specific Usage

```bash
# Include VS Code settings
export OTTER_EDITOR=vscode
otter build

# The editor-config layer will be applied to .config/
```

### Platform-Specific Usage

```bash
# On macOS, the macos-specific layer will be automatically applied
otter build

# Files will be copied to platform/ directory
```

## Development Workflow Benefits

### 1. **Rapid Iteration**

- Modify layer files directly in `./layers/`
- Run `otter build` immediately to test changes
- No need to commit and push to Git repositories

### 2. **Layer Development**

- Create new layers locally in `./layers/`
- Test them thoroughly before publishing
- Easy to share with team members via the project repository

### 3. **Debugging and Experimentation**

- Quickly test different configurations
- Debug layer interactions
- Prototype new layer structures

## Template Variables

The example demonstrates template variable substitution:

```dockerfile
# In Otterfile
VAR PROJECT_NAME=local-example
VAR ENVIRONMENT=development

# Template variables are passed to layers
LAYER ./layers/app-config TARGET app TEMPLATE project=${PROJECT_NAME} env=${ENVIRONMENT}
```

The `config.json` file uses these variables:

```json
{
  "app": {
    "name": "${project}",        # Becomes "local-example"
    "environment": "${env}"      # Becomes "development"
  }
}
```

## File URI Examples

You can also use absolute paths with the `file://` scheme:

```dockerfile
# Absolute path to shared layer
LAYER file:///Users/shared/company-layers/monitoring TARGET monitoring

# Network path (if supported by OS)
LAYER file://server/shared/layers/common TARGET common
```

## Best Practices

### 1. **Layer Organization**

- Keep each layer focused on a single concern
- Use descriptive directory names
- Document what each layer provides

### 2. **Development Workflow**

1. Create/modify layers locally
2. Test with `otter build`
3. Iterate until satisfied
4. Commit layers to project repository
5. Eventually extract to shared repositories if needed

### 3. **Mixed Usage**

```dockerfile
# Local layers for development
LAYER ./layers/base-config
LAYER ./layers/dev-tools IF env=development

# Remote layers for stable, shared configurations
LAYER git@github.com:company/prod-monitoring.git IF env=production
LAYER git@github.com:company/security-baseline.git
```

### 4. **Testing Different Environments**

```bash
# Test all environments quickly
for env in development staging production; do
  echo "Testing $env environment..."
  OTTER_ENV=$env otter build
done
```

## Migration from Remote to Local

If you have existing remote layers you want to modify:

1. **Clone the remote layer:**

   ```bash
   git clone git@github.com:company/my-layer.git ./layers/my-layer
   ```

2. **Update Otterfile:**

   ```dockerfile
   # Before
   LAYER git@github.com:company/my-layer.git

   # After
   LAYER ./layers/my-layer
   ```

3. **Test and modify locally**

4. **When ready, push changes back to remote repository**

This workflow gives you the best of both worlds: local development speed with remote sharing capabilities.

## Advanced Examples

### Conditional File URIs

```dockerfile
# Different shared layers for different teams
LAYER file:///company/shared/frontend-team IF team=frontend
LAYER file:///company/shared/backend-team IF team=backend
```

### Complex Template Scenarios

```dockerfile
VAR TEAM=backend
VAR SERVICE_TYPE=api
VAR SCALE=small

LAYER ./layers/${TEAM}-base
LAYER ./layers/${SERVICE_TYPE}-config TEMPLATE scale=${SCALE}
```

This example shows how local layers make development much faster and more flexible while maintaining the power of Otter's layered approach to environment configuration.
