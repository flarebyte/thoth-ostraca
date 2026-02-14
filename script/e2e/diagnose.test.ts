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

test("diagnose echo prints expected JSON and writes dumps", () => {
  const root = path.resolve(__dirname, "../..");
  const bin = buildBinary(root);
  const input = path.join(root, "testdata/diagnose/input.json");
  const expectedOutRaw = fs.readFileSync(path.join(root, "testdata/diagnose/out.golden.json"), "utf8");
  const expectedOut = JSON.stringify(JSON.parse(expectedOutRaw)) + "\n";

  const tmpRoot = path.join(root, "temp");
  fs.mkdirSync(tmpRoot, { recursive: true });
  const dumpDir = fs.mkdtempSync(path.join(tmpRoot, "diag-"));
  const dumpIn = path.join(dumpDir, "in.json");
  const dumpOut = path.join(dumpDir, "out.json");

  const run = spawnSync(
    bin,
    [
      "diagnose",
      "--stage",
      "echo",
      "--in",
      input,
      "--dump-in",
      dumpIn,
      "--dump-out",
      dumpOut,
    ],
    { encoding: "utf8", cwd: root }
  );
  expect(run.status).toBe(0);
  expect(run.stderr).toBe("");
  expect(run.stdout).toBe(expectedOut);

  // Dump files exist and contents match expected JSON (exact)
  expect(fs.existsSync(dumpIn)).toBe(true);
  expect(fs.existsSync(dumpOut)).toBe(true);

  const expectedDumpIn = JSON.stringify(JSON.parse(fs.readFileSync(input, "utf8")));
  const expectedDumpOut = JSON.stringify(JSON.parse(expectedOutRaw));
  expect(fs.readFileSync(dumpIn, "utf8")).toBe(expectedDumpIn);
  expect(fs.readFileSync(dumpOut, "utf8")).toBe(expectedDumpOut);
});
