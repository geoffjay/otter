# Opportunities

## Features

### 🎯 Conditional Layers & Environment-Aware Setup

State: Implemented

```
# Apply layers based on conditions
LAYER git@github.com:otter-layers/go-base.git
LAYER git@github.com:otter-layers/docker-compose.git IF env=development
LAYER git@github.com:otter-layers/kubernetes.git IF env=production
LAYER git@github.com:otter-layers/vscode-settings.git IF editor=vscode
LAYER git@github.com:otter-layers/cursor-rules.git IF editor=cursor

# OS-specific layers
LAYER git@github.com:otter-layers/macos-dev.git IF os=darwin
LAYER git@github.com:otter-layers/linux-dev.git IF os=linux
```

### 🔗 Layer Dependencies & Composition

State: Not implemented

```
# Declare layer dependencies
LAYER git@github.com:otter-layers/base-project.git AS base
LAYER git@github.com:otter-layers/go-setup.git DEPENDS base
LAYER git@github.com:otter-layers/testing-tools.git DEPENDS base
LAYER git@github.com:otter-layers/ci-cd.git DEPENDS base,go-setup
```

### 📋 Variables & Templating

State: Implemented

```
# Define variables
VAR PROJECT_NAME=my-api
VAR GO_VERSION=1.21
VAR DATABASE=postgres

# Use variables in layers
LAYER git@github.com:otter-layers/go-project.git TARGET src/${PROJECT_NAME}
LAYER git@github.com:otter-layers/${DATABASE}-setup.git
LAYER git@github.com:otter-layers/dockerfile.git TEMPLATE go_version=${GO_VERSION}
```

### 📌 Version Pinning & Layer Metadata

State: Not implemented

```
# Pin to specific versions/tags
LAYER git@github.com:otter-layers/react-base.git@v2.1.0
LAYER git@github.com:otter-layers/tailwind.git@latest
LAYER git@github.com:otter-layers/testing.git@main

# Layer with metadata
LAYER git@github.com:otter-layers/auth.git {
  version: ">=1.0.0",
  description: "Authentication setup with JWT",
  docs: "https://docs.example.com/auth-layer"
}
```

### 🎛️ Interactive Mode & Profiles

State: Not implemented

```
# Interactive setup
otter init --interactive
# Would prompt: "What type of project? [web, api, cli, library]"
# "Which database? [postgres, mysql, sqlite, none]"
# "Include testing setup? [yes, no]"

# Predefined profiles
otter init --profile fullstack-web
otter init --profile go-cli
otter init --profile react-app
```

### 🔄 Hooks & Lifecycle Events

State: Implemented

```
# Pre/post hooks for layers
LAYER git@github.com:otter-layers/database.git {
  before: ["chmod +x scripts/db-setup.sh"],
  after: ["./scripts/db-setup.sh", "go mod tidy"]
}

# Global hooks
ON_BEFORE_BUILD: ["echo 'Starting build...'"]
ON_AFTER_BUILD: ["make deps", "make test"]
ON_ERROR: ["echo 'Build failed, cleaning up...'", "make clean"]
```

### 🏗️ Multi-Stage & Grouped Layers

State: Not implemented

```
# Grouped layers for different stages
GROUP development {
  LAYER git@github.com:otter-layers/dev-tools.git
  LAYER git@github.com:otter-layers/hot-reload.git
  LAYER git@github.com:otter-layers/debug-config.git
}

GROUP production {
  LAYER git@github.com:otter-layers/prod-config.git
  LAYER git@github.com:otter-layers/monitoring.git
}

# Apply specific groups
# otter build --group development
```

### 🌐 Local Layers & Layer Discovery

State: Partially implemented (Local Layers ✅, Layer Discovery 🔄)

```
# Local directory layers
LAYER ./templates/custom-config TARGET config
LAYER file:///path/to/local/template

# Layer marketplace/registry
LAYER otter://registry/popular/go-web-api
LAYER otter://registry/user/my-custom-template

# Search and discover layers
# otter search "react typescript"
# otter info otter://registry/popular/go-web-api
```

### 🧪 Layer Validation & Testing

State: Not implemented

```
# Layer with validation
LAYER git@github.com:otter-layers/api-base.git {
  validate: {
    files_exist: ["main.go", "go.mod"],
    commands_work: ["go version", "go mod verify"]
  }
}
```

### ⚡ Performance & Caching Improvements

State: Not implemented

```
# Parallel layer processing
PARALLEL {
  LAYER git@github.com:otter-layers/frontend.git TARGET web
  LAYER git@github.com:otter-layers/backend.git TARGET api
  LAYER git@github.com:otter-layers/database.git TARGET db
}

# Smart caching with checksums
LAYER git@github.com:otter-layers/large-assets.git CACHE checksum
```

### 🔧 Enhanced CLI Commands

State: Not implemented

```
# Better build commands
otter build --dry-run                    # Show what would be applied
otter build --diff                       # Show file differences
otter build --watch                      # Watch for Otterfile changes
otter build --parallel                   # Enable parallel processing

# Layer management
otter layer list                         # Show applied layers
otter layer update                       # Update all cached layers
otter layer remove auth-layer            # Remove specific layer files
otter layer rollback                     # Undo last build

# Project templates
otter template create my-template        # Create template from current project
otter template publish my-template       # Publish to registry
```

### 📊 Configuration Inheritance & Overrides

State: Not implemented

```
# Base Otterfile
FROM git@github.com:otter-layers/base-otterfile.git

# Override specific layers
OVERRIDE database WITH git@github.com:otter-layers/mongodb.git
REMOVE layer testing-tools

# Environment-specific overrides
INCLUDE Otterfile.dev IF env=development
INCLUDE Otterfile.prod IF env=production
```
