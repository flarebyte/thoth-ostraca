export type Stickie = {
  id?: string;
  name?: string;
  note: string;
  labels?: string[];
  priority_level?: string;
};

export type UseCase = Stickie;

export type ComponentCall = Stickie & {
  title: string;
  directory?: string;
  level: number;
  useCases?: UseCase[];
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
