package gitdiff

import (
	"reflect"
	"testing"
)

const sample = `diff --git a/pb/demo.proto b/pb/demo.proto
index 1111111..2222222 100644
--- a/pb/demo.proto
+++ b/pb/demo.proto
@@ -178,0 +179 @@ service PaymentService {
+    rpc Refund(ChargeRequest) returns (ChargeResponse) {}
@@ -200,2 +201,3 @@
-a
-b
+c
diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
@@ -0,0 +1 @@
+package x
diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1 +0,0 @@
-package y
`

func TestParse(t *testing.T) {
	got, err := Parse(sample)
	if err != nil {
		t.Fatal(err)
	}
	want := []FileChange{
		{Path: "new.go", Added: true, Hunks: []Hunk{{Start: 0, Count: 0}}},
		{Path: "old.go", Deleted: true, Hunks: []Hunk{{Start: 1, Count: 1}}},
		{Path: "pb/demo.proto", Hunks: []Hunk{{Start: 178, Count: 0}, {Start: 200, Count: 2}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestTouches(t *testing.T) {
	c := FileChange{Path: "p", Hunks: []Hunk{{Start: 178, Count: 0}, {Start: 200, Count: 2}}}
	for _, tc := range []struct {
		start, end int
		want       bool
	}{
		{177, 179, true},  // insertion between 178 and 179, inside the block
		{170, 178, false}, // insertion right after the block's last line
		{179, 190, false}, // insertion right before the block's first line
		{201, 205, true},  // overlaps the modified 200-201
		{202, 205, false},
		{0, 0, true}, // no line range: whole file
	} {
		if got := c.Touches(tc.start, tc.end); got != tc.want {
			t.Errorf("Touches(%d,%d) = %t, want %t", tc.start, tc.end, got, tc.want)
		}
	}
}

func TestRunRejectsOptionLikeRange(t *testing.T) {
	if _, err := Run(t.Context(), ".", "--output=/tmp/x..y"); err == nil {
		t.Fatal("an option-like range must be rejected")
	}
}
