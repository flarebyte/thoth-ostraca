# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
thoth CLI root command
  Parse args for meta find
    Find *.thoth.yaml files
      Parse and validate YAML records
        Apply filter predicate
          Write JSON result (array/value/lines)
```

Supported use cases:

  - Helpful, well-documented flags
  - JSON output for CLI/CI/AI
  - Respect .gitignore by default
  - One file per locator
  - Validate {locator, meta} schema
  - Filter meta by locator
  - Script filter/map/reduce

Unsupported use cases (yet):

  - Map meta records
  - Reduce across meta set
  - Load action config file
  - Run shell using map output
  - Locators as file path or URL
  - Process in parallel
  - Create many meta files
  - Update many meta files
  - Diff meta files at scale

## Suggested Go Implementation
  - Module: go 1.22; command name: thoth
  - CLI: cobra for command tree; viper optional
  - Types: type Record struct { Locator string; Meta map[string]any }
  - YAML: gopkg.in/yaml.v3 for *.thoth.yaml
  - Discovery: filepath.WalkDir + gitignore filter (go-gitignore)
  - Schema: validate locator non-empty; meta is object
  - Filter/Map/Reduce: Lua scripts only (gopher-lua) for v1
  - Parallelism: bounded worker pool; channels for records
  - Output: aggregated JSON by default; --lines to stream; --pretty for humans
  - Commands: thoth find, thoth map, thoth reduce, thoth run (shell)
  - Flags: --root, --pattern, --no-gitignore, --workers, --script, --out
  - Tests: golden tests for I/O; fs testdata fixtures
  - Reduce: outputs a plain JSON value
  - Map: returns free-form JSON (any)

## Design Decisions
  - Filter: Lua-only (v1)
  - Map: free-form JSON (any)
  - Reduce: plain JSON value
  - Output: machine-oriented JSON by default (aggregate unless --lines)
  - Gitignore: always on; --no-gitignore to opt out

## Open Design Questions
  - Default worker pool size and tuning flags?
  - YAML schema strictness (unknown fields: error or ignore)?
  - Which shells to support for 'run' besides bash?
