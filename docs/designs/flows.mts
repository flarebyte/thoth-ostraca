import { FlowContext } from "./common.mts";
import { incrContext } from "./common.mts";
import { calls } from "./calls.mts";
import { ComponentCall } from "./common.mts";
import { useCases } from "./use_cases.mts";
import {
  findMetaLocators,
  parseYamlRecords,
  filterMetaLocators,
  mapMetaRecords,
  execShellFromMap,
  postMapShellResults,
  reduceMetaRecords,
  outputJsonResult,
  findFilesForCreate,
  enrichFilesWithOptionalInfo,
  filterFilenames,
  mapFilenames,
  postMapFromFiles,
  saveMetaFiles,
  findFilesForUpdate,
  loadExistingMeta,
  postMapUpdateFromFiles,
  updateMetaFiles,
  computeMetaDiffs,
  scanForOrphanMetas,
  validateMetaOnly,
  loadActionConfig,
} from "./steps.mts";

export const cliRoot = (context: FlowContext) => {
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

export const cliArgsRun = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "cli.run",
    title: "Parse args for run",
    directory: "cmd/thoth",
    note: "flags: --config (CUE .cue file). All other options belong in the action config.",
    level: context.level,
    useCases: [useCases.cliUX.name, useCases.outputJson.name],
  };
  calls.push(call);
  loadActionConfig(incrContext(context));
  routeByActionType(incrContext(context));
};

export const routeByActionType = (context: FlowContext) => {
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
  // Validate flow (schema/locator only; no transforms/shell)
  validateFlow(incrContext(context));
};

export const pipelineFlow = (context: FlowContext) => {
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

export const createFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.create",
    title: "Create meta files flow",
    level: context.level,
    useCases: [useCases.batchCreate.name],
  };
  calls.push(call);
  findFilesForCreate(incrContext(context));
  enrichFilesWithOptionalInfo(incrContext(context));
  filterFilenames(incrContext(context));
  mapFilenames(incrContext(context));
  postMapFromFiles(incrContext(context));
  saveMetaFiles(incrContext(context));
  outputJsonResult(incrContext(context));
};

export const updateFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.update",
    title: "Update meta files flow",
    level: context.level,
    useCases: [useCases.batchUpdate.name],
  };
  calls.push(call);
  findFilesForUpdate(incrContext(context));
  enrichFilesWithOptionalInfo(incrContext(context));
  filterFilenames(incrContext(context));
  mapFilenames(incrContext(context));
  loadExistingMeta(incrContext(context));
  postMapUpdateFromFiles(incrContext(context));
  updateMetaFiles(incrContext(context));
  outputJsonResult(incrContext(context));
};

export const diffFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.diff",
    title: "Diff meta files flow",
    level: context.level,
    useCases: [useCases.batchDiff.name],
  };
  calls.push(call);
  findFilesForUpdate(incrContext(context));
  enrichFilesWithOptionalInfo(incrContext(context));
  filterFilenames(incrContext(context));
  mapFilenames(incrContext(context));
  loadExistingMeta(incrContext(context));
  postMapUpdateFromFiles(incrContext(context));
  computeMetaDiffs(incrContext(context));
  scanForOrphanMetas(incrContext(context));
  outputJsonResult(incrContext(context));
};

export const validateFlow = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "flow.validate",
    title: "Validate meta files only",
    level: context.level,
    useCases: [useCases.metaValidateOnly.name, useCases.metaSchema.name],
  };
  calls.push(call);
  findMetaLocators(incrContext(context));
  parseYamlRecords(incrContext(context));
  validateMetaOnly(incrContext(context));
  outputJsonResult(incrContext(context));
};
