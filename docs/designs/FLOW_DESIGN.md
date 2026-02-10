# FLOW DESIGN OVERVIEW (Generated)

## Function calls tree

```
thoth CLI root command
  Parse args for run
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
        Enrich files with OS/Git info
        Filter filenames
        Map filenames
        Post-map from files
        Save meta files (*.thoth.yaml)
        Write JSON result (array/value/lines)
      Update meta files flow
        Find files recursively (update)
        Enrich files with OS/Git info
        Filter filenames
        Map filenames
        Load existing meta (if any)
        Post-map for update (with existing)
        Update meta files (merge/create)
        Write JSON result (array/value/lines)
      Diff meta files flow
        Find files recursively (update)
        Enrich files with OS/Git info
        Filter filenames
        Map filenames
        Load existing meta (if any)
        Post-map for update (with existing)
        Compute meta diffs
        Detect orphan meta files
        Write JSON result (array/value/lines)
      Validate meta files only
        Find *.thoth.yaml files
        Parse and validate YAML records
        Collect validation results
        Write JSON result (array/value/lines)
```

Supported use cases:

  - Helpful, well-documented flags
  - JSON output for CLI/CI/AI — Machine-oriented default; aggregated JSON; lines optional
  - Load action config file — Prefer YAML; allow JSON
  - Respect .gitignore by default — Always on; opt-out via --no-gitignore
  - One file per locator — Minimize merge conflicts
  - Validate {locator, meta} schema — Required fields: locator:string, meta:object; error on missing
  - Locators as file path or URL
  - Filter meta by locator — boolean predicate over {locator, meta}
  - Script filter/map/reduce — Lua only (v1): small + popular
  - Process in parallel — Goroutines + channels; bounded pool; default workers = CPU count
  - Map meta records — transform {locator, meta} → any
  - Run shell using map output — Support bash, sh, zsh early
  - Reduce across meta set — aggregate stream → single result
  - Create many meta files
  - Expose os.FileInfo for inputs — Include size, mode, modTime, isDir for filtering/mapping when enabled
  - Expose Git metadata for inputs — Use go-git to provide tracked/ignored, worktree status, and last commit info when enabled
  - Update many meta files
  - Diff meta files at scale
  - Validate meta files only — No transforms or shell; emit validation report


Unsupported use cases (yet):





## Suggested Go Implementation
  - Module: go 1.22; command name: thoth
  - CLI: cobra for command tree; viper optional
  - Types: type Record struct { Locator string; Meta map[string]any }
  - YAML: gopkg.in/yaml.v3 for *.thoth.yaml
  - Discovery: filepath.WalkDir + gitignore filter (go-gitignore); apply .gitignore even if not a git repo; do not follow symlinks by default
  - Schema: required fields (locator, meta); error on missing
  - Validation defaults: unknown top-level keys: error; meta.* keys: allowed
  - Validation config: validation.allowUnknownTopLevel (bool, default false)
  - Config schema: Ship JSON Schema at docs/schema/thoth.config.schema.json; validate YAML/JSON on load
  - Config versioning: configVersion string (e.g., '1'); breaking changes bump major; unknown version -> error
  - Config loader: unknown fields: error by default; allow lenient via env THOTH_CONFIG_LENIENT=true (dev only)
  - Filter/Map/Reduce: Lua scripts only (gopher-lua) for v1
  - Lua sandbox: Enable base/table/string/math; disable os/io/coroutine/debug by default, No filesystem/network access; no os.execute/io.popen, Per-script timeout + instruction limit via VM hooks, Deterministic math.random by default; configurable seed, Expose helpers under 'thoth.*' namespace (no global pollution)
  - Parallelism: bounded worker pool; default workers = runtime.NumCPU()
  - Output: aggregated JSON by default; --lines to stream; --pretty for humans
  - Ordering: Aggregated (array): sort deterministically by locator (pipeline) or relPath (create/update/diff), Lines: nondeterministic (parallel), each line is independent JSON value
  - Errors: Policy: errors.mode keep-going|fail-fast (default keep-going), Embed: errors.embedErrors=true includes per-item error objects; final exit non-zero if any error, Parse/validation errors: reported per-item when possible; fatal config/load errors abort early
  - Commands: thoth run (exec action config: pipeline/create/update/diff)
  - Flags: --config (YAML preferred; JSON accepted), --save (enable saving in create)
  - Tests: golden tests for I/O; fs testdata fixtures
  - Reduce: Lua fn(acc, value) -> acc; initial acc=nil (Lua sees nil), Applies in deterministic order (locator/relPath sort), Any JSON-serializable acc allowed (object/array/number/string/bool/null)
  - Map: returns free-form JSON (any)
  - Shells: support bash, sh, zsh early
  - Create flow: discover files (gitignore), filter/map/post-map over {file}
  - Save writer: if save.enabled or --save, write *.thoth.yaml
  - Filename: <sha256[:15]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml
  - Hash input: canonical root (CWD-based) + POSIX relPath
  - On exists: ignore (default) or error
  - Locator canonicalization: files: repo-relative POSIX path; use path.Clean + separator normalization, URLs: net/url parse; lowercase scheme/host; drop default ports; strip fragment, to_file_path: filepath.Join(root, posix->OS); reject absolute and '..' by default
  - Update flow: discover files, load existing meta if present, shallow-merge patch, create if missing
  - Merge strategy: config.update.merge: shallow|deep|jsonpatch (default shallow), shallow: replace top-level keys (objects); arrays replaced entirely, deep: recursive merge for objects; arrays replaced (v1), jsonpatch: apply user-provided RFC6902 patch from post-map { patch }, post-map may return { meta } (full desired) or { patch }; when both provided, { patch } takes precedence
  - Diff flow: same as update until patch; compute deep diff; do not write
  - Orphans: scan existing meta files; if locator path missing on disk, report
  - Diff output: RFC 6902 JSON Patch per item + summary (modified/unchanged/missing/orphan)
  - internal/diff: generate patches and optional before/after snapshots for debugging
  - Diff config: includeSnapshots (bool), output: patch|summary|both (default: both)
  - Exit codes: 0: success (no errors), 1: fatal setup/validation error (no output), 2: partial failures (some per-item errors present), 3: script/reduce failure (pipeline aborted)

