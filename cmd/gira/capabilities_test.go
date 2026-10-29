package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Frozen test for PLAN-0009 TASK-03: `gira capabilities --json` declares the
// context-provider protocol capabilities gira answers.

func TestCapabilitiesJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"capabilities", "--json"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("capabilities --json returned error: %v (stderr: %s)", err, stderr.String())
	}

	var doc struct {
		SchemaVersion int      `json:"schema_version"`
		Capabilities  []string `json:"capabilities"`
		Extras        []string `json:"extras"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatalf("capabilities --json output is not valid JSON: %v (got %q)", err, stdout.String())
	}

	if doc.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", doc.SchemaVersion)
	}
	wantCaps := []string{"outline", "slice", "impact", "read"}
	if !reflect.DeepEqual(doc.Capabilities, wantCaps) {
		t.Errorf("capabilities = %v, want %v", doc.Capabilities, wantCaps)
	}
	wantExtras := []string{"gira:tour", "gira:docs"}
	if !reflect.DeepEqual(doc.Extras, wantExtras) {
		t.Errorf("extras = %v, want %v", doc.Extras, wantExtras)
	}
}
