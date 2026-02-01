# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
thoth CLI root command
  Parse args for meta pipeline
    Load action config file (optional)
    Find *.thoth.yaml files
    Parse and validate YAML records
    Apply filter predicate
    Apply map transform
    Execute shell per mapped item
    Post-map shell results
    Apply reduce aggregate
    Write JSON result (array/value/lines)
```

Supported use cases:

  - Helpful, well-documented flags
  - JSON output for CLI/CI/AI
  - Load action config file
  - Respect .gitignore by default
  - One file per locator
  - Validate {locator, meta} schema
  - Locators as file path or URL
  - Filter meta by locator
  - Script filter/map/reduce
  - Process in parallel
  - Map meta records
  - Run shell using map output
  - Reduce across meta set

Unsupported use cases (yet):

  - Create many meta files
  - Update many meta files
  - Diff meta files at scale

## Suggested Go Implementation
  - Module: go 1.22; command name: thoth
  - CLI: cobra for command tree; viper optional
  - Types: type Record struct { Locator string; Meta map[string]any }
  - YAML: gopkg.in/yaml.v3 for *.thoth.yaml
  - Discovery: filepath.WalkDir + gitignore filter (go-gitignore)
  - Schema: required fields (locator, meta); error on missing
  - Filter/Map/Reduce: Lua scripts only (gopher-lua) for v1
  - Parallelism: bounded worker pool; default workers = runtime.NumCPU()
  - Output: aggregated JSON by default; --lines to stream; --pretty for humans
  - Commands: thoth meta (single pipeline incl. optional shell)
  - Flags: --root, --pattern, --no-gitignore, --workers, --filter-script, --map-script, --reduce-script, --run-shell, --shell, --post-map-script, --fail-fast, --capture-stdout, --capture-stderr, --config, --out
  - Tests: golden tests for I/O; fs testdata fixtures
  - Reduce: outputs a plain JSON value
  - Map: returns free-form JSON (any)
  - Shells: support bash, sh, zsh early

## Lua Data Contracts
  - Filter: fn({ locator, meta }) -> bool
  - Map: fn({ locator, meta }) -> any
  - Reduce: fn(acc, value) -> acc (single JSON value)
  - Post-map (shell): fn({ locator, input, shell: { cmd, exitCode, stdout, stderr, durationMs } }) -> any

## Design Decisions
  - Filter: Lua-only (v1)
  - Map: free-form JSON (any)
  - Reduce: plain JSON value
  - Output: machine-oriented JSON by default (aggregate unless --lines)
  - Gitignore: always on; --no-gitignore to opt out
  - Workers: default = CPU count (overridable via --workers)
  - YAML: error on missing required fields (locator, meta)
  - Shells: bash, sh, zsh supported early

## Open Design Questions
  - YAML strictness for unknown fields: error or ignore?
