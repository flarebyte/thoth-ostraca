import {
  ComponentCall,
  displayAsText,
  FlowContext,
  incrContext,
  UseCase,
} from "./common.mts";

const calls: ComponentCall[] = [];

const useCases = {
  filterByMeta: {
    name: "filter by meta",
    note: "Filter metadata associated with locator",
  },
  mapMeta: {
    name: "map meta",
    note: "Map metadata associated with locator",
  },
  reduceMeta: {
    name: "reduce meta",
    note: "reduce metadata for all locators",
  },
  mapReduceActionConfig: {
    name: "map reduce action config",
    note: "load map reduce action config from file (YAML)",
  },
  scripting: {
    name: "scripting",
    note: "filter map and reduce can be scripted (Lua)",
  },
  mapShell: {
    name: "map with shell",
    note: "use map metadata for running shell with locator name",
  },
  locatorSupport: {
    name: "locator support",
    note: "locator can be (relative) file or url",
  },
  parallelProcessing: {
    name: "parallel processing",
    note: "processing is done in parallel",
  },
  batchCreation: {
    name: "batch creation",
    note: "batch creation of locators metadata",
  },
  batchUpdate: {
    name: "batch update",
    note: "batch update of locators metadata",
  },
  batchDiff: {
    name: "batch diff",
    note: "batch diff of locators metadata",
  },
};

const mustUseCases: UseCase[] = [useCases.filterByMeta];

const cliArgsMetaFind = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta find",
    title: "Parse CLI args for metadata find",
    directory: "cmd",
    note: "Use cobra lib",
    level: context.level,
  };
  calls.push(call);
  findMetaLocators(incrContext(context));
};

const findMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "find meta locators",
    title: "Find individual meta locators",
    note: "yaml meta locator",
    level: context.level,
  };
  calls.push(call);
};

const filterMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "filter meta locators",
    title: "Filter metadata for a locator using a Lua script",
    note: "yaml meta file",
    level: context.level,
    useCases: mustUseCases,
  };
  calls.push(call);
};

cliArgsMetaFind({ level: 0 });

displayAsText(calls);
