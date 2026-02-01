import {
  appendSection,
  appendToReport,
  ComponentCall,
  displayCallsAsText,
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

const getTitlesForSet = (useCaseSet: Set<string>) =>
  [...useCaseSet].map((useCase) => getByName(useCase)?.title || useCase);

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
  action?: "pipeline" | "create" | "update"; // which flow to run
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
  cliArgsMetaPipeline(incrContext(context));
};

// Single pipeline command: discover → parse → filter → [map] → [shell] → [post-map] → [reduce] → output
const cliArgsMetaPipeline = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.meta",
    title: "Parse args for meta pipeline",
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
    note: "action: pipeline | create | update",
    level: context.level,
  };
  calls.push(call);
  // Pipeline flow (meta processing)
  pipelineFlow(incrContext(context));
  // Create flow (generate new meta files from filenames)
  createFlow(incrContext(context));
  // Update flow (update or create meta files from filenames)
  updateFlow(incrContext(context));
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

await appendToReport("Supported use cases:\n");

await appendToReport(toBulletPoints(getTitlesForSet(toUseCaseSet(calls))));

await appendToReport("\nUnsupported use cases (yet):\n");

await appendToReport(
  toBulletPoints(
    getTitlesForSet(getSetDifference(mustUseCases, toUseCaseSet(calls))),
  ),
);

// Suggested Go implementation outline
await appendSection("Suggested Go Implementation", [
  "Module: go 1.22; command name: thoth",
  "CLI: cobra for command tree; viper optional",
  "Types: type Record struct { Locator string; Meta map[string]any }",
  "YAML: gopkg.in/yaml.v3 for *.thoth.yaml",
  "Discovery: filepath.WalkDir + gitignore filter (go-gitignore)",
  "Schema: required fields (locator, meta); error on missing",
  "Filter/Map/Reduce: Lua scripts only (gopher-lua) for v1",
  "Parallelism: bounded worker pool; default workers = runtime.NumCPU()",
  "Output: aggregated JSON by default; --lines to stream; --pretty for humans",
  "Commands: thoth meta (single pipeline incl. optional shell and create)",
  "Flags: --config (YAML preferred; JSON accepted), --save (enable saving in create)",
  "Tests: golden tests for I/O; fs testdata fixtures",
  "Reduce: outputs a plain JSON value",
  "Map: returns free-form JSON (any)",
  "Shells: support bash, sh, zsh early",
  "Create flow: discover files (gitignore), filter/map/post-map over {file}",
  "Save writer: if save.enabled or --save, write *.thoth.yaml",
  "Filename: <sha256[:12]>-<lastdir>-<filename>.thoth.yaml",
  "Hash input: discovery relPath for stability",
  "On exists: ignore (default) or error",
  "Update flow: discover files, load existing meta if present, shallow-merge patch, create if missing",
  "Merge strategy: shallow merge (new keys override)",
]);

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

await appendSection("Go Package Outline", [
  "cmd/thoth: cobra wiring, --config parsing, action routing",
  "internal/config: load/validate YAML (inline Lua strings), defaults",
  "internal/fs: walk with gitignore, file info struct ({path, relPath, dir, base, name, ext})",
  "internal/meta: YAML read/write of {locator, meta}",
  "internal/lua: gopher-lua helpers to run inline scripts with typed inputs",
  "internal/pipeline: stages (filter/map/shell/post-map/reduce), worker pool",
  "internal/shell: exec with capture, timeouts, env, sh/bash/zsh",
  "internal/save: filename builder (<sha256[:12]>-<lastdir>-<filename>.thoth.yaml), onExists policy",
]);

await appendSection("Design Decisions", [
  "Filter: Lua-only (v1)",
  "Map: free-form JSON (any)",
  "Reduce: plain JSON value",
  "Output: machine-oriented JSON by default (aggregate unless --lines)",
  "Gitignore: always on; --no-gitignore to opt out",
  "Workers: default = CPU count (overridable via --workers)",
  "YAML: error on missing required fields (locator, meta)",
  "Shells: bash, sh, zsh supported early",
  "Save filename: sha256 prefix length = 12 by default",
]);

await appendSection("Open Design Questions", [
  "YAML strictness for unknown fields: error or ignore?",
]);
