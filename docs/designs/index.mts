import {
  ComponentCall,
  displayAsText,
  FlowContext,
  incrContext,
  toUseCaseSet,
} from "./common.mts";

const calls: ComponentCall[] = [];

const useCases = {
  filterByMeta: {
    name: "filter by meta",
    title: "Filter metadata associated with locator",
  },
  mapMeta: {
    name: "map meta",
    title: "Map metadata associated with locator",
  },
  reduceMeta: {
    name: "reduce meta",
    title: "Reduce metadata for all locators",
  },
  mapReduceActionConfig: {
    name: "map reduce action config",
    title: "Load map reduce action config from file (YAML)",
  },
  scripting: {
    name: "scripting",
    title: "Filter map and reduce can be scripted (Lua)",
  },
  mapShell: {
    name: "map with shell",
    title: "Use map metadata for running shell with locator name",
  },
  locatorSupport: {
    name: "locator support",
    title: "Locator can be (relative) file or url",
  },
  parallelProcessing: {
    name: "parallel processing",
    title: "Processing is done in parallel",
    note: "Uses Goroutines and channels.",
  },
  batchCreation: {
    name: "batch creation",
    title: "batch creation of locators metadata",
  },
  batchUpdate: {
    name: "batch update",
    title: "batch update of locators metadata",
  },
  batchDiff: {
    name: "batch diff",
    title: "batch diff of locators metadata",
  },
};

const getByName = (expectedName: string) =>
  Object.values(useCases).find(({ name }) => name === expectedName);

const getTitlesForSet = (useCaseSet: Set<string>) =>
  [...useCaseSet].map((useCase) => getByName(useCase)?.title);

const mustUseCases = new Set([
  ...Object.values(useCases).map(({ name }) => name),
]);

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
  filterMetaLocators(incrContext(context));
};

const filterMetaLocators = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "filter meta locators",
    title: "Filter metadata for a locator using a Lua script",
    note: "yaml meta file",
    level: context.level,
    useCases: [useCases.filterByMeta.name],
  };
  calls.push(call);
};

cliArgsMetaFind({ level: 0 });

displayAsText(calls);

console.log(getTitlesForSet(mustUseCases));

console.log(toUseCaseSet(calls));
