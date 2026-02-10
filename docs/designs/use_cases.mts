import { UseCase } from "./common.mts";

export const useCases: Record<string, UseCase> = {
  filesInfo: {
    name: "files.info",
    title: "Expose os.FileInfo for inputs",
    note: "Include size, mode, modTime, isDir for filtering/mapping when enabled",
  },
  filesGit: {
    name: "files.git",
    title: "Expose Git metadata for inputs",
    note: "Use go-git to provide tracked/ignored, worktree status, and last commit info when enabled",
  },
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
  metaValidateOnly: {
    name: "meta.validate",
    title: "Validate meta files only",
    note: "No transforms or shell; emit validation report",
  },
};

export const getByName = (expectedName: string) =>
  Object.values(useCases).find(({ name }) => name === expectedName);

export const mustUseCases = new Set([
  ...Object.values(useCases).map(({ name }) => name),
]);

export const useCaseCatalogByName: Record<string, { name: string; title: string; note?: string }> =
  Object.fromEntries(Object.values(useCases).map((u) => [u.name, u]));
