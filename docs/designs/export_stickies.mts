import { mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { stringify as yamlStringify } from "bun:yaml";

import type { Stickie } from "./common.mts";
import { useCases } from "./use_cases.mts";

// Convert internal dotted name (e.g., "meta.filter") to stickie slug (e.g., "meta-filter").
// Rules: lowercase, replace any non [a-z0-9] with '-', collapse repeats, trim '-'.
export const toStickieName = (internalName: string): string => {
  return internalName
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
};

const toStickie = (internalName: string, title: string, note?: string): Stickie => {
  const mergedNote = note ? `${title} — ${note}` : title;
  return {
    name: toStickieName(internalName),
    note: mergedNote,
    labels: ["usecase"],
  };
};

export const writeUseCaseStickies = async (outDir = "notes") => {
  await mkdir(outDir, { recursive: true });
  const entries = Object.values(useCases).sort((a, b) => a.name.localeCompare(b.name));
  for (const uc of entries) {
    const stickie = toStickie(uc.name, uc.title, uc.note);
    const yaml = yamlStringify(stickie, { indent: 2 });
    const base = toStickieName(uc.name);
    const outPath = join(outDir, `${base}.stickie.yaml`);
    await writeFile(outPath, yaml, "utf8");
  }
};

if (import.meta.main) {
  await writeUseCaseStickies("notes");
}
