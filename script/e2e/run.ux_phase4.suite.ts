import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import {
  buildBinary,
  expectedJSONFromGolden,
  projectRoot,
  runThoth,
  saveOutputs,
} from './helpers';

test('phase4 ux progress writes to stderr and keeps stdout deterministic', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'temp', 'ux-progress.cue');
  fs.mkdirSync(path.dirname(cfg), { recursive: true });
  fs.writeFileSync(
    cfg,
    `{
  configVersion: "1"
  action: "nop"
  discovery: { root: "testdata/repos/yaml1" }
  ui: { progress: true, progressIntervalMs: 1 }
}
`,
    'utf8',
  );
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/ux_progress_out.golden.json',
  );

  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-ux-progress', run);
  expect(run.status).toBe(0);
  expect(run.stdout).toBe(expectedOut);
  expect(
    run.stderr.split('\n').some((line) => line.startsWith('progress stage=')),
  ).toBe(true);
});

test('phase4 ux validate-config rejects workers<1', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'temp', 'ux-invalid-workers.cue');
  fs.mkdirSync(path.dirname(cfg), { recursive: true });
  fs.writeFileSync(
    cfg,
    `{
  configVersion: "1"
  action: "nop"
  workers: 0
}
`,
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-ux-invalid-workers', run);
  expect(run.status).not.toBe(0);
  expect(run.stderr.trim()).toBe('invalid workers: must be >= 1');
});

test('phase4 ux validate-config rejects maxRecordsInMemory<1', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'temp', 'ux-invalid-limits.cue');
  fs.mkdirSync(path.dirname(cfg), { recursive: true });
  fs.writeFileSync(
    cfg,
    `{
  configVersion: "1"
  action: "nop"
  limits: { maxRecordsInMemory: 0 }
}
`,
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-ux-invalid-limits', run);
  expect(run.status).not.toBe(0);
  expect(run.stderr.trim()).toBe(
    'invalid limits.maxRecordsInMemory: must be >= 1',
  );
});

test('phase4 ux validate-config rejects shell enabled without argsTemplate', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'temp', 'ux-invalid-shell-args.cue');
  fs.mkdirSync(path.dirname(cfg), { recursive: true });
  fs.writeFileSync(
    cfg,
    `{
  configVersion: "1"
  action: "nop"
  shell: { enabled: true }
}
`,
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-ux-invalid-shell-args', run);
  expect(run.status).not.toBe(0);
  expect(run.stderr.trim()).toBe(
    'invalid shell.argsTemplate: required when shell.enabled=true',
  );
});

test('phase4 ux output.out empty string is treated as stdout', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'temp', 'ux-output-empty-out.cue');
  fs.mkdirSync(path.dirname(cfg), { recursive: true });
  fs.writeFileSync(
    cfg,
    `{
  configVersion: "1"
  action: "nop"
  discovery: { root: "testdata/repos/yaml1" }
  output: { out: "" }
}
`,
    'utf8',
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-ux-output-empty-out', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout.startsWith('{"records":')).toBe(true);
});
