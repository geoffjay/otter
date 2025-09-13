# Otter

Otter simplifies development environment setup through a layer concept that pulls other templates containing files into the project it's run inside of.

## Features

- **Layer-based configuration**: Define layers from git repositories that are pulled into your project
- **Dockerfile-like syntax**: Familiar syntax for defining layers and targets
- **Intelligent file handling**: Support for `.otterignore` to exclude files from layers
- **Caching**: Git repositories are cached locally for faster subsequent builds
- **Flexible targeting**: Specify custom target directories for each layer

## Installation

### Option 1: Download Pre-built Binaries

Download the latest release for your platform from the [GitHub Releases](https://github.com/geoffjay/otter/releases) page.

### Option 2: Install with Go

```bash
go install github.com/geoffjay/otter@latest
```

### Option 3: Build from Source

1. Clone this repository:

```bash
git clone <this-repo-url>
cd otter
```

2. Build the binary:

```bash
make build
# or manually: go mod tidy && go build -o bin/otter
```

3. (Optional) Install globally:

```bash
make install
```

### Option 4: Docker

```bash
# Run directly
docker run --rm -it ghcr.io/geoffjay/otter:latest --help

# Use in a project directory
docker run --rm -v $(pwd):/workspace -w /workspace ghcr.io/geoffjay/otter:latest init
```

## Quick Start

1. **Initialize a project:**

```bash
otter init
```

This creates:

- `.otter/cache/` directory for caching git repositories
- `.otterignore` file with sensible defaults
- Sample `Otterfile` with example layer definitions

2. **Define layers in your `Otterfile`:**

```dockerfile
# Pull a Go CLI template
LAYER git@github.com:otter-layers/go-cobra-cli.git

# Pull Cursor rules to a specific directory
LAYER git@github.com:otter-layers/cursor-go-rules.git TARGET .cursor/rules

# Pull configuration files
LAYER https://github.com/user/dotfiles.git TARGET config
```

3. **Build your environment:**

```bash
otter build
```

## Commands

### `otter init`

Initialize the current directory for otter by creating:

- `.otter/cache/` directory for layer caching
- `.otterignore` file with default ignore patterns
- Sample `Otterfile` with example usage

### `otter build`

Read the `Otterfile` (or `Envfile`) and apply all defined layers to the current project.

**Options:**

- `-f, --file <path>`: Specify a custom Otterfile/Envfile path

## Otterfile Syntax

The `Otterfile` uses a Dockerfile-like syntax:

### LAYER Command

```dockerfile
LAYER <git-repository-url> [TARGET <target-directory>]
```

**Examples:**

```dockerfile
# Clone to project root
LAYER git@github.com:user/template.git

# Clone to specific directory
LAYER https://github.com/user/configs.git TARGET .config

# SSH repository with custom target
LAYER git@github.com:company/internal-template.git TARGET internal
```

## .otterignore File

The `.otterignore` file works similarly to `.gitignore` and specifies files and patterns to exclude when merging layers.

**Example `.otterignore`:**

```
# Version control
.git/
.svn/

# Otter internals
.otter/

# Dependencies
node_modules/
vendor/

# Temporary files
*.log
*.tmp
.DS_Store

# Specific files
secrets.yaml
local-config.json
```

## How It Works

1. **Initialization**: `otter init` sets up the `.otter/cache/` directory structure
2. **Layer Processing**: `otter build` reads your `Otterfile` and processes each `LAYER` command:
   - Clones git repositories to `.otter/cache/` (or updates if already cached)
   - Applies `.otterignore` patterns to filter files
   - Copies allowed files to the specified target directory
3. **File Merging**: Files from layers are merged into your project, with existing files being overwritten

## Repository Structure

```
your-project/
├── .otter/
│   └── cache/          # Cached git repositories
├── .otterignore        # File ignore patterns
├── Otterfile          # Layer definitions
└── [your project files]
```

## Layer Repository Best Practices

When creating layer repositories:

1. **Keep layers focused**: Each layer should serve a specific purpose
2. **Use descriptive README**: Document what the layer provides
3. **Avoid large files**: Layers should contain configuration and template files, not large assets
4. **Consider ignore patterns**: Structure your layer so common ignore patterns work well
5. **Version your layers**: Use git tags for stable layer versions

## Examples

### Basic Go Project Setup

```dockerfile
# Otterfile
LAYER git@github.com:otter-layers/go-mod.git
LAYER git@github.com:otter-layers/go-cobra-cli.git
LAYER git@github.com:otter-layers/go-gitignore.git
```

### Full Stack Development Environment

```dockerfile
# Backend setup
LAYER git@github.com:otter-layers/go-api.git TARGET backend
LAYER git@github.com:otter-layers/docker-compose.git

# Frontend setup
LAYER git@github.com:otter-layers/react-typescript.git TARGET frontend
LAYER git@github.com:otter-layers/tailwind-config.git TARGET frontend

# Development tools
LAYER git@github.com:otter-layers/vscode-settings.git TARGET .vscode
LAYER git@github.com:otter-layers/cursor-rules.git TARGET .cursor
```

## Development

### Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized builds)
- Make (for using the Makefile targets)

### Building and Testing

```bash
# Install dependencies
make deps

# Run linting
make lint

# Run tests
make test

# Build binary
make build

# Build for all platforms
make build-all

# Run example workflow
make run-example
```

### Docker Development

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# Get shell access in container
make docker-shell
```

## CI/CD

This project uses GitHub Actions for continuous integration and deployment:

- **Lint workflow** (`.github/workflows/lint.yml`): Runs on every push and PR
  - Go formatting checks
  - go vet
  - golangci-lint
  - Unit tests with coverage
- **Build workflow** (`.github/workflows/build.yml`): Builds for multiple platforms
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
  - Creates GitHub releases on tags
  - Builds and pushes Docker images

### Release Process

1. Create and push a git tag: `git tag v1.0.0 && git push origin v1.0.0`
2. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release with binaries
   - Build and push Docker image to GitHub Container Registry

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linting: `make lint`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

The CI/CD pipeline will automatically run tests and linting on your PR.
