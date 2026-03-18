#!/usr/bin/env bun

// Release helper for Go binaries using Bun/TypeScript
// - Reads version from main.project.yaml (tags.version)
// - Builds multi-platform binaries via build-go.ts
// - Creates a GitHub release v<version> with generated artifacts

import { promises as fs } from 'node:fs';
import path from 'node:path';
import { readVersionFromProjectYAML, runChecked } from './script_helpers';

async function ensureGhAvailable() {
  const which = Bun.which('gh');
  if (!which) throw new Error('gh (GitHub CLI) is required');
}

async function listBuildArtifacts(dir = 'build'): Promise<string[]> {
  try {
    const items = await fs.readdir(dir);
    const files: string[] = [];
    for (const item of items) {
      const fullPath = path.join(dir, item);
      const st = await fs.stat(fullPath);
      if (st.isFile()) {
        files.push(fullPath);
      }
    }
    files.sort();
    return files;
  } catch {
    return [];
  }
}

async function main() {
  const args = process.argv.slice(2);
  const dryRun = args.includes('--dry-run');

  if (!dryRun) {
    await ensureGhAvailable();
  }
  const version = (await readVersionFromProjectYAML()).trim();
  if (!version)
    throw new Error('version not found in main.project.yaml (tags.version)');

  if (dryRun) {
    console.log(`[dry-run] Would build binaries via: bun run build-go.ts`);
    console.log(
      `[dry-run] Would create GitHub release v${version} from ./build/*`,
    );
    return;
  }

  // fresh build
  try {
    await fs.rm('build', { recursive: true });
  } catch {}
  await runChecked(['bun', 'run', 'build-go.ts']);

  const artifacts = await listBuildArtifacts('build');
  if (!artifacts.length) throw new Error('no build artifacts found');

  console.log(`Creating GitHub release v${version}`);
  await runChecked([
    'gh',
    'release',
    'create',
    `v${version}`,
    ...artifacts,
    '--generate-notes',
  ]);
}

main().catch((err) => {
  console.error(err?.message || err);
  process.exitCode = 1;
});
