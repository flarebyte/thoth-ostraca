import { promises as fs } from 'node:fs';

export async function readFileSafe(p: string): Promise<string> {
  try {
    return await fs.readFile(p, 'utf8');
  } catch {
    return '';
  }
}

export async function readVersionFromProjectYAML(
  p = 'main.project.yaml',
): Promise<string> {
  const raw = await readFileSafe(p);
  if (!raw) return '';
  const lines = raw.split(/\r?\n/);
  let inTags = false;
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    if (!inTags) {
      if (/^tags\s*:\s*$/.test(trimmed)) inTags = true;
      continue;
    }
    if (/^\S/.test(line)) break;
    const m = line.match(/^\s*version\s*:\s*(.+)\s*$/);
    if (m) {
      let v = m[1].trim();
      if (
        (v.startsWith('"') && v.endsWith('"')) ||
        (v.startsWith("'") && v.endsWith("'"))
      ) {
        v = v.slice(1, -1);
      }
      return v;
    }
  }
  return '';
}

export async function runChecked(
  cmd: string[],
  opts: { cwd?: string; env?: Record<string, string | undefined> } = {},
) {
  const proc = Bun.spawn(cmd, {
    cwd: opts.cwd,
    env: opts.env,
    stdout: 'inherit',
    stderr: 'inherit',
  });
  const exitCode = await proc.exited;
  if (exitCode !== 0) {
    throw new Error(`Command failed (${exitCode}): ${cmd.join(' ')}`);
  }
}
