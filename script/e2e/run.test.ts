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

test("thoth run with valid config prints envelope JSON", () => {
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
  const parsed = JSON.parse(run.stdout);
  expect(typeof parsed).toBe("object");
  expect(Array.isArray(parsed.records)).toBe(true);
  expect(typeof parsed.meta.config.configVersion).toBe("string");
  expect(typeof parsed.meta.config.action).toBe("string");
});

test("thoth run executes discovery and respects gitignore by default", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/discovery1.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/discovery1_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("thoth run parses YAML records into {locator,meta}", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/yaml1.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/yaml1_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("thoth run fails on invalid YAML (missing meta)", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/yaml1_nogit.cue");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe("");
  expect(run.stderr.includes("missing required field: meta")).toBe(true);
});

test("thoth run filters records via Lua predicate", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/filter1.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/filter1_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("thoth run fails on invalid Lua", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/filter_bad.cue");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe("");
  expect(run.stderr.includes("lua-filter")).toBe(true);
});

test("thoth run maps records via Lua transform", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/map1.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/map1_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("thoth run fails on invalid map Lua", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/map_bad.cue");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe("");
  expect(run.stderr.includes("lua-map")).toBe(true);
});

test("thoth run executes shell and postmap", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/shell1.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/shell1_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("thoth run fails when shell program is missing", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/shell_bad_prog.cue");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe("");
  expect(run.stderr.includes("shell-exec")).toBe(true);
});

test("thoth run reduces to count and prints full envelope", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/reduce1.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/reduce1_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("thoth run prints NDJSON lines when output.lines is true", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/lines1.cue");
  const expectedLines = fs.readFileSync(path.join(root, "testdata/run/lines1_out.golden.ndjson"), "utf8");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedLines);
});

test("thoth run fails on invalid reduce Lua", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/reduce_bad.cue");
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).not.toBe(0);
  expect(run.stdout).toBe("");
  expect(run.stderr.includes("lua-reduce")).toBe(true);
});

test("keep-going with embedErrors=true embeds record errors and lists envelope errors", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/keep1_embed_true.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/keep1_embed_true_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
});

test("keep-going with embedErrors=false only lists envelope errors", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const cfg = path.join(root, "testdata/configs/keep1_embed_false.cue");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/run/keep1_embed_false_out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";
  const run = spawnSync(bin, ["run", "--config", cfg], { encoding: "utf8", cwd: root });
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);
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
  expect(run.stderr.includes("action")).toBe(true);
});
