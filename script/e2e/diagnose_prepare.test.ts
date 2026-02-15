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

test('diagnose prepare meta-files then parse-validate-yaml matches golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repoRoot = path.join(root, 'testdata/repos/discovery1');

  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/prepare_parse_validate_yaml_out.golden.json',
  );

  const run = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'parse-validate-yaml',
      '--prepare',
      'meta-files',
      '--root',
      repoRoot,
    ],
    root,
  );
  saveOutputs(root, 'diagnose-prepare-parse-validate', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('diagnose prepare input-files then discover-input-files matches golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repoRoot = path.join(root, 'testdata/repos/create1');

  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/prepare_discover_input_files_out.golden.json',
  );

  const run = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'discover-input-files',
      '--prepare',
      'input-files',
      '--root',
      repoRoot,
    ],
    root,
  );
  saveOutputs(root, 'diagnose-prepare-discover-inputs', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('diagnose prepare meta-files writes dump-in and dump-out that match expected JSON', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repoRoot = path.join(root, 'testdata/repos/discovery1');

  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const dumpDir = fs.mkdtempSync(path.join(tmpRoot, 'prep-'));
  const dumpIn = path.join(dumpDir, 'in.json');
  const dumpOut = path.join(dumpDir, 'out.json');

  const run = runThoth(
    bin,
    [
      'diagnose',
      '--stage',
      'parse-validate-yaml',
      '--prepare',
      'meta-files',
      '--root',
      repoRoot,
      '--dump-in',
      dumpIn,
      '--dump-out',
      dumpOut,
    ],
    root,
  );
  saveOutputs(root, 'diagnose-prepare-dumps', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');

  // Dump files exist
  expect(fs.existsSync(dumpIn)).toBe(true);
  expect(fs.existsSync(dumpOut)).toBe(true);

  // Expected dump-in: prepared input after discover-meta-files
  const relRoot = path.relative(root, repoRoot).split(path.sep).join('/');
  const expectedDumpInObj = {
    records: [{ locator: 'a.thoth.yaml' }],
    meta: { discovery: { root: relRoot } },
  } as const;
  const expectedDumpIn = JSON.stringify(expectedDumpInObj);
  expect(fs.readFileSync(dumpIn, 'utf8')).toBe(expectedDumpIn);

  // Expected dump-out: parsed single record
  const expectedDumpOutObj = {
    records: [{ locator: 'd1', meta: { kind: 'discovery1' } }],
    meta: { contractVersion: '1', discovery: { root: relRoot } },
  } as const;
  const expectedDumpOut = JSON.stringify(expectedDumpOutObj);
  expect(fs.readFileSync(dumpOut, 'utf8')).toBe(expectedDumpOut);
});
