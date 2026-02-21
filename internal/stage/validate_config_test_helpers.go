package stage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func runValidateConfigWithContent(t *testing.T, fileName, content string) (Envelope, error) {
	t.Helper()
	_ = os.MkdirAll("temp", 0o755)
	cfg := filepath.Join("temp", fileName)
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfg}}
	return Run(context.Background(), "validate-config", in, Deps{})
}
