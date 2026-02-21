package stage

import (
	"path/filepath"
	"testing"
)

func TestEnvelopeContractSnapshotV1(t *testing.T) {
	env := Envelope{
		Records: []Record{{
			Locator: "a",
			Meta:    map[string]any{"k": "v"},
			Mapped:  map[string]any{"x": 1.0},
			Shell: &ShellResult{
				ExitCode:        0,
				Stdout:          strPtr("ok\n"),
				Stderr:          strPtr(""),
				StdoutTruncated: false,
				StderrTruncated: false,
				TimedOut:        false,
			},
			Post: map[string]any{"locator": "a"},
		}},
		Meta: &Meta{ContractVersion: "1", Stage: "echo", Config: &ConfigMeta{ConfigVersion: "1", Action: "nop"}},
	}
	if err := ValidateEnvelope(env); err != nil {
		t.Fatalf("ValidateEnvelope failed: %v", err)
	}
	got := append(normalizeJSON(t, mustJSON(env)), '\n')
	want := mustRead(t, filepath.Join("..", "..", "testdata", "contracts", "envelope_v1.golden.json"))
	if string(got) != string(append(normalizeJSON(t, want), '\n')) {
		t.Fatalf("contract snapshot mismatch\nwant: %s\n got: %s", string(want), string(got))
	}
}

func TestRecordContractSnapshotV1(t *testing.T) {
	rec := Record{Locator: "x", Meta: map[string]any{"enabled": true, "name": "A"}}
	got := append(normalizeJSON(t, mustJSON(rec)), '\n')
	want := mustRead(t, filepath.Join("..", "..", "testdata", "contracts", "record_v1.golden.json"))
	if string(got) != string(append(normalizeJSON(t, want), '\n')) {
		t.Fatalf("record snapshot mismatch\nwant: %s\n got: %s", string(want), string(got))
	}
}
