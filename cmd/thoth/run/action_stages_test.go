package run

import (
	"testing"

	"github.com/flarebyte/thoth-ostraca/internal/stage"
)

func TestPreparedActionStagesAndFlags(t *testing.T) {
	meta := &stage.Meta{
		FileInfo: &stage.FileInfoMeta{Enabled: true},
		Git:      &stage.GitMeta{Enabled: true},
		PersistMeta: &stage.PersistMetaMeta{
			Enabled: true,
		},
		Lua: &stage.LuaMeta{ReduceInline: "x", FilterInline: "y"},
	}
	cases := []string{"pipeline", "nop", "input-pipeline", "validate", "create-meta", "update-meta", "diff-meta"}
	for _, c := range cases {
		stages, err := PreparedActionStages(c, meta)
		if err != nil {
			t.Fatalf("PreparedActionStages(%s) err: %v", c, err)
		}
		if len(stages) == 0 {
			t.Fatalf("PreparedActionStages(%s) empty", c)
		}
	}
	if _, err := PreparedActionStages("bad", meta); err == nil {
		t.Fatalf("expected invalid action error")
	}
	if !fileInfoEnabled(meta) || !gitEnabled(meta) || !persistMetaEnabled(meta) || !reduceEnabled(meta) || !filterEnabled(meta) {
		t.Fatalf("expected enabled helpers true")
	}
	if fileInfoEnabled(nil) || gitEnabled(nil) || persistMetaEnabled(nil) || reduceEnabled(nil) || filterEnabled(nil) {
		t.Fatalf("expected nil helpers false")
	}
}
