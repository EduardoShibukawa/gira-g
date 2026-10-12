package extractors

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"labs/gira-g/internal/domain"
)

// DataFormatExtractor handles Layer 1 syntax and format validation.
type DataFormatExtractor struct{}

// NewDataFormatExtractor creates a new Layer 1 data format extractor.
func NewDataFormatExtractor() *DataFormatExtractor {
	return &DataFormatExtractor{}
}

func (e *DataFormatExtractor) Name() string {
	return "Layer1DataFormat"
}

func (e *DataFormatExtractor) Layer() Layer {
	return LayerDataFormat
}

// CanExtract checks if the file is a generic structured format handled in Layer 1.
func (e *DataFormatExtractor) CanExtract(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".yaml" || ext == ".yml"
}

// Extract parses and validates the syntax of the JSON or YAML file.
// If valid, syntax validation succeeds without producing domain nodes.
// If invalid, an error is returned describing the malformed format.
func (e *DataFormatExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	ext := strings.ToLower(filepath.Ext(file.Path))
	switch ext {
	case ".json":
		var v any
		if err := json.Unmarshal(file.Content, &v); err != nil {
			return nil, nil, nil, fmt.Errorf("malformed JSON in %s: %w", file.Path, err)
		}
	case ".yaml", ".yml":
		var v any
		if err := yaml.Unmarshal(file.Content, &v); err != nil {
			return nil, nil, nil, fmt.Errorf("malformed YAML in %s: %w", file.Path, err)
		}
	}

	return nil, nil, nil, nil
}
