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

// HelmExtractor extracts services, components, and dependencies from Chart.yaml and values.yaml files.
type HelmExtractor struct{}

// NewHelmExtractor creates a new Helm values/chart extractor.
func NewHelmExtractor() *HelmExtractor {
	return &HelmExtractor{}
}

func (e *HelmExtractor) Name() string {
	return "Layer2Helm"
}

func (e *HelmExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches Chart.yaml, Chart.yml, values.yaml, and values.yml (or values-*.yaml).
func (e *HelmExtractor) CanExtract(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if base == "chart.yaml" || base == "chart.yml" {
		return true
	}
	if base == "values.yaml" || base == "values.yml" || strings.HasPrefix(base, "values-") || strings.HasPrefix(base, "values.") {
		return true
	}
	return false
}

// findYamlKeyLine attempts to locate the line where a top-level or sub-level key is defined.
func findYamlKeyLine(content string, key string) int {
	lines := strings.Split(content, "\n")
	re := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(key) + `\s*:`)
	for idx, line := range lines {
		if re.MatchString(line) {
			return idx + 1
		}
	}
	return 1
}

// Extract parses Chart.yaml or values.yaml and extracts domain entities.
func (e *HelmExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	base := strings.ToLower(filepath.Base(file.Path))
	discoveredAt := time.Now().UTC()

	if base == "chart.yaml" || base == "chart.yml" {
		return e.extractChart(file, discoveredAt)
	}

	return e.extractValues(file, discoveredAt)
}

