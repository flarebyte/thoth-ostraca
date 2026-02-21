import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth } from './helpers';
import {
  normalizeWorkersOut,
  seededShuffle,
  sortKey,
  writeCueConfig,
} from './test_utils';

test('stress ingestion determinism at scale is byte-identical across runs and workers', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-stress-ingest-'));

  const total = 720;
  const order = seededShuffle(
    Array.from({ length: total }, (_, i) => i),
    Date.now() & 0xffffffff,
  );
  for (const i of order) {
    const a = String(i % 13).padStart(2, '0');
    const b = String(Math.floor(i / 13) % 11).padStart(2, '0');
    const dir = path.join(repo, 'tree', a, b);
    fs.mkdirSync(dir, { recursive: true });
    const locator = `stress/${String(i).padStart(4, '0')}`;
    const file = path.join(dir, `f${String(i).padStart(4, '0')}.thoth.yaml`);
    fs.writeFileSync(
      file,
      `locator: "${locator}"\nmeta:\n  i: ${i}\n  enabled: true\n`,
      'utf8',
    );
  }

  const makeCfg = (workers: number): string => {
    const cfg = path.join(repo, `w${workers}.cue`);
    writeCueConfig(
      cfg,
      `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "fail-fast" }
  output: { lines: false }
  workers: ${workers}
}`,
    );
    return cfg;
  };

  const runRepeat = (cfg: string, reps: number) =>
    Array.from({ length: reps }, () =>
      runThoth(bin, ['run', '--config', cfg], root),
    );

  const runs1 = runRepeat(makeCfg(1), 3);
  const runs8 = runRepeat(makeCfg(8), 3);

  for (const r of [...runs1, ...runs8]) {
    expect(r.status).toBe(0);
    expect(r.stderr).toBe('');
  }

  expect(runs1[0].stdout).toBe(runs1[1].stdout);
  expect(runs1[1].stdout).toBe(runs1[2].stdout);
  expect(runs8[0].stdout).toBe(runs8[1].stdout);
  expect(runs8[1].stdout).toBe(runs8[2].stdout);
  expect(normalizeWorkersOut(runs1[0].stdout)).toBe(
    normalizeWorkersOut(runs8[0].stdout),
  );
});

test('stress mixed failures keep-going is deterministic with expected errors', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-stress-mixed-'));

  const validCount = 120;
  const invalidCount = 20;
  const oversizedCount = 20;
  const unreadableTarget = process.platform === 'win32' ? 0 : 5;
  let unreadableApplied = 0;

  for (let i = 0; i < validCount; i += 1) {
    const dir = path.join(repo, 'ok', String(i % 9));
    fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(
      path.join(dir, `v${String(i).padStart(3, '0')}.thoth.yaml`),
      `locator: "ok/${String(i).padStart(3, '0')}"\nmeta:\n  i: ${i}\n  enabled: true\n`,
      'utf8',
    );
  }
  for (let i = 0; i < invalidCount; i += 1) {
    const dir = path.join(repo, 'bad', String(i % 5));
    fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(
      path.join(dir, `inv${String(i).padStart(3, '0')}.thoth.yaml`),
      `locator: "inv/${String(i).padStart(3, '0')}"\n`,
      'utf8',
    );
  }
  for (let i = 0; i < oversizedCount; i += 1) {
    const dir = path.join(repo, 'big', String(i % 4));
    fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(
      path.join(dir, `big${String(i).padStart(3, '0')}.thoth.yaml`),
      `locator: "big/${String(i).padStart(3, '0')}"\nmeta:\n  payload: "${'x'.repeat(512)}"\n`,
      'utf8',
    );
  }
  if (unreadableTarget > 0) {
    for (let i = 0; i < unreadableTarget; i += 1) {
      const dir = path.join(repo, 'perm');
      fs.mkdirSync(dir, { recursive: true });
      const p = path.join(dir, `u${String(i).padStart(3, '0')}.thoth.yaml`);
      fs.writeFileSync(
        p,
        `locator: "perm/${String(i).padStart(3, '0')}"\nmeta:\n  i: ${i}\n  enabled: true\n`,
        'utf8',
      );
      try {
        fs.chmodSync(p, 0o000);
        try {
          fs.readFileSync(p, 'utf8');
          fs.chmodSync(p, 0o644);
        } catch {
          unreadableApplied += 1;
        }
      } catch {}
    }
  }

  const cfg = path.join(repo, 'mixed.cue');
  writeCueConfig(
    cfg,
    `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
  limits: { maxYAMLBytes: 128 }
  workers: 8
}`,
  );

  const runs = Array.from({ length: 3 }, () =>
    runThoth(bin, ['run', '--config', cfg], root),
  );
  for (const r of runs) {
    expect(r.status).toBe(0);
    expect(r.stderr).toBe('');
  }
  expect(runs[0].stdout).toBe(runs[1].stdout);
  expect(runs[1].stdout).toBe(runs[2].stdout);

  const out = JSON.parse(runs[0].stdout) as {
    records: Array<{
      locator: string;
      error?: { stage?: string; message?: string };
    }>;
    errors: Array<{ stage: string; locator: string; message: string }>;
  };
  const expectedFailures = invalidCount + oversizedCount + unreadableApplied;
  expect(out.errors.length).toBe(expectedFailures);
  expect(
    out.errors.every((e, i, arr) =>
      i === 0 ? true : sortKey(arr[i - 1]) <= sortKey(e),
    ),
  ).toBe(true);
  expect(
    out.records.some(
      (r) => !r.error && r.locator.startsWith('ok/') && r.locator.length > 3,
    ),
  ).toBe(true);
  expect(
    out.records.some((r) => r.error?.stage === 'parse-validate-yaml'),
  ).toBe(true);

  if (unreadableApplied > 0) {
    for (let i = 0; i < unreadableTarget; i += 1) {
      const p = path.join(
        repo,
        'perm',
        `u${String(i).padStart(3, '0')}.thoth.yaml`,
      );
      try {
        fs.chmodSync(p, 0o644);
      } catch {}
    }
  }
});
