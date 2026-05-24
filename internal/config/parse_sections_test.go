package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeCue(t *testing.T, name, body string) string {
	t.Helper()
	d := t.TempDir()
	p := filepath.Join(d, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write cue: %v", err)
	}
	return p
}

func TestCompileCUEBranches(t *testing.T) {
	if _, err := compileCUE("x.json"); err == nil {
		t.Fatalf("expected extension error")
	}
	if _, err := compileCUE("/no/such/file.cue"); err == nil {
		t.Fatalf("expected read error")
	}
	bad := writeCue(t, "bad.cue", "{")
	if _, err := compileCUE(bad); err == nil {
		t.Fatalf("expected invalid cue error")
	}
}

func TestParseMinimal_AllSections(t *testing.T) {
	cfg := writeCue(t, "all.cue", `
configVersion: "1"
action: "diff-meta"
discovery: {
  root: "src"
  include: ["**/*.go"]
  exclude: ["vendor/**"]
  noGitignore: true
  followSymlinks: true
}
validation: { allowUnknownTopLevel: true }
limits: { maxYAMLBytes: 999, maxRecordsInMemory: 77 }
lua: {
  timeoutMs: 123
  instructionLimit: 456
  memoryLimitBytes: 789
  deterministicRandom: false
  libs: { base: true, table: true, string: true, math: true }
}
locatorPolicy: {
  allowAbsolute: true
  allowParentRefs: true
  posixStyle: true
  allowURLs: true
}
filter: { inline: "return true" }
map: { inline: "return {}" }
shell: {
  enabled: true
  decodeJsonStdout: true
  program: "sh"
  argsTemplate: ["-c", "echo hi"]
  workingDir: "."
  env: { A: "B" }
  timeoutMs: 11
  capture: { stdout: true, stderr: true, maxBytes: 99 }
  strictTemplating: true
  killProcessGroup: true
  termGraceMs: 22
}
postMap: { inline: "return {}" }
reduce: { inline: "return {}" }
persistMeta: { enabled: true, dryRun: true, outDir: "meta" }
updateMeta: { expectedLua: { inline: "return {}" }, patch: { a: 1 } }
diffMeta: {
  format: "summary"
  only: "all"
  summary: true
  failOnChange: true
  expectedLua: { inline: "return {}" }
  expectedPatch: { b: 2 }
}
output: { out: "-", pretty: true, lines: false }
errors: { mode: "keep-going", embedErrors: true }
workers: 3
ui: { progress: true, progressIntervalMs: 10 }
fileInfo: { enabled: true }
git: { enabled: true }
`)
	m, err := ParseMinimal(cfg)
	if err != nil {
		t.Fatalf("ParseMinimal err: %v", err)
	}
	if !m.Discovery.HasRoot || !m.Discovery.HasInclude || !m.Discovery.HasExclude || !m.Discovery.HasNoGitignore || !m.Discovery.HasFollowSymlink {
		t.Fatalf("discovery not parsed fully: %+v", m.Discovery)
	}
	if !m.Validation.HasAllowUnknownTop || !m.Limits.HasMaxYAMLBytes || !m.Limits.HasMaxRecordsInMemory {
		t.Fatalf("validation/limits not parsed: %+v %+v", m.Validation, m.Limits)
	}
	if !m.LuaSandbox.HasSection || !m.LuaSandbox.HasTimeoutMs || !m.LuaSandbox.Libs.HasMath {
		t.Fatalf("lua sandbox not parsed: %+v", m.LuaSandbox)
	}
	if !m.LocatorPolicy.HasAllowAbs || !m.LocatorPolicy.HasAllowParent || !m.LocatorPolicy.HasPosix || !m.LocatorPolicy.HasAllowURLs {
		t.Fatalf("locator policy not parsed: %+v", m.LocatorPolicy)
	}
	if !m.Filter.HasInline || !m.Map.HasInline || !m.Shell.HasSection || !m.Shell.HasProgram || !m.Shell.HasArgs || !m.Shell.HasEnv || !m.Shell.HasCapture {
		t.Fatalf("filter/map/shell not parsed: %+v %+v %+v", m.Filter, m.Map, m.Shell)
	}
	if !m.PostMap.HasInline || !m.Reduce.HasInline || !m.PersistMeta.HasSection || !m.PersistMeta.HasOutDir {
		t.Fatalf("post/reduce/persist not parsed: %+v %+v %+v", m.PostMap, m.Reduce, m.PersistMeta)
	}
	if !m.UpdateMeta.HasSection || !m.UpdateMeta.HasPatch || !m.UpdateMeta.HasExpectedLuaCode {
		t.Fatalf("updateMeta not parsed: %+v", m.UpdateMeta)
	}
	if !m.DiffMeta.HasSection || !m.DiffMeta.HasExpectedPatch || !m.DiffMeta.HasExpectedLuaCode || !m.DiffMeta.HasFormat || !m.DiffMeta.HasOnly || !m.DiffMeta.HasSummary || !m.DiffMeta.HasFailOnChange {
		t.Fatalf("diffMeta not parsed: %+v", m.DiffMeta)
	}
	if !m.Output.HasOut || !m.Output.HasPretty || !m.Output.HasLines || !m.Errors.HasMode || !m.Errors.HasEmbed || !m.Workers.HasCount || !m.UI.HasSection || !m.FileInfo.HasEnabled || !m.Git.HasEnabled {
		t.Fatalf("output/misc not parsed: %+v %+v %+v %+v %+v %+v", m.Output, m.Errors, m.Workers, m.UI, m.FileInfo, m.Git)
	}
}

func TestParseSectionErrors(t *testing.T) {
	badDiscovery := writeCue(t, "d.cue", `configVersion: "1"
action: "validate"
discovery: { include: 1 }`)
	if _, err := ParseMinimal(badDiscovery); err == nil {
		t.Fatalf("expected discovery include error")
	}

	badUpdate := writeCue(t, "u.cue", `configVersion: "1"
action: "validate"
updateMeta: { expectedLua: { inline: 1 } }`)
	if _, err := ParseMinimal(badUpdate); err == nil {
		t.Fatalf("expected updateMeta expectedLua error")
	}

	badDiff := writeCue(t, "df.cue", `configVersion: "1"
action: "validate"
diffMeta: { format: "bad" }`)
	if _, err := ParseMinimal(badDiff); err == nil {
		t.Fatalf("expected diffMeta format error")
	}
}
