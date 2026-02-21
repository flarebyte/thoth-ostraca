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
import { normalizeWorkersOut, writeCueConfig } from './test_utils';

test('thoth run parse-validate-yaml unreadable file obeys keep-going and fail-fast', () => {
  if (process.platform === 'win32') {
    return;
  }
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-parse-perm-'));

  fs.writeFileSync(
    path.join(repo, 'good.thoth.yaml'),
    'locator: "good"\nmeta:\n  n: 1\n',
    'utf8',
  );
  const unreadable = path.join(repo, 'blocked.thoth.yaml');
  fs.writeFileSync(unreadable, 'locator: "blocked"\nmeta:\n  n: 2\n', 'utf8');

  try {
    fs.chmodSync(unreadable, 0o000);
    try {
      fs.readFileSync(unreadable, 'utf8');
      fs.chmodSync(unreadable, 0o644);
      return;
    } catch {
      // expected unreadable file
    }

    const cfgKeep = path.join(repo, 'keep.cue');
    writePipelineErrorModeConfig(cfgKeep, repo, 'keep-going', true);
    const runKeep = runThoth(bin, ['run', '--config', cfgKeep], root);
    expect(runKeep.status).toBe(0);
    expect(runKeep.stderr).toBe('');
    const outKeep = JSON.parse(runKeep.stdout) as {
      records: Array<{ locator: string; error?: { stage: string } }>;
      errors?: Array<{ stage: string; locator: string }>;
    };
    expect(outKeep.records.map((r) => r.locator)).toContain('good');
    expect(
      outKeep.records.some(
        (r) =>
          r.locator === 'blocked.thoth.yaml' &&
          r.error &&
          r.error.stage === 'parse-validate-yaml',
      ),
    ).toBe(true);
    expect(
      (outKeep.errors ?? []).some(
        (e) =>
          e.stage === 'parse-validate-yaml' &&
          e.locator === 'blocked.thoth.yaml',
      ),
    ).toBe(true);

    const cfgFail = path.join(repo, 'fail.cue');
    writePipelineErrorModeConfig(cfgFail, repo, 'fail-fast');
    const runFail = runThoth(bin, ['run', '--config', cfgFail], root);
    expect(runFail.status).not.toBe(0);
    expect(runFail.stdout).toBe('');
    expect(runFail.stderr.includes('blocked.thoth.yaml')).toBe(true);
  } finally {
    try {
      fs.chmodSync(unreadable, 0o644);
    } catch {}
  }
});

test('thoth run pipeline parse-validate-yaml fail-fast exits non-zero and includes invalid file path', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_yaml_failfast.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-p3-yaml-failfast', run);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(
    run.stderr.includes('testdata/repos/p3_yaml1/invalid/c.thoth.yaml'),
  ).toBe(true);
});

test('thoth run pipeline parse-validate-yaml keep-going embeds record errors and envelope errors', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_yaml_keepgoing.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_yaml_keepgoing_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-p3-yaml-keepgoing', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run strict top-level default rejects unknown keys', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_yaml_strict_default.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_yaml_strict_default_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-p3-yaml-strict-default', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run strict top-level allowUnknownTopLevel accepts unknown keys and drops them', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(
    root,
    'testdata/configs/p3_yaml_strict_allow_unknown.cue',
  );
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_yaml_strict_allow_unknown_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-p3-yaml-strict-allow-unknown', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run parses YAML records into {locator,meta}', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/yaml1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/yaml1_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run fails on invalid YAML (missing meta)', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/yaml1_nogit.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('missing required field: meta')).toBe(true);
});

test('parse-validate-yaml determinism: many files workers=1 equals workers=8', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-yaml-many-'));
  for (let i = 0; i < 50; i += 1) {
    const name = `f${String(i).padStart(2, '0')}.thoth.yaml`;
    const locator = `many/${String(i).padStart(2, '0')}`;
    fs.writeFileSync(
      path.join(repo, name),
      `locator: "${locator}"\nmeta:\n  n: ${i}\n`,
      'utf8',
    );
  }

  const cfg1 = path.join(repo, 'workers1.cue');
  writeCueConfig(
    cfg1,
    `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  workers: 1
  errors: { mode: "keep-going", embedErrors: true }
}`,
  );
  const run1 = runThoth(bin, ['run', '--config', cfg1], root);
  expect(run1.status).toBe(0);
  expect(run1.stderr).toBe('');

  const cfg8 = path.join(repo, 'workers8.cue');
  writeCueConfig(
    cfg8,
    `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  workers: 8
  errors: { mode: "keep-going", embedErrors: true }
}`,
  );
  const run8 = runThoth(bin, ['run', '--config', cfg8], root);
  expect(run8.status).toBe(0);
  expect(run8.stderr).toBe('');

  expect(normalizeWorkersOut(run1.stdout)).toBe(
    normalizeWorkersOut(run8.stdout),
  );
});

test('parse-validate-yaml maxYAMLBytes keep-going collects size error and fail-fast exits', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const repo = fs.mkdtempSync(path.join(tmpRoot, 'p3-yaml-size-'));

  fs.writeFileSync(
    path.join(repo, 'ok.thoth.yaml'),
    'locator: "ok"\nmeta:\n  k: 1\n',
    'utf8',
  );
  const hugeMeta = 'x'.repeat(300);
  fs.writeFileSync(
    path.join(repo, 'big.thoth.yaml'),
    `locator: "big"\nmeta:\n  payload: "${hugeMeta}"\n`,
    'utf8',
  );

  const keepCfg = path.join(repo, 'keep.cue');
  writeCueConfig(
    keepCfg,
    `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
  limits: { maxYAMLBytes: 128 }
}`,
  );
  const runKeep = runThoth(bin, ['run', '--config', keepCfg], root);
  expect(runKeep.status).toBe(0);
  expect(runKeep.stderr).toBe('');
  const outKeep = JSON.parse(runKeep.stdout) as {
    records: Array<{
      locator: string;
      error?: { stage: string; message: string };
    }>;
    errors?: Array<{ stage: string; locator: string; message: string }>;
  };
  expect(outKeep.records.some((r) => r.locator === 'ok')).toBe(true);
  expect(
    outKeep.records.some(
      (r) =>
        r.locator === 'big.thoth.yaml' &&
        r.error?.stage === 'parse-validate-yaml' &&
        r.error.message.includes('maxYAMLBytes'),
    ),
  ).toBe(true);
  expect(
    (outKeep.errors ?? []).some(
      (e) =>
        e.stage === 'parse-validate-yaml' &&
        e.locator === 'big.thoth.yaml' &&
        e.message.includes('maxYAMLBytes'),
    ),
  ).toBe(true);

  const failCfg = path.join(repo, 'fail.cue');
  writeCueConfig(
    failCfg,
    `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "fail-fast" }
  limits: { maxYAMLBytes: 128 }
}`,
  );
  const runFail = runThoth(bin, ['run', '--config', failCfg], root);
  expect(runFail.status).not.toBe(0);
  expect(runFail.stdout).toBe('');
  expect(runFail.stderr.includes('big.thoth.yaml')).toBe(true);
  expect(runFail.stderr.includes('maxYAMLBytes')).toBe(true);
});
