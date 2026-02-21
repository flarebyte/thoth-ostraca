import { expect, test } from 'bun:test';
import { spawnSync } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import { buildBinary, projectRoot, runThoth, saveOutputs } from './helpers';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function sh(
  cwd: string,
  cmd: string,
  args: string[],
  env?: Record<string, string>,
) {
  const run = spawnSync(cmd, args, {
    cwd,
    encoding: 'utf8',
    env: { ...process.env, ...(env ?? {}) },
  });
  if (run.status !== 0) {
    throw new Error(`${cmd} ${args.join(' ')} failed: ${run.stderr}`);
  }
}

type EnrichOut = {
  records: Array<{
    locator: string;
    git?: {
      tracked: boolean;
      ignored: boolean;
      status: string;
      lastCommit: null | { author: string; time: string; hash: string };
    };
  }>;
  meta?: Record<string, unknown>;
};

function setupGitFixture(repo: string): { subDir: string; commitHash: string } {
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(path.join(repo, 'sub', 'dir'), { recursive: true });
  sh(repo, 'git', ['init']);
  sh(repo, 'git', ['config', 'user.name', 'Test User']);
  sh(repo, 'git', ['config', 'user.email', 'test@example.com']);

  fs.writeFileSync(
    path.join(repo, '.gitignore'),
    'sub/dir/ignored.txt\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'sub', 'dir', 'tracked.txt'),
    'hello\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'sub', 'dir', 'modified.txt'),
    'v1\n',
    'utf8',
  );

  const env = {
    GIT_AUTHOR_DATE: '2024-01-02T03:04:05Z',
    GIT_COMMITTER_DATE: '2024-01-02T03:04:05Z',
  };
  sh(
    repo,
    'git',
    ['add', '.gitignore', 'sub/dir/tracked.txt', 'sub/dir/modified.txt'],
    env,
  );
  sh(repo, 'git', ['commit', '-m', 'init'], env);
  const commitHash = spawnSync('git', ['rev-parse', 'HEAD'], {
    cwd: repo,
    encoding: 'utf8',
  }).stdout.trim();

  fs.writeFileSync(
    path.join(repo, 'sub', 'dir', 'modified.txt'),
    'v2\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'sub', 'dir', 'untracked.txt'),
    'new\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(repo, 'sub', 'dir', 'ignored.txt'),
    'hidden\n',
    'utf8',
  );

  return { subDir: path.join(repo, 'sub', 'dir'), commitHash };
}

function writeInput(
  pathname: string,
  root: string,
  records: Array<{ locator: string }>,
  workers?: number,
): void {
  const meta: Record<string, unknown> = {
    contractVersion: '1',
    discovery: { root },
    git: { enabled: true },
    errors: { mode: 'keep-going', embedErrors: true },
  };
  if (workers !== undefined) meta.workers = workers;
  fs.mkdirSync(path.dirname(pathname), { recursive: true });
  fs.writeFileSync(
    pathname,
    JSON.stringify({
      records,
      meta,
    }),
    'utf8',
  );
}

function parseOut(stdout: string): EnrichOut {
  return JSON.parse(stdout) as EnrichOut;
}

function rec(out: EnrichOut, name: string) {
  const found = out.records.find((r) => r.locator === name);
  if (!found) throw new Error(`record not found: ${name}`);
  return found;
}

function normalizeWorkers(stdout: string): string {
  const out = JSON.parse(stdout) as EnrichOut;
  if (out.meta && typeof out.meta === 'object') {
    delete out.meta.workers;
  }
  return `${JSON.stringify(out)}\n`;
}

test('enrich-git resolves repo root above discovery.root and reports tracked/modified/untracked', () => {
  const root = projectRoot();
  const bin = buildBinary(root);

  const repo = path.join(root, 'temp', 'git-enrich-subdir-repo');
  const { subDir, commitHash } = setupGitFixture(repo);

  const inPath = path.join(root, 'temp', 'git-enrich-subdir.in.json');
  writeInput(inPath, subDir, [
    { locator: 'tracked.txt' },
    { locator: 'modified.txt' },
    { locator: 'untracked.txt' },
  ]);

  const run = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-git', '--in', inPath],
    root,
  );
  saveOutputs(root, 'enrich-git-subdir', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const out = parseOut(run.stdout);
  expect(rec(out, 'tracked.txt').git?.tracked).toBe(true);
  expect(rec(out, 'tracked.txt').git?.status).toBe('clean');
  expect(rec(out, 'tracked.txt').git?.lastCommit?.hash).toBe(commitHash);
  expect(rec(out, 'tracked.txt').git?.lastCommit?.author).toBe(
    'Test User <test@example.com>',
  );
  expect(rec(out, 'tracked.txt').git?.lastCommit?.time).toBe(
    '2024-01-02T03:04:05Z',
  );

  expect(rec(out, 'modified.txt').git?.tracked).toBe(true);
  expect(rec(out, 'modified.txt').git?.status).toBe('modified');
  expect(rec(out, 'modified.txt').git?.lastCommit?.author).toBe(
    'Test User <test@example.com>',
  );

  expect(rec(out, 'untracked.txt').git?.tracked).toBe(false);
  expect(rec(out, 'untracked.txt').git?.status).toBe('untracked');
  expect(rec(out, 'untracked.txt').git?.lastCommit).toBeNull();
});

test('enrich-git marks ignored files and discovery-input-files respects .gitignore', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'git-enrich-ignored-repo');
  const { subDir } = setupGitFixture(repo);

  const discover = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'discover-input-files',
      '--prepare',
      'input-files',
      '--root',
      repo,
    ],
    root,
  );
  expect(discover.status).toBe(0);
  expect(discover.stderr).toBe('');
  const discovered = JSON.parse(discover.stdout) as {
    records: Array<{ locator: string }>;
  };
  expect(
    discovered.records.some((r) => r.locator === 'sub/dir/ignored.txt'),
  ).toBe(false);

  const inPath = path.join(root, 'temp', 'git-enrich-ignored.in.json');
  writeInput(inPath, subDir, [{ locator: 'ignored.txt' }]);
  const run = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-git', '--in', inPath],
    root,
  );
  saveOutputs(root, 'enrich-git-ignored', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const out = parseOut(run.stdout);
  expect(rec(out, 'ignored.txt').git?.ignored).toBe(true);
});

test('enrich-git output is byte-identical across reruns and workers=1 vs workers=8', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'git-enrich-determinism-repo');
  const { subDir } = setupGitFixture(repo);

  const recs = [
    { locator: 'tracked.txt' },
    { locator: 'modified.txt' },
    { locator: 'untracked.txt' },
    { locator: 'ignored.txt' },
  ];
  const inPath1 = path.join(root, 'temp', 'git-enrich-det-1.in.json');
  const inPath8 = path.join(root, 'temp', 'git-enrich-det-8.in.json');
  writeInput(inPath1, subDir, recs, 1);
  writeInput(inPath8, subDir, recs, 8);

  const run1a = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-git', '--in', inPath1],
    root,
  );
  const run1b = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-git', '--in', inPath1],
    root,
  );
  expect(run1a.status).toBe(0);
  expect(run1b.status).toBe(0);
  expect(run1a.stderr).toBe('');
  expect(run1b.stderr).toBe('');
  expect(run1a.stdout).toBe(run1b.stdout);

  const run8 = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-git', '--in', inPath8],
    root,
  );
  expect(run8.status).toBe(0);
  expect(run8.stderr).toBe('');
  expect(normalizeWorkers(run1a.stdout)).toBe(normalizeWorkers(run8.stdout));
});
