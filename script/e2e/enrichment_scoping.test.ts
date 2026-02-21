import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { writeCfg } from './config_helpers';
import { buildBinary, projectRoot, runThoth } from './helpers';
import { sh } from './shell_helpers';

type RunOut = {
  records: Array<Record<string, unknown>>;
  meta?: Record<string, unknown>;
};

function copyTree(src: string, dest: string): string {
  fs.rmSync(dest, { recursive: true, force: true });
  fs.cpSync(src, dest, { recursive: true });
  return dest;
}

function parseOut(stdout: string): RunOut {
  return JSON.parse(stdout) as RunOut;
}

function hasKeyDeep(v: unknown, key: string): boolean {
  if (!v || typeof v !== 'object') return false;
  if (Array.isArray(v)) return v.some((x) => hasKeyDeep(x, key));
  for (const [k, x] of Object.entries(v as Record<string, unknown>)) {
    if (k === key) return true;
    if (hasKeyDeep(x, key)) return true;
  }
  return false;
}

function assertNoEnrichmentKeys(out: RunOut) {
  expect(hasKeyDeep(out.records, 'fileInfo')).toBe(false);
  expect(hasKeyDeep(out.records, 'git')).toBe(false);
}

function normalizeWorkers(stdout: string): string {
  const out = JSON.parse(stdout) as RunOut;
  if (out.meta && typeof out.meta === 'object') {
    delete out.meta.workers;
  }
  return `${JSON.stringify(out)}\n`;
}

function setupGitRepoFixture(repo: string): string {
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(path.join(repo, 'dir'), { recursive: true });
  sh(repo, 'git', ['init']);
  sh(repo, 'git', ['config', 'user.name', 'Test User']);
  sh(repo, 'git', ['config', 'user.email', 'test@example.com']);
  fs.writeFileSync(path.join(repo, 'a.txt'), 'a\n', 'utf8');
  fs.writeFileSync(path.join(repo, 'dir', 'b.txt'), 'b\n', 'utf8');
  const env = {
    GIT_AUTHOR_DATE: '2024-01-02T03:04:05Z',
    GIT_COMMITTER_DATE: '2024-01-02T03:04:05Z',
  };
  sh(repo, 'git', ['add', '.'], env);
  sh(repo, 'git', ['commit', '-m', 'init'], env);
  return repo;
}

test('no-enrichment contract across pipeline/validate/create-meta/update-meta/diff-meta', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const createSrc = path.join(root, 'testdata', 'repos', 'create1');
  const updateSrc = path.join(root, 'testdata', 'repos', 'update1');
  const diffSrc = path.join(root, 'testdata', 'repos', 'diff1');

  const cases: Array<{ name: string; cfg: Record<string, unknown> }> = [
    {
      name: 'scoping-pipeline',
      cfg: {
        configVersion: '1',
        action: 'pipeline',
        discovery: { root: 'testdata/repos/yaml1' },
        fileInfo: { enabled: false },
        git: { enabled: false },
      },
    },
    {
      name: 'scoping-validate',
      cfg: {
        configVersion: '1',
        action: 'validate',
        discovery: { root: 'testdata/repos/yaml1' },
        fileInfo: { enabled: false },
        git: { enabled: false },
      },
    },
    {
      name: 'scoping-create',
      cfg: {
        configVersion: '1',
        action: 'create-meta',
        discovery: {
          root: copyTree(
            createSrc,
            path.join(root, 'temp', 'scoping-create-copy'),
          ),
        },
        fileInfo: { enabled: false },
        git: { enabled: false },
      },
    },
    {
      name: 'scoping-update',
      cfg: {
        configVersion: '1',
        action: 'update-meta',
        discovery: {
          root: copyTree(
            updateSrc,
            path.join(root, 'temp', 'scoping-update-copy'),
          ),
        },
        fileInfo: { enabled: false },
        git: { enabled: false },
      },
    },
    {
      name: 'scoping-diff',
      cfg: {
        configVersion: '1',
        action: 'diff-meta',
        discovery: {
          root: copyTree(diffSrc, path.join(root, 'temp', 'scoping-diff-copy')),
        },
        fileInfo: { enabled: false },
        git: { enabled: false },
      },
    },
  ];

  for (const c of cases) {
    const cfgPath = writeCfg(root, c.name, c.cfg);
    const run = runThoth(bin, ['run', '--config', cfgPath], root);
    expect(run.status).toBe(0);
    expect(run.stderr).toBe('');
    assertNoEnrichmentKeys(parseOut(run.stdout));
  }
});

