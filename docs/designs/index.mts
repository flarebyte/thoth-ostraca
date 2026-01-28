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
    note: "Filter metadata associated with file",
    priority_level: "must",
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
    useCases: mustUseCases,
  };
  calls.push(call);
  findMetaFiles(incrContext(context));
};

const findMetaFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "find meta files",
    title: "Find individual meta files",
    note: "yaml meta file",
    level: context.level,
    useCases: mustUseCases,
  };
  calls.push(call);
};

cliArgsMetaFind({ level: 0 });

displayAsText(calls);
