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

// GitHubActionsExtractor extracts CI/CD jobs, steps, and referenced actions from GitHub Actions workflows.
type GitHubActionsExtractor struct{}

// NewGitHubActionsExtractor creates a new Layer 2 GitHub Actions extractor.
func NewGitHubActionsExtractor() *GitHubActionsExtractor {
	return &GitHubActionsExtractor{}
}

func (e *GitHubActionsExtractor) Name() string {
	return "Layer2GitHubActions"
}

func (e *GitHubActionsExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches GitHub Actions workflow files under .github/workflows/.
func (e *GitHubActionsExtractor) CanExtract(path string) bool {
	normalized := filepath.ToSlash(path)
	ext := strings.ToLower(filepath.Ext(normalized))
	if ext != ".yml" && ext != ".yaml" {
		return false
	}

	dir := filepath.ToSlash(filepath.Dir(normalized))
	return strings.HasSuffix(dir, ".github/workflows") ||
		strings.Contains(normalized, "/.github/workflows/") ||
		strings.HasPrefix(normalized, ".github/workflows/")
}

// Keywords signaling pipeline infrastructure jobs.
var ghaInfraKeywords = []string{
	"deploy", "release", "publish", "provision", "infra", "terraform",
	"docker", "helm", "k8s", "kubernetes", "cloud", "aws", "gcp", "azure", "push",
}

func isGitHubActionsInfra(jobID string, jobName string) bool {
	id := strings.ToLower(jobID)
	name := strings.ToLower(jobName)
	for _, kw := range ghaInfraKeywords {
		if strings.Contains(id, kw) || strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// sanitizeID replaces non-alphanumeric runes with underscores.
func sanitizeID(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_\-.:/]`)
	return re.ReplaceAllString(s, "_")
}

// getYamlNodeLineRange calculates the starting and ending line numbers of a YAML AST node.
func getYamlNodeLineRange(node *yaml.Node) (int, int) {
	if node == nil {
		return 1, 1
	}
	start := node.Line
	end := node.Line
	for _, child := range node.Content {
		cStart, cEnd := getYamlNodeLineRange(child)
		if cStart > 0 && (start == 0 || cStart < start) {
			start = cStart
		}
		if cEnd > end {
			end = cEnd
		}
	}
	if start <= 0 {
		start = 1
	}
	if end < start {
		end = start
	}
	return start, end
}

// Extract parses the workflow YAML and emits Service/Infrastructure jobs and Library actions.
func (e *GitHubActionsExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(file.Content, &root); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse GitHub Actions workflow in %s: %w", file.Path, err)
	}

	if len(root.Content) == 0 {
		return nil, nil, nil, nil
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, nil, nil, nil
	}

	var workflowName string
	var jobsNode *yaml.Node

	for i := 0; i < len(doc.Content); i += 2 {
		keyNode := doc.Content[i]
		valNode := doc.Content[i+1]

		switch keyNode.Value {
		case "name":
			workflowName = valNode.Value
		case "jobs":
			jobsNode = valNode
		}
	}

	if workflowName == "" {
		base := filepath.Base(file.Path)
		ext := filepath.Ext(base)
		workflowName = strings.TrimSuffix(base, ext)
	}

	if jobsNode == nil || jobsNode.Kind != yaml.MappingNode {
		return nil, nil, nil, nil
	}

	discoveredAt := time.Now().UTC()
	var nodes []domain.Node
	var relationships []domain.Relationship
	actionNodeMap := make(map[string]domain.Node)

	// Iterate over jobs
	for i := 0; i < len(jobsNode.Content); i += 2 {
		select {
		case <-ctx.Done():
			return nil, nil, nil, ctx.Err()
		default:
		}

		jobKeyNode := jobsNode.Content[i]
		jobValNode := jobsNode.Content[i+1]

		jobID := jobKeyNode.Value
		var jobName string
		var runsOn string
		var stepsNode *yaml.Node

		if jobValNode.Kind == yaml.MappingNode {
			for j := 0; j < len(jobValNode.Content); j += 2 {
				k := jobValNode.Content[j]
				v := jobValNode.Content[j+1]
				switch k.Value {
				case "name":
					jobName = v.Value
				case "runs-on":
					runsOn = v.Value
				case "steps":
					stepsNode = v
				}
			}
		}

		displayName := jobName
		if displayName == "" {
			displayName = jobID
		}

		jobStartLine, jobEndLine := getYamlNodeLineRange(jobValNode)
		if jobKeyNode.Line > 0 && jobKeyNode.Line < jobStartLine {
			jobStartLine = jobKeyNode.Line
		}

		jobKind := domain.NodeService
		if isGitHubActionsInfra(jobID, jobName) {
			jobKind = domain.NodeInfrastructure
		}

		jobNodeID := fmt.Sprintf("service:job:%s", jobID)
		jobMetadata := map[string]any{
			"job_id":   jobID,
			"workflow": workflowName,
			"file":     file.Path,
		}
		if runsOn != "" {
			jobMetadata["runs_on"] = runsOn
		}

		jobNode := domain.Node{
			ID:          jobNodeID,
			Kind:        jobKind,
			Name:        displayName,
			Description: fmt.Sprintf("GitHub Actions job %s in workflow %s", jobID, workflowName),
			Metadata:    jobMetadata,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: jobStartLine,
						EndLine:   jobEndLine,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "github-actions",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, jobNode)

		// Parse steps inside job
		if stepsNode != nil && stepsNode.Kind == yaml.SequenceNode {
			for _, stepNode := range stepsNode.Content {
				if stepNode.Kind != yaml.MappingNode {
					continue
				}

				var stepName string
				var usesStr string
				var usesNode *yaml.Node

				for k := 0; k < len(stepNode.Content); k += 2 {
					key := stepNode.Content[k]
					val := stepNode.Content[k+1]
					switch key.Value {
					case "name":
						stepName = val.Value
					case "uses":
						usesStr = val.Value
						usesNode = val
					}
				}

				if usesStr == "" {
					continue
				}

				// External action referenced via 'uses: org/repo@ref'
				actionName := usesStr
				version := ""
				if atIdx := strings.Index(usesStr, "@"); atIdx != -1 {
					actionName = usesStr[:atIdx]
					version = usesStr[atIdx+1:]
				}

				stepStart, stepEnd := getYamlNodeLineRange(stepNode)
				usesLine := stepStart
				if usesNode != nil && usesNode.Line > 0 {
					usesLine = usesNode.Line
				}

				actionNodeID := fmt.Sprintf("library:%s", actionName)
				if _, exists := actionNodeMap[actionNodeID]; !exists {
					actionMeta := map[string]any{
						"action": actionName,
						"uses":   usesStr,
					}
					if version != "" {
						actionMeta["version"] = version
					}

					actionNode := domain.Node{
						ID:          actionNodeID,
						Kind:        domain.NodeLibrary,
						Name:        actionName,
						Description: fmt.Sprintf("GitHub Action %s", usesStr),
						Metadata:    actionMeta,
						Evidence: []domain.EvidenceRecord{
							{
								Locator: domain.Locator{
									Repo:      file.Repo,
									CommitSHA: file.CommitSHA,
									Path:      file.Path,
									StartLine: stepStart,
									EndLine:   stepEnd,
								},
								Origin:            domain.OriginExtracted,
								DeploymentProfile: "github-actions",
								Status:            domain.StatusCurrent,
								DiscoveredAt:      discoveredAt,
							},
						},
					}
					actionNodeMap[actionNodeID] = actionNode
					nodes = append(nodes, actionNode)
				}

				// Relationship: Job -> Action
				relID := fmt.Sprintf("rel:%s-consumes-%s", sanitizeID(jobNodeID), sanitizeID(actionNodeID))
				rel := domain.Relationship{
					ID:       relID,
					SourceID: jobNodeID,
					TargetID: actionNodeID,
					Verb:     domain.VerbConsumes,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: usesLine,
								EndLine:   usesLine,
							},
							Origin:            domain.OriginExtracted,
							DeploymentProfile: "github-actions",
							Status:            domain.StatusCurrent,
							DiscoveredAt:      discoveredAt,
						},
					},
				}
				if stepName != "" {
					// We can store stepName in evidence if needed or retain clean rel
				}
				relationships = append(relationships, rel)
			}
		}
	}

	return nodes, relationships, nil, nil
}
