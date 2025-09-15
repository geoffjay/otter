# Hooks & Lifecycle Events Example

This example demonstrates the new hooks and lifecycle events feature in Otter, which allows you to execute commands at specific points during the build process.

## Features

### 🌍 Global Hooks

- **`ON_BEFORE_BUILD`** - Commands executed before processing any layers
- **`ON_AFTER_BUILD`** - Commands executed after all layers are processed
- **`ON_ERROR`** - Commands executed when any error occurs (cleanup, notifications)

### 🔗 Layer Hooks

- **`BEFORE`** - Commands executed before processing a specific layer
- **`AFTER`** - Commands executed after processing a specific layer

## Perfect for Documentation Site Generation!

The hooks feature is ideal for setting up static site generators like Hugo, Jekyll, or MkDocs:

```dockerfile
# Documentation layer with Hugo setup
LAYER ./hugo-docs AFTER ["hugo mod init", "npm install", "hugo server --bind 0.0.0.0"]

# Or Jekyll setup
LAYER ./jekyll-docs AFTER ["bundle install", "bundle exec jekyll serve"]

# Or MkDocs setup
LAYER ./mkdocs AFTER ["pip install -r requirements.txt", "mkdocs serve"]
```

## Syntax

### Global Hooks

```dockerfile
ON_BEFORE_BUILD: ["command1", "command2"]
ON_AFTER_BUILD: ["command1", "command2"]
ON_ERROR: ["cleanup-command1", "cleanup-command2"]
```

### Layer Hooks

```dockerfile
LAYER ./my-layer BEFORE ["setup-command"] AFTER ["post-command"]

# Can be combined with other layer options
LAYER ./my-layer TARGET docs TEMPLATE name=myproject BEFORE ["echo 'Starting'"] AFTER ["echo 'Done'"]
```

## Example Use Cases

### 📚 Documentation Site Generation

```dockerfile
LAYER ./hugo-theme TARGET docs AFTER ["hugo mod init docs", "hugo server --port 1313"]
```

### 🗄️ Database Setup

```dockerfile
LAYER ./database-config BEFORE ["docker-compose up -d postgres"] AFTER ["./migrate-db.sh"]
```

### 🧪 Development Environment

```dockerfile
LAYER ./dev-tools AFTER ["npm install", "chmod +x scripts/*.sh", "./scripts/setup.sh"]
```

### 🏗️ Build Automation

```dockerfile
ON_BEFORE_BUILD: ["make clean"]
ON_AFTER_BUILD: ["make test", "make package"]
ON_ERROR: ["make clean", "echo 'Build failed - workspace cleaned'"]
```

## Command Execution

- Commands are executed in the project's root directory
- Commands support full shell features (pipes, redirection, etc.)
- Commands are executed in sequence - if one fails, execution stops
- Error commands are executed if any step fails
- All command output is displayed during build

## Testing This Example

```bash
# This example shows the parsing and structure
# (Note: The referenced layers don't exist, so copying won't work)
otter init
otter build --dry-run  # Would show parsed hooks (if implemented)
```

The hooks feature makes Otter incredibly powerful for automating complex setup workflows!
