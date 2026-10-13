package extractors

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// MiseExtractor extracts toolchain definitions, environments, and tasks from mise configuration files.
type MiseExtractor struct{}

// NewMiseExtractor creates a new Layer 2 Mise extractor.
func NewMiseExtractor() *MiseExtractor {
	return &MiseExtractor{}
}

func (e *MiseExtractor) Name() string {
	return "Layer2Mise"
}

func (e *MiseExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches mise configuration files (mise.toml, .mise.toml, mise/config.toml, .mise/config.toml).
func (e *MiseExtractor) CanExtract(path string) bool {
	normalized := filepath.ToSlash(path)
	base := strings.ToLower(filepath.Base(normalized))

	if base == "mise.toml" || base == ".mise.toml" {
		return true
	}

	if base == "config.toml" {
		dir := strings.ToLower(filepath.Base(filepath.Dir(normalized)))
		if dir == "mise" || dir == ".mise" {
			return true
		}
	}

	return strings.HasSuffix(normalized, "mise/config.toml") ||
		strings.HasSuffix(normalized, ".mise/config.toml")
}

var miseInfraTools = map[string]bool{
	"docker":    true,
	"terraform": true,
	"helm":      true,
	"kubectl":   true,
	"k8s":       true,
	"cosign":    true,
	"syft":      true,
	"awscli":    true,
	"gcloud":    true,
	"az":        true,
	"vault":     true,
	"consul":    true,
}

func isMiseInfra(tool string) bool {
	t := strings.ToLower(tool)
	return miseInfraTools[t]
}

type miseToolDef struct {
	name      string
	version   string
	startLine int
	endLine   int
}

type miseTaskDef struct {
	name        string
	description string
	run         string
	depends     []string
	startLine   int
	endLine     int
}

// stripInlineComment removes TOML comments starting with '#' not inside quotes.
func stripInlineComment(line string) string {
	inQuote := false
	var quoteChar rune
	for i, r := range line {
		if (r == '"' || r == '\'') && (i == 0 || line[i-1] != '\\') {
			if !inQuote {
				inQuote = true
				quoteChar = r
			} else if r == quoteChar {
				inQuote = false
			}
		} else if r == '#' && !inQuote {
			return strings.TrimSpace(line[:i])
		}
	}
	return strings.TrimSpace(line)
}

func extractTomlStringValue(val string) string {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") && len(val) >= 2 {
		return val[1 : len(val)-1]
	}
	if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") && len(val) >= 2 {
		return val[1 : len(val)-1]
	}
	return val
}

func parseTomlStringArray(val string) []string {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		inner := val[1 : len(val)-1]
		parts := strings.Split(inner, ",")
		var res []string
		for _, p := range parts {
			cleaned := extractTomlStringValue(strings.TrimSpace(p))
			if cleaned != "" {
				res = append(res, cleaned)
			}
		}
		return res
	}
	return nil
}

