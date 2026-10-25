package extractors

import (
	"context"
	"testing"
)

func TestTypeScriptASTExtractor(t *testing.T) {
	ext := NewTypeScriptASTExtractor()

	if !ext.CanExtract("src/guards.ts") {
		t.Fatalf("expected CanExtract to match src/guards.ts")
	}
	if !ext.CanExtract("src/components/App.tsx") {
		t.Fatalf("expected CanExtract to match src/components/App.tsx")
	}
	if ext.CanExtract("node_modules/foo/index.ts") {
		t.Fatalf("expected CanExtract to ignore node_modules")
	}
	if ext.CanExtract("src/types.d.ts") {
		t.Fatalf("expected CanExtract to ignore .d.ts")
	}

	tsSource := `import { describe, expect } from "bun:test";
import { isProvisionalTest } from "./guards";

export interface GuardInput {
    task: string;
    touch?: string[];
    changed: string[];
}

export type GuardResult = {
    ok: boolean;
    reason?: string;
};

export class GuardEvaluator {
    private version: string;

    constructor(version: string) {
        this.version = version;
    }

    public evaluate(input: GuardInput): GuardResult {
        return { ok: true };
    }

    async reset(): Promise<void> {}
}

export function isProvisionalTest(path: string): boolean {
    return path.includes("scratch_");
}

export const checkStatus = (svc: string): boolean => {
    return svc !== "";
};
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc1234",
		Path:      "src/guards.ts",
		Content:   []byte(tsSource),
	}

	nodes, _, _, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	// Should extract Interface GuardInput, Type GuardResult, Class GuardEvaluator, plus Module Library node
	if len(nodes) < 3 {
		t.Fatalf("expected at least 3 nodes, got %d", len(nodes))
	}

	var foundClass, foundIface, foundType, foundLib bool
	for _, n := range nodes {
		switch n.Name {
		case "GuardEvaluator":
			foundClass = true
			methods, ok := n.Metadata["methods"].([]map[string]any)
			if !ok || len(methods) != 3 {
				t.Errorf("expected 3 methods on GuardEvaluator, got %v", n.Metadata["methods"])
			}
		case "GuardInput":
			foundIface = true
		case "GuardResult":
			foundType = true
		case "guards":
			foundLib = true
			funcs, ok := n.Metadata["functions"].([]map[string]any)
			if !ok || len(funcs) != 2 {
				t.Errorf("expected 2 top-level functions in guards module, got %v", n.Metadata["functions"])
			}
		}
	}

	if !foundClass {
		t.Errorf("expected class GuardEvaluator to be extracted")
	}
	if !foundIface {
		t.Errorf("expected interface GuardInput to be extracted")
	}
	if !foundType {
		t.Errorf("expected type GuardResult to be extracted")
	}
	if !foundLib {
		t.Errorf("expected library module guards to be extracted")
	}
}
