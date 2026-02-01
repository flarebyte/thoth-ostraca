import {
  appendSection,
  appendUseCases,
  appendToReport,
  appendKeyValueList,
  ComponentCall,
  displayCallsAsText,
  displayCallsDetailed,
  FlowContext,
  getSetDifference,
  incrContext,
  resetReport,
  toBulletPoints,
  toUseCaseSet,
} from "./common.mts";

const calls: ComponentCall[] = [];

// Use-cases represent end-user capabilities we aim to support in the Go CLI.
const useCases = {
  metaFilter: {
    name: "meta.filter",
    title: "Filter meta by locator",
    note: "boolean predicate over {locator, meta}",
  },
  metaMap: {
    name: "meta.map",
    title: "Map meta records",
    note: "transform {locator, meta} → any",
  },
  metaReduce: {
    name: "meta.reduce",
    title: "Reduce across meta set",
    note: "aggregate stream → single result",
  },
  actionConfig: {
    name: "action.config",
    title: "Load action config file",
    note: "Prefer YAML; allow JSON",
  },
  embeddedScripting: {
    name: "scripting.embed",
    title: "Script filter/map/reduce",
    note: "Lua only (v1): small + popular",
  },
  shellExecFromMap: {
    name: "map.shell",
    title: "Run shell using map output",
    note: "Support bash, sh, zsh early",
  },
  locatorKinds: {
    name: "locator.kinds",
    title: "Locators as file path or URL",
  },
  parallelism: {
    name: "exec.parallel",
    title: "Process in parallel",
    note: "Goroutines + channels; bounded pool; default workers = CPU count",
  },
  batchCreate: {
    name: "batch.create",
    title: "Create many meta files",
  },
  batchUpdate: {
    name: "batch.update",
    title: "Update many meta files",
  },
  batchDiff: {
    name: "batch.diff",
    title: "Diff meta files at scale",
  },
  gitConflictFriendly: {
    name: "vcs.conflict-friendly",
    title: "One file per locator",
    note: "Minimize merge conflicts",
  },
  cliUX: {
    name: "cli.ux",
    title: "Helpful, well-documented flags",
  },
  gitIgnore: {
    name: "fs.gitignore",
    title: "Respect .gitignore by default",
    note: "Always on; opt-out via --no-gitignore",
  },
  outputJson: {
    name: "output.json",
    title: "JSON output for CLI/CI/AI",
    note: "Machine-oriented default; aggregated JSON; lines optional",
  },
  metaSchema: {
    name: "meta.schema",
    title: "Validate {locator, meta} schema",
    note: "Required fields: locator:string, meta:object; error on missing",
  },
};

const getByName = (expectedName: string) =>
  Object.values(useCases).find(({ name }) => name === expectedName);

// Action/pipeline config model (TypeScript shape used for design)
// v1: scripts are inline only (path-based scripts planned for v2)
type InlineScript = { inline: string };
type DiscoveryOptions = {
  root?: string;
  noGitignore?: boolean; // default false (respect .gitignore)
};
type OutputOptions = {
  lines?: boolean; // default false (aggregate JSON)
  pretty?: boolean; // default false
  out?: string; // file path or "-" for stdout
};
type ShellCapture = { stdout?: boolean; stderr?: boolean; maxBytes?: number };
type ShellOptions = {
  enabled?: boolean; // set true to run shell step
  program?: "bash" | "sh" | "zsh";
  commandTemplate?: string; // e.g. "echo {value}"
  workingDir?: string;
  env?: Record<string, string>;
  timeoutMs?: number;
  failFast?: boolean;
  capture?: ShellCapture;
};
type SaveOptions = {
  enabled?: boolean; // when true, write meta files to disk
  onExists?: "ignore" | "error"; // behavior if target exists
  dir?: string; // optional output directory for meta files
  hashAlgo?: "sha256"; // future extension; default sha256
  hashLen?: number; // characters from hash prefix; default 12
};
type PipelineConfig = {
  configVersion?: string;
  action?: "pipeline" | "create" | "update" | "diff"; // which flow to run
  discovery?: DiscoveryOptions;
  workers?: number; // default: CPU count
  filter?: InlineScript; // skip stage if omitted
  map?: InlineScript; // skip if omitted
  shell?: ShellOptions; // optional shell execution
  postMap?: InlineScript; // maps shell results → any
  reduce?: InlineScript; // skip if omitted
  output?: OutputOptions;
  save?: SaveOptions; // create mode: write meta files
};

