import { expect, test } from 'bun:test';
import { spawnSync } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth } from './helpers';
import { normalizeWorkersOut } from './test_utils';

function mkShellRepo(root: string): string {
  const base = fs.mkdtempSync(path.join(root, 'temp', 'phase4-shell-'));
  fs.mkdirSync(path.join(base, 'sub'), { recursive: true });
  fs.writeFileSync(
    path.join(base, 'a.thoth.yaml'),
    'locator: "a"\nmeta:\n  name: "A"\n',
    'utf8',
  );
  fs.writeFileSync(path.join(base, 'sub', 'known.txt'), 'ok', 'utf8');
  return base;
}

function toCuePath(p: string): string {
  return p.replaceAll('\\', '\\\\');
}

function writeCfg(
  root: string,
  repo: string,
  workers: number,
  shellBlock: string,
): string {
  const cfg = path.join(
    root,
    'temp',
    `phase4-shell-${process.pid}-${Date.now()}-${Math.random().toString(16).slice(2)}.cue`,
  );
  const body = `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${toCuePath(repo)}" }
  workers: ${workers}
  shell: ${shellBlock}
}`;
  fs.writeFileSync(cfg, body, 'utf8');
  return cfg;
}

function firstShellStdout(stdout: string): string {
  const env = JSON.parse(stdout) as {
    records?: Array<{ shell?: { stdout?: string } }>;
  };
  return env.records?.[0]?.shell?.stdout ?? '';
}

function firstShell(stdout: string): Record<string, unknown> {
  const env = JSON.parse(stdout) as {
    records?: Array<{ shell?: Record<string, unknown> }>;
  };
  const sh = env.records?.[0]?.shell;
  if (!sh) {
    throw new Error('missing shell in first record');
  }
  return sh;
}

function shellAvailable(program: string): boolean {
  const p = spawnSync(program, ['-c', 'printf ok'], { encoding: 'utf8' });
  return !p.error;
}

test('phase4 shell strict templating enforces unknown placeholders', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  const strictCfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'%s\' \'{oops}\'"], strictTemplating: true }',
  );
  const strictRun = runThoth(bin, ['run', '--config', strictCfg], root);
  expect(strictRun.status).not.toBe(0);
  expect(strictRun.stdout).toBe('');
  expect(
    strictRun.stderr.includes('strict templating: invalid placeholder {oops}'),
  ).toBe(true);

  const looseCfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'%s\' \'{oops}\'"], strictTemplating: false }',
  );
  const looseRun = runThoth(bin, ['run', '--config', looseCfg], root);
  expect(looseRun.status).toBe(0);
  expect(looseRun.stderr).toBe('');
  expect(firstShellStdout(looseRun.stdout)).toBe('{oops}');
});

test('phase4 shell timeout terminates and returns stable error', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  const cfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "sleep 2"], timeoutMs: 30, termGraceMs: 10, killProcessGroup: true }',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('shell-exec: timeout')).toBe(true);
});

test('phase4 shell capture truncates stdout at maxBytes', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  const cfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'0123456789\'"], capture: { stdout: true, stderr: true, maxBytes: 5 } }',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(firstShellStdout(run.stdout)).toBe('01234');
  const shell = firstShell(run.stdout);
  expect(shell.stdoutTruncated).toBe(true);
  expect(shell.stderrTruncated).toBe(false);
  expect(shell.timedOut).toBe(false);
});

test('phase4 shell canonical schema on success and non-zero exit', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  const okCfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'ok\'"] }',
  );
  const okRun = runThoth(bin, ['run', '--config', okCfg], root);
  expect(okRun.status).toBe(0);
  expect(okRun.stderr).toBe('');
  const okShell = firstShell(okRun.stdout);
  expect(okShell.exitCode).toBe(0);
  expect(okShell.timedOut).toBe(false);
  expect(okShell.stdoutTruncated).toBe(false);
  expect(okShell.stderrTruncated).toBe(false);
  expect(okShell.error).toBeUndefined();

  const badCfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'bad\' >&2; exit 7"] }',
  );
  const badRun = runThoth(bin, ['run', '--config', badCfg], root);
  expect(badRun.status).toBe(0);
  expect(badRun.stderr).toBe('');
  const badShell = firstShell(badRun.stdout);
  expect(badShell.exitCode).toBe(7);
  expect(badShell.timedOut).toBe(false);
  expect(badShell.stdoutTruncated).toBe(false);
  expect(badShell.stderrTruncated).toBe(false);
  expect(badShell.error).toBeUndefined();
});

