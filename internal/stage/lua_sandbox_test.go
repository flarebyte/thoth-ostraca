package stage

import (
	"testing"
)

func defaultLuaSandboxForTest() *Meta {
	return &Meta{
		LuaSandbox: &LuaSandboxMeta{
			TimeoutMs:        2000,
			InstructionLimit: 1000000,
			MemoryLimitBytes: 8388608,
			Libs: LuaSandboxLibsMeta{
				Base:   true,
				Table:  true,
				String: true,
				Math:   true,
			},
			DeterministicRandom: true,
		},
	}
}

func TestLuaSandbox_Timeout(t *testing.T) {
	meta := defaultLuaSandboxForTest()
	meta.LuaSandbox.TimeoutMs = 10
	meta.LuaSandbox.InstructionLimit = 100000000
	_, _, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, "while true do end", "fail-fast", meta)
	if err == nil || err.Error() != "lua-map: sandbox timeout" {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestLuaSandbox_InstructionLimit(t *testing.T) {
	meta := defaultLuaSandboxForTest()
	meta.LuaSandbox.TimeoutMs = 2000
	meta.LuaSandbox.InstructionLimit = 10
	_, _, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, "while true do end", "fail-fast", meta)
	if err == nil || err.Error() != "lua-map: sandbox instruction limit" {
		t.Fatalf("expected instruction limit, got %v", err)
	}
}

func TestLuaSandbox_MemoryLimit(t *testing.T) {
	meta := defaultLuaSandboxForTest()
	meta.LuaSandbox.MemoryLimitBytes = 64
	meta.LuaSandbox.InstructionLimit = 100000000
	_, _, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, "return string.rep('a', 1024)", "fail-fast", meta)
	if err == nil || err.Error() != "lua-map: sandbox memory limit" {
		t.Fatalf("expected memory limit, got %v", err)
	}
}

func TestLuaSandbox_LibAllowlist(t *testing.T) {
	meta := defaultLuaSandboxForTest()
	meta.LuaSandbox.Libs.String = false
	_, envE, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, "return string.lower('A')", "keep-going", meta)
	if err != nil {
		t.Fatalf("unexpected fatal: %v", err)
	}
	if envE == nil {
		t.Fatalf("expected keep-going env error")
	}
	if envE.Stage != luaMapStage || envE.Locator != "a" {
		t.Fatalf("unexpected env error shape: %+v", envE)
	}
}

func TestLuaSandbox_DeterministicRandom(t *testing.T) {
	meta := defaultLuaSandboxForTest()
	code := "return { locator = locator, r = math.random(1, 1000000) }"
	r1, _, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, code, "fail-fast", meta)
	if err != nil {
		t.Fatalf("run1: %v", err)
	}
	r2, _, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, code, "fail-fast", meta)
	if err != nil {
		t.Fatalf("run2: %v", err)
	}
	r3, _, err := processLuaMapRecord(Record{Locator: "b", Meta: map[string]any{}}, code, "fail-fast", meta)
	if err != nil {
		t.Fatalf("run3: %v", err)
	}
	m1, _ := r1.Mapped.(map[string]any)
	m2, _ := r2.Mapped.(map[string]any)
	m3, _ := r3.Mapped.(map[string]any)
	if m1["r"] != m2["r"] {
		t.Fatalf("expected deterministic random for same locator")
	}
	if m1["r"] == m3["r"] {
		t.Fatalf("expected record-specific deterministic random")
	}
}

func TestLuaSandbox_KeepGoingViolationShape(t *testing.T) {
	meta := defaultLuaSandboxForTest()
	meta.LuaSandbox.InstructionLimit = 10
	rec, envE, err := processLuaMapRecord(Record{Locator: "a", Meta: map[string]any{}}, "while true do end", "keep-going", meta)
	if err != nil {
		t.Fatalf("unexpected fatal: %v", err)
	}
	if envE == nil || envE.Message != sandboxInstructionViolation {
		t.Fatalf("expected keep-going sandbox env error, got %+v", envE)
	}
	if rec.Error == nil || rec.Error.Message != sandboxInstructionViolation {
		t.Fatalf("expected keep-going record error, got %+v", rec.Error)
	}
}
