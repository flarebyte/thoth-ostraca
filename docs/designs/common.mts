export type ProcessingTime = {
  minMilli: number;
  maxMilli: number;
};

export type Characteristics = {
  evolution?: number;
  maintenance?: number;
  security?: number;
  operations?: number;
  monitoring?: number;
  accessibility?: number;
  internationalisation?: number;
};

export type ComponentCall = {
  name: string;
  title: string;
  directory?: string;
  mustHave?: string[];
  shouldHave?: string[];
  couldHave?: string[];
  wontHave?: string[];
  drawbacks?: string[];
  processingTime: ProcessingTime;
  characteristics: Characteristics;
};

export type FlowContext = {
    level: number;
}