export const ACTION_CONFIG_EXAMPLE: PipelineConfig = {
  configVersion: "1",
  action: "pipeline",
  discovery: {
    root: ".",
    noGitignore: false,
  },
  workers: 8,
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
    commandTemplate: "echo {value}",
    workingDir: ".",
    env: { CI: "true" },
    timeoutMs: 60000,
    failFast: true,
    capture: { stdout: true, stderr: true, maxBytes: 1048576 },
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
  save: { enabled: false, onExists: "ignore", hashLen: 12 },
};

// Everything listed here is expected to be supported long-term.
const mustUseCases = new Set([
  ...Object.values(useCases).map(({ name }) => name),
]);

// Build a catalog keyed by the canonical use-case name
const useCaseCatalogByName: Record<string, { name: string; title: string; note?: string }> =
  Object.fromEntries(Object.values(useCases).map((u) => [u.name, u]));

// Structured sections (TS model)
const GO_PACKAGE_OUTLINE: string[] = [
  "cmd/thoth: cobra wiring, --config parsing, action routing",
  "internal/config: load/validate YAML (inline Lua strings), defaults",
  "internal/fs: walk with gitignore, file info struct ({path, relPath, dir, base, name, ext})",
  "internal/meta: YAML read/write of {locator, meta}",
  "internal/lua: gopher-lua helpers to run inline scripts with typed inputs",
  "internal/pipeline: stages (filter/map/shell/post-map/reduce), worker pool",
  "internal/shell: exec with capture, timeouts, env, sh/bash/zsh",
  "internal/save: filename builder (<sha256[:12]>-<lastdir>-<filename>.thoth.yaml), onExists policy",
  "internal/diff: RFC6902 patch generation + item summary",
];

const DESIGN_DECISIONS: string[] = [
  "Filter: Lua-only (v1)",
  "Map: free-form JSON (any)",
  "Reduce: plain JSON value",
  "Output: machine-oriented JSON by default (aggregate unless --lines)",
  "Gitignore: always on; --no-gitignore to opt out",
  "Workers: default = CPU count (overridable via --workers)",
  "YAML: error on missing required fields (locator, meta)",
  "Shells: bash, sh, zsh supported early",
  "Save filename: sha256 prefix length = 12 by default",
];

// Helpers to suggest Go package, function, and file names based on call names
const toTokens = (s: string) => s.split(/[^a-zA-Z0-9]+/).filter(Boolean);
const toGoExported = (tokens: string[]) =>
  tokens.map((t) => t.charAt(0).toUpperCase() + t.slice(1)).join("");
const toSnake = (tokens: string[]) => tokens.map((t) => t.toLowerCase()).join("_");
const guessPkg = (call: ComponentCall) => {
  if (call.directory) return call.directory;
  const n = call.name;
  if (n.startsWith("cli.")) return "cmd/thoth";
  if (n.startsWith("fs.")) return "internal/fs";
  if (n.startsWith("meta.parse")) return "internal/meta";
  if (n.startsWith("meta.load")) return "internal/meta";
  if (n.startsWith("meta.save")) return "internal/save";
  if (n.startsWith("meta.update")) return "internal/save";
  if (n.startsWith("meta.diff")) return "internal/diff";
  if (n.startsWith("output.")) return "internal/output";
  if (n.startsWith("shell.")) return "internal/shell";
  if (n.startsWith("files.")) return "internal/pipeline";
  if (n.startsWith("flow.")) return "internal/pipeline";
  if (n.startsWith("action.")) return "internal/config";
  if (n.startsWith("meta.")) return "internal/pipeline";
  return "internal";
};
const suggestFor = (call: ComponentCall) => {
  const pkg = guessPkg(call);
  const tokens = toTokens(call.name);
  const func = toGoExported(tokens);
  const file = `${pkg}/${toSnake(tokens.slice(-2)) || toSnake(tokens)}.go`;
  return { pkg, func, file };
};

