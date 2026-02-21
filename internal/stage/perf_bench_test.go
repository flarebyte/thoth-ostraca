package stage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// Benchmarks in this file cover discovery walk/glob, YAML parsing, Lua filter/map,
// and shell execution overhead. They are intended for local profiling and regression checks.

const (
	benchFileCount = 1000
	benchLuaCount  = 1000
)

func writeBenchmarkYAMLFiles(b *testing.B, root string, n int) []string {
	b.Helper()
	locators := make([]string, 0, n)
	for i := 0; i < n; i++ {
		dir := filepath.Join(root, fmt.Sprintf("d%02d", i%20))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("mkdir failed: %v", err)
		}
		name := fmt.Sprintf("f%04d.thoth.yaml", i)
		rel := filepath.ToSlash(filepath.Join(fmt.Sprintf("d%02d", i%20), name))
		body := fmt.Sprintf("locator: item-%04d\nmeta:\n  enabled: true\n  value: %d\n", i, i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			b.Fatalf("write yaml failed: %v", err)
		}
		locators = append(locators, rel)
	}
	sort.Strings(locators)
	return locators
}

func benchmarkRootMeta(root string) *Meta {
	return &Meta{
		Discovery: &DiscoveryMeta{Root: root},
		Limits:    &LimitsMeta{MaxRecordsInMemory: defaultMaxRecordsInMemory},
		Workers:   1,
	}
}

func BenchmarkDiscoverMetaFiles(b *testing.B) {
	root := b.TempDir()
	_ = writeBenchmarkYAMLFiles(b, root, benchFileCount)
	in := Envelope{
		Records: []Record{},
		Meta:    benchmarkRootMeta(root),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := discoverRunner(context.Background(), in, Deps{})
		if err != nil {
			b.Fatalf("discover failed: %v", err)
		}
		if len(out.Records) != benchFileCount {
			b.Fatalf("unexpected record count: got %d want %d", len(out.Records), benchFileCount)
		}
	}
}

func BenchmarkParseValidateYAML(b *testing.B) {
	root := b.TempDir()
	locators := writeBenchmarkYAMLFiles(b, root, benchFileCount)
	records := make([]Record, 0, len(locators))
	for _, l := range locators {
		records = append(records, Record{Locator: l})
	}
	in := Envelope{
		Records: records,
		Meta:    benchmarkRootMeta(root),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := parseValidateYAMLRunner(context.Background(), in, Deps{})
		if err != nil {
			b.Fatalf("parse-validate-yaml failed: %v", err)
		}
		if len(out.Records) != benchFileCount {
			b.Fatalf("unexpected record count: got %d want %d", len(out.Records), benchFileCount)
		}
	}
}

func benchmarkLuaEnvelope() Envelope {
	records := make([]Record, 0, benchLuaCount)
	for i := 0; i < benchLuaCount; i++ {
		records = append(records, Record{
			Locator: fmt.Sprintf("item-%04d", i),
			Meta: map[string]any{
				"enabled": i%2 == 0,
				"value":   i,
			},
		})
	}
	return Envelope{
		Records: records,
		Meta: &Meta{
			Workers: 1,
			Lua: &LuaMeta{
				FilterInline: "return (meta and meta.enabled) == true",
				MapInline:    "return { locator = locator, value = (meta and meta.value) or 0 }",
			},
			Limits: &LimitsMeta{MaxRecordsInMemory: defaultMaxRecordsInMemory},
		},
	}
}

func BenchmarkLuaFilter(b *testing.B) {
	in := benchmarkLuaEnvelope()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := luaFilterRunner(context.Background(), in, Deps{})
		if err != nil {
			b.Fatalf("lua-filter failed: %v", err)
		}
		if len(out.Records) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

func BenchmarkLuaMap(b *testing.B) {
	in := benchmarkLuaEnvelope()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := luaMapRunner(context.Background(), in, Deps{})
		if err != nil {
			b.Fatalf("lua-map failed: %v", err)
		}
		if len(out.Records) != benchLuaCount {
			b.Fatalf("unexpected record count: got %d want %d", len(out.Records), benchLuaCount)
		}
	}
}

func BenchmarkShellExecNoOp(b *testing.B) {
	program := "sh"
	args := []string{"-c", "true"}
	if runtime.GOOS == "windows" {
		program = "cmd"
		args = []string{"/C", "exit 0"}
	}
	if _, err := exec.LookPath(program); err != nil {
		b.Skipf("%s not available", program)
	}

	records := make([]Record, 0, 200)
	for i := 0; i < 200; i++ {
		records = append(records, Record{Locator: fmt.Sprintf("r-%03d", i), Mapped: map[string]any{"id": i}})
	}
	in := Envelope{
		Records: records,
		Meta: &Meta{
			Workers: 1,
			Shell: &ShellMeta{
				Enabled:      true,
				Program:      program,
				ArgsTemplate: args,
				WorkingDir:   ".",
				TimeoutMs:    5000,
				Capture: ShellCaptureMeta{
					Stdout:   true,
					Stderr:   true,
					MaxBytes: 1024,
				},
				StrictTemplating: true,
				KillProcessGroup: false,
				TermGraceMs:      100,
			},
			Limits: &LimitsMeta{MaxRecordsInMemory: defaultMaxRecordsInMemory},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := shellExecRunner(context.Background(), in, Deps{})
		if err != nil {
			b.Fatalf("shell-exec failed: %v", err)
		}
		if len(out.Records) != len(records) {
			b.Fatalf("unexpected record count: got %d want %d", len(out.Records), len(records))
		}
	}
}
