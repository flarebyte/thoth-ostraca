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

test('diagnose prepared pipeline runs through until-stage and matches golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/yaml1.cue');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/diagnose/prepare_pipeline_until_parse_validate_yaml_out.golden.json',
  );

  const run = runThoth(
    bin,
    [
      'diagnose',
      '--prepare-pipeline',
      'pipeline',
      '--until-stage',
      'parse-validate-yaml',
      '--config',
      cfg,
    ],
    root,
  );
  saveOutputs(root, 'diagnose-prepare-pipeline-until-stage', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(run.stdout).toBe(expectedOut);
});

test('diagnose prepared pipeline stage-index matches stage name output', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/yaml1.cue');

  const byName = runThoth(
    bin,
    [
      'diagnose',
      '--prepare-pipeline',
      'pipeline',
      '--stage',
      'discover-meta-files',
      '--config',
      cfg,
    ],
    root,
  );
  const byIndex = runThoth(
    bin,
    [
      'diagnose',
      '--prepare-pipeline',
      'pipeline',
      '--stage',
      'parse-validate-yaml',
      '--stage-index',
      '0',
      '--config',
      cfg,
    ],
    root,
  );
  saveOutputs(root, 'diagnose-prepare-pipeline-stage-name', byName);
  saveOutputs(root, 'diagnose-prepare-pipeline-stage-index', byIndex);
  expect(byName.status).toBe(0);
  expect(byIndex.status).toBe(0);
  expect(byName.stderr).toBe('');
  expect(byIndex.stderr).toBe('');
  expect(byName.stdout).toBe(byIndex.stdout);
});

test('diagnose dump-dir writes stable stage boundary fixtures for prepared pipeline', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/yaml1.cue');

  const tmpRoot = path.join(root, 'temp');
  fs.mkdirSync(tmpRoot, { recursive: true });
  const dumpDir = fs.mkdtempSync(path.join(tmpRoot, 'diag-dump-dir-'));

  const run = runThoth(
    bin,
    [
      'diagnose',
      '--prepare-pipeline',
      'pipeline',
      '--until-index',
      '1',
      '--config',
      cfg,
      '--dump-dir',
      dumpDir,
    ],
    root,
  );
  saveOutputs(root, 'diagnose-prepare-pipeline-dump-dir', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');

  const actualFiles = fs.readdirSync(dumpDir).sort();
  const expectedFiles = [
    '001_discover-meta-files_in.json',
    '001_discover-meta-files_out.json',
    '002_parse-validate-yaml_in.json',
    '002_parse-validate-yaml_out.json',
  ];
  expect(actualFiles).toEqual(expectedFiles);

  for (const name of expectedFiles) {
    const actual = JSON.stringify(
      JSON.parse(fs.readFileSync(path.join(dumpDir, name), 'utf8')),
    );
    const expected = JSON.stringify(
      JSON.parse(
        fs.readFileSync(
          path.join(root, 'testdata/diagnose/dump_dir_pipeline', name),
          'utf8',
        ),
      ),
    );
    expect(actual).toBe(expected);
  }
});
