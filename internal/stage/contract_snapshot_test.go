package stage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ValidateEnvelopeShape asserts a minimal, strict public JSON contract.
func ValidateEnvelopeShape(e Envelope) error {
	// contractVersion
	if e.Meta == nil || e.Meta.ContractVersion != "1" {
		return errors.New("meta.contractVersion must be '1'")
	}
	// Encode to generic to validate keys and types
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	var g map[string]any
	if err := json.Unmarshal(b, &g); err != nil {
		return err
	}
	// Allowed top-level keys
	allowedTop := map[string]bool{"records": true, "meta": true, "errors": true}
	for k := range g {
		if !allowedTop[k] {
			return errors.New("unexpected top-level key: " + k)
		}
	}
	// records array
	recs, ok := g["records"].([]any)
	if !ok {
		return errors.New("records must be array")
	}
	for _, it := range recs {
		m, ok := it.(map[string]any)
		if !ok {
			return errors.New("record must be object")
		}
		// allowed record keys
		allowedRec := map[string]bool{
			"locator": true, "meta": true, "mapped": true, "shell": true, "post": true, "error": true, "fileInfo": true, "git": true,
		}
		for k := range m {
			if !allowedRec[k] {
				return errors.New("unexpected record key: " + k)
			}
		}
		if loc, ok := m["locator"]; ok {
			if _, ok := loc.(string); !ok {
				return errors.New("record.locator must be string")
			}
		}
		if errv, hasErr := m["error"]; hasErr && errv != nil {
			em, ok := errv.(map[string]any)
			if !ok {
				return errors.New("record.error must be object")
			}
			if _, ok := em["stage"].(string); !ok {
				return errors.New("record.error.stage must be string")
			}
			if _, ok := em["message"].(string); !ok {
				return errors.New("record.error.message must be string")
			}
			if loc, ok := em["locator"]; ok && loc != nil {
				if _, ok := loc.(string); !ok {
					return errors.New("record.error.locator must be string")
				}
			}
			for k := range em {
				if k != "stage" && k != "message" && k != "locator" {
					return errors.New("unexpected error key: " + k)
				}
			}
		}
	}
	return nil
}

func encodeEnvelopeContract(env Envelope) ([]byte, error) {
	// Prepare like write-output (compact aggregate)
	if env.Meta == nil {
		env.Meta = &Meta{}
	}
	env.Meta.ContractVersion = "1"
	SortEnvelopeErrors(&env)
	return encodeJSONCompact(env)
}

func runActionWithConfig(t *testing.T, cfgPath string) (Envelope, []byte) {
	t.Helper()
	ctx := context.Background()
	in := Envelope{Records: []Record{}, Meta: &Meta{ConfigPath: cfgPath}}
	out, err := Run(ctx, "validate-config", in, Deps{})
	if err != nil {
		t.Fatalf("validate-config: %v", err)
	}
	action := "pipeline"
	if out.Meta != nil && out.Meta.Config != nil && out.Meta.Config.Action != "" {
		action = out.Meta.Config.Action
	}
	var stages []string
	switch action {
	case "pipeline", "nop":
		stages = []string{"discover-meta-files", "parse-validate-yaml", "validate-locators", "lua-filter", "lua-map", "shell-exec", "lua-postmap", "lua-reduce"}
	case "validate":
		stages = []string{"discover-meta-files", "parse-validate-yaml", "validate-locators"}
	case "create-meta":
		stages = []string{"discover-input-files", "enrich-fileinfo", "enrich-git", "write-meta-files"}
	case "update-meta":
		stages = []string{"discover-input-files", "enrich-fileinfo", "enrich-git", "load-existing-meta", "merge-meta", "write-updated-meta-files"}
	case "diff-meta":
		stages = []string{"discover-input-files", "enrich-fileinfo", "enrich-git", "discover-meta-files", "compute-meta-diff"}
	default:
		t.Fatalf("unknown action: %s", action)
	}
	cur := out
	for _, s := range stages {
		cur, err = Run(ctx, s, cur, Deps{})
		if err != nil {
			t.Fatalf("stage %s: %v", s, err)
		}
	}
	b, err := encodeEnvelopeContract(cur)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return cur, b
}

