import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  buildBinary,
  expectedJSONFromGolden,
  projectRoot,
  runThoth,
  saveOutputs,
} from './helpers';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

type EnvelopeOut = Record<string, unknown> & { meta?: { workers?: unknown } };
function normalizeWorkersOut(stdout: string): string {
  const actual = JSON.parse(stdout) as EnvelopeOut;
  if (actual.meta && 'workers' in actual.meta) {
    delete actual.meta.workers;
  }
  return `${JSON.stringify(actual)}\n`;
}

function writeCueConfig(cfgPath: string, body: string): void {
  fs.mkdirSync(path.dirname(cfgPath), { recursive: true });
  fs.writeFileSync(cfgPath, body, 'utf8');
}

test('thoth run with valid config prints envelope JSON', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/minimal.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  // Save outputs for inspection; temp/ is git-ignored
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
    writeCueConfig(
      cfgKeep,
      `{
  configVersion: "v0"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
}`,
    );
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
    writeCueConfig(
      cfgFail,
      `{
  configVersion: "v0"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "fail-fast" }
}`,
    );
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
    writeCueConfig(
      cfgKeep,
      `{
  configVersion: "v0"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
}`,
    );
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
    writeCueConfig(
      cfgFail,
      `{
  configVersion: "v0"
  action: "pipeline"
  discovery: { root: "${repo.replaceAll('\\', '\\\\')}" }
  errors: { mode: "fail-fast" }
}`,
    );
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

test('thoth run filters records via Lua predicate', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/filter1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/filter1_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run fails on invalid Lua', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/filter_bad.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('lua-filter')).toBe(true);
});

test('thoth run maps records via Lua transform', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/map1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/map1_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run fails on invalid map Lua', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/map_bad.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('lua-map')).toBe(true);
});

test('thoth run executes shell and postmap', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/shell1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/shell1_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run fails when shell program is missing', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/shell_bad_prog.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('shell-exec')).toBe(true);
});

test('thoth run reduces to count and prints full envelope', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/reduce1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/reduce1_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('thoth run prints NDJSON lines when output.lines is true', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/lines1.cue');
  const expectedLines = fs.readFileSync(
    path.join(root, 'testdata/run/lines1_out.golden.ndjson'),
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedLines);
});

test('thoth run fails on invalid reduce Lua', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/reduce_bad.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('lua-reduce')).toBe(true);
});

test('validate-locators: default policy flags invalid parent refs and backslashes', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/locator_policy_default.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/locator_policy_default_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('validate-locators: relaxed policy allows parent refs and backslashes', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/locator_policy_relaxed.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/locator_policy_relaxed_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('validate-locators: URL locators are rejected by default policy', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_locator_urls_default.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_locator_urls_default_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('validate-locators: URL locators are accepted and normalized when enabled', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p3_locator_urls_allow.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p3_locator_urls_allow_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('keep-going with embedErrors=true embeds record errors and lists envelope errors', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/keep1_embed_true.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/keep1_embed_true_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('keep-going with embedErrors=false only lists envelope errors', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/keep1_embed_false.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/keep1_embed_false_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('determinism: workers=2 matches single-worker golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/workers2.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/keep1_embed_true_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const normalized = normalizeWorkersOut(run.stdout);
  expect(normalized).toBe(expectedOut);
});

test('determinism: workers=1 equals workers=2 golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/workers1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/keep1_embed_true_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const normalized = normalizeWorkersOut(run.stdout);
  expect(normalized).toBe(expectedOut);
});

test('thoth run with missing field fails with short error', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/missing_action.cue');
  const run = runThoth(bin, ['run', '-c', cfg], root);
  // Save outputs for inspection; temp/ is git-ignored
  saveOutputs(root, 'run-missing-action', run);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('action')).toBe(true);
});

test('write-output: pretty JSON to stdout matches pretty golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/output_stdout_pretty.cue');
  const expected = fs.readFileSync(
    path.join(root, 'testdata/run/output_stdout_pretty_out.golden.pretty.json'),
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-output-pretty', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expected);
});

test('write-output: file output writes to disk and stdout is empty', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/output_file.cue');
  const outPath = path.join(root, 'temp', 'out.json');
  try {
    fs.unlinkSync(outPath);
  } catch {}
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-output-file', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe('');
  expect(fs.existsSync(outPath)).toBe(true);
  const expected =
    '{"records":[],"meta":{"contractVersion":"1","config":{"configVersion":"v0","action":"nop"},"output":{"out":"temp/out.json"},"reduced":0}}\n';
  expect(fs.readFileSync(outPath, 'utf8')).toBe(expected);
});

test('validate-only: ok repo yields compact JSON with records', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/validate_only_ok.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/validate_only_ok_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('validate-only: mixed repo embeds record and envelope errors (keep-going)', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/validate_only_mixed.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/validate_only_mixed_out.golden.json',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('invalid action yields a short error mentioning allowed actions', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/action_unknown.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes("allowed 'pipeline' or 'validate'")).toBe(true);
});