## Exit Codes
  - 0: success (no errors)
  - 1: fatal setup/validation error (no output)
  - 2: partial failures (some per-item errors present)
  - 3: script/reduce failure (pipeline aborted)

## Ordering & Determinism
  - Aggregated output (array): deterministic sort
  - Sort key: 'locator' for pipeline; 'file.relPath' for create/update/diff
  - Reduce: consumes values in the same deterministic order as the aggregated array
  - Streaming (--lines): order is nondeterministic due to parallelism; each line is independent JSON value

## Action Config (YAML Example)
```yaml
configVersion: "1"
action: pipeline
discovery:
  root: .
  noGitignore: false
  followSymlinks: false
workers: 8
errors:
  mode: keep-going
  embedErrors: true
lua:
  timeoutMs: 2000
  instructionLimit: 1000000
  memoryLimitBytes: 8388608
  libs:
    base: true
    table: true
    string: true
    math: true
  allowOSExecute: false
  allowEnv: false
  deterministicRandom: true
validation:
  allowUnknownTopLevel: false
locatorPolicy:
  allowAbsolute: false
  allowParentRefs: false
  posixStyle: true
filter:
  inline: |-
    -- keep records with meta.enabled == true
    return (meta and meta.enabled) == true
map:
  inline: |-
    -- project selected fields
    return { locator = locator, name = meta and meta.name }
shell:
  enabled: true
  program: bash
  argsTemplate:
    - echo
    - "{json}"
  workingDir: .
  env:
    CI: "true"
  timeoutMs: 60000
  failFast: true
  capture:
    stdout: true
    stderr: true
    maxBytes: 1048576
  strictTemplating: true
  killProcessGroup: true
  termGraceMs: 2000
postMap:
  inline: |-
    -- summarize shell result
    return { locator = locator, exit = shell.exitCode }
reduce:
  inline: |-
    -- count items
    return (acc or 0) + 1
output:
  lines: false
  pretty: false
  out: "-"
```

## Action Config (Create Example)
```yaml
configVersion: "1"
action: create
discovery:
  root: .
  noGitignore: false
files:
  info: true
  git: true
workers: 8
filter:
  inline: |-
    -- only process markdown files
    return string.match(file.ext or "", "^%.md$") ~= nil
map:
  inline: |-
    -- produce initial meta from file info
    return { title = file.base, category = file.dir }
postMap:
  inline: |-
    -- finalize meta shape
    return { meta = { title = (input.title or file.base) } }
output:
  lines: false
  pretty: false
  out: "-"
save:
  enabled: false
  onExists: ignore
```

