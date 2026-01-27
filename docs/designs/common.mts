export type ProcessingTime = {
  minMilli: number;
  maxMilli: number;
};

export type ComponentCall = {
  name: string;
  title: string;
  directory: string | undefined;
  mustHave: string[] | undefined;
  shouldHave: string[] | undefined;
  couldHave: string[] | undefined;
  wontHave: string[] | undefined;
  drawbacks: string[] | undefined;
  processingTime: ProcessingTime;
};
