import { writeFile } from "node:fs/promises";
import { risks } from "./risks.mts";

const RISKS_PATH = "docs/RISKS.md";

export const generateRisksReport = async () => {
  const entries = Object.values(risks).sort((a, b) => a.name.localeCompare(b.name));

  const lines: string[] = [];
  lines.push("# Risks Overview (Generated)");
  lines.push("");
  lines.push("This document summarizes key risks and mitigations.");
  lines.push("");

  // Summary table
  lines.push("## Summary");
  for (const r of entries) {
    lines.push(`- ${r.title} [${r.name}]`);
  }
  lines.push("");

  // Detailed sections
  for (const r of entries) {
    lines.push(`## ${r.title} [${r.name}]`);
    lines.push("");
    lines.push(`- Description: ${r.description}`);
    lines.push(`- Mitigation: ${r.mitigation}`);
    lines.push("");
  }

  await writeFile(RISKS_PATH, lines.join("\n"), "utf8");
};

