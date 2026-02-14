import { calls } from './calls.mts';
import {
  appendKeyValueList,
  appendSection,
  appendToReport,
  appendUseCases,
  type ComponentCall,
  displayCallsAsText,
  displayCallsDetailed,
  getSetDifference,
  resetReport,
  toUseCaseSet,
} from './common.mts';
import { cliRoot } from './flows.mts';
import { suggestFor } from './suggestions.mts';
import { mustUseCases, useCaseCatalogByName } from './use_cases.mts';

// CUE is the preferred format for action configs; examples are embedded below

const GO_PACKAGE_OUTLINE: string[] = [
  'cmd/thoth: cobra wiring, --config parsing, action routing',
  'internal/config: load/validate CUE (inline Lua strings), defaults',
  'internal/fs: walk with gitignore, file info struct ({path, relPath, dir, base, name, ext} + optional {size, mode, modTime, isDir} when files.info=true; optional Git via go-git when files.git=true)',
  'internal/git: repository detection + file status and last-commit via go-git (files.git=true)',
  'internal/meta: YAML read/write of {locator, meta}',
  'internal/lua: gopher-lua helpers to run inline scripts with typed inputs',
  'internal/pipeline: stages (filter/map/shell/post-map/reduce), worker pool',
  'internal/shell: exec with capture, timeouts, env, sh/bash/zsh',
  'internal/save: filename builder (<sha256[:12]>-<lastdir>-<filename>.thoth.yaml), onExists policy',
  'internal/diff: RFC6902 patch generation + item summary',
];

const DESIGN_DECISIONS: string[] = [
  'Filter: Lua-only (v1)',
  'Map: free-form JSON (any)',
  'Reduce: plain JSON value',
  'Output: machine-oriented JSON by default (aggregate unless --lines)',
  'Gitignore: always on; --no-gitignore to opt out',
  'Workers: default = CPU count (overridable via --workers)',
  'YAML: error on missing required fields (locator, meta)',
  'Shells: bash, sh, zsh supported early',
  'Save filename: sha256 prefix length = 15 by default',
];