// Root CLI: thoth
const cliRoot = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.root",
    title: "thoth CLI root command",
    directory: "cmd/thoth",
    note: "cobra-based command tree",
    level: context.level,
    useCases: [useCases.cliUX.name],
  };
  calls.push(call);
  // Register commands under the root.
  cliArgsRun(incrContext(context));
};

// Single run command: executes the configured action (pipeline/create/update/diff)
// pipeline: discover → parse → filter → [map] → [shell] → [post-map] → [reduce] → output
const cliArgsRun = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.run",
    title: "Parse args for run",
    directory: "cmd/thoth",
    note: "flags: --config (YAML preferred; JSON accepted). All other options belong in the action config.",
    level: context.level,
    useCases: [useCases.cliUX.name, useCases.outputJson.name],
  };
  calls.push(call);
  loadActionConfig(incrContext(context));
  routeByActionType(incrContext(context));
};

// Route to the appropriate flow based on config.action
const routeByActionType = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "action.route",
    title: "Route by action type",
    note: "action: pipeline | create | update | diff",
    level: context.level,
  };
  calls.push(call);
  // Pipeline flow (meta processing)
  pipelineFlow(incrContext(context));
  // Create flow (generate new meta files from filenames)
  createFlow(incrContext(context));
  // Update flow (update or create meta files from filenames)
  updateFlow(incrContext(context));
  // Diff flow (show changes without writing; detect orphans)
  diffFlow(incrContext(context));
};

const pipelineFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.pipeline",
    title: "Meta pipeline flow",
    level: context.level,
  };
  calls.push(call);
  findMetaLocators(incrContext(context));
  parseYamlRecords(incrContext(context));
  filterMetaLocators(incrContext(context));
  mapMetaRecords(incrContext(context));
  execShellFromMap(incrContext(context));
  postMapShellResults(incrContext(context));
  reduceMetaRecords(incrContext(context));
  outputJsonResult(incrContext(context));
};

const createFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.create",
    title: "Create meta files flow",
    level: context.level,
    useCases: [useCases.batchCreate.name],
  };
  calls.push(call);
  findFilesForCreate(incrContext(context));
  filterFilenames(incrContext(context));
  mapFilenames(incrContext(context));
  postMapFromFiles(incrContext(context));
  saveMetaFiles(incrContext(context));
  outputJsonResult(incrContext(context));
};

const updateFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.update",
    title: "Update meta files flow",
    level: context.level,
    useCases: [useCases.batchUpdate.name],
  };
  calls.push(call);
  findFilesForUpdate(incrContext(context));
  filterFilenames(incrContext(context));
  mapFilenames(incrContext(context));
  loadExistingMeta(incrContext(context));
  postMapUpdateFromFiles(incrContext(context));
  updateMetaFiles(incrContext(context));
  outputJsonResult(incrContext(context));
};

const diffFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.diff",
    title: "Diff meta files flow",
    level: context.level,
    useCases: [useCases.batchDiff.name],
  };
  calls.push(call);
  findFilesForUpdate(incrContext(context));
  filterFilenames(incrContext(context));
  mapFilenames(incrContext(context));
  loadExistingMeta(incrContext(context));
  postMapUpdateFromFiles(incrContext(context));
  computeMetaDiffs(incrContext(context));
  scanForOrphanMetas(incrContext(context));
  outputJsonResult(incrContext(context));
};

