package jev_test

import (
	"context"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/jev"
)

func TestClassifyIntent(t *testing.T) {
	classifier := jev.NewLocalClassifier()

	tests := []struct {
		query          string
		expectedIntent jev.IntentType
		wantDiagram    bool
	}{
		{"Qual é o objetivo principal deste projeto?", jev.IntentPurpose, false},
		{"What is the purpose of this codebase?", jev.IntentPurpose, false},
		{"Como eu utilizo o sistema pela primeira vez?", jev.IntentQuickstart, false},
		{"How do I run the setup?", jev.IntentQuickstart, false},
		{"Como inicializo o projeto com diagrama?", jev.IntentQuickstart, true},
		{"O que quebra se eu alterar o pacote de billing?", jev.IntentImpact, false},
		{"Qual o impacto de modificar a função ProcessPayment?", jev.IntentImpact, false},
		{"Mostre o fluxo de arquitetura", jev.IntentComponentDetail, true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			res := classifier.ClassifyIntent(context.Background(), tt.query)
			if res.Intent != tt.expectedIntent {
				t.Errorf("for query %q, expected intent %q, got %q", tt.query, tt.expectedIntent, res.Intent)
			}
			if res.WantDiagram != tt.wantDiagram {
				t.Errorf("for query %q, expected wantDiagram %v, got %v", tt.query, tt.wantDiagram, res.WantDiagram)
			}
			if res.Confidence < 0.70 {
				t.Errorf("expected confidence >= 0.70, got %f", res.Confidence)
			}
		})
	}
}

func TestClassifyDocument(t *testing.T) {
	classifier := jev.NewLocalClassifier()

	tests := []struct {
		path         string
		content      string
		expectedRole jev.DocumentRole
	}{
		{"docs/adr/001-hexagonal.md", "## Decision\nWe will enforce hexagonal architecture.", jev.RolePolicy},
		{"docs/decisions/adr-auth.md", "Status: Accepted\nContext: JWT tokens.", jev.RolePolicy},
		{"tasks/T30/PLAN.md", "done_when: bun test\n## Passos\n- [ ] RED", jev.RoleTask},
		{".scratch/features/issues/task-01.md", "Status: Pending\nBlocked by: none", jev.RoleTask},
		{"specs/api/payments.proto", "service PaymentService {}", jev.RoleContract},
		{"README.md", "# Gira-G Architecture Harness\nWelcome to the docs.", jev.RoleDocumentation},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			res := classifier.ClassifyDocument(context.Background(), tt.path, []byte(tt.content))
			if res.Role != tt.expectedRole {
				t.Errorf("for path %q, expected role %q, got %q", tt.path, tt.expectedRole, res.Role)
			}
		})
	}
}

func TestClassifyDrift(t *testing.T) {
	classifier := jev.NewLocalClassifier()

	t.Run("zero diff", func(t *testing.T) {
		res := classifier.ClassifyDrift(context.Background(), "spec", "")
		if res.Severity != jev.SeverityNone {
			t.Errorf("expected none, got %s", res.Severity)
		}
	})

	t.Run("cosmetic comments", func(t *testing.T) {
		res := classifier.ClassifyDrift(context.Background(), "spec", "+ // updated comment")
		if res.Severity != jev.SeverityCosmetic {
			t.Errorf("expected cosmetic, got %s", res.Severity)
		}
	})

	t.Run("breaking forbidden import", func(t *testing.T) {
		res := classifier.ClassifyDrift(context.Background(), "Architecture policy: UI package cannot import database directly", "+import \"github.com/db/sql\"")
		if res.Severity != jev.SeverityBreaking {
			t.Errorf("expected breaking, got %s", res.Severity)
		}
	})
}

func TestRunnerClassifierFallback(t *testing.T) {
	local := jev.NewLocalClassifier()
	runner := jev.NewRunnerClassifier(local)

	res := runner.ClassifyIntent(context.Background(), "Qual o objetivo do sistema?")
	if res.Intent != jev.IntentPurpose {
		t.Errorf("expected IntentPurpose, got %s", res.Intent)
	}
	if res.Engine == "" {
		t.Errorf("expected engine identifier, got empty")
	}
}
