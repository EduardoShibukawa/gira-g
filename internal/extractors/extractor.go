package extractors

import (
	"context"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// Layer defines the 3-layer extractor classification.
type Layer int

const (
	LayerDataFormat  Layer = 1
	LayerToolMeaning Layer = 2
	LayerLanguageAST Layer = 3
)

// TargetFile describes a file candidate for extraction within a repository.
type TargetFile struct {
	Repo      string
	CommitSHA string
	Path      string
	Content   []byte
}

// Extractor defines the unified internal interface for all fact extractors.
type Extractor interface {
	Name() string
	Layer() Layer
	CanExtract(path string) bool
	Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error)
}
