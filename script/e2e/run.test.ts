import { expect, test } from 'bun:test';
import { spawnSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import {
  buildBinary,
  expectedJSONFromGolden,
  projectRoot,
  runThoth,
  saveOutputs,
} from './helpers';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

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
  const actual = JSON.parse(run.stdout);
  if (actual.meta) delete actual.meta.workers;
  const normalized = JSON.stringify(actual) + '\n';
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
  const actual = JSON.parse(run.stdout);
  if (actual.meta) delete actual.meta.workers;
  const normalized = JSON.stringify(actual) + '\n';
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
