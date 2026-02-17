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

test('diagnose echo prints expected JSON and writes dumps', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const input = path.join(root, 'testdata/diagnose/input.json');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/out.golden.json',
  );

  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const dumpDir = fs.mkdtempSync(path.join(tmpRoot, 'diag-'));
  const dumpIn = path.join(dumpDir, 'in.json');
  const dumpOut = path.join(dumpDir, 'out.json');

  const run = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'echo',
      '--in',
      input,
      '--dump-in',
      dumpIn,
      '--dump-out',
      dumpOut,
    ],
    root,
  );
  // Save outputs for inspection; temp/ is git-ignored
  saveOutputs(root, 'diagnose-echo', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);

  // Dump files exist and contents match expected JSON (exact)
  expect(fs.existsSync(dumpIn)).toBe(true);
  expect(fs.existsSync(dumpOut)).toBe(true);

  const expectedDumpIn = JSON.stringify(
    JSON.parse(fs.readFileSync(input, 'utf8')),
  );
  const expectedDumpOut = JSON.stringify(JSON.parse(expectedOut));
  expect(fs.readFileSync(dumpIn, 'utf8')).toBe(expectedDumpIn);
  expect(fs.readFileSync(dumpOut, 'utf8')).toBe(expectedDumpOut);
});

test('diagnose validate-config produces expected envelope when --config is provided', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/minimal.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/validate_config_out.golden.json',
  );

  const run = runThoth(
    bin,
    ['diagnose', '--stage', 'validate-config', '--config', cfg],
    root,
  );
  // Save outputs for inspection; temp/ is git-ignored
  saveOutputs(root, 'diagnose-validate-config', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('diagnose discover-meta-files respects gitignore by default and can be disabled', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repoRoot = path.join(root, 'testdata/repos/discovery1');

  // With no-gitignore: both files present
  const runAll = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'discover-meta-files',
      '--root',
      repoRoot,
      '--no-gitignore',
    ],
    root,
  );
  // Save outputs for inspection; temp/ is git-ignored
  saveOutputs(root, 'diagnose-discover-no-gitignore', runAll);
  expect(runAll.status).toBe(0);
  expect(runAll.stderr).toBe('');
  const parsed = JSON.parse(runAll.stdout);
  const recs = parsed.records as Array<{ locator: string }>;
  const locs = recs.map((r) => r.locator);
  // Exact order expected
  expect(JSON.stringify(locs)).toBe(
    JSON.stringify(['a.thoth.yaml', 'ignored/x.thoth.yaml']),
  );
});

test('diagnose discover-meta-files default excludes ignored files and keeps sorted locators', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/p3_discovery_default_discover_meta_files_out.golden.json',
  );
  const run = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'discover-meta-files',
      '--root',
      path.join(root, 'testdata/repos/p3_discovery1'),
    ],
    root,
  );
  saveOutputs(root, 'diagnose-p3-discover-default', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('diagnose discover-meta-files no-gitignore includes ignored files and keeps sorted locators', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/p3_discovery_no_gitignore_discover_meta_files_out.golden.json',
  );
  const run = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'discover-meta-files',
      '--root',
      path.join(root, 'testdata/repos/p3_discovery1'),
      '--no-gitignore',
    ],
    root,
  );
  saveOutputs(root, 'diagnose-p3-discover-no-gitignore', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});