## Action Config (Create Minimal Example)
```yaml
configVersion: "1"
action: create
discovery:
  root: .
  noGitignore: false
filter:
  inline: return true
map:
  inline: return { meta = { created = true } }
output:
  lines: false
  pretty: true
  out: "-"
save:
  enabled: false
  onExists: ignore
  hashLen: 15
```

## Action Config (Diff Example)
```yaml
configVersion: "1"
action: diff
discovery:
  root: .
  noGitignore: false
workers: 8
errors:
  mode: keep-going
  embedErrors: true
filter:
  inline: |-
    -- example: only .json files
    return string.match(file.ext or "", "^%.json$") ~= nil
map:
  inline: |-
    -- compute desired meta fields from filename
    return { category = file.dir }
diff:
  includeSnapshots: false
  output: both
output:
  lines: false
  pretty: true
  out: "-"
```

## Action Config (Lua Limits Example)
```yaml
configVersion: "1"
action: pipeline
discovery:
  root: .
  noGitignore: false
  followSymlinks: false
workers: 4
errors:
  mode: keep-going
  embedErrors: true
lua:
  timeoutMs: 500
  instructionLimit: 100000
  memoryLimitBytes: 2097152
  libs:
    base: true
    table: true
    string: true
    math: true
  deterministicRandom: true
  randomSeed: 1234
filter:
  inline: return true
map:
  inline: return { locator = locator, ok = true }
output:
  lines: false
  pretty: true
  out: "-"
```

## Lua Data Contracts
  - Filter: fn({ locator, meta }) -> bool
  - Map: fn({ locator, meta }) -> any
  - Reduce: fn(acc, value) -> acc (single JSON value)
  - Post-map (shell): fn({ locator, input, shell: { cmd, exitCode, stdout, stderr, durationMs } }) -> any
  - Create Filter: fn({ file: { path, relPath, dir, base, name, ext } }) -> bool
  - Create Map: fn({ file }) -> any
  - Create Post-map: fn({ file, input }) -> { meta }
  - Update Post-map: fn({ file, input, existing? }) -> { meta } | { patch } (RFC6902)

## Lua Input Examples
  - pipeline.filter/map: { locator = "path/or/url", meta = { ... } }
  - pipeline.post-map(shell): { locator, input = <map result>, shell = { cmd, exitCode, stdout, stderr, durationMs } }
  - create.filter/map: { file = { path, relPath, dir, base, name, ext } }
  - update.post-map: { file, input = <map result>, existing = { locator, meta }? }
  - diff.post-map: { file, input = <map result>, existing = { locator, meta }? }

## Reduce Behavior
  - pipeline.reduce: accumulates over map or post-map(shell) results in deterministic order (sorted by locator); returns a single JSON value
  - create.reduce (optional): accumulates over post-map results in deterministic order (sorted by file.relPath); dry-run friendly
  - update.reduce (optional): accumulates over post-map patches/simulated results in deterministic order (sorted by file.relPath)
  - acc initialization: starts as nil in Lua (use 'acc or <default>'); any JSON-serializable value allowed
  - when reduce is present: output is a single JSON value; --lines is ignored
  - diff: reduce not applicable (summary auto-generated)

## Error Handling
  - Modes: errors.mode = 'keep-going' (default) or 'fail-fast'
  - Keep-going: continue other items; embed per-item errors when errors.embedErrors=true; exit non-zero if any error
  - Fail-fast: stop processing on first error; still emit any already-produced results; exit non-zero
  - Per-item error shape: { error: { stage, code, message, details? }, context: { locator?|file? } }
  - Reduce receives only successful items; if all fail, reduce is skipped and an error is returned
  - Config/load-level errors: abort immediately (no output beyond an error message)

## Result Shapes
  - Aggregated (array): list of items with consistent envelope for CI parsing
  - Success item: { status: 'ok', context: { locator?|file? }, value: any, shell? }
  - Error item: { status: 'error', context: { locator?|file? }, error: { stage, code, message, details? } }
  - Lines (--lines): each line is a success or error item as above
  - Diff action: uses Diff Output Shape for success items; errors follow the error item schema

## Diff Output Shape
  - Per-item result: { file, status, patch?, before?, after? }
  - status ∈ { modified, unchanged, missing, orphan }
  - missing: file exists but no meta found (previously 'created')
  - orphan: meta exists but locator file is missing
  - patch: RFC 6902 JSON Patch array (ops: add/remove/replace/move/copy/test) transforming before -> after
  - before: existing meta object (if any); after: desired meta after applying patch
  - Top-level summary: counts per status and totals

