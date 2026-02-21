import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth } from './helpers';

type RunEnvelope = {
  records: Array<{
    locator: string;
    error?: { stage: string; message: string };
  }>;
  errors?: Array<{ stage: string; locator?: string; message: string }>;
  meta?: Record<string, unknown>;
};

function writeCfg(
  root: string,
  name: string,
  cfg: Record<string, unknown>,
): string {
  const p = path.join(root, 'temp', `${name}.cue`);
  fs.mkdirSync(path.dirname(p), { recursive: true });
  fs.writeFileSync(p, `${JSON.stringify(cfg, null, 2)}\n`, 'utf8');
  return p;
}

function isSortedErrors(
  errs: Array<{ stage: string; locator?: string; message: string }>,
): boolean {
  for (let i = 1; i < errs.length; i++) {
    const a = errs[i - 1];
    const b = errs[i];
    if (!a || !b) continue;
    const ak = `${a.stage}\u0000${a.locator ?? ''}\u0000${a.message}`;
    const bk = `${b.stage}\u0000${b.locator ?? ''}\u0000${b.message}`;
    if (ak > bk) return false;
  }
  return true;
}

function parseOut(stdout: string): RunEnvelope {
  return JSON.parse(stdout) as RunEnvelope;
}

function copyTree(src: string, dst: string): string {
  fs.rmSync(dst, { recursive: true, force: true });
  fs.cpSync(src, dst, { recursive: true });
  return dst;
}

test('error model keep-going+embedErrors for pipeline produces partial success and deterministic sorted errors', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'err-model-pipeline');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });

  fs.writeFileSync(
    path.join(repo, 'ok.thoth.yaml'),
    'locator: ok.txt\nmeta:\n  cmd: ok\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'bad_locator.thoth.yaml'),
    'locator: ../bad.txt\nmeta: {}\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'invalid.thoth.yaml'),
    'locator: x\nmeta:\n  -\n',
    'utf8',
  );

  const cfg = writeCfg(root, 'err-model-pipeline-keep', {
    configVersion: '1',
    action: 'pipeline',
    discovery: { root: repo },
    errors: { mode: 'keep-going', embedErrors: true },
    map: {
      inline: 'return { locator = locator, cmd = (meta and meta.cmd) or "ok" }',
    },
  });

  const r1 = runThoth(bin, ['run', '--config', cfg], root);
  const r2 = runThoth(bin, ['run', '--config', cfg], root);
  expect(r1.status).toBe(0);
  expect(r2.status).toBe(0);
  expect(r1.stderr).toBe('');
  expect(r2.stderr).toBe('');
  expect(r1.stdout).toBe(r2.stdout);

  const out = parseOut(r1.stdout);
  const errs = out.errors ?? [];
  expect(errs.length).toBeGreaterThanOrEqual(2);
  expect(isSortedErrors(errs)).toBe(true);
  expect(errs.some((e) => e.stage === 'parse-validate-yaml')).toBe(true);
  expect(errs.some((e) => e.stage === 'validate-locators')).toBe(true);
  expect(out.records.some((r) => r.locator === 'ok.txt' && !r.error)).toBe(
    true,
  );
});

test('error model fail-fast for pipeline aborts with non-zero exit', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'err-model-pipeline-fast');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });
  fs.writeFileSync(
    path.join(repo, 'bad.thoth.yaml'),
    'locator: ../bad.txt\nmeta: {}\n',
    'utf8',
  );

  const cfg = writeCfg(root, 'err-model-pipeline-failfast', {
    configVersion: '1',
    action: 'pipeline',
    discovery: { root: repo },
    errors: { mode: 'fail-fast', embedErrors: true },
  });
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('invalid locator')).toBe(true);
});

test('error model keep-going with shell missing program records shell errors and exits non-zero when all fail', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'err-model-pipeline-shell-missing');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });
  fs.writeFileSync(
    path.join(repo, 'a.thoth.yaml'),
    'locator: a\nmeta: {}\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'b.thoth.yaml'),
    'locator: b\nmeta: {}\n',
    'utf8',
  );

  const cfg = writeCfg(root, 'err-model-pipeline-shell-missing', {
    configVersion: '1',
    action: 'pipeline',
    discovery: { root: repo },
    errors: { mode: 'keep-going', embedErrors: true },
    shell: {
      enabled: true,
      program: 'definitely-missing-program-xyz',
      argsTemplate: ['{json}'],
    },
  });
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  const out = parseOut(run.stdout);
  const errs = out.errors ?? [];
  expect(errs.some((e) => e.stage === 'shell-exec')).toBe(true);
  expect(out.records.every((r) => r.error?.stage === 'shell-exec')).toBe(true);
});

