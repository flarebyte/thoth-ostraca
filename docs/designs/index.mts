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
  // Register subcommands under the root.
  cliArgsMetaFind(incrContext(context));
  cliArgsMetaMap(incrContext(context));
  cliArgsMetaReduce(incrContext(context));
  cliArgsRun(incrContext(context));
};

// Subcommand that discovers meta files and emits JSON.
const cliArgsMetaFind = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.meta.find",
    title: "Parse args for meta find",
    directory: "cmd/thoth",
    note: "flags: --root, --pattern, --no-gitignore, --json, --lines, --pretty",
    level: context.level,
    useCases: [useCases.cliUX.name, useCases.outputJson.name],
  };
  calls.push(call);
  findMetaLocators(incrContext(context));
};

// Subcommand: map
const cliArgsMetaMap = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.meta.map",
    title: "Parse args for meta map",
    directory: "cmd/thoth",
    note: "flags: --root, --pattern, --no-gitignore, --script, --json, --lines, --pretty",
    level: context.level,
    useCases: [useCases.cliUX.name, useCases.outputJson.name],
  };
  calls.push(call);
  findMetaLocatorsForMap(incrContext(context));
};

// Subcommand: reduce
const cliArgsMetaReduce = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.meta.reduce",
    title: "Parse args for meta reduce",
    directory: "cmd/thoth",
    note: "flags: --root, --pattern, --no-gitignore, --script, --json, --pretty",
    level: context.level,
    useCases: [useCases.cliUX.name, useCases.outputJson.name],
  };
  calls.push(call);
  findMetaLocatorsForReduce(incrContext(context));
};

// Subcommand: run (shell from map output)
const cliArgsRun = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.run",
    title: "Parse args for run (shell)",
    directory: "cmd/thoth",
    note: "flags: --root, --pattern, --no-gitignore, --script, --shell, --workers",
    level: context.level,
    useCases: [useCases.cliUX.name],
  };
  calls.push(call);
  findMetaLocatorsForRun(incrContext(context));
};

// File discovery: respects .gitignore and finds *.thoth.yaml files
const findMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery",
    title: "Find *.thoth.yaml files",
    note: "walk root; .gitignore ON by default; --no-gitignore to disable; pattern overrides",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
  parseYamlRecords(incrContext(context));
};

// Same discovery used by map
const findMetaLocatorsForMap = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.map",
    title: "Find *.thoth.yaml files (map)",
    note: "reuse discovery; supports patterns and .gitignore",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
  parseYamlRecordsForMap(incrContext(context));
};

// Same discovery used by reduce
const findMetaLocatorsForReduce = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.reduce",
    title: "Find *.thoth.yaml files (reduce)",
    note: "reuse discovery; supports patterns and .gitignore",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
  parseYamlRecordsForReduce(incrContext(context));
};

// Same discovery used by run
const findMetaLocatorsForRun = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.run",
    title: "Find *.thoth.yaml files (run)",
    note: "reuse discovery; supports patterns and .gitignore",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
  parseYamlRecordsForRun(incrContext(context));
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
  filterMetaLocators(incrContext(context));
};

// Parser for map flow
const parseYamlRecordsForMap = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.parse.map",
    title: "Parse and validate YAML (map)",
    note: "yaml.v3; strict fields; types; support file path or URL locator",
    level: context.level,
    useCases: [useCases.metaSchema.name, useCases.locatorKinds.name],
  };
  calls.push(call);
  loadActionConfig(incrContext(context));
  mapMetaRecords(incrContext(context));
};

// Parser for reduce flow
const parseYamlRecordsForReduce = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.parse.reduce",
    title: "Parse and validate YAML (reduce)",
    note: "yaml.v3; strict fields; types; support file path or URL locator",
    level: context.level,
    useCases: [useCases.metaSchema.name, useCases.locatorKinds.name],
  };
  calls.push(call);
  loadActionConfig(incrContext(context));
  reduceMetaRecords(incrContext(context));
};

// Parser for run flow
const parseYamlRecordsForRun = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.parse.run",
    title: "Parse and validate YAML (run)",
    note: "yaml.v3; strict fields; types; support file path or URL locator",
    level: context.level,
    useCases: [useCases.metaSchema.name, useCases.locatorKinds.name],
  };
  calls.push(call);
  loadActionConfig(incrContext(context));
  mapMetaForRun(incrContext(context));
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
  outputJsonResult(incrContext(context));
};

// Optional: load action config for map/reduce/run
const loadActionConfig = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "action.config.load",
    title: "Load action config file (optional)",
    note: "--config path; YAML preferred; JSON allowed",
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
  outputJsonResult(incrContext(context));
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
  outputJsonResult(incrContext(context));
};

// Run step: execute shell using map output
const mapMetaForRun = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.map.for-run",
    title: "Map for run (shell input)",
    note: "Lua-only map (v1) to build command args",
    level: context.level,
    useCases: [useCases.metaMap.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
  execShellFromMap(incrContext(context));
};

const execShellFromMap = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "shell.exec",
    title: "Execute shell per mapped item",
    note: "Supports bash, sh, zsh; parallel with bounded workers",
    level: context.level,
    useCases: [useCases.shellExecFromMap.name, useCases.parallelism.name],
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
  "Commands: thoth find, thoth map, thoth reduce, thoth run (shell)",
  "Flags: --root, --pattern, --no-gitignore, --workers, --script, --out",
  "Tests: golden tests for I/O; fs testdata fixtures",
  "Reduce: outputs a plain JSON value",
  "Map: returns free-form JSON (any)",
  "Shells: support bash, sh, zsh early",
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
]);

await appendSection("Open Design Questions", [
  "YAML strictness for unknown fields: error or ignore?",
]);
