package stage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return b
}

func normalizeJSON(t *testing.T, b []byte) []byte {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return append(out, '\n')
}

func TestEnvelopeContractSnapshotV1(t *testing.T) {
	env := Envelope{
		Records: []Record{{
			Locator: "a",
			Meta:    map[string]any{"k": "v"},
			Mapped:  map[string]any{"x": 1.0},
			Shell:   &ShellResult{ExitCode: 0, Stdout: "ok\n", Stderr: ""},
			Post:    map[string]any{"locator": "a"},
		}},
		Meta: &Meta{ContractVersion: "1", Stage: "echo", Config: &ConfigMeta{ConfigVersion: "v0", Action: "nop"}},
	}
	if err := ValidateEnvelope(env); err != nil {
		t.Fatalf("ValidateEnvelope failed: %v", err)
	}
	got := normalizeJSON(t, mustJSON(env))
	want := mustRead(t, filepath.Join("..", "..", "testdata", "contracts", "envelope_v1.golden.json"))
	if string(got) != string(normalizeJSON(t, want)) {
		t.Fatalf("contract snapshot mismatch\nwant: %s\n got: %s", string(want), string(got))
	}
}

func TestRecordContractSnapshotV1(t *testing.T) {
	rec := Record{Locator: "x", Meta: map[string]any{"enabled": true, "name": "A"}}
	got := normalizeJSON(t, mustJSON(rec))
	want := mustRead(t, filepath.Join("..", "..", "testdata", "contracts", "record_v1.golden.json"))
	if string(got) != string(normalizeJSON(t, want)) {
		t.Fatalf("record snapshot mismatch\nwant: %s\n got: %s", string(want), string(got))
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
