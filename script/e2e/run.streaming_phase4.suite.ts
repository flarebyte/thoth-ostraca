import { expect, test } from 'bun:test';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { buildBinary, projectRoot, runThoth } from './helpers';

function toCuePath(p: string): string {
  return p.replaceAll('\\', '\\\\');
}

function mkStreamingRepo(root: string, total: number): string {
  const base = fs.mkdtempSync(path.join(root, 'temp', 'phase4-stream-'));
  for (let i = 0; i < total; i += 1) {
    const dir = path.join(base, 'tree', (i % 17).toString().padStart(2, '0'));
    fs.mkdirSync(dir, { recursive: true });
    const file = path.join(dir, `f${i.toString().padStart(4, '0')}.thoth.yaml`);
    const body = `locator: "item/${i.toString().padStart(4, '0')}"\nmeta:\n  i: ${i}\n  enabled: true\n`;
    fs.writeFileSync(file, body, 'utf8');
  }
  return base;
}

function writeStreamingCfg(
  root: string,
  repo: string,
  workers: number,
  lines: boolean,
  maxRecords: number,
): string {
  const cfg = path.join(
    root,
    'temp',
    `phase4-stream-${process.pid}-${Date.now()}-${Math.random().toString(16).slice(2)}.cue`,
  );
  const body = `{
  configVersion: "1"
  action: "pipeline"
  discovery: { root: "${toCuePath(repo)}" }
  output: { lines: ${lines} }
  limits: { maxRecordsInMemory: ${maxRecords} }
  workers: ${workers}
  errors: { mode: "fail-fast" }
}`;
  fs.writeFileSync(cfg, body, 'utf8');
  return cfg;
}

test('phase4 streaming ndjson respects low maxRecordsInMemory and is deterministic', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkStreamingRepo(root, 200);
  const cfg1 = writeStreamingCfg(root, repo, 1, true, 50);
  const cfg8 = writeStreamingCfg(root, repo, 8, true, 50);

  const r1 = runThoth(bin, ['run', '--config', cfg1], root);
  const r8 = runThoth(bin, ['run', '--config', cfg8], root);

  expect(r1.status).toBe(0);
  expect(r8.status).toBe(0);
  expect(r1.stderr).toBe('');
  expect(r8.stderr).toBe('');

  const lines = r1.stdout.trim().split('\n');
  expect(lines.length).toBe(200);
  expect(r1.stdout).toBe(r8.stdout);
});

test('phase4 buffered mode fails when maxRecordsInMemory would overflow', () => {
  const root = projectRoot();
  const bin = buildBinary(root);
  const repo = mkStreamingRepo(root, 200);
  const cfg = writeStreamingCfg(root, repo, 1, false, 50);

  const run = runThoth(bin, ['run', '--config', cfg], root);
  expect(run.status).not.toBe(0);
  expect(run.stderr.includes('maxRecordsInMemory')).toBe(true);
  expect(run.stderr.includes('output.lines=true')).toBe(true);
});
