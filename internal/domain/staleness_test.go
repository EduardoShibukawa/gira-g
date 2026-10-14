package domain

import "testing"

// SC3: the index reports the commits it was built from and can tell whether a
// local HEAD is one of them.
func TestSCStaleness(t *testing.T) {
	mm := &MacroMap{
		Nodes: []Node{{Evidence: []EvidenceRecord{
			{Locator: Locator{Repo: "r", CommitSHA: "aaa", Path: "x"}},
		}}},
		Relationships: []Relationship{{Evidence: []EvidenceRecord{
			{Locator: Locator{Repo: "r2", CommitSHA: "bbb", Path: "y"}},
		}}},
	}
	got := mm.IndexedCommits()
	if len(got) != 2 || got[0] != "r2@bbb" || got[1] != "r@aaa" {
		t.Fatalf("IndexedCommits = %v", got)
	}
	if !mm.HasCommit("aaa") || !mm.HasCommit("bbb") {
		t.Fatal("HasCommit must find an indexed commit")
	}
	if mm.HasCommit("zzz") || mm.HasCommit("") {
		t.Fatal("HasCommit must reject unknown/empty shas")
	}
}
