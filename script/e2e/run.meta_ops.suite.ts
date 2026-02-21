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
  configVersion: "1"
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
  const srcRepo = path.join(root, 'testdata/repos/create1');
  const tempRepo = path.join(root, 'temp', 'create1_repo_second');
  fs.rmSync(tempRepo, { recursive: true, force: true });
  fs.mkdirSync(path.dirname(tempRepo), { recursive: true });
  fs.cpSync(srcRepo, tempRepo, { recursive: true });
  const cfgPath = path.join(root, 'temp', 'create1_tmp_second.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "create-meta"
  discovery: { root: "${path.join('temp', 'create1_repo_second').replaceAll('\\', '\\\\')}" }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const first = runThoth(bin, ['run', '--config', cfgPath], root);
  expect(first.status).toBe(0);
  expect(first.stderr).toBe('');
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
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update1_repo').replaceAll('\\', '\\\\')}" }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-update-meta', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const metaA = path.join(tempRepo, 'a.txt.thoth.yaml');
  const metaB = path.join(tempRepo, 'b.txt.thoth.yaml');
  expect(fs.readFileSync(metaA, 'utf8')).toBe(
    'locator: a.txt\nmeta:\n  x: 1\n',
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
  const srcRepo = path.join(root, 'testdata/repos/update1');
  const tempRepo = path.join(root, 'temp', 'update1_repo_keep');
  fs.rmSync(tempRepo, { recursive: true, force: true });
  fs.mkdirSync(path.dirname(tempRepo), { recursive: true });
  fs.cpSync(srcRepo, tempRepo, { recursive: true });
  fs.writeFileSync(
    path.join(tempRepo, 'a.txt.thoth.yaml'),
    'locator: a.txt\n# missing meta\n',
    'utf8',
  );
  const cfgPath = path.join(root, 'temp', 'update1_tmp_keep.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update1_repo_keep').replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  saveOutputs(root, 'run-update-meta-keep', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  expect(fs.readFileSync(path.join(tempRepo, 'b.txt.thoth.yaml'), 'utf8')).toBe(
    'locator: b.txt\nmeta: {}\n',
  );
  const out = JSON.parse(run.stdout) as {
    records: Array<{ locator: string; error?: unknown }>;
    errors: Array<unknown>;
  };
  const recA = out.records.find((r) => r.locator === 'a.txt');
  expect(recA?.error).toBeTruthy();
  expect(out.errors.length).toBeGreaterThan(0);
});

test('update-meta: invalid existing meta fails fast', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const srcRepo = path.join(root, 'testdata/repos/update1');
  const tempRepo = path.join(root, 'temp', 'update1_repo_fail');
  fs.rmSync(tempRepo, { recursive: true, force: true });
  fs.mkdirSync(path.dirname(tempRepo), { recursive: true });
  fs.cpSync(srcRepo, tempRepo, { recursive: true });
  fs.writeFileSync(
    path.join(tempRepo, 'a.txt.thoth.yaml'),
    'locator: a.txt\n# missing meta\n',
    'utf8',
  );
  const cfgPath = path.join(root, 'temp', 'update1_tmp_fail.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update1_repo_fail').replaceAll('\\', '\\\\')}" }
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
  expect(out.meta.diff.pairedCount).toBe(2);
  expect(out.meta.diff.orphanCount).toBe(1);
  expect(out.meta.diff.changedCount).toBe(0);
  expect(JSON.stringify(out.meta.diff.orphanMetaFiles)).toBe(
    JSON.stringify(['orphan.txt.thoth.yaml']),
  );
});

test('diff-meta: content diff v2 matches golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/diff2.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-diff-meta-v2', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/diff2_out.golden.json',
  );
  expect(run.stdout).toBe(expectedOut);
});

test('diff-meta: structural diff v3 arrays and type changes matches golden', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const cfg = path.join(root, 'testdata/configs/p5_diff_arrays1.cue');
  const run = runThoth(bin, ['run', '--config', cfg], root);
  saveOutputs(root, 'run-diff-meta-v3', run);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const expectedOut = expectedJSONFromGolden(
    root,
    'testdata/run/p5_diff_arrays1_out.golden.json',
  );
  expect(run.stdout).toBe(expectedOut);
});

