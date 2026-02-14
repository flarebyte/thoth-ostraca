import { stringify as yamlStringify } from 'bun:yaml';
import { appendFile, mkdir, writeFile } from 'node:fs/promises';

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
 * Canonical description of a risk in the design.
 * - name: short machine-friendly identifier
 * - title: concise human-friendly description
 * - description: what could go wrong and why
 * - mitigation: actions or controls to reduce impact/likelihood
 */
export type Risk = {
  name: string;
  title: string;
  description: string;
  mitigation: string;
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
 * One stickie per idea by default
 */
export type Stickie = {
  //format for name: blackboard-export-format
  name: string;
  // Keep each note concise and focused on what is not obvious, with emphasis
  // on the “why” (intent, trade-offs, constraints, and rationale). Prefer neutral,
  // implementation-agnostic phrasing.
  note?: string;
  //code in programming language
  code?: string;
  // Use labels from this controlled set when
  // relevant: usecase, example, flow, design, implementation, decision, security, operations,
  // compliance, glossary, principle, validation, howto, faq, library, wip, stable.
  labels: string[];
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
    .filter((useCase) => typeof useCase === 'string');
  return new Set(allUseCases);
};

/**
 * Reset the generated design report.
 */
export const resetReport = async () => {
  await writeFile('docs/designs/FLOW_DESIGN.md', '');
};

/**
 * Append a single line to the generated design report.
 */
export const appendToReport = async (line: string) => {
  await appendFile('docs/designs/FLOW_DESIGN.md', `${line}\n`, 'utf8');
};

/**
 * Render the flow-tree as indented lines of titles.
 */
export const displayCallsAsText = async (calls: ComponentCall[]) => {
  for (const call of calls) {
    const spaces = ' '.repeat(call.level * 2);
    await appendToReport(`${spaces}${call.title}`);
  }
};

/**
 * Render a detailed view of the call tree with notes and suggestions.
 */
export const displayCallsDetailed = async (calls: ComponentCall[]) => {
  for (const call of calls) {
    const base = ' '.repeat(call.level * 2);
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
  lines.map((line) => `  - ${line}`).join('\n');

/**
 * Convert a free-form title/name into a dash-lower slug suitable for stickie names and filenames.
 * Rules: lowercase, replace non [a-z0-9] with '-', collapse repeats, trim edges.
 */
export const toStickieName = (s: string): string =>
  s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');

/**
 * Create a stickie YAML file for a section under `notes/`.
 * Non-fatal: any error while writing the stickie is swallowed so docs generation continues.
 */
export const writeSectionStickie = async (
  title: string,
  lines: string[] | string,
  outDir = 'notes',
) => {
  try {
    await mkdir(outDir, { recursive: true });
    const name = toStickieName(title);
    const note = Array.isArray(lines) ? toBulletPoints(lines) : lines;
    const stickie: Stickie = {
      name,
      note,
      labels: ['design'],
    };
    const yaml = yamlStringify(stickie, { indent: 2 });
    await writeFile(`${outDir}/${name}.stickie.yaml`, yaml, 'utf8');
  } catch (_) {
    // Best-effort: ignore errors (e.g., sandboxed environments)
  }
};

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
  // Also emit a stickie for this section (best-effort)
  await writeSectionStickie(title, lines);
};

/**
 * Build human-readable lines for a set of use-case names using a catalog.
 */
export const toUseCaseLines = (
  useCaseNames: Set<string>,
  catalogByName: Record<string, UseCase>,
): string[] => {
  return [...useCaseNames].map((name) => {
    const uc = catalogByName[name];
    if (!uc) return name;
    return uc.note ? `${uc.title} — ${uc.note}` : uc.title;
  });
};

/**
 * Append a small heading then bullet list for use-cases (title + note).
 */
export const appendUseCases = async (
  heading: string,
  useCaseNames: Set<string>,
  catalogByName: Record<string, UseCase>,
) => {
  await appendToReport(`${heading}\n`);
  await appendToReport(
    toBulletPoints(toUseCaseLines(useCaseNames, catalogByName)),
  );
  await appendToReport('\n');
};

/**
 * Append a key/value list as a section, preserving entry order.
 */
export const appendKeyValueList = async (
  title: string,
  entries: Array<[string, string | string[]]>,
) => {
  await appendToReport(`\n## ${title}`);
  const lines = entries.map(([k, v]) => {
    const value = Array.isArray(v) ? v.join(', ') : v;
    return `${k}: ${value}`;
  });
  await appendToReport(toBulletPoints(lines));
};

/**
 * Baseline risk catalogue for the project.
 * Add or refine entries as the design evolves.
 */
// Risk records live in docs/designs/risks.mts