## Update Merge Strategy
  - config.update.merge: 'shallow' | 'deep' | 'jsonpatch' (default 'shallow')
  - shallow: replace top-level keys; arrays replaced entirely
  - deep: recursive merge for objects; arrays replaced (v1 semantics)
  - jsonpatch: apply RFC6902 operations from post-map { patch }
  - Post-map return: may return { meta } (full desired) or { patch }; if both present, { patch } is applied
  - Validation: patch must apply cleanly; otherwise per-item error

## Lua Builtins (thoth namespace)
  - thoth.locator.kind(locator) -> 'file' | 'url'
  - thoth.locator.normalize(locator, root?) -> string (canonical: file=repo-relative POSIX path; url=lowercase scheme/host, strip default port)
  - thoth.locator.to_file_path(locator, root) -> string|nil (nil for URLs; validates policy; cleans and joins; rejects absolute and '..' by default)
  - thoth.path.clean_posix(s) -> string (collapse '.', remove redundant '/', no '..')
  - thoth.url.is_url(s) -> bool (http/https schemes)
  - thoth.lua.version() -> string (Lua VM version)
  - thoth.runtime.limits() -> { timeoutMs, instructionLimit, memoryLimitBytes }

## Locator Normalization
  - File locators: canonical form is repo-relative POSIX-style path (no leading './', '/' forbidden by default)
  - Disallow '..' segments and absolute paths by default (config.locatorPolicy controls exceptions)
  - Normalization: collapse '.', remove duplicate '/', convert OS separators to '/' for storage
  - URL locators: lowercase scheme and host; strip default ports (http:80, https:443); preserve path/query; remove fragment
  - locator.to_file_path: returns OS-native absolute path under 'root' after validation and clean join
  - Security: reject traversal (..), absolute inputs, and non-http(s) URLs by default

## Discovery Semantics
  - .gitignore: honored by default even when not in a git repo (local .gitignore files are parsed)
  - Symlinks: do not follow by default (discovery.followSymlinks=false)
  - Exclusions: no magic exclusions beyond .gitignore rules

## Lua Execution Environment
  - Allowed libs: base, table, string, math (by default)
  - Disabled libs: os, io, coroutine, debug (by default)
  - Filesystem/Network: not accessible from Lua (only via thoth.* helpers)
  - Env: not accessible by default; may be allowed via lua.allowEnv + envAllowlist
  - Timeouts: per-script timeout (lua.timeoutMs) and instructionLimit enforced via VM hooks
  - Memory: soft limit (lua.memoryLimitBytes); abort on large allocations when feasible
  - Randomness: math.random seeded deterministically by default; override via lua.randomSeed; set deterministicRandom=false to use time-based seed
  - Helpers: exposed under thoth.* namespace; avoid global pollution

## Shell Execution Spec
  - Templating: placeholders {name} with optional transforms {name|json} and {name|sh}
  - Placeholders: {value} (map result, string only), {json} (JSON of map result), {locator}, {index}, {file.path}, {file.relPath}, {file.dir}, {file.base}, {file.name}, {file.ext}
  - Strict mode (default): unknown placeholders -> error; {value} must be string or use {value|json}
  - Escaping: in commandTemplate (string), all placeholders are shell-escaped by default; {..|sh} forces quoting explicitly
  - Security: prefer argsTemplate (argv form) to avoid shell parsing; each arg templated independently
  - Timeout: on timeout, send SIGTERM to process group, wait termGraceMs, then SIGKILL; killProcessGroup=true by default
  - Exit codes: non-zero → record error; if failFast=true, abort remaining work
  - Env: explicit env entries merged with process env; no implicit interpolation in templates (use {env.VAR} not supported v1)

## Function calls details

