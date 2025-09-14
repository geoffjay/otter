# Layer-Specific .otterignore Example

This example demonstrates how `.otterignore` files work with layers:

## How It Works

1. **Project-level `.otterignore`**: Ignores files across ALL layers
2. **Layer-specific `.otterignore`**: Ignores files only within that specific layer
3. **Combined filtering**: Files are ignored if they match EITHER pattern set
4. **`.otterignore` files themselves**: Always ignored (never copied to target)

## File Structure

```
layer-ignore-example/
├── .otterignore              # Project-level ignores: README.md, project-specific.txt
├── Otterfile                 # Defines the layers
├── test-layer/
│   ├── .otterignore          # Layer ignores: LICENSE, *.tmp
│   ├── LICENSE               # ❌ Ignored by layer .otterignore
│   ├── README.md             # ❌ Ignored by project .otterignore
│   ├── FOO.md                # ✅ Copied (not ignored by either)
│   ├── temp.tmp              # ❌ Ignored by layer *.tmp pattern
│   └── config.yaml           # ✅ Copied (not ignored by either)
└── another-layer/
    ├── .otterignore          # Layer ignores: secrets.env, *.bak
    ├── secrets.env           # ❌ Ignored by layer .otterignore
    └── app.config            # ✅ Copied (not ignored by either)
```

## Expected Results

When you run `otter build`, only these files will be copied:

- `output/FOO.md` (from test-layer)
- `output/config.yaml` (from test-layer)
- `more-output/app.config` (from another-layer)

## Usage

```bash
# Initialize the project
otter init

# Apply the layers (this will respect both project and layer .otterignore files)
otter build

# Check the results
ls output/          # Should contain: FOO.md, config.yaml
ls more-output/     # Should contain: app.config
```

## Files That Should Be Ignored

- `LICENSE` - ignored by `test-layer/.otterignore`
- `README.md` - ignored by project `.otterignore`
- `temp.tmp` - ignored by `test-layer/.otterignore` (\*.tmp pattern)
- `secrets.env` - ignored by `another-layer/.otterignore`
- `.otterignore` files themselves - always ignored

This demonstrates the exact behavior you requested: layer-specific ignore patterns are combined with project-level patterns for comprehensive file filtering!
