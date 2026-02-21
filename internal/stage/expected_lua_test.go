package stage

import (
	"encoding/json"
	"testing"
)

func TestRunExpectedLuaInline_ReturnsFunction(t *testing.T) {
	_, _, err := runExpectedLuaInline("diff-meta-expectedLua", &Meta{}, "a.txt", map[string]any{"x": 1}, "return 1")
	if err == nil || err.Error() != "expectedLua must return function" {
		t.Fatalf("expected function error, got: %v", err)
	}
}

func TestRunExpectedLuaInline_ReturnsObject(t *testing.T) {
	_, _, err := runExpectedLuaInline("diff-meta-expectedLua", &Meta{}, "a.txt", map[string]any{"x": 1}, "return function(locator, existingMeta) return 1 end")
	if err == nil || err.Error() != "expectedLua must return object" {
		t.Fatalf("expected object error, got: %v", err)
	}
}

func TestRunExpectedLuaInline_Deterministic(t *testing.T) {
	code := "return function(locator, existingMeta) return { locator = locator, x = existingMeta and existingMeta.x or 0 } end"
	a, _, err := runExpectedLuaInline("diff-meta-expectedLua", &Meta{}, "a.txt", map[string]any{"x": 1}, code)
	if err != nil {
		t.Fatalf("run 1: %v", err)
	}
	b, _, err := runExpectedLuaInline("diff-meta-expectedLua", &Meta{}, "a.txt", map[string]any{"x": 1}, code)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	if string(aj) != string(bj) {
		t.Fatalf("nondeterministic output: %s vs %s", string(aj), string(bj))
	}
}