test('phase4 shell keep-going timeout and capture-disabled schema', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);
  const cfg = path.join(
    root,
    'temp',
    `phase4-shell-timeout-${process.pid}-${Date.now()}-${Math.random().toString(16).slice(2)}.cue`,
  );
  const body = `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${toCuePath(repo)}" }
  workers: 1
  errors: { mode: "keep-going", embedErrors: true }
  shell: {
    enabled: true
    program: "sh"
    timeoutMs: 30
    termGraceMs: 10
    capture: { stdout: false, stderr: true, maxBytes: 16 }
    argsTemplate: ["-c", "sleep 2"]
  }
}`;
  fs.writeFileSync(cfg, body, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(1);
  expect(run.stderr.includes('keep-going: no successful records')).toBe(true);
  const shell = firstShell(run.stdout);
  expect(shell.exitCode).toBe(-2);
  expect(shell.timedOut).toBe(true);
  expect(shell.stdout).toBeUndefined();
  expect(shell.stdoutTruncated).toBe(false);
  expect(shell.stderrTruncated).toBe(false);
});

test('phase4 shell uses configured workingDir and env overlay', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  const dirCfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", workingDir: "sub", argsTemplate: ["-c", "pwd"] }',
  );
  const dirRun = runThoth(bin, ['run', '--config', dirCfg], root);
  expect(dirRun.status).toBe(0);
  expect(dirRun.stderr).toBe('');
  const cwdOut = firstShellStdout(dirRun.stdout).trim().replaceAll('\\\\', '/');
  expect(cwdOut.endsWith('/sub')).toBe(true);

  const envCfg = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", env: { THOTH_E2E_ENV: "phase4" }, argsTemplate: ["-c", "printf \'%s\' $THOTH_E2E_ENV"] }',
  );
  const envRun = runThoth(bin, ['run', '--config', envCfg], root);
  expect(envRun.status).toBe(0);
  expect(envRun.stderr).toBe('');
  expect(firstShellStdout(envRun.stdout)).toBe('phase4');
});

test('phase4 shell supports sh/bash/zsh (skip if unavailable)', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  for (const program of ['sh', 'bash', 'zsh']) {
    if (!shellAvailable(program)) {
      if (program === 'sh') {
        throw new Error('required shell sh not available');
      }
      continue;
    }
    const cfg = writeCfg(
      root,
      repo,
      1,
      `{ enabled: true, program: "${program}", argsTemplate: ["-c", "printf 'ok'"] }`,
    );
    const run = runThoth(bin, ['run', '--config', cfg], root);
    expect(run.status).toBe(0);
    expect(run.stderr).toBe('');
    expect(firstShellStdout(run.stdout)).toBe('ok');
  }
});

test('phase4 shell determinism across reruns and workers', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkShellRepo(root);

  const cfg1 = writeCfg(
    root,
    repo,
    1,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'%s\' \'{json}\'"] }',
  );
  const cfg8 = writeCfg(
    root,
    repo,
    8,
    '{ enabled: true, program: "sh", argsTemplate: ["-c", "printf \'%s\' \'{json}\'"] }',
  );

  const r11 = runThoth(bin, ['run', '--config', cfg1], root);
  const r12 = runThoth(bin, ['run', '--config', cfg1], root);
  const r8 = runThoth(bin, ['run', '--config', cfg8], root);

  expect(r11.status).toBe(0);
  expect(r12.status).toBe(0);
  expect(r8.status).toBe(0);
  expect(r11.stderr).toBe('');
  expect(r12.stderr).toBe('');
  expect(r8.stderr).toBe('');

  expect(r11.stdout).toBe(r12.stdout);
  expect(normalizeWorkersOut(r11.stdout)).toBe(normalizeWorkersOut(r8.stdout));
});
