export type Stickie = {
  id?: string;
  name?: string;
  note: string;
  labels?: string[];
  priority_level?: "must" | "should" | "could";
};

export type UseCase = Stickie;

export type ComponentCall = Stickie & {
  title: string;
  directory?: string;
};

export type FlowContext = {
  level: number;
};

export const incrContext = (flowContext: FlowContext) => ({
  level: flowContext.level + 1,
});
