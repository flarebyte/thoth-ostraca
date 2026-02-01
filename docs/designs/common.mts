import { appendFile, writeFile } from "node:fs/promises";

export type UseCase = {
  name: string;
  title: string;
  note: string;
};

export type ComponentCall = {
  name: string;
  title: string;
  note: string;
  directory?: string;
  level: number;
  useCases?: string[];
};

export type FlowContext = {
  level: number;
};

export const incrContext = (flowContext: FlowContext) => ({
  level: flowContext.level + 1,
});

export const toUseCaseSet = (calls: ComponentCall[]) => {
  const allUseCases = calls
    .flatMap(({ useCases }) => useCases)
    .filter((useCase) => typeof useCase === "string");
  return new Set(allUseCases);
};

export const resetReport = async () => {
  await writeFile("docs/designs/FLOW_DESIGN.md", "");
};

export const appendToReport = async (line: string) => {
  await appendFile("docs/designs/FLOW_DESIGN.md", line + "\n", "utf8");
};

export const displayCallsAsText = async (calls: ComponentCall[]) => {
  for (const call of calls) {
    const spaces = " ".repeat(call.level * 2);
    await appendToReport(`${spaces}${call.title}`);
  }
};

export const getSetDifference = (
  setA: Set<string>,
  setB: Set<string>,
): Set<string> => {
  return new Set([...setA].filter((item) => !setB.has(item)));
};
export const toBulletPoints = (lines: string[]) =>
  lines.map((line) => `  - ${line}`).join("\n");
