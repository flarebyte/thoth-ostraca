import { ComponentCall } from "./common.mts";

const calls: ComponentCall[] = [];

const cliArgsMetaFind = () => {
  const call: ComponentCall = {
    name: "meta find",
    title: "Parse CLI args for metadata find",
    directory: "cmd",
    mustHave: ["Use cobra lib"],
    processingTime: {
      minMilli: 1,
      maxMilli: 1,
    },
    characteristics: {
      evolution: 0.8,
      maintenance: 0.8,
      security: 0.8,
      operations: 0.8,
    },
  };
  calls.push(call);
  findMetaFiles();
};

const findMetaFiles = () => {
  const call: ComponentCall = {
    name: "find meta files",
    title: "Find individual meta files",
    mustHave: ["yaml meta file"],
    processingTime: {
      minMilli: 1,
      maxMilli: 1000,
    },
    characteristics: {
      evolution: 0.8,
      maintenance: 0.8,
      security: 0.8,
      operations: 0.8,
    },
  };
  calls.push(call);
};

cliArgsMetaFind();

console.log(calls);
