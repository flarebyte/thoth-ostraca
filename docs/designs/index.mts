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
    note: "Lua preferred: small + popular",
  },
  shellExecFromMap: {
    name: "map.shell",
    title: "Run shell using map output",
    note: "Default bash; allow others",
  },
  locatorKinds: {
    name: "locator.kinds",
    title: "Locators as file path or URL",
  },
  parallelism: {
    name: "exec.parallel",
    title: "Process in parallel",
    note: "Goroutines + channels; bounded pool",
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
  },
  outputJson: {
    name: "output.json",
    title: "JSON output for humans/CI/AI",
    note: "Pretty/compact/lines variants",
  },
  metaSchema: {
    name: "meta.schema",
    title: "Validate {locator, meta} schema",
    note: "locator:string; meta:object",
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
  cliArgsMetaFind(incrContext(context));
};

// Subcommand that discovers meta files and emits JSON.
const cliArgsMetaFind = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.meta.find",
    title: "Parse args for meta find",
    directory: "cmd/thoth",
    note: "flags: --root, --pattern, --ignore, --json",
    level: context.level,
    useCases: [useCases.cliUX.name, useCases.outputJson.name],
  };
  calls.push(call);
  findMetaLocators(incrContext(context));
};

// File discovery: respects .gitignore and finds *.thoth.yaml files
const findMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery",
    title: "Find *.thoth.yaml files",
    note: "walk root; respect .gitignore; pattern overrides",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
  parseYamlRecords(incrContext(context));
};

// Parse and validate each YAML meta file → {locator, meta}
const parseYamlRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.parse",
    title: "Parse and validate YAML records",
    note: "yaml.v3; strict fields; types",
    level: context.level,
    useCases: [useCases.metaSchema.name],
  };
  calls.push(call);
  filterMetaLocators(incrContext(context));
};

// Filtering step: predicate over stream of {locator, meta}
const filterMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.filter.step",
    title: "Apply filter predicate",
    note: "built-in or Lua predicate",
    level: context.level,
    useCases: [useCases.metaFilter.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
  outputJsonStream(incrContext(context));
};

// Output JSON: friendly for humans, CI, and AI
const outputJsonStream = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "output.json.lines",
    title: "Write JSON (pretty/compact/lines)",
    note: "default: json lines for streams",
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
  "Schema: validate locator non-empty; meta is object",
  "Filter/Map/Reduce: built-in funcs + optional gopher-lua scripts",
  "Parallelism: bounded worker pool; channels for records",
  "Output: JSON lines (default), pretty via --pretty, compact via --compact",
  "Commands: thoth find, thoth map, thoth reduce, thoth run (shell)",
  "Flags: --root, --pattern, --ignore, --workers, --script, --out",
  "Tests: golden tests for I/O; fs testdata fixtures",
]);

await appendSection("Open Design Questions", [
  "Filter expression: prefer small DSL or go with Lua first?",
  "Map output shape: free-form any vs constrained fields?",
  "Reduce outputs: single JSON value vs object with metadata?",
  "Default output: JSON lines or pretty JSON when writing to TTY?",
  "Gitignore behavior: always on, or opt-out flag --no-gitignore?",
]);
