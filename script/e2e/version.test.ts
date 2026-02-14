import { test, expect } from "bun:test";
import { spawnSync } from "child_process";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

test("thoth version prints dev", () => {
  const root = path.resolve(__dirname, "../..");

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

  const run = spawnSync(bin, ["version"], { encoding: "utf8" });
  expect(run.status).toBe(0);
  expect(run.stdout).toBe("thoth dev\n");
});
