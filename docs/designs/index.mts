import { generateFlowDesignReport } from "./write_report.mts";
import { generateRisksReport } from "./write_risks.mts";

// Re-export config examples to preserve the previous API surface
export {
  ACTION_CONFIG_EXAMPLE,
  ACTION_CONFIG_CREATE_EXAMPLE,
  ACTION_CONFIG_CREATE_MINIMAL,
} from "./config_model.mts";

await generateFlowDesignReport();
await generateRisksReport();
