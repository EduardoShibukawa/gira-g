package extractors

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// GoModExtractor extracts module declarations and library dependencies from go.mod files.
type GoModExtractor struct{}

// NewGoModExtractor creates a new Layer 2 go.mod extractor.
func NewGoModExtractor() *GoModExtractor {
	return &GoModExtractor{}
}

func (e *GoModExtractor) Name() string {
	return "Layer2GoMod"
}

func (e *GoModExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches go.mod files by filename.
func (e *GoModExtractor) CanExtract(path string) bool {
	return filepath.Base(path) == "go.mod"
}

// Extract parses the go.mod file and emits Library nodes for direct dependencies.
func (e *GoModExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	scanner := bufio.NewScanner(bytes.NewReader(file.Content))

	var modulePath string
	var goVersion string

	type depEntry struct {
		pkg       string
		version   string
		startLine int
		endLine   int
		indirect  bool
	}

	var deps []depEntry

	inRequireBlock := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		select {
		case <-ctx.Done():
			return nil, nil, nil, ctx.Err()
		default:
		}

		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		if trimmed == "" {
			continue
		}

		// Check for block termination
		if inRequireBlock {
			if strings.HasPrefix(trimmed, ")") {
				inRequireBlock = false
				continue
			}
		}

		// Handle comments outside of inline comments
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Module directive: module <path>
		if strings.HasPrefix(trimmed, "module ") || trimmed == "module" {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				modulePath = fields[1]
			}
			continue
		}

		// Go directive: go <version>
		if strings.HasPrefix(trimmed, "go ") || trimmed == "go" {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				goVersion = fields[1]
			}
			continue
		}

		// Require block start: require (
		if strings.HasPrefix(trimmed, "require (") || trimmed == "require (" {
			inRequireBlock = true
			continue
		}

		// Single line require: require <pkg> <version>
		if strings.HasPrefix(trimmed, "require ") {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "require "))
			if strings.HasPrefix(rest, "(") {
				inRequireBlock = true
				continue
			}
			fields := strings.Fields(rest)
			if len(fields) >= 2 {
				pkg := fields[0]
				version := fields[1]
				isIndirect := strings.Contains(rawLine, "// indirect")
				deps = append(deps, depEntry{
					pkg:       pkg,
					version:   version,
					startLine: lineNum,
					endLine:   lineNum,
					indirect:  isIndirect,
				})
			}
			continue
		}

		// Inside require block: <pkg> <version> [// indirect]
		if inRequireBlock {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				pkg := fields[0]
				version := fields[1]
				isIndirect := strings.Contains(rawLine, "// indirect")
				deps = append(deps, depEntry{
					pkg:       pkg,
					version:   version,
					startLine: lineNum,
					endLine:   lineNum,
					indirect:  isIndirect,
				})
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to scan go.mod in %s: %w", file.Path, err)
	}

	discoveredAt := time.Now().UTC()

	var nodes []domain.Node

	for _, dep := range deps {
		// Only direct dependencies emit Library nodes
		if dep.indirect {
			continue
		}

		metadata := map[string]any{
			"version": dep.version,
		}
		if goVersion != "" {
			metadata["go_version"] = goVersion
		}
		if modulePath != "" {
			metadata["module"] = modulePath
		}

		nodeID := fmt.Sprintf("library:%s", dep.pkg)
		node := domain.Node{
			ID:       nodeID,
			Kind:     domain.NodeLibrary,
			Name:     dep.pkg,
			Metadata: metadata,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: dep.startLine,
						EndLine:   dep.endLine,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "gomod",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}

		nodes = append(nodes, node)
	}

	return nodes, nil, nil, nil
}
