import * as fs from 'node:fs';
import * as path from 'node:path';

export function writeCfg(
  root: string,
  name: string,
  cfg: Record<string, unknown>,
): string {
  const p = path.join(root, 'temp', `${name}.cue`);
  fs.mkdirSync(path.dirname(p), { recursive: true });
  fs.writeFileSync(p, `${JSON.stringify(cfg, null, 2)}\n`, 'utf8');
  return p;
}
