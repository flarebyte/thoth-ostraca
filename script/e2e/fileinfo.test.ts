import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import { buildBinary, projectRoot, runThoth, saveOutputs } from './helpers';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function setAllFileTimes(root: string, iso: string): void {
  const ts = Math.floor(new Date(iso).getTime() / 1000);
  const walk = (dir: string) => {
    for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
      const p = path.join(dir, ent.name);
      if (ent.isDirectory()) {
        walk(p);
      } else {
        fs.utimesSync(p, ts, ts);
      }
    }
  };
  walk(root);
}

test('enrich-fileinfo attaches deterministic fields for create-meta', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const srcRepo = path.join(root, 'testdata/repos/create1');
  const tempRepo = path.join(root, 'temp', 'fileinfo1_repo');
  fs.rmSync(tempRepo, { recursive: true, force: true });
  fs.mkdirSync(path.dirname(tempRepo), { recursive: true });
  fs.cpSync(srcRepo, tempRepo, { recursive: true });
  // Set deterministic mod times
  const fixed = '2024-01-02T03:04:05Z';
  setAllFileTimes(tempRepo, fixed);

  const cfgPath = path.join(root, 'temp', 'fileinfo1_tmp.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "create-meta"
  discovery: { root: "${path.join('temp', 'fileinfo1_repo').replaceAll('\\', '\\\\')}" }
  fileInfo: { enabled: true }
  errors: { mode: "keep-going", embedErrors: true }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');

  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'fileinfo-create-meta', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');

  const out = JSON.parse(run.stdout) as {
    records: Array<{
      locator: string;
      fileInfo?: {
        size: number;
        mode: string;
        modTime: string;
        isDir: boolean;
      };
    }>;
  };
  const recA = out.records.find((r) => r.locator === 'a.txt');
  const recB = out.records.find((r) => r.locator === 'dir/b.txt');
  expect(recA?.fileInfo).toBeTruthy();
  expect(recB?.fileInfo).toBeTruthy();
  expect(recA?.fileInfo?.modTime).toBe('2024-01-02T03:04:05Z');
  expect(recB?.fileInfo?.modTime).toBe('2024-01-02T03:04:05Z');
  expect(typeof recA?.fileInfo?.size).toBe('number');
  expect(typeof recA?.fileInfo?.isDir).toBe('boolean');
  expect(typeof recA?.fileInfo?.mode).toBe('string');
});

test('enrich-fileinfo keep-going embeds record error and envelope error on missing file', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const env = {
    records: [{ locator: 'missing.txt' }, { locator: 'README.md' }],
    meta: {
      contractVersion: '1',
      discovery: { root },
      fileInfo: { enabled: true },
      errors: { mode: 'keep-going', embedErrors: true },
    },
  };
  const dumpIn = path.join(root, 'temp', 'fileinfo-missing-in.json');
  fs.mkdirSync(path.dirname(dumpIn), { recursive: true });
  fs.writeFileSync(dumpIn, JSON.stringify(env), 'utf8');
  const run = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-fileinfo', '--in', dumpIn],
    root,
  );
  saveOutputs(root, 'fileinfo-missing', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const out = JSON.parse(run.stdout) as {
    records: Array<{ locator: string; error?: unknown }>;
    errors: Array<{ stage: string; locator: string; message: string }>;
  };
  const recMissing = out.records.find((r) => r.locator === 'missing.txt');
  expect(recMissing?.error).toBeTruthy();
  expect(
    out.errors.find(
      (e) => e.stage === 'enrich-fileinfo' && e.locator === 'missing.txt',
    ),
  ).toBeTruthy();
});
