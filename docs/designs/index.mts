import { ComponentCall, FlowContext, incrContext } from "./common.mts";

const calls: ComponentCall[] = [];

const cliArgsMetaFind = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "meta find",
    title: "Parse CLI args for metadata find",
    directory: "cmd",
    note: "Use cobra lib",
  };
  calls.push(call);
  findMetaFiles(incrContext(context));
};

const findMetaFiles = (context: FlowContext) => {
  const call: ComponentCall = {
    name: "find meta files",
    title: "Find individual meta files",
    note: "yaml meta file",
  };
  calls.push(call);
};

cliArgsMetaFind({level: 0});

console.log(calls);
