package extractors

import (
	"context"
	"strings"
	"testing"
)

func TestPipelineSkipsMalformedDataFile(t *testing.T) {
	pipe := NewPipeline(NewDataFormatExtractor())
	mm, err := pipe.ProcessFiles(context.Background(), "sys", []TargetFile{
		{Repo: "r", CommitSHA: "c", Path: "tasks/0021/evidence/e1.json", Content: []byte("{\"a\":1}\n{\"a\":2}\n")},
		{Repo: "r", CommitSHA: "c", Path: "valid.json", Content: []byte(`{"ok":true}`)},
	})
	if err != nil {
		t.Fatalf("index must not abort on one malformed data file: %v", err)
	}
	found := false
	for _, u := range mm.Unsupported {
		if u.Path == "valid.json" {
			t.Fatalf("valid file must not be marked unsupported")
		}
		if u.Path == "tasks/0021/evidence/e1.json" && strings.Contains(u.Reason, "malformed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected UnsupportedBoundary diagnostic for malformed file, got %+v", mm.Unsupported)
	}
}