```
thoth CLI root command [cli.root]
  - note: cobra-based command tree
  - pkg: cmd/thoth
  - func: CliRoot
  - file: cmd/thoth/cli_root.go
  Parse args for run [cli.run]
    - note: flags: --config (YAML preferred; JSON accepted). All other options belong in the action config.
    - pkg: cmd/thoth
    - func: CliRun
    - file: cmd/thoth/cli_run.go
    Load action config file [action.config.load]
      - note: --config path; YAML preferred; JSON accepted; drives entire pipeline
      - pkg: internal/config
      - func: ActionConfigLoad
      - file: internal/config/config_load.go
    Route by action type [action.route]
      - note: action: pipeline | create | update | diff
      - pkg: internal/config
      - func: ActionRoute
      - file: internal/config/action_route.go
      Meta pipeline flow [flow.pipeline]
        - pkg: internal/pipeline
        - func: FlowPipeline
        - file: internal/pipeline/flow_pipeline.go
        Find *.thoth.yaml files [fs.discovery]
          - note: walk root; .gitignore ON by default even outside git repos; --no-gitignore to disable; do not follow symlinks by default
          - pkg: internal/fs
          - func: FsDiscovery
          - file: internal/fs/fs_discovery.go
        Parse and validate YAML records [meta.parse]
          - note: yaml.v3; strict fields; types; locator canonicalization; top-level unknown = error (unless validation.allowUnknownTopLevel); inside meta: unknown allowed
          - pkg: internal/meta
          - func: MetaParse
          - file: internal/meta/meta_parse.go
        Apply filter predicate [meta.filter.step]
          - note: Lua-only predicate (v1)
          - pkg: internal/pipeline
          - func: MetaFilterStep
          - file: internal/pipeline/meta_filter_step.go
        Apply map transform [meta.map.step]
          - note: Lua-only mapping (v1); parallel by default
          - pkg: internal/pipeline
          - func: MetaMapStep
          - file: internal/pipeline/meta_map_step.go
        Execute shell per mapped item [shell.exec]
          - note: Conditional: --run-shell; argv templates preferred (no shell parsing); string templates auto-escape; supports bash/sh/zsh; parallel with bounded workers; feeds post-map/reduce; timeout kills process group
          - pkg: internal/shell
          - func: ShellExec
          - file: internal/shell/shell_exec.go
        Post-map shell results [meta.map.post-shell]
          - note: Conditional: --post-map-script; Lua transforms {locator,input,shell:{cmd,exitCode,stdout,stderr,durationMs}}
          - pkg: internal/pipeline
          - func: MetaMapPostShell
          - file: internal/pipeline/meta_post_shell.go
        Apply reduce aggregate [meta.reduce.step]
          - note: Lua-only reduce (v1); parallel feed; single JSON value
          - pkg: internal/pipeline
          - func: MetaReduceStep
          - file: internal/pipeline/meta_reduce_step.go
        Write JSON result (array/value/lines) [output.json.result]
          - note: default: aggregated JSON array (sorted by locator/relPath); --lines streams nondeterministically; reduce → single value; embed per-item errors when configured
          - pkg: internal/output
          - func: OutputJsonResult
          - file: internal/output/json_result.go
      Create meta files flow [flow.create]
        - pkg: internal/pipeline
        - func: FlowCreate
        - file: internal/pipeline/flow_create.go
        Find files recursively (gitignore) [fs.discovery.files]
          - note: walk root; .gitignore ON by default (even if not a git repo); no patterns; do not follow symlinks by default; filenames as inputs
          - pkg: internal/fs
          - func: FsDiscoveryFiles
          - file: internal/fs/discovery_files.go
        Enrich files with OS/Git info [files.enrich]
          - note: Conditional: files.info and/or files.git; attach file.info (os.Stat) and file.git (go-git status/last commit)
          - pkg: internal/pipeline
          - func: FilesEnrich
          - file: internal/pipeline/files_enrich.go
        Filter filenames [files.filter.step]
          - note: Lua-only predicate (v1) over {file}
          - pkg: internal/pipeline
          - func: FilesFilterStep
          - file: internal/pipeline/files_filter_step.go
        Map filenames [files.map.step]
          - note: Lua-only map (v1) over {file}
          - pkg: internal/pipeline
          - func: FilesMapStep
          - file: internal/pipeline/files_map_step.go
        Post-map from files [files.map.post]
          - note: Conditional: inline Lua transforms {file,input} -> any
          - pkg: internal/pipeline
          - func: FilesMapPost
          - file: internal/pipeline/files_map_post.go
        Save meta files (*.thoth.yaml) [meta.save]
          - note: Conditional: config.save.enabled or --save; name = <sha256[:15]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml; sanitize components; if path exists and belongs to different locator -> error; onExists: ignore|error
          - pkg: internal/save
          - func: MetaSave
          - file: internal/save/meta_save.go
        Write JSON result (array/value/lines) [output.json.result]
          - note: default: aggregated JSON array (sorted by locator/relPath); --lines streams nondeterministically; reduce → single value; embed per-item errors when configured
          - pkg: internal/output
          - func: OutputJsonResult
          - file: internal/output/json_result.go
      Update meta files flow [flow.update]
        - pkg: internal/pipeline
        - func: FlowUpdate
        - file: internal/pipeline/flow_update.go
        Find files recursively (update) [fs.discovery.files.update]
          - note: walk root; .gitignore ON by default (even if not a git repo); do not follow symlinks by default; filenames as inputs
          - pkg: internal/fs
          - func: FsDiscoveryFilesUpdate
          - file: internal/fs/files_update.go
        Enrich files with OS/Git info [files.enrich]
          - note: Conditional: files.info and/or files.git; attach file.info (os.Stat) and file.git (go-git status/last commit)
          - pkg: internal/pipeline
          - func: FilesEnrich
          - file: internal/pipeline/files_enrich.go
        Filter filenames [files.filter.step]
          - note: Lua-only predicate (v1) over {file}
          - pkg: internal/pipeline
          - func: FilesFilterStep
          - file: internal/pipeline/files_filter_step.go
        Map filenames [files.map.step]
          - note: Lua-only map (v1) over {file}
          - pkg: internal/pipeline
          - func: FilesMapStep
          - file: internal/pipeline/files_map_step.go
        Load existing meta (if any) [meta.load.existing]
          - note: compute expected path by naming convention; read YAML if exists
          - pkg: internal/meta
          - func: MetaLoadExisting
          - file: internal/meta/load_existing.go
        Post-map for update (with existing) [files.map.post.update]
          - note: Lua receives {file,input,existing?}; returns either { meta } (full desired) or { patch } (RFC6902)
          - pkg: internal/pipeline
          - func: FilesMapPostUpdate
          - file: internal/pipeline/files_post_update.go
        Update meta files (merge/create) [meta.update]
          - note: merge strategy via config.update.merge: shallow|deep|jsonpatch (default shallow); if post-map returns patch, apply RFC6902; else merge existing with returned meta; missing -> create new by naming convention; verify filename hash against current root+relPath (mismatch -> error)
          - pkg: internal/save
          - func: MetaUpdate
          - file: internal/save/meta_update.go
        Write JSON result (array/value/lines) [output.json.result]
          - note: default: aggregated JSON array (sorted by locator/relPath); --lines streams nondeterministically; reduce → single value; embed per-item errors when configured
          - pkg: internal/output
          - func: OutputJsonResult
          - file: internal/output/json_result.go
      Diff meta files flow [flow.diff]
        - pkg: internal/pipeline
        - func: FlowDiff
        - file: internal/pipeline/flow_diff.go
        Find files recursively (update) [fs.discovery.files.update]
          - note: walk root; .gitignore ON by default (even if not a git repo); do not follow symlinks by default; filenames as inputs
          - pkg: internal/fs
          - func: FsDiscoveryFilesUpdate
          - file: internal/fs/files_update.go
        Enrich files with OS/Git info [files.enrich]
          - note: Conditional: files.info and/or files.git; attach file.info (os.Stat) and file.git (go-git status/last commit)
          - pkg: internal/pipeline
          - func: FilesEnrich
          - file: internal/pipeline/files_enrich.go
        Filter filenames [files.filter.step]
          - note: Lua-only predicate (v1) over {file}
          - pkg: internal/pipeline
          - func: FilesFilterStep
          - file: internal/pipeline/files_filter_step.go
        Map filenames [files.map.step]
          - note: Lua-only map (v1) over {file}
          - pkg: internal/pipeline
          - func: FilesMapStep
          - file: internal/pipeline/files_map_step.go
        Load existing meta (if any) [meta.load.existing]
          - note: compute expected path by naming convention; read YAML if exists
          - pkg: internal/meta
          - func: MetaLoadExisting
          - file: internal/meta/load_existing.go
        Post-map for update (with existing) [files.map.post.update]
          - note: Lua receives {file,input,existing?}; returns either { meta } (full desired) or { patch } (RFC6902)
          - pkg: internal/pipeline
          - func: FilesMapPostUpdate
          - file: internal/pipeline/files_post_update.go
        Compute meta diffs [meta.diff.compute]
          - note: deep diff existing vs patch-applied result; output RFC6902 JSON Patch + summary
          - pkg: internal/diff
          - func: MetaDiffCompute
          - file: internal/diff/diff_compute.go
        Detect orphan meta files [meta.diff.orphans]
          - note: iterate *.thoth.yaml; if locator is file path and does not exist, flag
          - pkg: internal/diff
          - func: MetaDiffOrphans
          - file: internal/diff/diff_orphans.go
        Write JSON result (array/value/lines) [output.json.result]
          - note: default: aggregated JSON array (sorted by locator/relPath); --lines streams nondeterministically; reduce → single value; embed per-item errors when configured
          - pkg: internal/output
          - func: OutputJsonResult
          - file: internal/output/json_result.go
      Validate meta files only [flow.validate]
        - pkg: internal/pipeline
        - func: FlowValidate
        - file: internal/pipeline/flow_validate.go
        Find *.thoth.yaml files [fs.discovery]
          - note: walk root; .gitignore ON by default even outside git repos; --no-gitignore to disable; do not follow symlinks by default
          - pkg: internal/fs
          - func: FsDiscovery
          - file: internal/fs/fs_discovery.go
        Parse and validate YAML records [meta.parse]
          - note: yaml.v3; strict fields; types; locator canonicalization; top-level unknown = error (unless validation.allowUnknownTopLevel); inside meta: unknown allowed
          - pkg: internal/meta
          - func: MetaParse
          - file: internal/meta/meta_parse.go
        Collect validation results [meta.validate.only]
          - note: Schema + locator checks only; no filter/map/reduce/shell
          - pkg: internal/pipeline
          - func: MetaValidateOnly
          - file: internal/pipeline/validate_only.go
        Write JSON result (array/value/lines) [output.json.result]
          - note: default: aggregated JSON array (sorted by locator/relPath); --lines streams nondeterministically; reduce → single value; embed per-item errors when configured
          - pkg: internal/output
          - func: OutputJsonResult
          - file: internal/output/json_result.go
```

