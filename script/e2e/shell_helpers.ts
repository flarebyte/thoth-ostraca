import { spawnSync } from 'node:child_process';

export function sh(
  cwd: string,
  cmd: string,
  args: string[],
  env?: Record<string, string>,
): string {
  const run = spawnSync(cmd, args, {
    cwd,
    encoding: 'utf8',
    env: { ...process.env, ...(env ?? {}) },
  });
  if (run.status !== 0) {
    throw new Error(`${cmd} ${args.join(' ')} failed: ${run.stderr}`);
  }
  return run.stdout;
}
