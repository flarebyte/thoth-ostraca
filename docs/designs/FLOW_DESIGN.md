# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
thoth CLI root command
  Parse args for meta find
    Find *.thoth.yaml files
      Parse and validate YAML records
        Apply filter predicate
          Write JSON result (array/value/lines)
  Parse args for meta map
    Find *.thoth.yaml files (map)
      Parse and validate YAML (map)
        Load action config file (optional)
        Apply map transform
          Write JSON result (array/value/lines)
  Parse args for meta reduce
    Find *.thoth.yaml files (reduce)
      Parse and validate YAML (reduce)
        Load action config file (optional)
        Apply reduce aggregate
          Write JSON result (array/value/lines)
  Parse args for run (shell)
    Find *.thoth.yaml files (run)
      Parse and validate YAML (run)
        Load action config file (optional)
        Map for run (shell input)
          Execute shell per mapped item
```

Supported use cases:

  - Helpful, well-documented flags
  - JSON output for CLI/CI/AI
  - Respect .gitignore by default
  - One file per locator
  - Validate {locator, meta} schema
  - Locators as file path or URL
  - Filter meta by locator
  - Script filter/map/reduce
  - Process in parallel
  - Load action config file
  - Map meta records
  - Reduce across meta set
  - Run shell using map output

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
  - Commands: thoth find, thoth map, thoth reduce, thoth run (shell)
  - Flags: --root, --pattern, --no-gitignore, --workers, --script, --out
  - Tests: golden tests for I/O; fs testdata fixtures
  - Reduce: outputs a plain JSON value
  - Map: returns free-form JSON (any)
  - Shells: support bash, sh, zsh early

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