## Action Script Scope
```
Action     Input                      Filter     Map        Post-Map     Reduce     Output                                    
------------------------------------------------------------------------------------------------------------------------------
pipeline   { locator, meta }          Lua (yes)  Lua (yes)  Lua (shell)  Lua (yes)  array of records or single value (reduce) 
create     { file }                   Lua (yes)  Lua (yes)  Lua (yes)    Lua (opt)  array of post-map results; save if enabled
update     { file, existing? }        Lua (yes)  Lua (yes)  Lua (patch|meta) Lua (opt)  array of updates (dry-run) or write changes
diff       { file, existing? }        Lua (yes)  Lua (yes)  Lua (patch)  N/A        patch list (RFC6902) + summary; orphans flagged
validate   { locator, meta }          N/A        N/A        N/A          N/A        validation report array                   
```

## Pure helper functions

```
Compute POSIX relPath from root+path [fs.relpath.posix]
  - note: pure: joins, cleans, converts to '/' separators
  - pkg: internal/fs
  - func: FsRelpathPosix
  - file: internal/fs/relpath_posix.go
Sanitize filename components [locator.component.sanitize]
  - note: pure: lowercase, replace invalid chars, collapse dashes
  - pkg: internal/save
  - func: LocatorComponentSanitize
  - file: internal/save/component_sanitize.go
Compute sha256 prefix + optional rootTag [locator.hash.tag]
  - note: pure: hash over canonical root + relPath; prefix length configurable
  - pkg: internal/save
  - func: LocatorHashTag
  - file: internal/save/hash_tag.go
Build meta filename from relPath [save.build_filename]
  - note: pure: <sha256[:N]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml
  - pkg: internal/save
  - func: SaveBuildFilename
  - file: internal/save/build_filename.go
Sort records by locator [output.sort.records]
  - note: pure: stable, deterministic order for pipeline
  - pkg: internal/output
  - func: OutputSortRecords
  - file: internal/output/sort_records.go
Sort files by relPath [output.sort.files]
  - note: pure: stable, deterministic order for create/update/diff
  - pkg: internal/output
  - func: OutputSortFiles
  - file: internal/output/sort_files.go
Shallow merge objects [merge.shallow]
  - note: pure: replace top-level keys; arrays replaced entirely
  - pkg: internal/save
  - func: MergeShallow
  - file: internal/save/merge_shallow.go
Deep merge objects [merge.deep]
  - note: pure: recursive object merge; arrays replaced (v1)
  - pkg: internal/save
  - func: MergeDeep
  - file: internal/save/merge_deep.go
Compute RFC6902 patch from before/after [diff.jsonpatch.from]
  - note: pure: deterministic patch generation
  - pkg: internal/diff
  - func: DiffJsonpatchFrom
  - file: internal/diff/jsonpatch_from.go
Apply argv placeholders with transforms [template.args.substitute]
  - note: pure: resolve {json},{locator},{file.*}, enforce strictTemplating
  - pkg: internal/shell
  - func: TemplateArgsSubstitute
  - file: internal/shell/args_substitute.go
Validate top-level meta schema [validate.meta.top_level]
  - note: pure: 'locator' string, 'meta' object; unknown top-level guard
  - pkg: internal/meta
  - func: ValidateMetaTopLevel
  - file: internal/meta/top_level.go
```

