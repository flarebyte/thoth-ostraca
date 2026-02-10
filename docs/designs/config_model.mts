// Action/pipeline config model (TypeScript shape used for design)
// v1: scripts are inline only (path-based scripts planned for v2)
export type InlineScript = { inline: string };
export type DiscoveryOptions = {
  root?: string;
  noGitignore?: boolean; // default false (respect .gitignore)
  followSymlinks?: boolean; // default false (do not follow for safety)
};
export type OutputOptions = {
  lines?: boolean; // default false (aggregate JSON)
  pretty?: boolean; // default false
  out?: string; // file path or "-" for stdout
};
// Options controlling what fields are available on the {file} input object
// used by files.filter/map/postMap. When info=true, os.FileInfo-derived fields
// are populated in addition to path components.
export type FilesInputOptions = {
  info?: boolean; // default false: include size/mode/modTime/isDir when true
};

// File input object shape (Lua-visible as `file`):
// - Always present: { path, relPath, dir, base, name, ext }
// - When files.info=true: { size (int), mode (string or numeric), modTime (RFC3339 string), isDir (bool) }
//   Exact runtime encoding tbd; intent is to mirror key os.FileInfo fields for filtering/mapping convenience.
export type ShellCapture = { stdout?: boolean; stderr?: boolean; maxBytes?: number };
export type ShellOptions = {
  enabled?: boolean; // set true to run shell step
  program?: "bash" | "sh" | "zsh";
  // Exactly one of commandTemplate or argsTemplate should be used.
  commandTemplate?: string; // single command string passed to shell, with placeholders
  argsTemplate?: string[]; // argv form: avoids shell parsing hazards; placeholders resolved per-arg
  workingDir?: string;
  env?: Record<string, string>;
  timeoutMs?: number;
  failFast?: boolean;
  capture?: ShellCapture;
  strictTemplating?: boolean; // default true: enforce transforms/escaping; unknown placeholders -> error
  killProcessGroup?: boolean; // default true: signal entire group on timeout/error
  termGraceMs?: number; // default e.g. 2000ms before SIGKILL
};
export type ErrorPolicy = {
  mode?: "keep-going" | "fail-fast"; // default keep-going
  embedErrors?: boolean; // default true: include per-item error objects in output
};
export type SaveOptions = {
  enabled?: boolean; // when true, write meta files to disk
  onExists?: "ignore" | "error"; // behavior if target exists
  dir?: string; // optional output directory for meta files
  hashAlgo?: "sha256"; // future extension; default sha256
  hashLen?: number; // characters from hash prefix; default 15
};
export type ValidationOptions = {
  allowUnknownTopLevel?: boolean; // default false: error on unknown top-level keys
};
export type LocatorPolicy = {
  allowAbsolute?: boolean; // default false: reject absolute file paths
  allowParentRefs?: boolean; // default false: reject ".." segments
  posixStyle?: boolean; // default true: canonical POSIX separators in locators
};
export type DiffOptions = {
  includeSnapshots?: boolean; // include before/after in per-item results
  output?: "patch" | "summary" | "both"; // default: both
};
export type UpdateOptions = {
  merge?: "shallow" | "deep" | "jsonpatch"; // default shallow
};
export type LuaOptions = {
  timeoutMs?: number; // per-script soft timeout (hook-driven)
  instructionLimit?: number; // max instructions per script (hook)
  memoryLimitBytes?: number; // soft cap (best-effort), may abort on large allocations
  libs?: {
    base?: boolean; // basic language features (enabled)
    table?: boolean; // enabled
    string?: boolean; // enabled
    math?: boolean; // enabled (deterministic random if configured)
    os?: boolean; // disabled by default
    io?: boolean; // disabled by default
    coroutine?: boolean; // disabled by default
    debug?: boolean; // disabled by default
  };
  allowOSExecute?: boolean; // default false (no os.execute / io.popen)
  allowEnv?: boolean; // default false (no direct os.getenv)
  envAllowlist?: string[]; // if allowEnv, restrict to these keys
  deterministicRandom?: boolean; // default true: seed math.random with fixed seed
  randomSeed?: number; // optional fixed seed override
};
export type PipelineConfig = {
  configVersion?: string;
  action?: "pipeline" | "create" | "update" | "diff" | "validate"; // which flow to run
  discovery?: DiscoveryOptions;
  files?: FilesInputOptions; // control available {file} fields for filtering/mapping
  workers?: number; // default: CPU count
  errors?: ErrorPolicy; // error handling strategy
  lua?: LuaOptions; // lua sandbox and runtime options
  filter?: InlineScript; // skip stage if omitted
  map?: InlineScript; // skip if omitted
  shell?: ShellOptions; // optional shell execution
  postMap?: InlineScript; // maps shell results → any
  reduce?: InlineScript; // skip if omitted
  output?: OutputOptions;
  save?: SaveOptions; // create mode: write meta files
  diff?: DiffOptions; // diff mode: output tuning
  validation?: ValidationOptions; // YAML strictness
  locatorPolicy?: LocatorPolicy; // locator normalization & security
  update?: UpdateOptions; // update mode: merge strategy
};

