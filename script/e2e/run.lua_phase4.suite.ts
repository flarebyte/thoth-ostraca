import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth } from './helpers';
import { normalizeWorkersOut } from './test_utils';

function toCuePath(p: string): string {
  return p.replaceAll('\\', '\\\\');
}

function mkLuaRepo(root: string, total: number): string {
  const base = fs.mkdtempSync(path.join(root, 'temp', 'phase4-lua-'));
  for (let i = 0; i < total; i += 1) {
    const dir = path.join(base, 'data');
    fs.mkdirSync(dir, { recursive: true });
    const file = path.join(dir, `f${i.toString().padStart(4, '0')}.thoth.yaml`);
    const body = `locator: "item/${i.toString().padStart(4, '0')}"\nmeta:\n  i: ${i}\n  enabled: true\n`;
    fs.writeFileSync(file, body, 'utf8');
  }
  return base;
}

function writeLuaCfg(
  root: string,
  repo: string,
  workers: number,
  mapInline: string,
  timeoutMs: number,
  instructionLimit: number,
): string {
  const cfg = path.join(
    root,
    'temp',
    `phase4-lua-${process.pid}-${Date.now()}-${Math.random().toString(16).slice(2)}.cue`,
  );
  const body = `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${toCuePath(repo)}" }
  workers: ${workers}
  errors: { mode: "fail-fast" }
  lua: {
    timeoutMs: ${timeoutMs}
    instructionLimit: ${instructionLimit}
    memoryLimitBytes: 8388608
    deterministicRandom: true
    libs: { base: true, table: true, string: true, math: true }
  }
  map: { inline: ${JSON.stringify(mapInline)} }
}`;
  fs.writeFileSync(cfg, body, 'utf8');
  return cfg;
}

test('phase4 lua fail-fast infinite loop triggers timeout', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkLuaRepo(root, 2);
  const cfg = writeLuaCfg(
    root,
    repo,
    1,
    'return (function() while true do end end)()',
    20,
    100000000,
  );
  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe('');
  expect(run.stderr.includes('lua-map: sandbox timeout')).toBe(true);
});

test('phase4 lua deterministic random is stable across workers', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkLuaRepo(root, 50);
  const code = 'return { locator = locator, r = math.random(1, 1000000) }';
  const cfg1 = writeLuaCfg(root, repo, 1, code, 2000, 1000000);
  const cfg8 = writeLuaCfg(root, repo, 8, code, 2000, 1000000);

  const r1 = runThoth(bin, ['run', '--config', cfg1], root);
  const r8 = runThoth(bin, ['run', '--config', cfg8], root);

  expect(r1.status).toBe(0);
  expect(r8.status).toBe(0);
  expect(r1.stderr).toBe('');
  expect(r8.stderr).toBe('');
  expect(normalizeWorkersOut(r1.stdout)).toBe(normalizeWorkersOut(r8.stdout));
});
