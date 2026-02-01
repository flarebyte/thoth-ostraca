# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
thoth CLI root command
  Parse args for meta find
    Find *.thoth.yaml files
      Parse and validate YAML records
        Apply filter predicate
          Write JSON (pretty/compact/lines)
```

Supported use cases:

  - Helpful, well-documented flags
  - JSON output for humans/CI/AI
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
  - Filter/Map/Reduce: built-in funcs + optional gopher-lua scripts
  - Parallelism: bounded worker pool; channels for records
  - Output: JSON lines (default), pretty via --pretty, compact via --compact
  - Commands: thoth find, thoth map, thoth reduce, thoth run (shell)
  - Flags: --root, --pattern, --ignore, --workers, --script, --out
  - Tests: golden tests for I/O; fs testdata fixtures

## Open Design Questions
  - Filter expression: prefer small DSL or go with Lua first?
  - Map output shape: free-form any vs constrained fields?
  - Reduce outputs: single JSON value vs object with metadata?
  - Default output: JSON lines or pretty JSON when writing to TTY?
  - Gitignore behavior: always on, or opt-out flag --no-gitignore?