## Go Package Outline
  - cmd/thoth: cobra wiring, --config parsing, action routing
  - internal/config: load/validate YAML (inline Lua strings), defaults
  - internal/fs: walk with gitignore, file info struct ({path, relPath, dir, base, name, ext} + optional {size, mode, modTime, isDir} when files.info=true; optional Git via go-git when files.git=true)
  - internal/git: repository detection + file status and last-commit via go-git (files.git=true)
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
  - Save filename: sha256 prefix length = 15 by default

## Filename Collision & Stability
  - Sanitization: lowercase ASCII; replace non [a-z0-9._-] with '-', collapse repeats; trim '-'
  - Format: <sha256[:15]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml (rootTag=hash of canonical root when root!='.')
  - Hash input: canonical root (CWD-based) + POSIX relPath; stable across OS; renames change hash
  - Collision: extremely unlikely; if computed path exists but locator differs -> error (do not overwrite)
  - Root changes: recommended to keep root at '.'; if different, include rootTag and enforce hash match; otherwise error
  - Orphans: renames create new meta file; detection handled by orphan scan in diff flow

## Schema Validation
  - Top-level: required keys 'locator' (string, non-empty) and 'meta' (object)
  - Top-level: unknown keys -> error by default; can allow via validation.allowUnknownTopLevel = true
  - Meta object: unknown keys are allowed (user data)
  - Locator: accept file paths (relative/absolute) and URLs (http/https)