// File discovery: respects .gitignore and finds *.thoth.yaml files
const findMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery",
    title: "Find *.thoth.yaml files",
    note: "walk root; .gitignore ON by default; --no-gitignore to disable",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
};

// Discovery for create flow: all files under root respecting .gitignore
const findFilesForCreate = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.files",
    title: "Find files recursively (gitignore)",
    note: "walk root; .gitignore ON by default; no patterns; filenames as inputs",
    level: context.level,
    useCases: [useCases.gitIgnore.name],
  };
  calls.push(call);
};

// Discovery for update flow: reuse create discovery semantics
const findFilesForUpdate = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.files.update",
    title: "Find files recursively (update)",
    note: "walk root; .gitignore ON by default; filenames as inputs",
    level: context.level,
    useCases: [useCases.gitIgnore.name],
  };
  calls.push(call);
};

// Parse and validate each YAML meta file → {locator, meta}
const parseYamlRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.parse",
    title: "Parse and validate YAML records",
    note: "yaml.v3; strict fields; types; support file path or URL locator",
    level: context.level,
    useCases: [useCases.metaSchema.name, useCases.locatorKinds.name],
  };
  calls.push(call);
};


// Filtering step: predicate over stream of {locator, meta}
const filterMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.filter.step",
    title: "Apply filter predicate",
    note: "Lua-only predicate (v1)",
    level: context.level,
    useCases: [useCases.metaFilter.name, useCases.embeddedScripting.name, useCases.parallelism.name],
  };
  calls.push(call);
};

// Create: filter over filenames
const filterFilenames = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.filter.step",
    title: "Filter filenames",
    note: "Lua-only predicate (v1) over {file}",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
};

// Create: map filenames to structures
const mapFilenames = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.map.step",
    title: "Map filenames",
    note: "Lua-only map (v1) over {file}",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
};

// Create: optional post-map to produce final {meta}
const postMapFromFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.map.post",
    title: "Post-map from files",
    note: "Conditional: inline Lua transforms {file,input} -> any",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
};

// Create: save meta files using naming convention
const saveMetaFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.save",
    title: "Save meta files (*.thoth.yaml)",
    note: "Conditional: config.save.enabled or --save; name = <hash>-<lastdir>-<filename>.thoth.yaml; onExists: ignore|error",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
};

// Update: load existing meta if present for each filename
const loadExistingMeta = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.load.existing",
    title: "Load existing meta (if any)",
    note: "compute expected path by naming convention; read YAML if exists",
    level: context.level,
    useCases: [useCases.batchUpdate.name],
  };
  calls.push(call);
};

// Update: post-map with access to existing meta
const postMapUpdateFromFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.map.post.update",
    title: "Post-map for update (with existing)",
    note: "Lua receives {file,input,existing?}; returns { meta } patch",
    level: context.level,
    useCases: [useCases.batchUpdate.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
};

// Update: merge and write meta (create if missing)
const updateMetaFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.update",
    title: "Update meta files (merge/create)",
    note: "shallow merge: new keys override existing; missing -> create new by naming convention",
    level: context.level,
    useCases: [useCases.batchUpdate.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
};

// Diff: compute differences between existing and would-be-updated meta
const computeMetaDiffs = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.diff.compute",
    title: "Compute meta diffs",
    note: "deep diff existing vs patch-applied result; output RFC6902 JSON Patch + summary",
    level: context.level,
    useCases: [useCases.batchDiff.name],
  };
  calls.push(call);
};

// Diff: scan for orphan meta files (locator path missing)
const scanForOrphanMetas = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.diff.orphans",
    title: "Detect orphan meta files",
    note: "iterate *.thoth.yaml; if locator is file path and does not exist, flag",
    level: context.level,
    useCases: [useCases.batchDiff.name, useCases.locatorKinds.name],
  };
  calls.push(call);
};

// Optional: load action config for map/reduce/run
const loadActionConfig = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "action.config.load",
    title: "Load action config file",
    note: "--config path; YAML preferred; JSON accepted; drives entire pipeline",
    level: context.level,
    useCases: [useCases.actionConfig.name],
  };
  calls.push(call);
};

