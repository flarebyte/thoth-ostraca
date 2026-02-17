import { expect, test } from 'bun:test';
import * as path from 'node:path';
import {
  buildBinary,
  expectedJSONFromGolden,
  projectRoot,
  runThoth,
} from './helpers';

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