// Extract parses the mise.toml file and emits Library/Infrastructure tools and Service tasks.
func (e *MiseExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	scanner := bufio.NewScanner(bytes.NewReader(file.Content))

	var tools []miseToolDef
	var tasks []*miseTaskDef
	var currentTask *miseTaskDef
	envVars := make(map[string]string)

	currentSection := ""
	lineNum := 0

	reHeader := regexp.MustCompile(`^\[([A-Za-z0-9_.\-]+)\]$`)

	for scanner.Scan() {
		lineNum++

		select {
		case <-ctx.Done():
			return nil, nil, nil, ctx.Err()
		default:
		}

		rawLine := scanner.Text()
		trimmed := stripInlineComment(rawLine)

		if trimmed == "" {
			continue
		}

		// Section header, e.g. [tools], [tasks.check], [env]
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if currentTask != nil {
				currentTask.endLine = lineNum - 1
				currentTask = nil
			}

			matches := reHeader.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				headerName := matches[1]
				if headerName == "tools" || headerName == "env" || headerName == "tasks" {
					currentSection = headerName
				} else if strings.HasPrefix(headerName, "tasks.") {
					currentSection = "tasks"
					taskName := strings.TrimPrefix(headerName, "tasks.")
					currentTask = &miseTaskDef{
						name:      taskName,
						startLine: lineNum,
						endLine:   lineNum,
					}
					tasks = append(tasks, currentTask)
				} else if strings.HasPrefix(headerName, "tools.") {
					currentSection = "tools"
					toolName := strings.TrimPrefix(headerName, "tools.")
					tools = append(tools, miseToolDef{
						name:      toolName,
						startLine: lineNum,
						endLine:   lineNum,
					})
				} else {
					currentSection = headerName
				}
			}
			continue
		}

		// Key-Value parsing
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])

			switch {
			case currentSection == "tools":
				version := extractTomlStringValue(v)
				if strings.HasPrefix(v, "{") && strings.Contains(v, "version") {
					reVer := regexp.MustCompile(`version\s*=\s*["']([^"']+)["']`)
					verMatch := reVer.FindStringSubmatch(v)
					if len(verMatch) > 1 {
						version = verMatch[1]
					}
				}
				tools = append(tools, miseToolDef{
					name:      k,
					version:   version,
					startLine: lineNum,
					endLine:   lineNum,
				})

			case currentTask != nil:
				currentTask.endLine = lineNum
				switch k {
				case "description":
					currentTask.description = extractTomlStringValue(v)
				case "run":
					currentTask.run = extractTomlStringValue(v)
				case "depends":
					currentTask.depends = parseTomlStringArray(v)
				}

			case currentSection == "tasks":
				// Inline tasks under [tasks]: check = "go test ..."
				runCmd := extractTomlStringValue(v)
				tasks = append(tasks, &miseTaskDef{
					name:      k,
					run:       runCmd,
					startLine: lineNum,
					endLine:   lineNum,
				})

			case currentSection == "env":
				envVars[k] = extractTomlStringValue(v)
			}
		} else if currentTask != nil {
			currentTask.endLine = lineNum
		}
	}

	if currentTask != nil && currentTask.endLine < lineNum {
		currentTask.endLine = lineNum
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to scan mise configuration in %s: %w", file.Path, err)
	}

	discoveredAt := time.Now().UTC()
	var nodes []domain.Node
	var relationships []domain.Relationship

	toolMap := make(map[string]domain.Node)

	// Emit tool nodes
	for _, tool := range tools {
		kind := domain.NodeLibrary
		if isMiseInfra(tool.name) {
			kind = domain.NodeInfrastructure
		}

		nodeID := fmt.Sprintf("library:%s", tool.name)
		meta := map[string]any{
			"tool": tool.name,
		}
		if tool.version != "" {
			meta["version"] = tool.version
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        kind,
			Name:        tool.name,
			Description: fmt.Sprintf("Mise toolchain %s", tool.name),
			Metadata:    meta,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: tool.startLine,
						EndLine:   tool.endLine,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "mise",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}

		toolMap[tool.name] = node
		nodes = append(nodes, node)
	}

	// Emit task nodes
	for _, task := range tasks {
		nodeID := fmt.Sprintf("service:task:%s", task.name)
		taskMeta := map[string]any{
			"task": task.name,
		}
		if task.run != "" {
			taskMeta["run"] = task.run
		}
		if len(task.depends) > 0 {
			taskMeta["depends"] = task.depends
		}
		if len(envVars) > 0 {
			taskMeta["env_vars_count"] = len(envVars)
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeService,
			Name:        task.name,
			Description: task.description,
			Metadata:    taskMeta,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: task.startLine,
						EndLine:   task.endLine,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "mise",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)

		// Link task -> tool dependencies if the run command invokes a configured tool
		if task.run != "" {
			runFields := strings.Fields(task.run)
			if len(runFields) > 0 {
				invokedCmd := runFields[0]
				if toolNode, hasTool := toolMap[invokedCmd]; hasTool {
					relID := fmt.Sprintf("rel:%s-consumes-%s", sanitizeID(nodeID), sanitizeID(toolNode.ID))
					rel := domain.Relationship{
						ID:       relID,
						SourceID: nodeID,
						TargetID: toolNode.ID,
						Verb:     domain.VerbConsumes,
						Evidence: []domain.EvidenceRecord{
							{
								Locator: domain.Locator{
									Repo:      file.Repo,
									CommitSHA: file.CommitSHA,
									Path:      file.Path,
									StartLine: task.startLine,
									EndLine:   task.endLine,
								},
								Origin:            domain.OriginExtracted,
								DeploymentProfile: "mise",
								Status:            domain.StatusCurrent,
								DiscoveredAt:      discoveredAt,
							},
						},
					}
					relationships = append(relationships, rel)
				}
			}
		}

		// Link task -> depends tasks
		for _, depTaskName := range task.depends {
			depTaskID := fmt.Sprintf("service:task:%s", depTaskName)
			relID := fmt.Sprintf("rel:%s-consumes-%s", sanitizeID(nodeID), sanitizeID(depTaskID))
			rel := domain.Relationship{
				ID:       relID,
				SourceID: nodeID,
				TargetID: depTaskID,
				Verb:     domain.VerbConsumes,
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      file.Repo,
							CommitSHA: file.CommitSHA,
							Path:      file.Path,
							StartLine: task.startLine,
							EndLine:   task.endLine,
						},
						Origin:            domain.OriginExtracted,
						DeploymentProfile: "mise",
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