export const ACTION_CONFIG_EXAMPLE: PipelineConfig = {
  configVersion: "1",
  action: "pipeline",
  discovery: {
    root: ".",
    noGitignore: false,
    followSymlinks: false,
  },
  workers: 8,
  errors: { mode: "keep-going", embedErrors: true },
  lua: {
    timeoutMs: 2000,
    instructionLimit: 1_000_000,
    memoryLimitBytes: 8 * 1024 * 1024,
    libs: { base: true, table: true, string: true, math: true },
    allowOSExecute: false,
    allowEnv: false,
    deterministicRandom: true,
  },
  validation: { allowUnknownTopLevel: false },
  locatorPolicy: { allowAbsolute: false, allowParentRefs: false, posixStyle: true },
  // Lua inline scripts (concise and self-contained)
  filter: {
    inline: `-- keep records with meta.enabled == true
return (meta and meta.enabled) == true`,
  },
  map: {
    inline: `-- project selected fields
return { locator = locator, name = meta and meta.name }`,
  },
  shell: {
    enabled: true,
    program: "bash",
    // Recommend argv form to avoid shell parsing hazards
    argsTemplate: ["echo", "{json}"],
    workingDir: ".",
    env: { CI: "true" },
    timeoutMs: 60000,
    failFast: true,
    capture: { stdout: true, stderr: true, maxBytes: 1048576 },
    strictTemplating: true,
    killProcessGroup: true,
    termGraceMs: 2000,
  },
  postMap: {
    inline: `-- summarize shell result
return { locator = locator, exit = shell.exitCode }`,
  },
  reduce: {
    inline: `-- count items
return (acc or 0) + 1`,
  },
  output: { lines: false, pretty: false, out: "-" },
};

export const ACTION_CONFIG_CREATE_EXAMPLE: PipelineConfig = {
  configVersion: "1",
  action: "create",
  discovery: {
    root: ".",
    noGitignore: false,
  },
  // When true, expose os.FileInfo-derived fields on {file}
  files: { info: true },
  workers: 8,
  // Filter filenames (receive { file })
  filter: {
    inline: `-- only process markdown files
return string.match(file.ext or "", "^%.md$") ~= nil`,
  },
  // Map from filename to initial structure
  map: {
    inline: `-- produce initial meta from file info
return { title = file.base, category = file.dir }`,
  },
  // Optional post-map to finalize meta
  postMap: {
    inline: `-- finalize meta shape
return { meta = { title = (input.title or file.base) } }`,
  },
  // No reduce by default in create
  output: { lines: false, pretty: false, out: "-" },
  // Saving behavior
  save: { enabled: false, onExists: "ignore" },
};

export const ACTION_CONFIG_CREATE_MINIMAL: PipelineConfig = {
  configVersion: "1",
  action: "create",
  discovery: { root: ".", noGitignore: false },
  // Process all files (example); dry-run (no save)
  filter: { inline: `return true` },
  map: { inline: `return { meta = { created = true } }` },
  output: { lines: false, pretty: true, out: "-" },
  save: { enabled: false, onExists: "ignore", hashLen: 15 },
};

// Additional examples used in the generated docs
export const ACTION_CONFIG_DIFF_EXAMPLE: PipelineConfig = {
  configVersion: "1",
  action: "diff",
  discovery: { root: ".", noGitignore: false },
  workers: 8,
  errors: { mode: "keep-going", embedErrors: true },
  filter: { inline: `-- example: only .json files
return string.match(file.ext or "", "^%.json$") ~= nil` },
  map: { inline: `-- compute desired meta fields from filename
return { category = file.dir }` },
  diff: { includeSnapshots: false, output: "both" },
  // update-style post-map available as needed
  output: { lines: false, pretty: true, out: "-" },
};

export const ACTION_CONFIG_LUA_LIMITS_EXAMPLE: PipelineConfig = {
  configVersion: "1",
  action: "pipeline",
  discovery: { root: ".", noGitignore: false, followSymlinks: false },
  workers: 4,
  errors: { mode: "keep-going", embedErrors: true },
  lua: {
    timeoutMs: 500,
    instructionLimit: 100_000,
    memoryLimitBytes: 2 * 1024 * 1024,
    libs: { base: true, table: true, string: true, math: true },
    deterministicRandom: true,
    randomSeed: 1234,
  },
  filter: { inline: `return true` },
  map: { inline: `return { locator = locator, ok = true }` },
  output: { lines: false, pretty: true, out: "-" },
};