test('create-meta: creates .thoth.yaml files and prints expected envelope', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const srcRepo = path.join(root, 'testdata/repos/create1');
  const tempRepo = path.join(root, 'temp', 'create1_repo');
  fs.rmSync(tempRepo, { recursive: true, force: true });
  fs.mkdirSync(path.dirname(tempRepo), { recursive: true });
  fs.cpSync(srcRepo, tempRepo, { recursive: true });
  const cfgPath = path.join(root, 'temp', 'create1_tmp.cue');
  const cfgContent = `{
  configVersion: "v0"
  action: "create-meta"
  discovery: { root: "${path.join('temp', 'create1_repo').replaceAll('\\', '\\\\')}" }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-create-meta', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const metaA = path.join(tempRepo, 'a.txt.thoth.yaml');
  const metaB = path.join(tempRepo, 'dir', 'b.txt.thoth.yaml');
  expect(fs.existsSync(metaA)).toBe(true);
  expect(fs.existsSync(metaB)).toBe(true);
  const metaIgnored = path.join(tempRepo, 'ignored.txt.thoth.yaml');
  const metaC = path.join(tempRepo, 'skipdir', 'c.txt.thoth.yaml');
  expect(fs.existsSync(metaIgnored)).toBe(false);
  expect(fs.existsSync(metaC)).toBe(false);
  const expectA = `locator: a.txt\nmeta: {}\n`;
  const expectB = `locator: dir/b.txt\nmeta: {}\n`;
  expect(fs.readFileSync(metaA, 'utf8')).toBe(expectA);
  expect(fs.readFileSync(metaB, 'utf8')).toBe(expectB);
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/create1_out.golden.json',
  );
  expect(run.stdout).toBe(expectedOut);
});

test('create-meta: second run fails-fast when meta exists', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfgPath = path.join(root, 'temp', 'create1_tmp.cue');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-create-meta-second', run);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('a.txt.thoth.yaml')).toBe(true);
});

test('update-meta: preserves existing meta and creates missing', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const srcRepo = path.join(root, 'testdata/repos/update1');
  const tempRepo = path.join(root, 'temp', 'update1_repo');
  fs.rmSync(tempRepo, { recursive: true, force: true });
  fs.mkdirSync(path.dirname(tempRepo), { recursive: true });
  fs.cpSync(srcRepo, tempRepo, { recursive: true });
  const cfgPath = path.join(root, 'temp', 'update1_tmp.cue');
  const cfgContent = `{
  configVersion: "v0"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update1_repo').replaceAll('\\', '\\\\')}" }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-update-meta', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  // a.txt meta preserved
  const metaA = path.join(tempRepo, 'a.txt.thoth.yaml');
  const metaB = path.join(tempRepo, 'b.txt.thoth.yaml');
  expect(fs.readFileSync(metaA, 'utf8')).toBe(
    'locator: a.txt\nmeta: { x: 1 }\n',
  );
  expect(fs.readFileSync(metaB, 'utf8')).toBe('locator: b.txt\nmeta: {}\n');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/update1_out.golden.json',
  );
  expect(run.stdout).toBe(expectedOut);
});

test('update-meta: invalid existing meta embeds errors in keep-going and still creates others', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const tempRepo = path.join(root, 'temp', 'update1_repo');
  // write invalid meta for a.txt
  fs.writeFileSync(
    path.join(tempRepo, 'a.txt.thoth.yaml'),
    'locator: a.txt\n# missing meta\n',
    'utf8',
  );
  const cfgPath = path.join(root, 'temp', 'update1_tmp_keep.cue');
  const cfgContent = `{
  configVersion: "v0"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update1_repo').replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-update-meta-keep', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  // b.txt still created
  expect(fs.readFileSync(path.join(tempRepo, 'b.txt.thoth.yaml'), 'utf8')).toBe(
    'locator: b.txt\nmeta: {}\n',
  );
  const out = JSON.parse(run.stdout) as {
    records: Array<{ locator: string; error?: unknown }>;
    errors: Array<unknown>;
  };
  const recA = out.records.find((r) => r.locator === 'a.txt');
  expect(recA.error).toBeTruthy();
  expect(out.errors.length).toBeGreaterThan(0);
});

test('update-meta: invalid existing meta fails fast', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfgPath = path.join(root, 'temp', 'update1_tmp_fail.cue');
  const cfgContent = `{
  configVersion: "v0"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update1_repo').replaceAll('\\', '\\\\')}" }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-update-meta-fail', run);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
});

test('diff-meta: computes orphans and counts deterministically', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/diff1.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-diff-meta', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const out = JSON.parse(run.stdout);
  expect(out.meta.diff.presentCount).toBe(2);
  expect(out.meta.diff.orphanCount).toBe(1);
  expect(JSON.stringify(out.meta.diff.orphans)).toBe(
    JSON.stringify(['orphan.txt.thoth.yaml']),
  );
});
