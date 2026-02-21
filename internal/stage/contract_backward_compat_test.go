package stage

import (
	"encoding/json"
	"os"
	"testing"
)

func TestContractBackwardCompatSnapshots(t *testing.T) {
	files := []string{
		"pipeline.golden.json",
		"validate.golden.json",
		"create-meta.golden.json",
		"update-meta.golden.json",
		"diff-meta.golden.json",
	}
	for _, name := range files {
		b, err := os.ReadFile(testdataPath("contracts", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var env Envelope
		if err := json.Unmarshal(b, &env); err != nil {
			t.Fatalf("unmarshal %s: %v", name, err)
		}
		if env.Meta == nil || env.Meta.ContractVersion != "1" {
			t.Fatalf("%s: meta.contractVersion must be 1", name)
		}
		if err := ValidateEnvelopeShape(env); err != nil {
			t.Fatalf("%s: shape: %v", name, err)
		}
	}
}
