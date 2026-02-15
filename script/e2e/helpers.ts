import { type SpawnSyncReturns, spawnSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

export function projectRoot(): string {
  // helpers.ts lives in script/e2e; project root is two levels up
  return path.resolve(__dirname, '../..');
}

export function buildBinary(root: string): string {
  const binDir = path.join(root, '.e2e-bin');
  fs.mkdirSync(binDir, { recursive: true });
  const bin = path.join(
    binDir,
    process.platform === 'win32' ? 'thoth.exe' : 'thoth',
  );
  const env = {
    ...process.env,
    CGO_ENABLED: '0',
    GOCACHE: path.join(root, '.gocache'),
    GOFLAGS: '-mod=mod',
  } as Record<string, string>;
  const build = spawnSync('go', ['build', '-o', bin, './cmd/thoth'], {
    cwd: root,
    env,
    encoding: 'utf8',
  });
  if (build.status !== 0) {
    throw new Error(
      `build failed (status ${build.status})\n${build.stdout}\n${build.stderr}`,
    );
  }
  return bin;
}

export function runThoth(
  bin: string,
  args: string[],
  root: string,
): SpawnSyncReturns<string> {
  return spawnSync(bin, args, { encoding: 'utf8', cwd: root });
}

export function expectedJSONFromGolden(root: string, relPath: string): string {
  const raw = fs.readFileSync(path.join(root, relPath), 'utf8');
  return JSON.stringify(JSON.parse(raw)) + '\n';
}

export function saveOutputs(
  root: string,
  base: string,
  run: SpawnSyncReturns<string>,
): void {
  const tempDir = path.join(root, 'temp');
  fs.mkdirSync(tempDir, { recursive: true });
  fs.writeFileSync(path.join(tempDir, `${base}.out.txt`), run.stdout);
  fs.writeFileSync(
    path.join(tempDir, `${base}.err.txt`),
    (run as any).stderr ?? '',
  );
}
