package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunHello(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	RunHello()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "Hello from internal/app") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
