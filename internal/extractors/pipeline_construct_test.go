package extractors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// F7: the pipeline separates an unsupported construct inside a parsed Go file
// from an unsupported file, keeps both rows in Unsupported, and the split
// survives the canonical JSON round trip used by index files.
func TestPipelineSeparatesConstructFromFile(t *testing.T) {
	files := []TargetFile{
		{Repo: "r", CommitSHA: "c", Path: "src/catalog/main.go", Content: []byte("package main\n\nfunc main() {\n\tstartOpAMPClient()\n}\n")},
		{Repo: "r", CommitSHA: "c", Path: "src/worker/app.py", Content: []byte("print('hi')\n")},
	}
	mm, err := DefaultPipeline().ProcessFiles(context.Background(), "t", files)
	if err != nil {
		t.Fatal(err)
	}
	data, err := mm.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var back domain.MacroMap
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	for _, m := range []*domain.MacroMap{mm, &back} {
		if len(m.Unsupported) != 2 {
			t.Errorf("both rows stay in Unsupported, got %+v", m.Unsupported)
		}
		if !m.IsUnsupportedConstruct("src/catalog/main.go") || m.IsUnsupportedConstruct("src/worker/app.py") {
			t.Errorf("construct split wrong: %v", m.UnsupportedConstructs)
		}
		files := m.UnsupportedFiles()
		if len(files) != 1 || files[0].Path != "src/worker/app.py" {
			t.Errorf("UnsupportedFiles = %+v, want only src/worker/app.py", files)
		}
	}
}
