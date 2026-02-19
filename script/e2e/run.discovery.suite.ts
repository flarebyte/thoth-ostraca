import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import {
  buildBinary,
  expectedJSONFromGolden,
  projectRoot,
  runThoth,
  saveOutputs,
  writePipelineErrorModeConfig,
} from './helpers';
import { writeCueConfig } from './test_utils';

test('thoth run with valid config prints envelope JSON', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/minimal.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-minimal', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const parsed = JSON.parse(run.stdout);
  expect(typeof parsed).toBe('object');
  expect(Array.isArray(parsed.records)).toBe(true);
  expect(typeof parsed.meta.config.configVersion).toBe('string');
  expect(typeof parsed.meta.config.action).toBe('string');
});

test('thoth run executes discovery and respects gitignore by default', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/discovery1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/discovery1_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run pipeline discovery excludes ignored meta files by default', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_discovery_default.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_discovery_default_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run pipeline discovery includes ignored meta files when noGitignore=true', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_discovery_no_gitignore.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_discovery_no_gitignore_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run discovery followSymlinks handles symlink dirs and loops deterministically', () => {
  if (process.platform === 'win32') {
    return;
  }
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-symlink-'));
  fs.mkdirSync(path.join(repo, 'real'), { recursive: true });
  fs.writeFileSync(
    path.join(repo, 'real', 'a.thoth.yaml'),
    'locator: "loc-real"\nmeta:\n  ok: true\n',
    'utf8',
  );
  try {
    fs.symlinkSync('real', path.join(repo, 'alias'), 'dir');
    fs.symlinkSync('../real', path.join(repo, 'real', 'loop'), 'dir');
  } catch {
    return;
  }

  const cfgNoFollow = path.join(repo, 'no-follow.cue');
  writeCueConfig(
    cfgNoFollow,
    `{
  configVersion: "v0"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
}`,
  );
  const runNoFollow = runThoth(bin, ['run', '--config', cfgNoFollow], root);
  expect(runNoFollow.status).toBe(0);
  expect(runNoFollow.stderr).toBe('');
  const outNoFollow = JSON.parse(runNoFollow.stdout) as {
    records: Array<{ locator: string }>;
  };
  expect(outNoFollow.records.map((r) => r.locator)).toEqual(['loc-real']);

  const cfgFollow = path.join(repo, 'follow.cue');
  writeCueConfig(
    cfgFollow,
    `{
  configVersion: "v0"
  action: "pipeline"
  discovery: {
    root: "${repo.replaceAll('\\', '\\\\')}"
    followSymlinks: true
  }
  errors: { mode: "keep-going", embedErrors: true }
}`,
  );
  const runFollow = runThoth(bin, ['run', '--config', cfgFollow], root);
  expect(runFollow.status).toBe(0);
  expect(runFollow.stderr).toBe('');
  const outFollow = JSON.parse(runFollow.stdout) as {
    records: Array<{ locator: string }>;
  };
  const locs = outFollow.records.map((r) => r.locator);
  expect(locs).toEqual(['loc-real']);
  expect(new Set(locs).size).toBe(locs.length);
});

test('thoth run discovery/parse permission errors obey keep-going and fail-fast', () => {
  if (process.platform === 'win32') {
    return;
  }
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-perm-'));

  const goodDir = path.join(repo, 'good');
  const deniedDir = path.join(repo, 'denied');
  fs.mkdirSync(goodDir, { recursive: true });
  fs.mkdirSync(deniedDir, { recursive: true });
  fs.writeFileSync(
    path.join(goodDir, 'ok.thoth.yaml'),
    'locator: "ok"\nmeta:\n  n: 1\n',
    'utf8',
  );
  fs.writeFileSync(
    path.join(deniedDir, 'bad.thoth.yaml'),
    'locator: "bad"\nmeta:\n  n: 2\n',
    'utf8',
  );

  try {
    fs.chmodSync(deniedDir, 0o000);
    try {
      fs.readdirSync(deniedDir);
      fs.chmodSync(deniedDir, 0o755);
      return;
    } catch {
      // expected path is unreadable
    }

    const cfgKeep = path.join(repo, 'keep.cue');
    writePipelineErrorModeConfig(cfgKeep, repo, 'keep-going', true);
    const runKeep = runThoth(bin, ['run', '--config', cfgKeep], root);
    expect(runKeep.status).toBe(0);
    expect(runKeep.stderr).toBe('');
    const outKeep = JSON.parse(runKeep.stdout) as {
      records: Array<{ locator: string }>;
      errors?: Array<{ stage: string; locator: string }>;
    };
    expect(outKeep.records.map((r) => r.locator)).toContain('ok');
    expect(
      (outKeep.errors ?? []).some(
        (e) =>
          e.stage === 'discover-meta-files' && e.locator.includes('denied'),
      ),
    ).toBe(true);

    const cfgFail = path.join(repo, 'fail.cue');
    writePipelineErrorModeConfig(cfgFail, repo, 'fail-fast');
    const runFail = runThoth(bin, ['run', '--config', cfgFail], root);
    expect(runFail.status).not.toBe(0);
    expect(runFail.stdout).toBe('');
    expect(runFail.stderr.includes('denied')).toBe(true);
  } finally {
    try {
      fs.chmodSync(deniedDir, 0o755);
    } catch {}
  }
});
