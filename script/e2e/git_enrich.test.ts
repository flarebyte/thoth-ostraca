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

test('enrich-git attaches tracked/ignored/status/lastCommit deterministically', () => {
  const root = projectRoot();
  const bin = buildBinary(root);

  const repo = path.join(root, 'temp', 'git-enrich-repo');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });

  sh(repo, 'git', ['init']);
  sh(repo, 'git', ['config', 'user.name', 'Test User']);
  sh(repo, 'git', ['config', 'user.email', 'test@example.com']);

  fs.writeFileSync(path.join(repo, 'tracked.txt'), 'hello\n', 'utf8');
  fs.writeFileSync(path.join(repo, 'modified.txt'), 'v1\n', 'utf8');
  fs.writeFileSync(path.join(repo, '.gitignore'), 'ignored.txt\n', 'utf8');
  fs.writeFileSync(path.join(repo, 'ignored.txt'), 'zzz\n', 'utf8');

  const env = {
    GIT_AUTHOR_DATE: '2024-01-02T03:04:05Z',
    GIT_COMMITTER_DATE: '2024-01-02T03:04:05Z',
  };
  sh(repo, 'git', ['add', 'tracked.txt', 'modified.txt'], env);
  sh(repo, 'git', ['commit', '-m', 'init'], env);

  // modify modified.txt
  fs.writeFileSync(path.join(repo, 'modified.txt'), 'v2\n', 'utf8');
  // create untracked
  fs.writeFileSync(path.join(repo, 'untracked.txt'), 'new\n', 'utf8');

  const inPath = path.join(root, 'temp', 'git-enrich-in.json');
  const input = {
    records: [
      { locator: 'tracked.txt' },
      { locator: 'modified.txt' },
      { locator: 'untracked.txt' },
      { locator: 'ignored.txt' },
    ],
    meta: {
      contractVersion: '1',
      discovery: { root: repo },
      git: { enabled: true },
      errors: { mode: 'keep-going', embedErrors: true },
    },
  } as const;
  fs.mkdirSync(path.dirname(inPath), { recursive: true });
  fs.writeFileSync(inPath, JSON.stringify(input), 'utf8');

  const run = runThoth(
    bin,
    ['diagnose', '--stage', 'enrich-git', '--in', inPath],
    root,
  );
  saveOutputs(root, 'enrich-git', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const out = JSON.parse(run.stdout) as {
    records: Array<{
      locator: string;
      git?: {
        tracked: boolean;
        ignored: boolean;
        status: string;
        lastCommit: null | { author: string; time: string; hash: string };
      };
    }>;
  };
  const rec = (name: string) => {
    const found = out.records.find((r) => r.locator === name);
    if (!found) throw new Error(`record not found: ${name}`);
    return found;
  };
  expect(rec('tracked.txt').git?.tracked).toBe(true);
  expect(rec('tracked.txt').git?.status).toBe('clean');
  expect(rec('tracked.txt').git?.lastCommit?.author).toBe(
    'Test User <test@example.com>',
  );
  expect(rec('tracked.txt').git?.lastCommit?.time).toBe('2024-01-02T03:04:05Z');

  expect(rec('modified.txt').git?.tracked).toBe(true);
  expect(rec('modified.txt').git?.status).toBe('modified');
  expect(rec('modified.txt').git?.lastCommit?.author).toBe(
    'Test User <test@example.com>',
  );

  expect(rec('untracked.txt').git?.tracked).toBe(false);
  expect(rec('untracked.txt').git?.status).toBe('untracked');
  expect(rec('untracked.txt').git?.lastCommit).toBeNull();

  expect(rec('ignored.txt').git?.ignored).toBe(true);
});
