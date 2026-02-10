import { FlowContext, ComponentCall } from "./common.mts";
import { incrContext } from "./common.mts";
import { calls } from "./calls.mts";
import { useCases } from "./use_cases.mts";

export const validateMetaOnly = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.validate.only",
    title: "Collect validation results",
    note: "Schema + locator checks only; no filter/map/reduce/shell",
    level: context.level,
    useCases: [useCases.metaValidateOnly.name, useCases.metaSchema.name],
  };
  calls.push(call);
};

export const findMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery",
    title: "Find *.thoth.yaml files",
    note: "walk root; .gitignore ON by default even outside git repos; --no-gitignore to disable; do not follow symlinks by default",
    level: context.level,
    useCases: [useCases.gitIgnore.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
};

export const findFilesForCreate = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.files",
    title: "Find files recursively (gitignore)",
    note: "walk root; .gitignore ON by default (even if not a git repo); no patterns; do not follow symlinks by default; filenames as inputs",
    level: context.level,
    useCases: [useCases.gitIgnore.name],
  };
  calls.push(call);
};

export const enrichFilesWithOptionalInfo = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.enrich",
    title: "Enrich files with OS/Git info",
    note: "Conditional: files.info and/or files.git; attach file.info (os.Stat) and file.git (go-git status/last commit)",
    level: context.level,
    useCases: [useCases.filesInfo.name, useCases.filesGit.name],
  };
  calls.push(call);
};

export const findFilesForUpdate = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "fs.discovery.files.update",
    title: "Find files recursively (update)",
    note: "walk root; .gitignore ON by default (even if not a git repo); do not follow symlinks by default; filenames as inputs",
    level: context.level,
    useCases: [useCases.gitIgnore.name],
  };
  calls.push(call);
};

export const parseYamlRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.parse",
    title: "Parse and validate YAML records",
    note: "yaml.v3; strict fields; types; locator canonicalization; top-level unknown = error (unless validation.allowUnknownTopLevel); inside meta: unknown allowed",
    level: context.level,
    useCases: [useCases.metaSchema.name, useCases.locatorKinds.name],
  };
  calls.push(call);
};

export const filterMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.filter.step",
    title: "Apply filter predicate",
    note: "Lua-only predicate (v1)",
    level: context.level,
    useCases: [useCases.metaFilter.name, useCases.embeddedScripting.name, useCases.parallelism.name],
  };
  calls.push(call);
};

export const filterFilenames = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.filter.step",
    title: "Filter filenames",
    note: "Lua-only predicate (v1) over {file}",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.embeddedScripting.name, useCases.filesInfo.name, useCases.filesGit.name],
  };
  calls.push(call);
};

export const mapFilenames = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.map.step",
    title: "Map filenames",
    note: "Lua-only map (v1) over {file}",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.embeddedScripting.name, useCases.filesInfo.name, useCases.filesGit.name],
  };
  calls.push(call);
};

export const postMapFromFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.map.post",
    title: "Post-map from files",
    note: "Conditional: inline Lua transforms {file,input} -> any",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.embeddedScripting.name, useCases.filesInfo.name, useCases.filesGit.name],
  };
  calls.push(call);
};

export const saveMetaFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.save",
    title: "Save meta files (*.thoth.yaml)",
    note: "Conditional: config.save.enabled or --save; name = <sha256[:15]>[-r<rootTag>]-<lastdir>-<filename>.thoth.yaml; sanitize components; if path exists and belongs to different locator -> error; onExists: ignore|error",
    level: context.level,
    useCases: [useCases.batchCreate.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
};

export const loadExistingMeta = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.load.existing",
    title: "Load existing meta (if any)",
    note: "compute expected path by naming convention; read YAML if exists",
    level: context.level,
    useCases: [useCases.batchUpdate.name],
  };
  calls.push(call);
};

export const postMapUpdateFromFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "files.map.post.update",
    title: "Post-map for update (with existing)",
    note: "Lua receives {file,input,existing?}; returns either { meta } (full desired) or { patch } (RFC6902)",
    level: context.level,
    useCases: [useCases.batchUpdate.name, useCases.embeddedScripting.name, useCases.filesInfo.name, useCases.filesGit.name],
  };
  calls.push(call);
};

export const updateMetaFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.update",
    title: "Update meta files (merge/create)",
    note: "merge strategy via config.update.merge: shallow|deep|jsonpatch (default shallow); if post-map returns patch, apply RFC6902; else merge existing with returned meta; missing -> create new by naming convention; verify filename hash against current root+relPath (mismatch -> error)",
    level: context.level,
    useCases: [useCases.batchUpdate.name, useCases.gitConflictFriendly.name],
  };
  calls.push(call);
};

export const computeMetaDiffs = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.diff.compute",
    title: "Compute meta diffs",
    note: "deep diff existing vs patch-applied result; output RFC6902 JSON Patch + summary",
    level: context.level,
    useCases: [useCases.batchDiff.name],
  };
  calls.push(call);
};

export const scanForOrphanMetas = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.diff.orphans",
    title: "Detect orphan meta files",
    note: "iterate *.thoth.yaml; if locator is file path and does not exist, flag",
    level: context.level,
    useCases: [useCases.batchDiff.name, useCases.locatorKinds.name],
  };
  calls.push(call);
};

export const loadActionConfig = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "action.config.load",
    title: "Load action config file",
    note: "--config path; YAML preferred; JSON accepted; drives entire pipeline",
    level: context.level,
    useCases: [useCases.actionConfig.name],
  };
  calls.push(call);
};

export const mapMetaRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.map.step",
    title: "Apply map transform",
    note: "Lua-only mapping (v1); parallel by default",
    level: context.level,
    useCases: [useCases.metaMap.name, useCases.embeddedScripting.name, useCases.parallelism.name],
  };
  calls.push(call);
};

export const reduceMetaRecords = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.reduce.step",
    title: "Apply reduce aggregate",
    note: "Lua-only reduce (v1); parallel feed; single JSON value",
    level: context.level,
    useCases: [useCases.metaReduce.name, useCases.embeddedScripting.name, useCases.parallelism.name],
  };
  calls.push(call);
};

export const execShellFromMap = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "shell.exec",
    title: "Execute shell per mapped item",
    note: "Conditional: --run-shell; argv templates preferred (no shell parsing); string templates auto-escape; supports bash/sh/zsh; parallel with bounded workers; feeds post-map/reduce; timeout kills process group",
    level: context.level,
    useCases: [useCases.shellExecFromMap.name, useCases.parallelism.name],
  };
  calls.push(call);
};

export const postMapShellResults = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta.map.post-shell",
    title: "Post-map shell results",
    note: "Conditional: --post-map-script; Lua transforms {locator,input,shell:{cmd,exitCode,stdout,stderr,durationMs}}",
    level: context.level,
    useCases: [useCases.metaMap.name, useCases.embeddedScripting.name],
  };
  calls.push(call);
};

export const outputJsonResult = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "output.json.result",
    title: "Write JSON result (array/value/lines)",
    note: "default: aggregated JSON array (sorted by locator/relPath); --lines streams nondeterministically; reduce → single value; embed per-item errors when configured",
    level: context.level,
    useCases: [useCases.outputJson.name],
  };
  calls.push(call);
};
