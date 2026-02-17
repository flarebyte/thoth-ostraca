import { expect, test } from 'bun:test';
import * as path from 'node:path';
import {
  buildBinary,
  expectedJSONFromGolden,
  projectRoot,
  runThoth,
  saveOutputs,
} from './helpers';
import { normalizeWorkersOut } from './test_utils';

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

test('thoth run fails on invalid reduce Lua', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/reduce_bad.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('lua-reduce')).toBe(true);
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
  saveOutputs(root, 'run-missing-action', run);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('action')).toBe(true);
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
