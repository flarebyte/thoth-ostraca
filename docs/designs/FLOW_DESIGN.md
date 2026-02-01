# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
thoth CLI root command
  Parse args for meta pipeline
    Load action config file
    Route by action type
      Meta pipeline flow
        Find *.thoth.yaml files
        Parse and validate YAML records
        Apply filter predicate
        Apply map transform
        Execute shell per mapped item
        Post-map shell results
        Apply reduce aggregate
        Write JSON result (array/value/lines)
      Create meta files flow
        Find files recursively (gitignore)
        Filter filenames
        Map filenames
        Post-map from files
        Save meta files (*.thoth.yaml)
        Write JSON result (array/value/lines)
      Update meta files flow
        Find files recursively (update)
        Filter filenames
        Map filenames
        Load existing meta (if any)
        Post-map for update (with existing)
        Update meta files (merge/create)
        Write JSON result (array/value/lines)
      Diff meta files flow
        Find files recursively (update)
        Filter filenames
        Map filenames
        Load existing meta (if any)
        Post-map for update (with existing)
        Compute meta diffs
        Detect orphan meta files
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
  - Create many meta files
  - Update many meta files
  - Diff meta files at scale

Unsupported use cases (yet):



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
  - Commands: thoth meta (single pipeline incl. optional shell and create)
  - Flags: --config (YAML preferred; JSON accepted), --save (enable saving in create)
  - Tests: golden tests for I/O; fs testdata fixtures
  - Reduce: outputs a plain JSON value
  - Map: returns free-form JSON (any)
  - Shells: support bash, sh, zsh early
  - Create flow: discover files (gitignore), filter/map/post-map over {file}
  - Save writer: if save.enabled or --save, write *.thoth.yaml
  - Filename: <sha256[:12]>-<lastdir>-<filename>.thoth.yaml
  - Hash input: discovery relPath for stability
  - On exists: ignore (default) or error
  - Update flow: discover files, load existing meta if present, shallow-merge patch, create if missing
  - Merge strategy: shallow merge (new keys override)
  - Diff flow: same as update until patch; compute deep diff; do not write
  - Orphans: scan existing meta files; if locator path missing on disk, report
  - Diff output: RFC 6902 JSON Patch per item + summary (created/modified/deleted/orphan/unchanged)
  - internal/diff: generate patches and optional before/after snapshots for debugging

## Action Config (JSON Example)
```json
{
  "configVersion": "1",
  "action": "pipeline",
  "discovery": {
    "root": ".",
    "noGitignore": false
  },
  "workers": 8,
  "filter": {
    "inline": "-- keep records with meta.enabled == true\nreturn (meta and meta.enabled) == true"
  },
  "map": {
    "inline": "-- project selected fields\nreturn { locator = locator, name = meta and meta.name }"
  },
  "shell": {
    "enabled": true,
    "program": "bash",
    "commandTemplate": "echo {value}",
    "workingDir": ".",
    "env": {
      "CI": "true"
    },
    "timeoutMs": 60000,
    "failFast": true,
    "capture": {
      "stdout": true,
      "stderr": true,
      "maxBytes": 1048576
    }
  },
  "postMap": {
    "inline": "-- summarize shell result\nreturn { locator = locator, exit = shell.exitCode }"
  },
  "reduce": {
    "inline": "-- count items\nreturn (acc or 0) + 1"
  },
  "output": {
    "lines": false,
    "pretty": false,
    "out": "-"
  }
}
```

## Action Config (Create Example)
```json
{
  "configVersion": "1",
  "action": "create",
  "discovery": {
    "root": ".",
    "noGitignore": false
  },
  "workers": 8,
  "filter": {
    "inline": "-- only process markdown files\nreturn string.match(file.ext or \"\", \"^%.md$\") ~= nil"
  },
  "map": {
    "inline": "-- produce initial meta from file info\nreturn { title = file.base, category = file.dir }"
  },
  "postMap": {
    "inline": "-- finalize meta shape\nreturn { meta = { title = (input.title or file.base) } }"
  },
  "output": {
    "lines": false,
    "pretty": false,
    "out": "-"
  },
  "save": {
    "enabled": false,
    "onExists": "ignore"
  }
}
```

## Action Config (Create Minimal Example)
```json
{
  "configVersion": "1",
  "action": "create",
  "discovery": {
    "root": ".",
    "noGitignore": false
  },
  "filter": {
    "inline": "return true"
  },
  "map": {
    "inline": "return { meta = { created = true } }"
  },
  "output": {
    "lines": false,
    "pretty": true,
    "out": "-"
  },
  "save": {
    "enabled": false,
    "onExists": "ignore",
    "hashLen": 12
  }
}
```

## Action Config (Diff Example)
```json
{
  "configVersion": "1",
  "action": "diff",
  "discovery": {
    "root": ".",
    "noGitignore": false
  },
  "workers": 8,
  "filter": {
    "inline": "-- example: only .json files\nreturn string.match(file.ext or \"\", \"^%.json$\") ~= nil"
  },
  "map": {
    "inline": "-- compute desired meta fields from filename\nreturn { category = file.dir }"
  },
  "output": {
    "lines": false,
    "pretty": true,
    "out": "-"
  }
}
```

## Lua Data Contracts
  - Filter: fn({ locator, meta }) -> bool
  - Map: fn({ locator, meta }) -> any
  - Reduce: fn(acc, value) -> acc (single JSON value)
  - Post-map (shell): fn({ locator, input, shell: { cmd, exitCode, stdout, stderr, durationMs } }) -> any
  - Create Filter: fn({ file: { path, relPath, dir, base, name, ext } }) -> bool
  - Create Map: fn({ file }) -> any
  - Create Post-map: fn({ file, input }) -> { meta }
  - Update Post-map: fn({ file, input, existing? }) -> { meta } (patch)

## Diff Output Shape
  - Per-item result: { file, status, patch?, before?, after? }
  - status ∈ { created, modified, deleted, unchanged, orphan }
  - patch: RFC 6902 JSON Patch array (ops: add/remove/replace/move/copy/test)
  - before/after: optional full meta snapshots for debugging (disabled by default)
  - Top-level summary: counts per status and totals

## Lua Builtins
  - locator.kind(locator) -> 'file' | 'url'
  - locator.to_file_path(locator, root) -> string|nil (nil for URLs)
  - url.is_url(s) -> bool (http/https schemes)

## Go Package Outline
  - cmd/thoth: cobra wiring, --config parsing, action routing
  - internal/config: load/validate YAML (inline Lua strings), defaults
  - internal/fs: walk with gitignore, file info struct ({path, relPath, dir, base, name, ext})
  - internal/meta: YAML read/write of {locator, meta}
  - internal/lua: gopher-lua helpers to run inline scripts with typed inputs
  - internal/pipeline: stages (filter/map/shell/post-map/reduce), worker pool
  - internal/shell: exec with capture, timeouts, env, sh/bash/zsh
  - internal/save: filename builder (<sha256[:12]>-<lastdir>-<filename>.thoth.yaml), onExists policy
  - internal/diff: RFC6902 patch generation + item summary

## Design Decisions
  - Filter: Lua-only (v1)
  - Map: free-form JSON (any)
  - Reduce: plain JSON value
  - Output: machine-oriented JSON by default (aggregate unless --lines)
  - Gitignore: always on; --no-gitignore to opt out
  - Workers: default = CPU count (overridable via --workers)
  - YAML: error on missing required fields (locator, meta)
  - Shells: bash, sh, zsh supported early
  - Save filename: sha256 prefix length = 12 by default

## Open Design Questions
  - YAML strictness for unknown fields: error or ignore?
