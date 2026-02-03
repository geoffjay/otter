# Structured Format Plan: Block Syntax for Layer Configuration

## Overview

This document outlines a plan for implementing a structured block syntax using `{}` for complex layer configurations.
This would provide an alternative to the current inline syntax with `BEFORE`/`AFTER` keywords and line continuation with
`\`.

## Proposed Syntax

```dockerfile
LAYER git@github.com:otter-layers/database.git TARGET config {
  before: ["chmod +x scripts/*.sh", "mkdir -p config"],
  after: ["./scripts/db-setup.sh", "go mod tidy"]
}
```

### Full Syntax Specification

```dockerfile
LAYER <repository> [TARGET <path>] [IF <condition>] [TEMPLATE <key=value>...] {
  before: [<command-array>],
  after: [<command-array>],
  # Future extensibility:
  # validate: { files_exist: [...], commands_work: [...] },
  # cache: "checksum" | "always" | "never",
  # timeout: 300,
  # retry: 3
}
```

## Comparison: Current vs. Proposed

### Current Inline Syntax

```dockerfile
LAYER git@github.com:example/repo.git \
  TARGET config \
  IF env=production \
  BEFORE ["mkdir -p config"] \
  AFTER ["./scripts/setup.sh", "go mod tidy"]
```

**Pros:**
- Simple, Dockerfile-like syntax
- No special parsing for block structures
- Familiar to users of shell scripts and Dockerfiles

**Cons:**
- Long lines even with continuation
- Limited extensibility for future features
- Hooks visually blend with other parameters

### Proposed Block Syntax

```dockerfile
LAYER git@github.com:example/repo.git TARGET config IF env=production {
  before: ["mkdir -p config"],
  after: ["./scripts/setup.sh", "go mod tidy"]
}
```

**Pros:**
- Clear visual separation of hooks from layer definition
- Easier to read for complex configurations
- More extensible for future features (validation, caching, timeouts)
- JSON-like structure familiar to modern developers
- Better IDE support potential (syntax highlighting, folding)

**Cons:**
- More complex parser implementation
- Two ways to do the same thing (inline vs block)
- Multi-line parsing complexity increases

## Implementation Approach

### Phase 1: Parser Enhancement

1. **Detect block opening**: When parsing a LAYER command, check if line ends with `{`
2. **Multi-line block collection**: Continue reading lines until matching `}`
3. **Block content parsing**: Parse the collected content as JSON-like key-value pairs

```go
// Pseudocode for block detection
func parseLayerCommand(args []string, config *OtterfileConfig, scanner *bufio.Scanner) error {
    // Check if last arg is "{" or ends with "{"
    if hasBlockStart(args) {
        blockContent, err := collectBlockContent(scanner)
        if err != nil {
            return err
        }
        return parseLayerWithBlock(args, blockContent, config)
    }
    // Fall back to current inline parsing
    return parseLayerInline(args, config)
}
```

### Phase 2: Block Content Parser

The block content would be parsed as a simplified JSON-like structure:

```go
type LayerBlock struct {
    Before   []string          `json:"before,omitempty"`
    After    []string          `json:"after,omitempty"`
    // Future fields:
    // Validate *ValidationConfig `json:"validate,omitempty"`
    // Cache    string            `json:"cache,omitempty"`
    // Timeout  int               `json:"timeout,omitempty"`
}
```

### Phase 3: Validation and Error Handling

- Balanced brace checking
- Unknown key detection with helpful errors
- Type validation for values
- Line number tracking for errors within blocks

## Technical Considerations

### 1. Scanner Modification

The current parser uses line-by-line scanning. Block syntax requires either:
- **Option A**: Buffer lines until block closes
- **Option B**: Switch to character-by-character parsing for blocks
- **Option C**: Pre-process file to normalize blocks to single lines

Recommendation: **Option A** - Buffer lines until block closes, similar to how line continuation works now.

### 2. JSON vs Custom Parser

For the block content:
- **JSON**: Use `encoding/json` directly (requires strict JSON syntax)
- **Relaxed JSON**: Allow trailing commas, unquoted keys (requires custom parser or library)
- **YAML subset**: More readable but adds dependency complexity

Recommendation: Start with **strict JSON** for simplicity, consider relaxed JSON later.

### 3. Backward Compatibility

- Inline syntax (`BEFORE [...]`, `AFTER [...]`) must continue working
- Block syntax is additive, not replacing
- Same Layer struct used internally regardless of syntax

### 4. Error Messages

Block syntax requires enhanced error messages:

```
Error on line 15-18: failed to parse layer block
  at line 17: unknown key "befor" (did you mean "before"?)
```

## Migration Path

1. **v1.x**: Current inline syntax only
2. **v2.0**: Add block syntax support (both syntaxes work)
3. **v2.x**: Recommend block syntax in documentation for complex layers
4. **v3.0**: Consider deprecating inline hooks (or keep both forever)

## Future Extensibility

The block syntax enables future features without syntax changes:

```dockerfile
LAYER git@github.com:example/database.git {
  before: ["./pre-check.sh"],
  after: ["./post-setup.sh"],

  # Future: Validation
  validate: {
    files_exist: ["schema.sql", "migrations/"],
    commands_work: ["psql --version"]
  },

  # Future: Caching control
  cache: "checksum",

  # Future: Execution control
  timeout: 300,
  retry: 3,
  continue_on_error: false,

  # Future: Dependencies
  depends_on: ["base-layer"],

  # Future: Metadata
  description: "Database configuration layer",
  version: ">=1.0.0"
}
```

## Estimated Effort

| Task | Complexity | Estimated Time |
|------|------------|----------------|
| Block detection in parser | Low | 1-2 hours |
| Block content collection | Medium | 2-3 hours |
| Block content parsing (JSON) | Low | 1-2 hours |
| Error handling enhancement | Medium | 2-3 hours |
| Tests for new syntax | Medium | 2-3 hours |
| Documentation updates | Low | 1-2 hours |
| **Total** | | **9-15 hours** |

## Decision Points

Before implementation, decide:

1. **JSON strictness**: Require strict JSON or allow relaxed syntax (trailing commas, unquoted keys)?
2. **Nesting depth**: Allow only flat key-value or nested structures for future features?
3. **Mixed syntax**: Allow both inline and block hooks in same LAYER, or mutually exclusive?
4. **Validation timing**: Validate block keys at parse time or execution time?

## Recommendation

Implement block syntax as an **optional enhancement** that coexists with inline syntax:

1. Start with strict JSON parsing for simplicity
2. Keep inline syntax for simple cases (single hook, short commands)
3. Recommend block syntax for complex configurations (multiple hooks, long commands)
4. Design block parser with extensibility in mind for future features

The implementation can be done incrementally without breaking existing Otterfiles.