test('update-meta: rewrite-stable canonical YAML (run twice, exact golden)', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'update_rewrite_repo');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });
  fs.writeFileSync(path.join(repo, 'z.txt'), 'z', 'utf8');
  fs.writeFileSync(
    path.join(repo, 'z.txt.thoth.yaml'),
    `locator: z.txt
meta:
  z: 1
  a:
    d: 4
    b: 2
    c:
      y: 2
      x: 1
`,
    'utf8',
  );
  const cfgPath = path.join(root, 'temp', 'update_rewrite_tmp.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update_rewrite_repo').replaceAll('\\', '\\\\')}" }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');

  const run1 = runThoth(bin, ['run', '--config', cfgPath], root);
  expect(run1.status).toBe(0);
  expect(run1.stderr).toBe('');
  const first = fs.readFileSync(path.join(repo, 'z.txt.thoth.yaml'), 'utf8');

  const run2 = runThoth(bin, ['run', '--config', cfgPath], root);
  expect(run2.status).toBe(0);
  expect(run2.stderr).toBe('');
  const second = fs.readFileSync(path.join(repo, 'z.txt.thoth.yaml'), 'utf8');

  expect(second).toBe(first);
  const golden = fs.readFileSync(
    path.join(
      root,
      'testdata',
      'golden',
      'meta',
      'update_nested_expected.thoth.yaml',
    ),
    'utf8',
  );
  expect(second).toBe(golden);
});

test('update-meta: deep-merge patch updates existing and creates missing meta', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'update_patch_repo');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });
  fs.writeFileSync(path.join(repo, 'a.txt'), 'a', 'utf8');
  fs.writeFileSync(path.join(repo, 'b.txt'), 'b', 'utf8');
  fs.writeFileSync(
    path.join(repo, 'a.txt.thoth.yaml'),
    `locator: a.txt
meta:
  keep: 1
  obj:
    y: 2
    x: 1
  arr:
    - 1
    - 2
`,
    'utf8',
  );
  const cfgPath = path.join(root, 'temp', 'update_patch_tmp.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update_patch_repo').replaceAll('\\', '\\\\')}" }
  updateMeta: {
    patch: {
      add: { k: 1 }
      obj: { y: 9, z: 3 }
      arr: [7]
    }
  }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');

  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');

  const gotA = fs.readFileSync(path.join(repo, 'a.txt.thoth.yaml'), 'utf8');
  const gotB = fs.readFileSync(path.join(repo, 'b.txt.thoth.yaml'), 'utf8');
  const wantA = fs.readFileSync(
    path.join(
      root,
      'testdata',
      'golden',
      'meta',
      'update_patch_existing_expected.thoth.yaml',
    ),
    'utf8',
  );
  const wantB = fs.readFileSync(
    path.join(
      root,
      'testdata',
      'golden',
      'meta',
      'update_patch_missing_expected.thoth.yaml',
    ),
    'utf8',
  );
  expect(gotA).toBe(wantA);
  expect(gotB).toBe(wantB);
});

test('update-meta: keep-going with invalid existing meta still updates valid file with patch', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = path.join(root, 'temp', 'update_patch_keep_repo');
  fs.rmSync(repo, { recursive: true, force: true });
  fs.mkdirSync(repo, { recursive: true });
  fs.writeFileSync(path.join(repo, 'bad.txt'), 'bad', 'utf8');
  fs.writeFileSync(path.join(repo, 'good.txt'), 'good', 'utf8');
  fs.writeFileSync(
    path.join(repo, 'bad.txt.thoth.yaml'),
    'locator: bad.txt\n# missing meta\n',
    'utf8',
  );
  const cfgPath = path.join(root, 'temp', 'update_patch_keep_tmp.cue');
  const cfgContent = `{
  configVersion: "1"
  action: "update-meta"
  discovery: { root: "${path.join('temp', 'update_patch_keep_repo').replaceAll('\\', '\\\\')}" }
  errors: { mode: "keep-going", embedErrors: true }
  updateMeta: {
    patch: {
      add: { k: 1 }
      obj: { y: 9, z: 3 }
      arr: [7]
    }
  }
}`;
  fs.writeFileSync(cfgPath, cfgContent, 'utf8');
  const run = runThoth(bin, ['run', '--config', cfgPath], root);
  expect(run.status).toBe(0);
  expect(run.stderr).toBe('');
  const out = JSON.parse(run.stdout) as {
    records: Array<{ locator: string; error?: unknown }>;
    errors: Array<unknown>;
  };
  const bad = out.records.find((r) => r.locator === 'bad.txt');
  expect(bad?.error).toBeTruthy();
  expect(out.errors.length).toBeGreaterThan(0);

  const gotGood = fs.readFileSync(
    path.join(repo, 'good.txt.thoth.yaml'),
    'utf8',
  );
  const wantGood = fs.readFileSync(
    path.join(
      root,
      'testdata',
      'golden',
      'meta',
      'update_patch_keep_valid_expected.thoth.yaml',
    ),
    'utf8',
  );
  expect(gotGood).toBe(wantGood);
});