func (e *HelmExtractor) extractChart(file TargetFile, discoveredAt time.Time) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	var chart struct {
		Name         string `yaml:"name"`
		Version      string `yaml:"version"`
		Description  string `yaml:"description"`
		Dependencies []struct {
			Name       string `yaml:"name"`
			Version    string `yaml:"version"`
			Repository string `yaml:"repository"`
			Condition  string `yaml:"condition"`
		} `yaml:"dependencies"`
	}

	if err := yaml.Unmarshal(file.Content, &chart); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse Chart.yaml in %s: %w", file.Path, err)
	}

	if chart.Name == "" {
		return nil, nil, nil, nil
	}

	chartNodeID := fmt.Sprintf("service:%s", chart.Name)
	chartNode := domain.Node{
		ID:          chartNodeID,
		Kind:        domain.NodeService,
		Name:        chart.Name,
		Description: chart.Description,
		Metadata: map[string]any{
			"version": chart.Version,
		},
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      file.Repo,
					CommitSHA: file.CommitSHA,
					Path:      file.Path,
					StartLine: 1,
					EndLine:   1,
				},
				Origin:            domain.OriginDeclared,
				DeploymentProfile: "helm:chart",
				Status:            domain.StatusCurrent,
				DiscoveredAt:      discoveredAt,
			},
		},
	}

	nodes := []domain.Node{chartNode}
	relationships := make([]domain.Relationship, 0)
	contentStr := string(file.Content)

	for _, dep := range chart.Dependencies {
		depNodeID := fmt.Sprintf("service:%s", dep.Name)
		kind := domain.NodeService
		if isInfrastructure(dep.Name, "") {
			kind = domain.NodeInfrastructure
		}

		line := findYamlKeyLine(contentStr, dep.Name)

		depNode := domain.Node{
			ID:   depNodeID,
			Kind: kind,
			Name: dep.Name,
			Metadata: map[string]any{
				"version":    dep.Version,
				"repository": dep.Repository,
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
					Origin:            domain.OriginDeclared,
					DeploymentProfile: "helm:chart",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, depNode)

		var verb domain.RelationVerb = domain.VerbConsumes
		if kind == domain.NodeInfrastructure {
			verb = domain.VerbAccesses
		}

		rel := domain.Relationship{
			ID:       fmt.Sprintf("rel:%s-%s-%s", chart.Name, string(verb), dep.Name),
			SourceID: chartNodeID,
			TargetID: depNodeID,
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
					Origin:            domain.OriginDeclared,
					DeploymentProfile: "helm:chart",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		relationships = append(relationships, rel)
	}

	return nodes, relationships, nil, nil
}

func (e *HelmExtractor) extractValues(file TargetFile, discoveredAt time.Time) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	var root map[string]any
	if err := yaml.Unmarshal(file.Content, &root); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse values.yaml in %s: %w", file.Path, err)
	}

	contentStr := string(file.Content)
	profile := fmt.Sprintf("helm:%s", filepath.Base(file.Path))

	var nodes []domain.Node
	nodeMap := make(map[string]domain.Node)

	// Top-level sections or components: in Helm values (especially OTel demo), services or components are objects
	// e.g. components: { cartservice: { ... }, ... } or top level: cartservice: { ... }, frontend: { ... }
	type candidateComp struct {
		name string
		data map[string]any
	}
	var candidates []candidateComp

	// Check if "components" key exists
	if comps, ok := root["components"].(map[string]any); ok {
		for name, val := range comps {
			if vMap, ok := val.(map[string]any); ok {
				candidates = append(candidates, candidateComp{name: name, data: vMap})
			}
		}
	}

	// Check other top-level keys that look like services/components (maps with image, ports, env, or enabled)
	for k, v := range root {
		if k == "global" || k == "components" || k == "default" || k == "imagePullSecrets" {
			continue
		}
		if vMap, ok := v.(map[string]any); ok {
			// If it has typical component fields
			_, hasImage := vMap["image"]
			_, hasPorts := vMap["ports"]
			_, hasPort := vMap["port"]
			_, hasEnv := vMap["env"]
			_, hasEnvOverrides := vMap["envOverrides"]
			_, hasEnabled := vMap["enabled"]
			if hasImage || hasPorts || hasPort || hasEnv || hasEnvOverrides || hasEnabled {
				candidates = append(candidates, candidateComp{name: k, data: vMap})
			}
		}
	}

	for _, cand := range candidates {
		svcName := cand.name
		nodeID := fmt.Sprintf("service:%s", svcName)

		imgStr := ""
		if img, ok := cand.data["image"].(string); ok {
			imgStr = img
		} else if imgMap, ok := cand.data["image"].(map[string]any); ok {
			if repo, ok := imgMap["repository"].(string); ok {
				imgStr = repo
			}
		}

		kind := domain.NodeService
		if isInfrastructure(svcName, imgStr) {
			kind = domain.NodeInfrastructure
		}

		line := findYamlKeyLine(contentStr, svcName)

		node := domain.Node{
			ID:   nodeID,
			Kind: kind,
			Name: svcName,
			Metadata: map[string]any{
				"image": imgStr,
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

	// Next, extract relationships from env / envOverrides or endpoints referencing other candidates
	var relationships []domain.Relationship
	for _, cand := range candidates {
		sourceID := fmt.Sprintf("service:%s", cand.name)
		sourceLine := findYamlKeyLine(contentStr, cand.name)

		// Inspect env / envOverrides
		var envList []string
		collectEnvStrings := func(envAny any) {
			switch envVal := envAny.(type) {
			case []any:
				for _, item := range envVal {
					if m, ok := item.(map[string]any); ok {
						if name, ok := m["name"].(string); ok {
							if val, ok := m["value"].(string); ok {
								envList = append(envList, fmt.Sprintf("%s=%s", name, val))
							}
						}
					} else if s, ok := item.(string); ok {
						envList = append(envList, s)
					}
				}
			case map[string]any:
				for k, v := range envVal {
					envList = append(envList, fmt.Sprintf("%s=%v", k, v))
				}
			}
		}

		if envAny, ok := cand.data["env"]; ok {
			collectEnvStrings(envAny)
		}
		if envAny, ok := cand.data["envOverrides"]; ok {
			collectEnvStrings(envAny)
		}

		// Look for references to other known candidates or infra
		for _, envEntry := range envList {
			for targetName, targetNode := range nodeMap {
				if targetName == cand.name {
					continue
				}
				// e.g. CART_SERVICE_ADDR, REDIS_ADDR, KAFKA_ADDR or value contains targetName
				if strings.Contains(strings.ToLower(envEntry), strings.ToLower(targetName)) {
					verb := domain.VerbConsumes
					if targetNode.Kind == domain.NodeInfrastructure {
						verb = domain.VerbAccesses
					}

					relID := fmt.Sprintf("rel:%s-%s-%s", cand.name, string(verb), targetName)
					rel := domain.Relationship{
						ID:       relID,
						SourceID: sourceID,
						TargetID: targetNode.ID,
						Verb:     verb,
						Evidence: []domain.EvidenceRecord{
							{
								Locator: domain.Locator{
									Repo:      file.Repo,
									CommitSHA: file.CommitSHA,
									Path:      file.Path,
									StartLine: sourceLine,
									EndLine:   sourceLine,
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
		}
	}

	return nodes, relationships, nil, nil
}
