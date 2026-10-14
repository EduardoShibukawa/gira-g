package extractors

import (
	"bufio"
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// EnvFileExtractor reads a Compose project's .env file. It emits no facts of its
// own: each KEY=VALUE line is a resolution hint that compose interpolation uses
// (ADR 0013).
type EnvFileExtractor struct{}

// NewEnvFileExtractor creates a .env extractor.
func NewEnvFileExtractor() *EnvFileExtractor { return &EnvFileExtractor{} }

func (e *EnvFileExtractor) Name() string { return "Layer2EnvFile" }

func (e *EnvFileExtractor) Layer() Layer { return LayerToolMeaning }

// CanExtract matches the Compose project environment file.
func (e *EnvFileExtractor) CanExtract(path string) bool { return filepath.Base(path) == ".env" }

// Extract emits nothing: the file's meaning is only in its hints.
func (e *EnvFileExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	return nil, nil, nil, nil
}

// Hints returns one env_value hint per variable, with the value as Compose reads
// it: optional `export`, quotes removed, and an unquoted value's trailing
// ` # comment` stripped.
func (e *EnvFileExtractor) Hints(ctx context.Context, file TargetFile) ([]Hint, error) {
	var hints []Hint
	sc := bufio.NewScanner(bytes.NewReader(file.Content))
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		text = strings.TrimPrefix(text, "export ")
		key, value, ok := strings.Cut(text, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		switch {
		case len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0]:
			value = value[1 : len(value)-1]
		default:
			if i := strings.Index(value, " #"); i >= 0 {
				value = strings.TrimSpace(value[:i])
			}
		}
		hints = append(hints, Hint{
			Kind:    HintEnvValue,
			Locator: domain.Locator{Repo: file.Repo, CommitSHA: file.CommitSHA, Path: file.Path, StartLine: line, EndLine: line},
			Key:     key,
			Value:   value,
		})
	}
	return hints, sc.Err()
}
