import * as fs from 'node:fs';
import * as path from 'node:path';

type EnvelopeOut = Record<string, unknown> & { meta?: { workers?: unknown } };

export function normalizeWorkersOut(stdout: string): string {
  const actual = JSON.parse(stdout) as EnvelopeOut;
  if (actual.meta && 'workers' in actual.meta) {
    delete actual.meta.workers;
  }
  return `${JSON.stringify(actual)}\n`;
}

export function writeCueConfig(cfgPath: string, body: string): void {
  fs.mkdirSync(path.dirname(cfgPath), { recursive: true });
  fs.writeFileSync(cfgPath, body, 'utf8');
}

export function seededShuffle(values: number[], seed: number): number[] {
  let s = seed >>> 0;
  const rnd = (): number => {
    s = (s * 1664525 + 1013904223) >>> 0;
    return s / 0x100000000;
  };
  for (let i = values.length - 1; i > 0; i -= 1) {
    const j = Math.floor(rnd() * (i + 1));
    const tmp = values[i];
    values[i] = values[j];
    values[j] = tmp;
  }
  return values;
}

export function sortKey(
  e: { stage?: string; locator?: string; message?: string } | undefined,
): string {
  if (!e) return '';
  return `${e.stage ?? ''}\u0000${e.locator ?? ''}\u0000${e.message ?? ''}`;
}
