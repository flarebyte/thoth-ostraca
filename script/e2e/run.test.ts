import { test, expect } from "bun:test";
import { spawnSync } from "child_process";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function buildBinary(root: string) {
  const binDir = path.join(root, ".e2e-bin");
  fs.mkdirSync(binDir, { recursive: true });
  const bin = path.join(binDir, process.platform === "win32" ? "thoth.exe" : "thoth");
  const env = {
    ...process.env,
    CGO_ENABLED: "0",
    GOCACHE: path.join(root, ".gocache"),
    GOFLAGS: "-mod=mod",
  } as Record<string, string>;
  const build = spawnSync("go", ["build", "-o", bin, "./cmd/thoth"], {
    cwd: root,
    env,
    encoding: "utf8",
  });
  if (build.status !== 0) {
    throw new Error(`build failed (status ${build.status})\n${build.stdout}\n${build.stderr}`);
  }
  return bin;
}

test("thoth run with valid config prints ok JSON", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/minimal.cue");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  // Save outputs for inspection; temp/ is git-ignored
  const tempDir = path.join(root, "temp");
  fs.mkdirSync(tempDir, { recursive: true });
  fs.writeFileSync(path.join(tempDir, "out.txt"), run.stdout);
  fs.writeFileSync(path.join(tempDir, "err.txt"), (run as any).stderr ?? "");
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe("{\"ok\":true}\n");
});

test("thoth run with missing field fails with short error", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/missing_action.cue");
  const run = spawnSync(bin, ["run", "-c", cfg], { encoding: "utf8", cwd: root });
  // Save outputs for inspection; temp/ is git-ignored
  const tempDir2 = path.join(root, "temp");
  fs.mkdirSync(tempDir2, { recursive: true });
  fs.writeFileSync(path.join(tempDir2, "out.txt"), run.stdout);
  fs.writeFileSync(path.join(tempDir2, "err.txt"), (run as any).stderr ?? "");
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe("");
  expect(run.stderr.includes("missing required field: action")).toBe(true);
});
