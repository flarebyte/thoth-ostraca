#!/usr/bin/env bun

// Bun/TypeScript rewrite of build-go.mjs (zx)
// Builds Go binaries for multiple platforms and writes checksums

import crypto from 'node:crypto';
import { promises as fs } from 'node:fs';
import path from 'node:path';

async function readFileSafe(p: string): Promise<string> {
  try {
    return await fs.readFile(p, 'utf8');
  } catch {
    return '';
  }
}

async function ensureDir(p: string): Promise<void> {
  await fs.mkdir(p, { recursive: true });
}

async function runChecked(
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

async function runCapture(
  cmd: string[],
  opts: { cwd?: string; env?: Record<string, string | undefined> } = {},
): Promise<string> {
  const proc = Bun.spawn(cmd, {
    cwd: opts.cwd,
    env: opts.env,
    stdout: 'pipe',
    stderr: 'pipe',
  });
  const exitCode = await proc.exited;
  if (exitCode !== 0) {
    return '';
  }
  return (await new Response(proc.stdout).text()).trim();
}

async function sha256File(filePath: string): Promise<string> {
  const hash = crypto.createHash('sha256');
  const file = Bun.file(filePath);
  const stream = file.stream();
  const reader = stream.getReader();
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    if (value) hash.update(value);
  }
  return hash.digest('hex');
}

async function main() {
  const version =
    (process.env.VERSION ?? (await readFileSafe('VERSION'))).trim() || '0.0.0';

  const currentDirectory = process.cwd();
  const folderName = path.basename(currentDirectory);
  const modulePath =
    (await runCapture(['go', 'list', '-m'], { cwd: currentDirectory })) ||
    `github.com/flarebyte/${folderName}`;
  const commitFromGit = await runCapture(
    ['git', 'rev-parse', '--short=12', 'HEAD'],
    { cwd: currentDirectory },
  );
  const commit = process.env.COMMIT || commitFromGit || 'unknown';
  const currentDate =
    process.env.DATE ?? new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');

  // Keep cli.* for compatibility and set internal/buildinfo.* for `thoth version`.
  const ldflags = [
    `-X ${modulePath}/cli.Version=${version}`,
    `-X ${modulePath}/cli.Date=${currentDate}`,
    `-X ${modulePath}/internal/buildinfo.Version=${version}`,
    `-X ${modulePath}/internal/buildinfo.Commit=${commit}`,
    `-X ${modulePath}/internal/buildinfo.Date=${currentDate}`,
  ].join(' ');

  const platforms = [
    { label: 'Linux (amd64)', os: 'linux', arch: 'amd64' },
    { label: 'Linux (arm64)', os: 'linux', arch: 'arm64' },
    { label: 'macOS (Apple Silicon)', os: 'darwin', arch: 'arm64' },
  ] as const;

  await ensureDir('build');

  const builtFiles: string[] = [];

  for (const p of platforms) {
    console.log(p.label);
    const env: Record<string, string> = { ...process.env } as Record<
      string,
      string
    >;
    env.GOOS = p.os;
    env.GOARCH = p.arch;
    if (p.os === 'darwin') {
      const macArch = p.arch === 'amd64' ? 'x86_64' : 'arm64';
      env.CGO_ENABLED = '1';
      env.CC = 'clang';
      env.CGO_CFLAGS = `-arch ${macArch}`;
      env.CGO_LDFLAGS = `-arch ${macArch}`;
      env.MACOSX_DEPLOYMENT_TARGET = env.MACOSX_DEPLOYMENT_TARGET || '11.0';
    }

    const out = path.join('build', `${folderName}-${p.os}-${p.arch}`);
    await runChecked(
      ['go', 'build', '-o', out, '-ldflags', ldflags, './cmd/thoth'],
      { env },
    );
    builtFiles.push(out);
  }

  // checksums (sha256), format: "<hex>  <path>" like shasum
  const lines: string[] = [];
  for (const f of builtFiles) {
    const digest = await sha256File(f);
    lines.push(`${digest}  ${f}`);
  }
  await fs.writeFile('build/checksums.txt', `${lines.join('\n')}\n`, 'utf8');
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