const SUGGESTED_GO_IMPLEMENTATION: Array<[string, string | string[]]> = [
  ['Module', 'go 1.22; command name: thoth'],
  ['CLI', 'cobra for command tree; viper optional'],
  ['Types', 'type Record struct { Locator string; Meta map[string]any }'],
  ['YAML', 'gopkg.in/yaml.v3 for *.thoth.yaml (meta records)'],
  [
    'Discovery',
    'filepath.WalkDir + gitignore filter (go-gitignore); apply .gitignore even if not a git repo; do not follow symlinks by default',
  ],
  ['Schema', 'required fields (locator, meta); error on missing'],
  [
    'Validation defaults',
    'unknown top-level keys: error; meta.* keys: allowed',
  ],
  [
    'Validation config',
    'validation.allowUnknownTopLevel (bool, default false)',
  ],
  [
    'Config schema',
    'Ship CUE schema at docs/schema/thoth.config.cue; validate .cue on load',
  ],
  [
    'Config versioning',
    "configVersion string (e.g., '1'); breaking changes bump major; unknown version -> error",
  ],
  [
    'Config loader',
    'unknown fields: error by default; allow lenient via env THOTH_CONFIG_LENIENT=true (dev only)',
  ],
  ['Filter/Map/Reduce', 'Lua scripts only (gopher-lua) for v1'],
  [
    'Lua sandbox',
    [
      'Enable base/table/string/math; disable os/io/coroutine/debug by default',
      'No filesystem/network access; no os.execute/io.popen',
      'Per-script timeout + instruction limit via VM hooks',
      'Deterministic math.random by default; configurable seed',
      "Expose helpers under 'thoth.*' namespace (no global pollution)",
    ],
  ],
  ['Parallelism', 'bounded worker pool; default workers = runtime.NumCPU()'],
  [
    'Output',
    'aggregated JSON by default; --lines to stream; --pretty for humans',
  ],
  [
    'Ordering',
    [
      'Aggregated (array): sort deterministically by locator (pipeline) or relPath (create/update/diff)',
      'Lines: nondeterministic (parallel), each line is independent JSON value',
    ],
  ],
  [
    'Errors',
    [
      'Policy: errors.mode keep-going|fail-fast (default keep-going)',
      'Embed: errors.embedErrors=true includes per-item error objects; final exit non-zero if any error',
      'Parse/validation errors: reported per-item when possible; fatal config/load errors abort early',
    ],
  ],
  [
    'Commands',
    'thoth run (exec action config: pipeline/create/update/diff); thoth diagnose (execute one stage)',
  ],
  ['Flags', '--config (.cue CUE file), --save (enable saving in create)'],
  ['Tests', 'golden tests for I/O; fs testdata fixtures'],
  [
    'Reduce',
    [
      'Lua fn(acc, value) -> acc; initial acc=nil (Lua sees nil)',
      'Applies in deterministic order (locator/relPath sort)',
      'Any JSON-serializable acc allowed (object/array/number/string/bool/null)',
    ],
  ],
  ['Map', 'returns free-form JSON (any)'],
  ['Shells', 'support bash, sh, zsh early'],
  [
    'Create flow',
    'discover files (gitignore), filter/map/post-map over {file}',
  ],
  ['Save writer', 'if save.enabled or --save, write *.thoth.yaml'],
  ['Filename', '<sha256[:15]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml'],
  ['Hash input', 'canonical root (CWD-based) + POSIX relPath'],
  ['On exists', 'ignore (default) or error'],
  [
    'Locator canonicalization',
    [
      'files: repo-relative POSIX path; use path.Clean + separator normalization',
      'URLs: net/url parse; lowercase scheme/host; drop default ports; strip fragment',
      "to_file_path: filepath.Join(root, posix->OS); reject absolute and '..' by default",
    ],
  ],
  [
    'Update flow',
    'discover files, load existing meta if present, shallow-merge patch, create if missing',
  ],
  [
    'Merge strategy',
    [
      'config.update.merge: shallow|deep|jsonpatch (default shallow)',
      'shallow: replace top-level keys (objects); arrays replaced entirely',
      'deep: recursive merge for objects; arrays replaced (v1)',
      'jsonpatch: apply user-provided RFC6902 patch from post-map { patch }',
      'post-map may return { meta } (full desired) or { patch }; when both provided, { patch } takes precedence',
    ],
  ],
  ['Diff flow', 'same as update until patch; compute deep diff; do not write'],
  [
    'Orphans',
    'scan existing meta files; if locator path missing on disk, report',
  ],
  [
    'Diff output',
    'RFC 6902 JSON Patch per item + summary (modified/unchanged/missing/orphan)',
  ],
  [
    'internal/diff',
    'generate patches and optional before/after snapshots for debugging',
  ],
  [
    'Diff config',
    'includeSnapshots (bool), output: patch|summary|both (default: both)',
  ],
  [
    'Exit codes',
    [
      '0: success (no errors)',
      '1: fatal setup/validation error (no output)',
      '2: partial failures (some per-item errors present)',
      '3: script/reduce failure (pipeline aborted)',
    ],
  ],
];