test('positive scoping: create-meta and update-meta include fileInfo/git only when enabled', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const createFileInfoRepo = setupGitRepoFixture(
    path.join(root, 'temp', 'scoping-create-fileinfo-repo'),
  );

  const createFileInfoCfg = writeCfg(root, 'scoping-create-fileinfo', {
    configVersion: '1',
    action: 'create-meta',
    discovery: { root: createFileInfoRepo },
    fileInfo: { enabled: true },
    git: { enabled: false },
  });
  const createFileInfoRun = runThoth(
    bin,
    ['run', '--config', createFileInfoCfg],
    root,
  );
  expect(createFileInfoRun.status).toBe(0);
  const createFileInfoOut = parseOut(createFileInfoRun.stdout);
  expect(
    createFileInfoOut.records.some((r) => Object.hasOwn(r, 'fileInfo')),
  ).toBe(true);
  expect(createFileInfoOut.records.some((r) => Object.hasOwn(r, 'git'))).toBe(
    false,
  );

  const createGitRepo = setupGitRepoFixture(
    path.join(root, 'temp', 'scoping-create-git-repo'),
  );
  const createGitCfg = writeCfg(root, 'scoping-create-git', {
    configVersion: '1',
    action: 'create-meta',
    discovery: { root: createGitRepo },
    fileInfo: { enabled: false },
    git: { enabled: true },
  });
  const createGitRun = runThoth(bin, ['run', '--config', createGitCfg], root);
  expect(createGitRun.status).toBe(0);
  const createGitOut = parseOut(createGitRun.stdout);
  expect(createGitOut.records.some((r) => Object.hasOwn(r, 'git'))).toBe(true);
  expect(createGitOut.records.some((r) => Object.hasOwn(r, 'fileInfo'))).toBe(
    false,
  );

  const updateFileInfoRepo = setupGitRepoFixture(
    path.join(root, 'temp', 'scoping-update-fileinfo-repo'),
  );
  const updateFileInfoCfg = writeCfg(root, 'scoping-update-fileinfo', {
    configVersion: '1',
    action: 'update-meta',
    discovery: { root: updateFileInfoRepo },
    fileInfo: { enabled: true },
    git: { enabled: false },
  });
  const updateFileInfoRun = runThoth(
    bin,
    ['run', '--config', updateFileInfoCfg],
    root,
  );
  expect(updateFileInfoRun.status).toBe(0);
  const updateFileInfoOut = parseOut(updateFileInfoRun.stdout);
  expect(
    updateFileInfoOut.records.some((r) => Object.hasOwn(r, 'fileInfo')),
  ).toBe(true);
  expect(updateFileInfoOut.records.some((r) => Object.hasOwn(r, 'git'))).toBe(
    false,
  );

  const updateGitRepo = setupGitRepoFixture(
    path.join(root, 'temp', 'scoping-update-git-repo'),
  );
  const updateGitCfg = writeCfg(root, 'scoping-update-git', {
    configVersion: '1',
    action: 'update-meta',
    discovery: { root: updateGitRepo },
    fileInfo: { enabled: false },
    git: { enabled: true },
  });
  const updateGitRun = runThoth(bin, ['run', '--config', updateGitCfg], root);
  expect(updateGitRun.status).toBe(0);
  const updateGitOut = parseOut(updateGitRun.stdout);
  expect(updateGitOut.records.some((r) => Object.hasOwn(r, 'git'))).toBe(true);
  expect(updateGitOut.records.some((r) => Object.hasOwn(r, 'fileInfo'))).toBe(
    false,
  );
});

test('determinism with fileInfo+git enabled: workers=1 vs workers=8', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const srcRepo = setupGitRepoFixture(
    path.join(root, 'temp', 'scoping-det-git-repo-src'),
  );
  const workRepo = path.join(root, 'temp', 'scoping-det-git-repo-work');

  const cfg1 = writeCfg(root, 'scoping-det-workers-1', {
    configVersion: '1',
    action: 'update-meta',
    discovery: { root: workRepo },
    fileInfo: { enabled: true },
    git: { enabled: true },
    workers: 1,
  });
  const cfg8 = writeCfg(root, 'scoping-det-workers-8', {
    configVersion: '1',
    action: 'update-meta',
    discovery: { root: workRepo },
    fileInfo: { enabled: true },
    git: { enabled: true },
    workers: 8,
  });

  copyTree(srcRepo, workRepo);
  const run1 = runThoth(bin, ['run', '--config', cfg1], root);
  copyTree(srcRepo, workRepo);
  const run8 = runThoth(bin, ['run', '--config', cfg8], root);
  expect(run1.status).toBe(0);
  expect(run8.status).toBe(0);
  expect(run1.stderr).toBe('');
  expect(run8.stderr).toBe('');
  expect(normalizeWorkers(run1.stdout)).toBe(normalizeWorkers(run8.stdout));
});
