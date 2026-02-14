#!/usr/bin/env bun
/**
 * Minimal Go CLI scaffold (Hello World) for this template.
 *
 * Behavior
 * - Creates a tiny Go CLI entry and an internal app package.
 * - Does NOT overwrite existing files.
 * - Uses zero external dependencies (no cobra, etc.).
 *
 * Usage
 * - bun run scratch/scaffold.mts
 * - Then run: `go run .` (prints Hello World), or `go run . hello`.
 *
 * Notes
 * - Generic files like .gitignore or build scripts are intentionally NOT created here.
 */

import { promises as fs } from 'node:fs';
import { dirname, join } from 'node:path';

async function ensureDir(p: string) {
  await fs.mkdir(p, { recursive: true });
}

async function writeIfMissing(relPath: string, content: string) {
  try {
    await fs.stat(relPath);
    console.log(`skip existing: ${relPath}`);
    return false;
  } catch {
    const dir = dirname(relPath);
    if (dir && dir !== '.') await ensureDir(dir);
    await fs.writeFile(relPath, content, { encoding: 'utf8', flag: 'wx' });
    console.log(`created: ${relPath}`);
    return true;
  }
}

const mainGo = `package main

import (
    "os"
    "fmt"
)

// TODO: CLI Wiring
// - This is a minimal entry point without external deps.
// - For subcommands, switch on os.Args and call functions in internal/app.
// - Later, you can replace this with a CLI framework if desired.
func main() {
    args := os.Args[1:]
    if len(args) > 0 {
        switch args[0] {
        case "hello":
            fmt.Println("Hello World")
            return
        }
    }
    fmt.Println("Hello World")
}
`;

const appGo = `package app

import (
    "fmt"
)

// Run is a simple placeholder for application logic.
// TODO: Replace prints with your actual command handlers.
func RunHello() {
    fmt.Println("Hello from internal/app")
}
`;

async function main() {
  const created: string[] = [];
  if (await writeIfMissing('main.go', mainGo)) created.push('main.go');
  if (await writeIfMissing(join('internal', 'app', 'app.go'), appGo))
    created.push('internal/app/app.go');

  if (created.length === 0) {
    console.log("No files created (all present). You're good to go.");
  } else {
    console.log('\nNext steps:');
    console.log('- go run .          # prints Hello World');
    console.log(
      '- go run . hello    # prints Hello World (explicit subcommand)',
    );
    console.log('\nTODOs:');
    console.log('- Wire flags/subcommands (in main.go) as needed.');
    console.log('- Move logic into internal/app and call it from main.');
    console.log('- Add your own build/test scripts as needed.');
  }
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
