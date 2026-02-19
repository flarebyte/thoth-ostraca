package stage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/testutil"
)

func testdataPath(parts ...string) string {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		root = filepath.Join("..", "..")
	}
	base := append([]string{root, "testdata"}, parts...)
	return filepath.Join(base...)
}

func rootTempPath(parts ...string) string {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		root = filepath.Join("..", "..")
	}
	base := append([]string{root, "temp"}, parts...)
	return filepath.Join(base...)
}

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
	if err := validateTopLevel(g); err != nil {
		return err
	}
	if err := validateRecordsSection(g["records"]); err != nil {
		return err
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

func canonicalJSON(t *testing.T, b []byte) []byte {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return out
}

func assertJSONEqual(t *testing.T, actual, expected []byte) {
	t.Helper()
	a := canonicalJSON(t, actual)
	e := canonicalJSON(t, expected)
	if string(a) != string(e) {
		t.Fatalf("mismatch\nactual: %s\nexpected: %s", string(actual), string(expected))
	}
}

func runActionWithConfig(t *testing.T, cfgPath string) (Envelope, []byte) {
	t.Helper()
	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Join("..", "..")
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}
	defer func() {
		_ = os.Chdir(prevDir)
	}()

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
	stages, selErr := selectStages(action)
	if selErr != nil {
		t.Fatalf("%v", selErr)
	}
	cur, failedStage, runErr := runStagesTest(ctx, stages, out)
	if runErr != nil {
		t.Fatalf("stage %s: %v", failedStage, runErr)
	}
	b, err := encodeEnvelopeContract(cur)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return cur, b
}

func TestContract_Pipeline(t *testing.T) {
	_, b := runActionWithConfig(t, testdataPath("configs", "keep1_embed_true.cue"))
	g, _ := os.ReadFile(testdataPath("contracts", "pipeline.golden.json"))
	validateContractBytes(t, b)
	assertJSONEqual(t, b, g)
}

func TestContract_Validate(t *testing.T) {
	_, b := runActionWithConfig(t, testdataPath("configs", "validate_only_ok.cue"))
	g, _ := os.ReadFile(testdataPath("contracts", "validate.golden.json"))
	validateContractBytes(t, b)
	assertJSONEqual(t, b, g)
}

func validateContractBytes(t *testing.T, b []byte) {
	t.Helper()
	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
}

func TestContract_CreateMeta(t *testing.T) {
	repoAbs := rootTempPath("create1_repo_contract")
	repoRel := filepath.ToSlash(filepath.Join("temp", "create1_repo_contract"))
	if err := testutil.CopyTree(testdataPath("repos", "create1"), repoAbs); err != nil {
		t.Fatalf("copy: %v", err)
	}
	cfg := rootTempPath("create1_contract.cue")
	_ = os.MkdirAll(rootTempPath(), 0o755)
	if err := os.WriteFile(cfg, []byte("{\n  configVersion: \"v0\"\n  action: \"create-meta\"\n  discovery: { root: \""+repoRel+"\" }\n}\n"), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	_, b := runActionWithConfig(t, cfg)
	g, _ := os.ReadFile(testdataPath("contracts", "create-meta.golden.json"))
	var env Envelope
	_ = json.Unmarshal(b, &env)
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	assertJSONEqual(t, b, g)
}

func TestContract_UpdateMeta(t *testing.T) {
	repoAbs := rootTempPath("update1_repo_contract")
	repoRel := filepath.ToSlash(filepath.Join("temp", "update1_repo_contract"))
	if err := testutil.CopyTree(testdataPath("repos", "update1"), repoAbs); err != nil {
		t.Fatalf("copy: %v", err)
	}
	cfg := rootTempPath("update1_contract.cue")
	_ = os.MkdirAll(rootTempPath(), 0o755)
	if err := os.WriteFile(cfg, []byte("{\n  configVersion: \"v0\"\n  action: \"update-meta\"\n  discovery: { root: \""+repoRel+"\" }\n}\n"), 0o644); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	_, b := runActionWithConfig(t, cfg)
	g, _ := os.ReadFile(testdataPath("contracts", "update-meta.golden.json"))
	var env Envelope
	_ = json.Unmarshal(b, &env)
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	assertJSONEqual(t, b, g)
}

func TestContract_DiffMeta(t *testing.T) {
	_, b := runActionWithConfig(t, testdataPath("configs", "diff1.cue"))
	g, _ := os.ReadFile(testdataPath("contracts", "diff-meta.golden.json"))
	var env Envelope
	_ = json.Unmarshal(b, &env)
	if err := ValidateEnvelopeShape(env); err != nil {
		t.Fatalf("shape: %v", err)
	}
	assertJSONEqual(t, b, g)
}

func TestContract_Goldens_Schema(t *testing.T) {
	files := []string{"pipeline.golden.json", "validate.golden.json", "create-meta.golden.json", "update-meta.golden.json", "diff-meta.golden.json"}
	for _, f := range files {
		b, err := os.ReadFile(testdataPath("contracts", f))
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