// Map step: transform {locator, meta} -> any
const mapMetaRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.map.step",
    title: "Apply map transform",
    note: "Lua-only mapping (v1); parallel by default",
    level: context.level,
    useCases: [useCases.metaMap.name, useCases.embeddedScripting.name, useCases.parallelism.name],
  };
  calls.push(call);
};

// Reduce step: aggregate stream -> single value
const reduceMetaRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.reduce.step",
    title: "Apply reduce aggregate",
    note: "Lua-only reduce (v1); parallel feed; single JSON value",
    level: context.level,
    useCases: [useCases.metaReduce.name, useCases.embeddedScripting.name, useCases.parallelism.name],
  };
  calls.push(call);
};


const execShellFromMap = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "shell.exec",
    title: "Execute shell per mapped item",
    note: "Conditional: --run-shell; supports bash, sh, zsh; parallel with bounded workers; feeds post-map/reduce when provided",
    level: context.level,
    useCases: [useCases.shellExecFromMap.name, useCases.parallelism.name],
  };
  calls.push(call);
};

// Optional: map the shell result to structured data for downstream reduce/output
const postMapShellResults = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.map.post-shell",
    title: "Post-map shell results",
    note: "Conditional: --post-map-script; Lua transforms {locator,input,shell:{cmd,exitCode,stdout,stderr,durationMs}}",
    level: context.level,
    useCases: [useCases.metaMap.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
};

// Output JSON: machine-oriented by default; aggregated or lines
const outputJsonResult = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "output.json.result",
    title: "Write JSON result (array/value/lines)",
    note: "default: aggregated JSON array; --lines to stream; reduce → single value",
    level: context.level,
    useCases: [useCases.outputJson.name],
  };
  calls.push(call);
};

// Start tree at root command.
cliRoot({ level: 0 });

await resetReport();
await appendToReport("# FLOW DESIGN OVERVIEW (Generated)\n");
await appendToReport("## Function calls tree\n");
await appendToReport("```");
await displayCallsAsText(calls);
await appendToReport("```\n");

await appendUseCases("Supported use cases:", toUseCaseSet(calls), useCaseCatalogByName);

await appendUseCases(
  "Unsupported use cases (yet):",
  getSetDifference(mustUseCases, toUseCaseSet(calls)),
  useCaseCatalogByName,
);

// Suggested Go implementation outline (structured)
const SUGGESTED_GO_IMPLEMENTATION: Array<[string, string | string[]]> = [
  ["Module", "go 1.22; command name: thoth"],
  ["CLI", "cobra for command tree; viper optional"],
  ["Types", "type Record struct { Locator string; Meta map[string]any }"],
  ["YAML", "gopkg.in/yaml.v3 for *.thoth.yaml"],
  ["Discovery", "filepath.WalkDir + gitignore filter (go-gitignore)"],
  ["Schema", "required fields (locator, meta); error on missing"],
  ["Filter/Map/Reduce", "Lua scripts only (gopher-lua) for v1"],
  ["Parallelism", "bounded worker pool; default workers = runtime.NumCPU()"],
  ["Output", "aggregated JSON by default; --lines to stream; --pretty for humans"],
  ["Commands", "thoth run (exec action config: pipeline/create/update/diff)"],
  ["Flags", "--config (YAML preferred; JSON accepted), --save (enable saving in create)"],
  ["Tests", "golden tests for I/O; fs testdata fixtures"],
  ["Reduce", "outputs a plain JSON value"],
  ["Map", "returns free-form JSON (any)"],
  ["Shells", "support bash, sh, zsh early"],
  ["Create flow", "discover files (gitignore), filter/map/post-map over {file}"],
  ["Save writer", "if save.enabled or --save, write *.thoth.yaml"],
  ["Filename", "<sha256[:12]>-<lastdir>-<filename>.thoth.yaml"],
  ["Hash input", "discovery relPath for stability"],
  ["On exists", "ignore (default) or error"],
  ["Update flow", "discover files, load existing meta if present, shallow-merge patch, create if missing"],
  ["Merge strategy", "shallow merge (new keys override)"],
  ["Diff flow", "same as update until patch; compute deep diff; do not write"],
  ["Orphans", "scan existing meta files; if locator path missing on disk, report"],
  ["Diff output", "RFC 6902 JSON Patch per item + summary (created/modified/deleted/orphan/unchanged)"],
  ["internal/diff", "generate patches and optional before/after snapshots for debugging"],
];

