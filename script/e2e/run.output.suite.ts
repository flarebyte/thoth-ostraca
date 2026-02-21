import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth, saveOutputs } from './helpers';
import { writeCueConfig } from './test_utils';

function expectJSONEqual(actual: string, expected: string): void {
  expect(JSON.parse(actual)).toEqual(JSON.parse(expected));
}

test('write-output: compact envelope to stdout matches golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/output_stdout_compact.cue');
  const expected = fs.readFileSync(
    path.join(root, 'testdata/run/output_stdout_compact_out.golden.json'),
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expectJSONEqual(run.stdout, expected);
});

test('write-output: pretty envelope to stdout matches pretty golden', () => {
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

test('write-output: NDJSON lines to stdout matches golden exactly', () => {
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

test('write-output: compact envelope to file writes expected golden and keeps stdout empty', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/output_file.cue');
  const outPath = path.join(root, 'temp', 'out.json');
  const expected = fs.readFileSync(
    path.join(root, 'testdata/run/output_file_compact_out.golden.json'),
    'utf8',
  );
  try {
    fs.unlinkSync(outPath);
  } catch {}
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-output-file', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe('');
  expect(fs.existsSync(outPath)).toBe(true);
  expectJSONEqual(fs.readFileSync(outPath, 'utf8'), expected);
});

test('write-output: pretty envelope to file writes expected golden and keeps stdout empty', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/output_file_pretty.cue');
  const outPath = path.join(root, 'temp', 'out.pretty.json');
  const expected = fs.readFileSync(
    path.join(root, 'testdata/run/output_file_pretty_out.golden.pretty.json'),
    'utf8',
  );
  try {
    fs.unlinkSync(outPath);
  } catch {}
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe('');
  expect(fs.existsSync(outPath)).toBe(true);
  expect(fs.readFileSync(outPath, 'utf8')).toBe(expected);
});

test('write-output: NDJSON lines to file writes expected golden and keeps stdout empty', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/output_file_lines.cue');
  const outPath = path.join(root, 'temp', 'out.lines.ndjson');
  const expected = fs.readFileSync(
    path.join(root, 'testdata/run/output_file_lines_out.golden.ndjson'),
    'utf8',
  );
  try {
    fs.unlinkSync(outPath);
  } catch {}
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe('');
  expect(fs.existsSync(outPath)).toBe(true);
  expect(fs.readFileSync(outPath, 'utf8')).toBe(expected);
});

test('write-output determinism: compact/pretty/lines stdout are byte-identical across reruns', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfgs = [
    path.join(root, 'testdata/configs/output_stdout_compact.cue'),
    path.join(root, 'testdata/configs/output_stdout_pretty.cue'),
    path.join(root, 'testdata/configs/lines1.cue'),
  ];
  for (const cfg of cfgs) {
    const r1 = runThoth(bin, ['run', '--config', cfg], root);
    const r2 = runThoth(bin, ['run', '--config', cfg], root);
    expect(r1.status).toBe(0);
    expect(r2.status).toBe(0);
    expect(r1.stderr).toBe('');
    expect(r2.stderr).toBe('');
    expect(r1.stdout).toBe(r2.stdout);
  }
});

test('write-output determinism: file modes are byte-identical across reruns', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cases = [
    {
      cfg: path.join(root, 'testdata/configs/output_file.cue'),
      outPath: path.join(root, 'temp', 'out.json'),
    },
    {
      cfg: path.join(root, 'testdata/configs/output_file_pretty.cue'),
      outPath: path.join(root, 'temp', 'out.pretty.json'),
    },
    {
      cfg: path.join(root, 'testdata/configs/output_file_lines.cue'),
      outPath: path.join(root, 'temp', 'out.lines.ndjson'),
    },
  ];
  for (const c of cases) {
    try {
      fs.unlinkSync(c.outPath);
    } catch {}
    const r1 = runThoth(bin, ['run', '--config', c.cfg], root);
    const first = fs.readFileSync(c.outPath, 'utf8');
    const r2 = runThoth(bin, ['run', '--config', c.cfg], root);
    const second = fs.readFileSync(c.outPath, 'utf8');
    expect(r1.status).toBe(0);
    expect(r2.status).toBe(0);
    expect(r1.stderr).toBe('');
    expect(r2.stderr).toBe('');
    expect(r1.stdout).toBe('');
    expect(r2.stdout).toBe('');
    expect(first).toBe(second);
  }
});

test('write-output determinism: lines stdout workers=1 and workers=8 are byte-identical', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg1 = path.join(root, 'temp', 'output-lines-workers1.cue');
  const cfg8 = path.join(root, 'temp', 'output-lines-workers8.cue');
  const body = (workers: number): string => `{
  configVersion: "1"
  action: "nop"
  discovery: {
    root: "testdata/repos/yaml1"
  }
  filter: {
    inline: "return (meta and meta.enabled) == true"
  }
  map: {
    inline: "return { locator = locator, name = meta and meta.name }"
  }
  workers: ${workers}
  output: {
    lines: true
  }
}`;
  writeCueConfig(cfg1, body(1));
  writeCueConfig(cfg8, body(8));
  const r1 = runThoth(bin, ['run', '--config', cfg1], root);
  const r8 = runThoth(bin, ['run', '--config', cfg8], root);
  expect(r1.status).toBe(0);
  expect(r8.status).toBe(0);
  expect(r1.stderr).toBe('');
  expect(r8.stderr).toBe('');
  expect(r1.stdout).toBe(r8.stdout);
});