## Config Schema & Versioning
  - Format: YAML preferred; JSON accepted (schema is JSON Schema)
  - Versioning: configVersion: '1'; breaking changes bump major (e.g., '2'); unknown version -> error
  - Unknown fields: error by default; dev-only lenient mode via env THOTH_CONFIG_LENIENT=true (ignored fields)
  - Defaults: applied at load time; loader returns normalized config (no runtime mutation)
  - Inline Lua (YAML): use block scalar (|) to avoid quoting/escaping issues; mind indentation

## YAML Tips (Inline Lua)
Example:

```yaml
configVersion: '1'
action: pipeline
filter:
  inline: |
    -- keep when meta.enabled is true
    return (meta and meta.enabled) == true
map:
  inline: |
    return { locator = locator, name = meta and meta.name }
```

## Stage Contracts
  - Record: struct { Locator string; Meta map[string]any }
  - FileInfo: struct { Path, RelPath, Dir, Base, Name, Ext string } + optional { Size int64; Mode os.FileMode; Mod time.Time; IsDir bool } when files.info=true
  - GitInfo (file.git when files.git=true): struct { Tracked bool; Ignored bool; LastCommitSha string; LastAuthorName string; LastAuthorEmail string; LastCommitTime time.Time; Status string; WorktreeStatus string; StagingStatus string }
  - ShellResult: struct { Cmd []string; ExitCode int; Stdout []byte; Stderr []byte; Duration time.Duration }
  - JSONPatch: []PatchOp (RFC6902)
  - MetaOut: struct { Meta map[string]any }
  - UpdateOut: struct { Meta map[string]any; Patch JSONPatch } (one of)
  - 
  - MetaFilter: func(Record) (bool, error)
  - MetaMap: func(Record) (any, error)
  - MetaPostShell: func(Record, ShellResult) (any, error)
  - MetaReduce: func(acc any, value any) (any, error)
  - 
  - FilesFilter: func(FileInfo) (bool, error)
  - FilesMap: func(FileInfo) (any, error)
  - FilesPostMap: func(FileInfo, input any) (MetaOut, error)
  - FilesPostUpdate: func(FileInfo, input any, existing *Record) (UpdateOut, error)

## Open Design Questions
  - YAML strictness for unknown fields: error or ignore?
