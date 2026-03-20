package stage

import (
	"strings"
	"testing"
)

func TestFormatLuaError_WithLineExcerpt(t *testing.T) {
	t.Parallel()

	code := "local a = 1\nlocal b = nil\nreturn b.missing\nreturn a"
	msg := "<string>:3: attempt to index nil with key 'missing'"

	got := formatLuaError("lua-map", "x.go", code, msg)

	if !strings.Contains(got, msg) {
		t.Fatalf("expected original message to be preserved, got %q", got)
	}
	if !strings.Contains(got, "  2 | local b = nil") {
		t.Fatalf("expected context line 2, got %q", got)
	}
	if !strings.Contains(got, "> 3 | return b.missing") {
		t.Fatalf("expected highlighted line 3, got %q", got)
	}
	if !strings.Contains(got, "  4 | return a") {
		t.Fatalf("expected context line 4, got %q", got)
	}
}

func TestFormatLuaError_WithoutLineExcerpt(t *testing.T) {
	t.Parallel()

	got := formatLuaError("lua-map", "x.go", "return true", "sandbox timeout")
	if got != "sandbox timeout" {
		t.Fatalf("expected unchanged message, got %q", got)
	}
}