export const generateFlowDesignReport = async () => {
  // Build tree and header
  cliRoot({ level: 0 });
  await resetReport();
  await appendToReport('# FLOW DESIGN OVERVIEW (Generated)\n');
  await appendToReport('## Function calls tree\n');
  await appendToReport('```');
  await displayCallsAsText(calls);
  await appendToReport('```\n');

  // Use-case coverage
  await appendUseCases(
    'Supported use cases:',
    toUseCaseSet(calls),
    useCaseCatalogByName,
  );
  await appendUseCases(
    'Unsupported use cases (yet):',
    getSetDifference(mustUseCases, toUseCaseSet(calls)),
    useCaseCatalogByName,
  );

  // Implementation guidance and examples
  await appendKeyValueList(
    'Suggested Go Implementation',
    SUGGESTED_GO_IMPLEMENTATION,
  );
  await appendKeyValueList('Exit Codes', [
    ['0', 'success (no errors)'],
    ['1', 'fatal setup/validation error (no output)'],
    ['2', 'partial failures (some per-item errors present)'],
    ['3', 'script/reduce failure (pipeline aborted)'],
  ]);
  await appendSection('Ordering & Determinism', [
    'Aggregated output (array): deterministic sort',
    "Sort key: 'locator' for pipeline; 'file.relPath' for create/update/diff",
    'Reduce: consumes values in the same deterministic order as the aggregated array',
    'Streaming (--lines): order is nondeterministic due to parallelism; each line is independent JSON value',
  ]);
  await appendSection(
    'Action Config (CUE Example)',
    '```cue\n' +
      [
        'configVersion: "1"',
        'action: "pipeline"',
        'discovery: { root: ".", noGitignore: false, followSymlinks: false }',
        'workers: 8',
        'errors: { mode: "keep-going", embedErrors: true }',
        'lua: { timeoutMs: 2000, instructionLimit: 1000000, memoryLimitBytes: 8388608, libs: { base: true, table: true, string: true, math: true }, deterministicRandom: true }',
        'validation: { allowUnknownTopLevel: false }',
        'locatorPolicy: { allowAbsolute: false, allowParentRefs: false, posixStyle: true }',
        'filter: { inline: "return (meta and meta.enabled) == true" }',
        'map:    { inline: "return { locator = locator, name = meta and meta.name }" }',
        'shell:  { enabled: true, program: "bash", argsTemplate: ["echo", "{json}"], workingDir: ".", env: { CI: "true" }, timeoutMs: 60000, failFast: true, capture: { stdout: true, stderr: true, maxBytes: 1048576 }, strictTemplating: true, killProcessGroup: true, termGraceMs: 2000 }',
        'postMap:{ inline: "return { locator = locator, exit = shell.exitCode }" }',
        'reduce: { inline: "return (acc or 0) + 1" }',
        'output: { lines: false, pretty: false, out: "-" }',
      ].join('\n') +
      '\n```',
  );
  await appendSection(
    'Action Config (Create Example, CUE)',
    '```cue\n' +
      [
        'configVersion: "1"',
        'action: "create"',
        'discovery: { root: ".", noGitignore: false }',
        'files: { info: true, git: true }',
        'workers: 8',
        'filter: { inline: "return string.match(file.ext or "", "^%.md$") ~= nil" }',
        'map:    { inline: "return { title = file.base, category = file.dir }" }',
        'postMap:{ inline: "return { meta = { title = (input.title or file.base) } }" }',
        'output: { lines: false, pretty: false, out: "-" }',
        'save:   { enabled: false, onExists: "ignore" }',
      ].join('\n') +
      '\n```',
  );
  await appendSection(
    'Action Config (Create Minimal, CUE)',
    '```cue\n' +
      [
        'configVersion: "1"',
        'action: "create"',
        'discovery: { root: ".", noGitignore: false }',
        'filter: { inline: "return true" }',
        'map:    { inline: "return { meta = { created = true } }" }',
        'output: { lines: false, pretty: true, out: "-" }',
        'save:   { enabled: false, onExists: "ignore", hashLen: 15 }',
      ].join('\n') +
      '\n```',
  );
  await appendSection(
    'Action Config (Diff Example, CUE)',
    '```cue\n' +
      [
        'configVersion: "1"',
        'action: "diff"',
        'discovery: { root: ".", noGitignore: false }',
        'workers: 8',
        'errors: { mode: "keep-going", embedErrors: true }',
        'filter: { inline: "return string.match(file.ext or "", "^%.json$") ~= nil" }',
        'map:    { inline: "return { category = file.dir }" }',
        'diff:   { includeSnapshots: false, output: "both" }',
        'output: { lines: false, pretty: true, out: "-" }',
      ].join('\n') +
      '\n```',
  );
  await appendSection(
    'Action Config (Lua Limits, CUE)',
    '```cue\n' +
      [
        'configVersion: "1"',
        'action: "pipeline"',
        'discovery: { root: ".", noGitignore: false, followSymlinks: false }',
        'workers: 4',
        'errors: { mode: "keep-going", embedErrors: true }',
        'lua: { timeoutMs: 500, instructionLimit: 100000, memoryLimitBytes: 2097152, libs: { base: true, table: true, string: true, math: true }, deterministicRandom: true, randomSeed: 1234 }',
        'filter: { inline: "return true" }',
        'map:    { inline: "return { locator = locator, ok = true }" }',
        'output: { lines: false, pretty: true, out: "-" }',
      ].join('\n') +
      '\n```',
  );

  // Narrative sections
  await appendSection('Lua Data Contracts', [
    'Filter: fn({ locator, meta }) -> bool',
    'Map: fn({ locator, meta }) -> any',
    'Reduce: fn(acc, value) -> acc (single JSON value)',
    'Post-map (shell): fn({ locator, input, shell: { cmd, exitCode, stdout, stderr, durationMs } }) -> any',
    'Create Filter: fn({ file: { path, relPath, dir, base, name, ext } }) -> bool',
    'Create Map: fn({ file }) -> any',
    'Create Post-map: fn({ file, input }) -> { meta }',
    'Update Post-map: fn({ file, input, existing? }) -> { meta } | { patch } (RFC6902)',
  ]);
  await appendSection('Diagnose Stage Boundary Types (Examples)', [
    'pipeline.filter/map input: { locator, meta }',
    'pipeline.shell input: <map output record>',
    'pipeline.post-map-shell input: { mapped, shell, locator? } (implementation-defined)',
    'create/update/diff discovery/enrich output: { file: { ... }, git?: { ... }, os?: { ... } }',
    'update/diff post-map input includes existing?: { locator, meta }',
    'Each stage declares input/output schema and stream type (items vs single value)',
  ]);
  await appendSection('Lua Input Examples', [
    'pipeline.filter/map: { locator = "path/or/url", meta = { ... } }',
    'pipeline.post-map(shell): { locator, input = <map result>, shell = { cmd, exitCode, stdout, stderr, durationMs } }',
    'create.filter/map: { file = { path, relPath, dir, base, name, ext } }',
    'update.post-map: { file, input = <map result>, existing = { locator, meta }? }',
    'diff.post-map: { file, input = <map result>, existing = { locator, meta }? }',
  ]);
  await appendSection('Reduce Behavior', [
    'pipeline.reduce: accumulates over map or post-map(shell) results in deterministic order (sorted by locator); returns a single JSON value',
    'create.reduce (optional): accumulates over post-map results in deterministic order (sorted by file.relPath); dry-run friendly',
    'update.reduce (optional): accumulates over post-map patches/simulated results in deterministic order (sorted by file.relPath)',
    "acc initialization: starts as nil in Lua (use 'acc or <default>'); any JSON-serializable value allowed",
    'when reduce is present: output is a single JSON value; --lines is ignored',
    'diff: reduce not applicable (summary auto-generated)',
  ]);
  await appendSection('Error Handling', [
    "Modes: errors.mode = 'keep-going' (default) or 'fail-fast'",
    'Keep-going: continue other items; embed per-item errors when errors.embedErrors=true; exit non-zero if any error',
    'Fail-fast: stop processing on first error; still emit any already-produced results; exit non-zero',
    'Per-item error shape: { error: { stage, code, message, details? }, context: { locator?|file? } }',
    'Reduce receives only successful items; if all fail, reduce is skipped and an error is returned',
    'Config/load-level errors: abort immediately (no output beyond an error message)',
  ]);
  await appendSection('Result Shapes', [
    'Aggregated (array): list of items with consistent envelope for CI parsing',
    "Success item: { status: 'ok', context: { locator?|file? }, value: any, shell? }",
    "Error item: { status: 'error', context: { locator?|file? }, error: { stage, code, message, details? } }",
    'Lines (--lines): each line is a success or error item as above',
    'Diff action: uses Diff Output Shape for success items; errors follow the error item schema',
  ]);
  await appendSection('Diff Output Shape', [
    'Per-item result: { file, status, patch?, before?, after? }',
    'status ∈ { modified, unchanged, missing, orphan }',
    "missing: file exists but no meta found (previously 'created')",
    'orphan: meta exists but locator file is missing',
    'patch: RFC 6902 JSON Patch array (ops: add/remove/replace/move/copy/test) transforming before -> after',
    'before: existing meta object (if any); after: desired meta after applying patch',
    'Top-level summary: counts per status and totals',
  ]);
  await appendSection('Update Merge Strategy', [
    "config.update.merge: 'shallow' | 'deep' | 'jsonpatch' (default 'shallow')",
    'shallow: replace top-level keys; arrays replaced entirely',
    'deep: recursive merge for objects; arrays replaced (v1 semantics)',
    'jsonpatch: apply RFC6902 operations from post-map { patch }',
    'Post-map return: may return { meta } (full desired) or { patch }; if both present, { patch } is applied',
    'Validation: patch must apply cleanly; otherwise per-item error',
  ]);
  await appendSection('Lua Builtins (thoth namespace)', [
    "thoth.locator.kind(locator) -> 'file' | 'url'",
    'thoth.locator.normalize(locator, root?) -> string (canonical: file=repo-relative POSIX path; url=lowercase scheme/host, strip default port)',
    "thoth.locator.to_file_path(locator, root) -> string|nil (nil for URLs; validates policy; cleans and joins; rejects absolute and '..' by default)",
    "thoth.path.clean_posix(s) -> string (collapse '.', remove redundant '/', no '..')",
    'thoth.url.is_url(s) -> bool (http/https schemes)',
    'thoth.lua.version() -> string (Lua VM version)',
    'thoth.runtime.limits() -> { timeoutMs, instructionLimit, memoryLimitBytes }',
  ]);
  await appendSection('Locator Normalization', [
    "File locators: canonical form is repo-relative POSIX-style path (no leading './', '/' forbidden by default)",
    "Disallow '..' segments and absolute paths by default (config.locatorPolicy controls exceptions)",
    "Normalization: collapse '.', remove duplicate '/', convert OS separators to '/' for storage",
    'URL locators: lowercase scheme and host; strip default ports (http:80, https:443); preserve path/query; remove fragment',
    "locator.to_file_path: returns OS-native absolute path under 'root' after validation and clean join",
    'Security: reject traversal (..), absolute inputs, and non-http(s) URLs by default',
  ]);
  await appendSection('Discovery Semantics', [
    '.gitignore: honored by default even when not in a git repo (local .gitignore files are parsed)',
    'Symlinks: do not follow by default (discovery.followSymlinks=false)',
    'Exclusions: no magic exclusions beyond .gitignore rules',
  ]);
  await appendSection('Lua Execution Environment', [
    'Allowed libs: base, table, string, math (by default)',
    'Disabled libs: os, io, coroutine, debug (by default)',
    'Filesystem/Network: not accessible from Lua (only via thoth.* helpers)',
    'Env: not accessible by default; may be allowed via lua.allowEnv + envAllowlist',
    'Timeouts: per-script timeout (lua.timeoutMs) and instructionLimit enforced via VM hooks',
    'Memory: soft limit (lua.memoryLimitBytes); abort on large allocations when feasible',
    'Randomness: math.random seeded deterministically by default; override via lua.randomSeed; set deterministicRandom=false to use time-based seed',
    'Helpers: exposed under thoth.* namespace; avoid global pollution',
  ]);
  await appendSection('Shell Execution Spec', [
    'Templating: placeholders {name} with optional transforms {name|json} and {name|sh}',
    'Placeholders: {value} (map result, string only), {json} (JSON of map result), {locator}, {index}, {file.path}, {file.relPath}, {file.dir}, {file.base}, {file.name}, {file.ext}',
    'Strict mode (default): unknown placeholders -> error; {value} must be string or use {value|json}',
    'Escaping: in commandTemplate (string), all placeholders are shell-escaped by default; {..|sh} forces quoting explicitly',
    'Security: prefer argsTemplate (argv form) to avoid shell parsing; each arg templated independently',
    'Timeout: on timeout, send SIGTERM to process group, wait termGraceMs, then SIGKILL; killProcessGroup=true by default',
    'Exit codes: non-zero → record error; if failFast=true, abort remaining work',
    'Env: explicit env entries merged with process env; no implicit interpolation in templates (use {env.VAR} not supported v1)',
  ]);

  // Detailed calls and action scope
  const detailedCalls = calls.map((c) => ({ ...c, suggest: suggestFor(c) }));
  await appendToReport('\n## Function calls details\n');
  await appendToReport('```');
  await displayCallsDetailed(detailedCalls);
  await appendToReport('```');

  const pad = (s: string, n: number) => s.padEnd(n);
  const actionRows: string[] = [];
  const header = [
    pad('Action', 10),
    pad('Input', 26),
    pad('Filter', 10),
    pad('Map', 10),
    pad('Post-Map', 12),
    pad('Reduce', 10),
    pad('Output', 42),
  ].join(' ');
  const sep = '-'.repeat(header.length);
  actionRows.push(header);
  actionRows.push(sep);
  actionRows.push(
    [
      pad('pipeline', 10),
      pad('{ locator, meta }', 26),
      pad('Lua (yes)', 10),
      pad('Lua (yes)', 10),
      pad('Lua (shell)', 12),
      pad('Lua (yes)', 10),
      pad('array of records or single value (reduce)', 42),
    ].join(' '),
  );
  actionRows.push(
    [
      pad('create', 10),
      pad('{ file }', 26),
      pad('Lua (yes)', 10),
      pad('Lua (yes)', 10),
      pad('Lua (yes)', 12),
      pad('Lua (opt)', 10),
      pad('array of post-map results; save if enabled', 42),
    ].join(' '),
  );
  actionRows.push(
    [
      pad('update', 10),
      pad('{ file, existing? }', 26),
      pad('Lua (yes)', 10),
      pad('Lua (yes)', 10),
      pad('Lua (patch|meta)', 12),
      pad('Lua (opt)', 10),
      pad('array of updates (dry-run) or write changes', 42),
    ].join(' '),
  );
  actionRows.push(
    [
      pad('diff', 10),
      pad('{ file, existing? }', 26),
      pad('Lua (yes)', 10),
      pad('Lua (yes)', 10),
      pad('Lua (patch)', 12),
      pad('N/A', 10),
      pad('patch list (RFC6902) + summary; orphans flagged', 42),
    ].join(' '),
  );
  actionRows.push(
    [
      pad('validate', 10),
      pad('{ locator, meta }', 26),
      pad('N/A', 10),
      pad('N/A', 10),
      pad('N/A', 12),
      pad('N/A', 10),
      pad('validation report array', 42),
    ].join(' '),
  );
  await appendSection(
    'Action Script Scope',
    '```\n' + actionRows.join('\n') + '\n```',
  );

  // Helper functions and remaining sections
  const helperCalls: ComponentCall[] = [
    {
      name: 'fs.relpath.posix',
      title: 'Compute POSIX relPath from root+path',
      note: "pure: joins, cleans, converts to '/' separators",
      directory: 'internal/fs',
      level: 0,
    },
    {
      name: 'locator.component.sanitize',
      title: 'Sanitize filename components',
      note: 'pure: lowercase, replace invalid chars, collapse dashes',
      directory: 'internal/save',
      level: 0,
    },
    {
      name: 'locator.hash.tag',
      title: 'Compute sha256 prefix + optional rootTag',
      note: 'pure: hash over canonical root + relPath; prefix length configurable',
      directory: 'internal/save',
      level: 0,
    },
    {
      name: 'save.build_filename',
      title: 'Build meta filename from relPath',
      note: 'pure: <sha256[:N]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml',
      directory: 'internal/save',
      level: 0,
    },
    {
      name: 'output.sort.records',
      title: 'Sort records by locator',
      note: 'pure: stable, deterministic order for pipeline',
      directory: 'internal/output',
      level: 0,
    },
    {
      name: 'output.sort.files',
      title: 'Sort files by relPath',
      note: 'pure: stable, deterministic order for create/update/diff',
      directory: 'internal/output',
      level: 0,
    },
    {
      name: 'merge.shallow',
      title: 'Shallow merge objects',
      note: 'pure: replace top-level keys; arrays replaced entirely',
      directory: 'internal/save',
      level: 0,
    },
    {
      name: 'merge.deep',
      title: 'Deep merge objects',
      note: 'pure: recursive object merge; arrays replaced (v1)',
      directory: 'internal/save',
      level: 0,
    },
    {
      name: 'diff.jsonpatch.from',
      title: 'Compute RFC6902 patch from before/after',
      note: 'pure: deterministic patch generation',
      directory: 'internal/diff',
      level: 0,
    },
    {
      name: 'template.args.substitute',
      title: 'Apply argv placeholders with transforms',
      note: 'pure: resolve {json},{locator},{file.*}, enforce strictTemplating',
      directory: 'internal/shell',
      level: 0,
    },
    {
      name: 'validate.meta.top_level',
      title: 'Validate top-level meta schema',
      note: "pure: 'locator' string, 'meta' object; unknown top-level guard",
      directory: 'internal/meta',
      level: 0,
    },
  ];
  const helperDetailedCalls: ComponentCall[] = helperCalls.map((c) => ({
    ...c,
    suggest: suggestFor(c),
  }));
  await appendToReport('\n## Pure helper functions\n');
  await appendToReport('```');
  await displayCallsDetailed(helperDetailedCalls);
  await appendToReport('```');

  await appendSection('Go Package Outline', GO_PACKAGE_OUTLINE);
  await appendSection('Design Decisions', DESIGN_DECISIONS);
  await appendSection('Diagnose Command', [
    'Subcommand: thoth diagnose',
    'Required flags: -c/--config <path> (CUE), --step <name>',
    "Input selection (mutually exclusive): --input-file <path> | --input-inline '<json>' | --input-stdin",
    'Default input mode: prepare upstream stages to boundary to generate realistic input',
    'Fixture capture: --dump-in [<path>|-], --dump-out [<path>|-] (JSON/NDJSON)',
    'Debug flags: --limit <N>, --seed <int>, --dry-shell (shell stage only)',
    'Semantics: execute only target step; upstream may run in prepare mode; no downstream',
    'Observability: emit header { action, executedStep, preparedStages, inputMode, limits } to stderr',
    "Streams: normal stage outputs -> stdout; diagnostics/logs/errors -> stderr; when dumping to '-', use it as stdout output",
    'Errors: non-zero on invalid config/step/json or stage failure; include action, step, input mode, item context',
  ]);
  await appendSection('Filename Collision & Stability', [
    "Sanitization: lowercase ASCII; replace non [a-z0-9._-] with '-', collapse repeats; trim '-'",
    "Format: <sha256[:15]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml (rootTag=hash of canonical root when root!='.')",
    'Hash input: canonical root (CWD-based) + POSIX relPath; stable across OS; renames change hash',
    'Collision: extremely unlikely; if computed path exists but locator differs -> error (do not overwrite)',
    "Root changes: recommended to keep root at '.'; if different, include rootTag and enforce hash match; otherwise error",
    'Orphans: renames create new meta file; detection handled by orphan scan in diff flow',
  ]);
  await appendSection('Schema Validation', [
    "Top-level: required keys 'locator' (string, non-empty) and 'meta' (object)",
    'Top-level: unknown keys -> error by default; can allow via validation.allowUnknownTopLevel = true',
    'Meta object: unknown keys are allowed (user data)',
    'Locator: accept file paths (relative/absolute) and URLs (http/https)',
  ]);
  await appendSection('Config Schema & Versioning', [
    'Format: CUE (.cue) preferred; schema lives at docs/schema/thoth.config.cue',
    'Validation: configs are evaluated and validated against CUE schema; unknown fields rejected by default',
    "Versioning: configVersion: '1'; breaking changes bump major (e.g., '2'); unknown version -> error",
    'Defaults: encoded directly in CUE schema; loader evaluates to a normalized config',
    'Inline Lua (CUE): use quoted strings for short code or triple-quoted strings for multi-line scripts',
  ]);
  await appendSection(
    'CUE Tips (Inline Lua)',
    'Example:\n\n```cue\n' +
      [
        'configVersion: "1"',
        'action: "pipeline"',
        'filter: { inline: "return (meta and meta.enabled) == true" }',
        'map:    { inline: "return { locator = locator, name = meta and meta.name }" }',
      ].join('\n') +
      '\n```',
  );
  await appendSection('Stage Contracts', [
    'Record: struct { Locator string; Meta map[string]any }',
    'FileInfo: struct { Path, RelPath, Dir, Base, Name, Ext string } + optional { Size int64; Mode os.FileMode; Mod time.Time; IsDir bool } when files.info=true',
    'GitInfo (file.git when files.git=true): struct { Tracked bool; Ignored bool; LastCommitSha string; LastAuthorName string; LastAuthorEmail string; LastCommitTime time.Time; Status string; WorktreeStatus string; StagingStatus string }',
    'ShellResult: struct { Cmd []string; ExitCode int; Stdout []byte; Stderr []byte; Duration time.Duration }',
    'JSONPatch: []PatchOp (RFC6902)',
    'MetaOut: struct { Meta map[string]any }',
    'UpdateOut: struct { Meta map[string]any; Patch JSONPatch } (one of)',
    '',
    'MetaFilter: func(Record) (bool, error)',
    'MetaMap: func(Record) (any, error)',
    'MetaPostShell: func(Record, ShellResult) (any, error)',
    'MetaReduce: func(acc any, value any) (any, error)',
    '',
    'FilesFilter: func(FileInfo) (bool, error)',
    'FilesMap: func(FileInfo) (any, error)',
    'FilesPostMap: func(FileInfo, input any) (MetaOut, error)',
    'FilesPostUpdate: func(FileInfo, input any, existing *Record) (UpdateOut, error)',
  ]);
  await appendSection('Open Design Questions', [
    'YAML strictness for unknown fields: error or ignore?',
  ]);
};
