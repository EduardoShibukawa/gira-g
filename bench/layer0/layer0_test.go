package layer0

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lifter(t *testing.T) Lifter {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "components.json"))
	if err != nil {
		t.Fatal(err)
	}
	var table []ComponentPath
	if err := json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	return NewLifter(table)
}

// SC13: each adapter maps a synthetic sample, written to the tool's documented
// schema, to the hand-written expected NAG.
func TestAdaptersGolden(t *testing.T) {
	l := lifter(t)
	for _, tc := range []struct {
		dir, input string
		adapt      func([]byte) (*NAG, error)
	}{
		{"gira", "macro-map.json", FromGIRA},
		{"graphify", "graph.json", func(b []byte) (*NAG, error) { return FromGraphify(b, l) }},
		{"ua", "knowledge-graph.json", func(b []byte) (*NAG, error) { return FromUA(b, l) }},
	} {
		t.Run(tc.dir, func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join("testdata", tc.dir, tc.input))
			if err != nil {
				t.Fatal(err)
			}
			nag, err := tc.adapt(in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := nag.CanonicalJSON()
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("testdata", tc.dir, "expected.nag.json"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("NAG differs from golden\n got:\n%s\nwant:\n%s", got, want)
			}
			again, _ := tc.adapt(in)
			b, _ := again.CanonicalJSON()
			if !bytes.Equal(b, got) {
				t.Fatal("adapter is not deterministic")
			}
		})
	}
}

// AC1, AC2: FromGIRA adapter correctly processes Schema 2 MacroMaps with embedded source_identity.
func TestFromGIRASchema2(t *testing.T) {
	in, err := os.ReadFile(filepath.Join("testdata", "gira", "macro-map.json"))
	if err != nil {
		t.Fatal(err)
	}

	// Transform golden sample into Schema 2 with embedded source_identity
	var raw map[string]any
	if err := json.Unmarshal(in, &raw); err != nil {
		t.Fatal(err)
	}
	raw["schema_version"] = 2
	raw["source_identity"] = map[string]any{
		"schema_version":  2,
		"manifest_sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"extractors": []map[string]any{
			{"id": "go", "version": "1.0.0"},
		},
		"sources": []map[string]any{
			{
				"repo_id":    "opentelemetry-demo",
				"sha":        "dedc0178918e260823323b8d95005a8cb924b007",
				"path_scope": ".",
				"url_sha256": "c83c589000000000000000000000000000000000000000000000000000000000",
			},
		},
	}
	schema2Bytes, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}

	nag, err := FromGIRA(schema2Bytes)
	if err != nil {
		t.Fatalf("FromGIRA failed on Schema 2 index: %v", err)
	}
	got, err := nag.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.Join("testdata", "gira", "expected.nag.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Schema 2 NAG differs from golden\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestLifterLongestPrefix(t *testing.T) {
	l := NewLifter([]ComponentPath{{"src/frontend/", "service:frontend"}, {"src/frontend-proxy/", "service:frontend-proxy"}})
	for p, want := range map[string]string{
		"src/frontend-proxy/envoy.yaml": "service:frontend-proxy",
		"./src/frontend/app.ts":         "service:frontend",
		"pb/demo.proto":                 "",
	} {
		if got := l.Component(p); got != want {
			t.Errorf("Component(%q) = %q, want %q", p, got, want)
		}
	}
}

// SC13: the locator checker accepts valid locators and rejects missing files,
// out-of-bounds ranges and ranges that do not contain the anchor.
func TestLocatorChecker(t *testing.T) {
	files := map[string]string{"src/checkout/main.go": "package main\n\nimport \"fmt\"\n\nfunc PlaceOrder() {\n\tfmt.Println(\"order\")\n}\n"}
	read := func(p string) ([]byte, error) {
		if c, ok := files[p]; ok {
			return []byte(c), nil
		}
		return nil, os.ErrNotExist
	}
	for _, tc := range []struct {
		ev     Evidence
		anchor string
		want   string // "" = accepted
	}{
		{Evidence{"src/checkout/main.go", 5, 7}, "func PlaceOrder", ""},
		{Evidence{"src/checkout/main.go", 1, 1}, "", ""},
		{Evidence{"src/checkout/main.go", 7, 9}, "", "out of bounds"},
		{Evidence{"src/checkout/main.go", 0, 2}, "", "out of bounds"},
		{Evidence{"src/checkout/main.go", 3, 2}, "", "out of bounds"},
		{Evidence{"src/checkout/main.go", 1, 3}, "func PlaceOrder", "does not contain anchor"},
		{Evidence{"src/checkout/missing.go", 1, 1}, "", "does not exist"},
	} {
		err := CheckLocator(read, tc.ev, tc.anchor)
		switch {
		case tc.want == "" && err != nil:
			t.Errorf("%+v rejected: %v", tc.ev, err)
		case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
			t.Errorf("%+v: got %v, want %q", tc.ev, err, tc.want)
		}
	}
}

// The lifting table is the ground truth's, shared by every tool.
func TestGroundTruthComponentTable(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "fixtures", "evaluation", "reference-corpus", "ground-truth.v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	l, err := LoadComponentTable(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := l.Component("src/checkout/kafka/producer.go"); got != "service:checkout" {
		t.Fatalf("checkout producer lifted to %q", got)
	}
	if got := l.Component("pb/demo.proto"); got != "" {
		t.Fatalf("the shared proto belongs to no Component, got %q", got)
	}
}
