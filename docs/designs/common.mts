export type ProcessingTime = {
  minMilli: number;
  maxMilli: number;
};

export type Characteristics = {
  evolution: number | undefined;
  maintenance: number | undefined;
  security: number | undefined;
  operations: number | undefined;
  monitoring: number | undefined;
  accessibility: number | undefined;
  internationalisation: number | undefined;
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
  characteristics: Characteristics;
};
