import { appendFile, writeFile } from "node:fs/promises";

/**
 * Canonical description of a capability or behavior we want to support.
 * - name: short machine-friendly identifier
 * - title: concise human-friendly description
 * - note: optional extra context or constraint
 */
export type UseCase = {
  name: string;
  title: string;
  note?: string;
};

/**
 * A single step in the flow-tree that the CLI will execute.
 * - name/title: brief identifiers for the step
 * - note: optional clarifying detail
 * - directory: logical Go package or path where code will live
 * - level: depth for tree rendering
 * - useCases: names of use-cases this step satisfies
 */
export type ComponentCall = {
  name: string;
  title: string;
  note?: string;
  directory?: string;
  level: number;
  useCases?: string[];
  suggest?: {
    pkg?: string;
    func?: string;
    file?: string;
  };
};

/**
 * Carry indentation depth while walking the design tree.
 */
export type FlowContext = {
  level: number;
};

/**
 * Increase nesting level for child calls.
 */
export const incrContext = (flowContext: FlowContext) => ({
  level: flowContext.level + 1,
});

/**
 * Extract all referenced use-case names from a call list.
 */
export const toUseCaseSet = (calls: ComponentCall[]) => {
  const allUseCases = calls
    .flatMap(({ useCases }) => useCases)
    .filter((useCase) => typeof useCase === "string");
  return new Set(allUseCases);
};

/**
 * Reset the generated design report.
 */
export const resetReport = async () => {
  await writeFile("docs/designs/FLOW_DESIGN.md", "");
};

/**
 * Append a single line to the generated design report.
 */
export const appendToReport = async (line: string) => {
  await appendFile("docs/designs/FLOW_DESIGN.md", line + "\n", "utf8");
};

/**
 * Render the flow-tree as indented lines of titles.
 */
export const displayCallsAsText = async (calls: ComponentCall[]) => {
  for (const call of calls) {
    const spaces = " ".repeat(call.level * 2);
    await appendToReport(`${spaces}${call.title}`);
  }
};

/**
 * Render a detailed view of the call tree with notes and suggestions.
 */
export const displayCallsDetailed = async (calls: ComponentCall[]) => {
  for (const call of calls) {
    const base = " ".repeat(call.level * 2);
    await appendToReport(`${base}${call.title} [${call.name}]`);
    if (call.note) {
      await appendToReport(`${base}  - note: ${call.note}`);
    }
    const pkg = call.directory || call.suggest?.pkg;
    if (pkg) {
      await appendToReport(`${base}  - pkg: ${pkg}`);
    }
    if (call.suggest?.func) {
      await appendToReport(`${base}  - func: ${call.suggest.func}`);
    }
    if (call.suggest?.file) {
      await appendToReport(`${base}  - file: ${call.suggest.file}`);
    }
  }
};

/**
 * Pure set difference: items in A not in B.
 */
export const getSetDifference = (
  setA: Set<string>,
  setB: Set<string>,
): Set<string> => {
  return new Set([...setA].filter((item) => !setB.has(item)));
};

/**
 * Render bullet points suitable for markdown.
 */
export const toBulletPoints = (lines: string[]) =>
  lines.map((line) => `  - ${line}`).join("\n");

/**
 * Convenience to add a titled section to the report.
 */
export const appendSection = async (
  title: string,
  lines: string[] | string,
) => {
  await appendToReport(`\n## ${title}`);
  if (Array.isArray(lines)) {
    await appendToReport(toBulletPoints(lines));
  } else {
    await appendToReport(lines);
  }
};