test('error model keep-going+embedErrors for update-meta keeps successes and records writer/load failures', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const src = path.join(root, 'testdata', 'repos', 'update1');
  const repo = copyTree(src, path.join(root, 'temp', 'err-model-update-keep'));

  fs.rmSync(path.join(repo, 'a.txt.thoth.yaml'), { force: true });
  fs.mkdirSync(path.join(repo, 'a.txt.thoth.yaml'), { recursive: true });

  const cfg = writeCfg(root, 'err-model-update-keep', {
    configVersion: '1',
    action: 'update-meta',
    discovery: { root: repo },
    errors: { mode: 'keep-going', embedErrors: true },
    fileInfo: { enabled: true },
  });
  const runOnce = () => {
    copyTree(src, repo);
    fs.rmSync(path.join(repo, 'a.txt.thoth.yaml'), { force: true });
    fs.mkdirSync(path.join(repo, 'a.txt.thoth.yaml'), { recursive: true });
    return runThoth(bin, ['run', '--config', cfg], root);
  };
  const r1 = runOnce();
  const r2 = runOnce();
  expect(r1.status).toBe(0);
  expect(r2.status).toBe(0);
  expect(r1.stderr).toBe('');
  expect(r2.stderr).toBe('');
  expect(r1.stdout).toBe(r2.stdout);

  const out = parseOut(r1.stdout);
  const errs = out.errors ?? [];
  expect(errs.length).toBeGreaterThan(0);
  expect(isSortedErrors(errs)).toBe(true);
  expect(errs.some((e) => e.stage === 'load-existing-meta')).toBe(true);
  expect(out.records.some((r) => r.error?.stage === 'load-existing-meta')).toBe(
    true,
  );
  expect(out.records.some((r) => r.locator === 'b.txt' && !r.error)).toBe(true);
});

test('error model fail-fast for update-meta aborts with non-zero exit', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const src = path.join(root, 'testdata', 'repos', 'update1');
  const repo = copyTree(src, path.join(root, 'temp', 'err-model-update-fast'));
  fs.rmSync(path.join(repo, 'a.txt.thoth.yaml'), { force: true });
  fs.mkdirSync(path.join(repo, 'a.txt.thoth.yaml'), { recursive: true });

  const cfg = writeCfg(root, 'err-model-update-failfast', {
    configVersion: '1',
    action: 'update-meta',
    discovery: { root: repo },
    errors: { mode: 'fail-fast', embedErrors: true },
  });
  copyTree(src, repo);
  fs.rmSync(path.join(repo, 'a.txt.thoth.yaml'), { force: true });
  fs.mkdirSync(path.join(repo, 'a.txt.thoth.yaml'), { recursive: true });
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('load-existing-meta')).toBe(true);
});

test('error model keep-going for diff-meta returns valid output with sorted errors and deterministic bytes', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const src = path.join(root, 'testdata', 'repos', 'diff1');
  const repo = copyTree(src, path.join(root, 'temp', 'err-model-diff-keep'));

  fs.writeFileSync(
    path.join(repo, 'bad.thoth.yaml'),
    'locator: bad\nmeta:\n  -\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'badloc.thoth.yaml'),
    'locator: ../x\nmeta: {}\n',
    'utf8',
  );

  const cfg = writeCfg(root, 'err-model-diff-keep', {
    configVersion: '1',
    action: 'diff-meta',
    discovery: { root: repo },
    errors: { mode: 'keep-going', embedErrors: true },
  });
  const r1 = runThoth(bin, ['run', '--config', cfg], root);
  const r2 = runThoth(bin, ['run', '--config', cfg], root);
  expect(r1.status).toBe(0);
  expect(r2.status).toBe(0);
  expect(r1.stderr).toBe('');
  expect(r2.stderr).toBe('');
  expect(r1.stdout).toBe(r2.stdout);

  const out = parseOut(r1.stdout);
  const errs = out.errors ?? [];
  expect(errs.length).toBeGreaterThan(0);
  expect(isSortedErrors(errs)).toBe(true);
  expect(out.meta && typeof out.meta === 'object' && 'diff' in out.meta).toBe(
    true,
  );
});
