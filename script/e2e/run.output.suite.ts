import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth, saveOutputs } from './helpers';

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
    '{"records":[],"meta":{"contractVersion":"1","config":{"configVersion":"v0","action":"nop"},"limits":{"maxRecordsInMemory":10000},"output":{"out":"temp/out.json"},"reduced":0}}\n';
  expect(fs.readFileSync(outPath, 'utf8')).toBe(expected);
});
