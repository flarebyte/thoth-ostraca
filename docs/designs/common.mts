export type UseCase = {
  name: string;
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

export const displayAsText = (calls: ComponentCall[]) => {
  for (const call of calls) {
    const spaces = " ".repeat(call.level * 2);
    console.log(`${spaces}${call.title}`);
  }
};
