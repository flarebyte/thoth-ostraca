package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeCfg(t *testing.T, body string) string {
	t.Helper()
	d := t.TempDir()
	p := filepath.Join(d, "cfg.cue")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	return p
}

func TestLoadAndValidateAndParseMinimal(t *testing.T) {
	cfg := writeCfg(t, `configVersion: "1"
action: "validate"`)
	if err := LoadAndValidate(cfg); err != nil {
		t.Fatalf("LoadAndValidate err: %v", err)
	}
	m, err := ParseMinimal(cfg)
	if err != nil {
		t.Fatalf("ParseMinimal err: %v", err)
	}
	if m.Action != "validate" {
		t.Fatalf("unexpected action: %s", m.Action)
	}
}

func TestLoadAndValidateErrors(t *testing.T) {
	cfg := writeCfg(t, `configVersion: "1"`)
	if err := LoadAndValidate(cfg); err == nil {
		t.Fatalf("expected missing action error")
	}
	cfgType1 := writeCfg(t, `configVersion: 1
action: "validate"`)
	if err := LoadAndValidate(cfgType1); err == nil {
		t.Fatalf("expected configVersion type error")
	}
	cfgType2 := writeCfg(t, `configVersion: "1"
action: 1`)
	if err := LoadAndValidate(cfgType2); err == nil {
		t.Fatalf("expected action type error")
	}
	cfg2 := writeCfg(t, `configVersion: "9"
action: "validate"`)
	if err := LoadAndValidate(cfg2); err == nil {
		t.Fatalf("expected unsupported config version error")
	}
}

func TestRequireStringField(t *testing.T) {
	p := writeCfg(t, `x: 1`)
	v, err := compileCUE(p)
	if err != nil {
		t.Fatalf("compileCUE err: %v", err)
	}
	if err := requireStringField(v, "missing"); err == nil {
		t.Fatalf("expected missing field error")
	}
}