func TestContract_Pipeline(t *testing.T) {
	_, b := runActionWithConfig(t, filepath.Join("testdata", "configs", "keep1_embed_true.cue"))
	g, _ := os.ReadFile(filepath.Join("testdata", "contracts", "pipeline.golden.json"))
	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	if string(b) != string(g) {
		t.Fatalf("mismatch\nactual: %s\nexpected: %s", string(b), string(g))
	}
}

func TestContract_Validate(t *testing.T) {
	_, b := runActionWithConfig(t, filepath.Join("testdata", "configs", "validate_only_ok.cue"))
	g, _ := os.ReadFile(filepath.Join("testdata", "contracts", "validate.golden.json"))
	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	if string(b) != string(g) {
		t.Fatalf("mismatch\nactual: %s\nexpected: %s", string(b), string(g))
	}
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	_ = os.RemoveAll(dst)
	if err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		out := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	}); err != nil {
		t.Fatalf("copy: %v", err)
	}
}

func TestContract_CreateMeta(t *testing.T) {
	repo := filepath.Join("temp", "create1_repo_contract")
	copyTree(t, filepath.Join("testdata", "repos", "create1"), repo)
	cfg := filepath.Join("temp", "create1_contract.cue")
	_ = os.MkdirAll("temp", 0o755)
	if err := os.WriteFile(cfg, []byte("{\n  configVersion: \"v0\"\n  action: \"create-meta\"\n  discovery: { root: \""+filepath.ToSlash(repo)+"\" }\n}\n"), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	_, b := runActionWithConfig(t, cfg)
	g, _ := os.ReadFile(filepath.Join("testdata", "contracts", "create-meta.golden.json"))
	var env Envelope
	_ = json.Unmarshal(b, &env)
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	if string(b) != string(g) {
		t.Fatalf("mismatch\nactual: %s\nexpected: %s", string(b), string(g))
	}
}

func TestContract_UpdateMeta(t *testing.T) {
	repo := filepath.Join("temp", "update1_repo_contract")
	copyTree(t, filepath.Join("testdata", "repos", "update1"), repo)
	cfg := filepath.Join("temp", "update1_contract.cue")
	_ = os.MkdirAll("temp", 0o755)
	if err := os.WriteFile(cfg, []byte("{\n  configVersion: \"v0\"\n  action: \"update-meta\"\n  discovery: { root: \""+filepath.ToSlash(repo)+"\" }\n}\n"), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	_, b := runActionWithConfig(t, cfg)
	g, _ := os.ReadFile(filepath.Join("testdata", "contracts", "update-meta.golden.json"))
	var env Envelope
	_ = json.Unmarshal(b, &env)
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	if string(b) != string(g) {
		t.Fatalf("mismatch\nactual: %s\nexpected: %s", string(b), string(g))
	}
}

func TestContract_DiffMeta(t *testing.T) {
	_, b := runActionWithConfig(t, filepath.Join("testdata", "configs", "diff1.cue"))
	g, _ := os.ReadFile(filepath.Join("testdata", "contracts", "diff-meta.golden.json"))
	var env Envelope
	_ = json.Unmarshal(b, &env)
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	if string(b) != string(g) {
		t.Fatalf("mismatch\nactual: %s\nexpected: %s", string(b), string(g))
	}
}

func TestContract_Goldens_Schema(t *testing.T) {
	files := []string{"pipeline.golden.json", "validate.golden.json", "create-meta.golden.json", "update-meta.golden.json", "diff-meta.golden.json"}
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join("testdata", "contracts", f))
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var env Envelope
		if err := json.Unmarshal(b, &env); err != nil {
			t.Fatalf("unmarshal %s: %v", f, err)
		}
		if err := ValidateEnvelopeShape(env); err != nil {
			t.Fatalf("shape %s: %v", f, err)
		}
	}
}