await appendKeyValueList("Suggested Go Implementation", SUGGESTED_GO_IMPLEMENTATION);

// Emit example action config as JSON for easy viewing
await appendSection(
  "Action Config (JSON Example)",
  "```json\n" + JSON.stringify(ACTION_CONFIG_EXAMPLE, null, 2) + "\n```",
);

await appendSection(
  "Action Config (Create Example)",
  "```json\n" + JSON.stringify(ACTION_CONFIG_CREATE_EXAMPLE, null, 2) + "\n```",
);

await appendSection(
  "Action Config (Create Minimal Example)",
  "```json\n" + JSON.stringify(ACTION_CONFIG_CREATE_MINIMAL, null, 2) + "\n```",
);

// Provide a diff config example (reuses update contracts; no writes)
export const ACTION_CONFIG_DIFF_EXAMPLE: PipelineConfig = {
  configVersion: "1",
  action: "diff",
  discovery: { root: ".", noGitignore: false },
  workers: 8,
  filter: { inline: `-- example: only .json files\nreturn string.match(file.ext or "", "^%.json$") ~= nil` },
  map: { inline: `-- compute desired meta fields from filename\nreturn { category = file.dir }` },
  // update-style post-map available as needed
  output: { lines: false, pretty: true, out: "-" },
};

await appendSection(
  "Action Config (Diff Example)",
  "```json\n" + JSON.stringify(ACTION_CONFIG_DIFF_EXAMPLE, null, 2) + "\n```",
);

await appendSection("Lua Data Contracts", [
  "Filter: fn({ locator, meta }) -> bool",
  "Map: fn({ locator, meta }) -> any",
  "Reduce: fn(acc, value) -> acc (single JSON value)",
  "Post-map (shell): fn({ locator, input, shell: { cmd, exitCode, stdout, stderr, durationMs } }) -> any",
  "Create Filter: fn({ file: { path, relPath, dir, base, name, ext } }) -> bool",
  "Create Map: fn({ file }) -> any",
  "Create Post-map: fn({ file, input }) -> { meta }",
  "Update Post-map: fn({ file, input, existing? }) -> { meta } (patch)",
]);

await appendSection("Diff Output Shape", [
  "Per-item result: { file, status, patch?, before?, after? }",
  "status ∈ { created, modified, deleted, unchanged, orphan }",
  "patch: RFC 6902 JSON Patch array (ops: add/remove/replace/move/copy/test)",
  "before/after: optional full meta snapshots for debugging (disabled by default)",
  "Top-level summary: counts per status and totals",
]);

await appendSection("Lua Builtins", [
  "locator.kind(locator) -> 'file' | 'url'",
  "locator.to_file_path(locator, root) -> string|nil (nil for URLs)",
  "url.is_url(s) -> bool (http/https schemes)",
]);

// Detailed calls section with notes and suggestions
const detailedCalls: ComponentCall[] = calls.map((c) => ({
  ...c,
  suggest: suggestFor(c),
}));

await appendToReport("\n## Function calls details\n");
await appendToReport("```");
await displayCallsDetailed(detailedCalls);
await appendToReport("```");

await appendSection("Go Package Outline", GO_PACKAGE_OUTLINE);

await appendSection("Design Decisions", DESIGN_DECISIONS);

await appendSection("Open Design Questions", [
  "YAML strictness for unknown fields: error or ignore?",
]);
