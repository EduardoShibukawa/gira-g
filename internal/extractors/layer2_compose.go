package extractors

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// ComposeExtractor extracts architectural nodes and dependencies from Docker Compose files.
type ComposeExtractor struct{}

// NewComposeExtractor creates a new Docker Compose extractor.
func NewComposeExtractor() *ComposeExtractor {
	return &ComposeExtractor{}
}

func (e *ComposeExtractor) Name() string {
	return "Layer2DockerCompose"
}

func (e *ComposeExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches docker-compose and compose YAML files.
func (e *ComposeExtractor) CanExtract(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == "docker-compose.yaml" || base == "docker-compose.yml" ||
		base == "compose.yaml" || base == "compose.yml" ||
		strings.HasPrefix(base, "docker-compose.") || strings.HasPrefix(base, "compose.")
}

// Known infrastructure names or substring patterns.
var infraKeywords = []string{
	"postgres", "postgresql", "redis", "valkey", "kafka", "zookeeper",
	"jaeger", "otelcol", "otel-collector", "opentelemetry-collector",
	"prometheus", "grafana", "flagd", "elasticsearch", "rabbitmq",
	"mysql", "mariadb", "mongo", "mongodb", "cassandra", "memcached",
	"envoy", "nginx", "consul", "vault",
}

func isInfrastructure(serviceName string, image string) bool {
	s := strings.ToLower(serviceName)
	img := strings.ToLower(image)
	for _, kw := range infraKeywords {
		if strings.Contains(s, kw) || strings.Contains(img, kw) {
			return true
		}
	}
	return false
}

// findServiceLine attempts to locate the line where a service key is defined.
func findServiceLine(content string, svcName string) int {
	lines := strings.Split(content, "\n")
	svcRegex := regexp.MustCompile(`^\s{2,4}` + regexp.QuoteMeta(svcName) + `\s*:`)
	for idx, line := range lines {
		if svcRegex.MatchString(line) {
			return idx + 1
		}
	}
	// Fallback: look for svcName:
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, svcName+":") || trimmed == svcName+":" {
			return idx + 1
		}
	}
	return 1
}

// Extract parses the compose file and generates Service/Infrastructure nodes and relationships.
func (e *ComposeExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(file.Content, &root); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse compose YAML in %s: %w", file.Path, err)
	}

	contentStr := string(file.Content)
	profile := fmt.Sprintf("compose:%s", filepath.Base(file.Path))

	var rawData struct {
		Services map[string]struct {
			Image       string   `yaml:"image"`
			DependsOn   any      `yaml:"depends_on"`
			Ports       []any    `yaml:"ports"`
			Environment any      `yaml:"environment"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(file.Content, &rawData); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to map compose structures in %s: %w", file.Path, err)
	}

	if len(rawData.Services) == 0 {
		return nil, nil, nil, nil
	}

	nodes := make([]domain.Node, 0, len(rawData.Services))
	relationships := make([]domain.Relationship, 0)
	nodeMap := make(map[string]domain.Node)

	discoveredAt := time.Now().UTC()

	for svcName, svcDef := range rawData.Services {
		nodeKind := domain.NodeService
		if isInfrastructure(svcName, svcDef.Image) {
			nodeKind = domain.NodeInfrastructure
		}

		nodeID := fmt.Sprintf("service:%s", svcName)
		line := findServiceLine(contentStr, svcName)

		node := domain.Node{
			ID:   nodeID,
			Kind: nodeKind,
			Name: svcName,
			Metadata: map[string]any{
				"image": svcDef.Image,
			},
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: line,
						EndLine:   line,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: profile,
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}

		nodes = append(nodes, node)
		nodeMap[svcName] = node
	}

	// Process dependencies
	for svcName, svcDef := range rawData.Services {
		var depList []string

		switch v := svcDef.DependsOn.(type) {
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					depList = append(depList, s)
				}
			}
		case map[string]any:
			for dep := range v {
				depList = append(depList, dep)
			}
		}

		for _, dep := range depList {
			sourceID := fmt.Sprintf("service:%s", svcName)
			targetID := fmt.Sprintf("service:%s", dep)

			targetNode, targetExists := nodeMap[dep]
			var verb domain.RelationVerb = domain.VerbConsumes
			if targetExists && targetNode.Kind == domain.NodeInfrastructure {
				verb = domain.VerbAccesses
			}

			line := findServiceLine(contentStr, svcName)

			rel := domain.Relationship{
				ID:       fmt.Sprintf("rel:%s-%s-%s", svcName, string(verb), dep),
				SourceID: sourceID,
				TargetID: targetID,
				Verb:     verb,
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      file.Repo,
							CommitSHA: file.CommitSHA,
							Path:      file.Path,
							StartLine: line,
							EndLine:   line,
						},
						Origin:            domain.OriginExtracted,
						DeploymentProfile: profile,
						Status:            domain.StatusCurrent,
						DiscoveredAt:      discoveredAt,
					},
				},
			}
			relationships = append(relationships, rel)
		}
	}

	return nodes, relationships, nil, nil
}

// Hints returns, for every compose service with a build section, its
// build.dockerfile and build.context before interpolation, located at the
// dockerfile line (or the build line when only a context is given).
func (e *ComposeExtractor) Hints(ctx context.Context, file TargetFile) ([]Hint, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(file.Content, &root); err != nil || len(root.Content) == 0 {
		return nil, nil
	}
	services := mappingValue(root.Content[0], "services")
	if services == nil || services.Kind != yaml.MappingNode {
		return nil, nil
	}
	var hints []Hint
	for i := 0; i+1 < len(services.Content); i += 2 {
		name, def := services.Content[i].Value, services.Content[i+1]
		build := mappingValue(def, "build")
		if build == nil {
			continue
		}
		h := Hint{Kind: HintComposeBuild, Service: "service:" + name, Value: "Dockerfile", Context: "."}
		line := build.Line
		switch build.Kind {
		case yaml.ScalarNode:
			h.Context = build.Value
		case yaml.MappingNode:
			if c := mappingValue(build, "context"); c != nil {
				h.Context = c.Value
			}
			if d := mappingValue(build, "dockerfile"); d != nil {
				h.Value, line = d.Value, d.Line
			}
		}
		h.Locator = domain.Locator{Repo: file.Repo, CommitSHA: file.CommitSHA, Path: file.Path, StartLine: line, EndLine: line}
		hints = append(hints, h)
	}
	return hints, nil
}

// mappingValue returns the value node of key in a YAML mapping, or nil.
func mappingValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}
