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

	"gopkg.in/yaml.v3"

	"labs/gira-g/internal/domain"
)

// MarkdownExtractor extracts structured documentation, tasks, skills, ADRs, and policies.
type MarkdownExtractor struct{}

// NewMarkdownExtractor creates a new Layer 2 Markdown and Policy extractor.
func NewMarkdownExtractor() *MarkdownExtractor {
	return &MarkdownExtractor{}
}

func (e *MarkdownExtractor) Name() string {
	return "Layer2Markdown"
}

func (e *MarkdownExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract determines whether a file qualifies for structured markdown extraction.
func (e *MarkdownExtractor) CanExtract(path string) bool {
	normalized := filepath.ToSlash(path)
	ext := strings.ToLower(filepath.Ext(normalized))
	if ext != ".md" && ext != ".markdown" {
		return false
	}

	lower := strings.ToLower(normalized)
	base := strings.ToLower(filepath.Base(normalized))

	// Explicit ADR / Decision paths
	if strings.HasPrefix(lower, "docs/decisions/") || strings.Contains(lower, "/docs/decisions/") ||
		strings.HasPrefix(lower, "docs/adr/") || strings.Contains(lower, "/docs/adr/") {
		return true
	}

	// Well-known documentation files
	if base == "agents.md" || base == "claude.md" || base == "gemini.md" ||
		base == "context.md" || base == "skill.md" {
		return true
	}

	// Task / Spec patterns
	if strings.HasPrefix(lower, "tasks/") || strings.Contains(lower, "/tasks/") ||
		strings.HasPrefix(lower, "specs/") || strings.Contains(lower, "/specs/") {
		return true
	}

	return false
}

// frontmatterInfo captures parsed YAML frontmatter metadata and line boundaries.
type frontmatterInfo struct {
	hasFrontmatter bool
	startLine      int
	endLine        int
	data           map[string]any
}

// parseFrontmatter extracts YAML frontmatter delimited by --- lines at the start of the file.
func parseFrontmatter(content []byte) frontmatterInfo {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0
	inFrontmatter := false
	startLine := 0
	var fmLines []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if lineNum == 1 && trimmed == "---" {
			inFrontmatter = true
			startLine = 1
			continue
		}

		if inFrontmatter {
			if trimmed == "---" {
				// End of frontmatter
				var data map[string]any
				if err := yaml.Unmarshal([]byte(strings.Join(fmLines, "\n")), &data); err == nil && data != nil {
					return frontmatterInfo{
						hasFrontmatter: true,
						startLine:      startLine,
						endLine:        lineNum,
						data:           data,
					}
				}
				return frontmatterInfo{
					hasFrontmatter: true,
					startLine:      startLine,
					endLine:        lineNum,
					data:           make(map[string]any),
				}
			}
			fmLines = append(fmLines, line)
		} else {
			// If first line wasn't '---', no frontmatter
			break
		}
	}

	return frontmatterInfo{}
}

// findDocumentTitle searches for the first Markdown '# <Title>' header.
func findDocumentTitle(content []byte) (string, int) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0
	reTitle := regexp.MustCompile(`^#\s+(.+)$`)

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		matches := reTitle.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1]), lineNum
		}
	}
	return "", 1
}

// cleanSlug normalizes a filename or title into a kebab-case slug.
func cleanSlug(s string) string {
	s = strings.TrimSuffix(s, filepath.Ext(s))
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9_\-]+`)
	cleaned := reg.ReplaceAllString(s, "-")
	cleaned = strings.Trim(cleaned, "-")
	return cleaned
}

// Extract parses the markdown file, extracts structured entities and relationships.
func (e *MarkdownExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	normalized := filepath.ToSlash(file.Path)
	lower := strings.ToLower(normalized)
	base := strings.ToLower(filepath.Base(normalized))

	fm := parseFrontmatter(file.Content)
	title, titleLine := findDocumentTitle(file.Content)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
	}

	discoveredAt := time.Now().UTC()
	var nodes []domain.Node
	var relationships []domain.Relationship

	// Total lines count in file
	totalLines := bytes.Count(file.Content, []byte("\n"))
	if len(file.Content) > 0 && !bytes.HasSuffix(file.Content, []byte("\n")) {
		totalLines++
	}
	if totalLines == 0 {
		totalLines = 1
	}

	// 1. Classification
	isTask := false
	isSkill := false
	isPolicy := false

	// Check schema in frontmatter
	if fm.hasFrontmatter {
		if schemaVal, ok := fm.data["schema"].(string); ok {
			if strings.HasPrefix(schemaVal, "task") {
				isTask = true
			} else if strings.HasPrefix(schemaVal, "dec") || strings.HasPrefix(schemaVal, "policy") || strings.HasPrefix(schemaVal, "adr") {
				isPolicy = true
			}
		}
	}

	if !isTask && !isPolicy {
		if base == "skill.md" || (fm.hasFrontmatter && fm.data["name"] != nil && strings.Contains(lower, "skill")) {
			isSkill = true
		} else if strings.HasPrefix(lower, "tasks/") || strings.Contains(lower, "/tasks/") {
			isTask = true
		} else if strings.HasPrefix(lower, "docs/decisions/") || strings.Contains(lower, "/docs/decisions/") ||
			strings.HasPrefix(lower, "docs/adr/") || strings.Contains(lower, "/docs/adr/") {
			isPolicy = true
		} else if base == "agents.md" || base == "claude.md" || base == "gemini.md" || base == "context.md" {
			isPolicy = true
		} else if strings.HasPrefix(lower, "specs/") || strings.Contains(lower, "/specs/") {
			isPolicy = true
		}
	}

	// 2. Node Generation based on type
	if isTask {
		slug := cleanSlug(filepath.Base(file.Path))
		nodeID := fmt.Sprintf("task:%s", slug)

		meta := make(map[string]any)
		meta["path"] = file.Path
		meta["title"] = title
		if fm.hasFrontmatter {
			for k, v := range fm.data {
				meta[k] = v
			}
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeTask,
			Name:        title,
			Description: fmt.Sprintf("Task documentation: %s", title),
			Metadata:    meta,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: 1,
						EndLine:   totalLines,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "task",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)

		// Check for relationships inside task markdown:
		// e.g. "Blocked by: ...", "implements ...", verification table lines
		scanner := bufio.NewScanner(bytes.NewReader(file.Content))
		currentLine := 0
		reBlockedBy := regexp.MustCompile(`(?i)Blocked\s+by:\s*([^\r\n]+)`)
		reImplements := regexp.MustCompile(`(?i)implements\s*:\s*([^\r\n]+)`)
		reVerTable := regexp.MustCompile(`^\|\s*([A-Za-z0-9_.\-]+)\s*\|\s*([^|]+)\|\s*(pass|fail|blocked|not run)\s*\|`)

		for scanner.Scan() {
			currentLine++
			select {
			case <-ctx.Done():
				return nil, nil, nil, ctx.Err()
			default:
			}

			lineText := scanner.Text()

			// Check blocked by (VerbGoverns or VerbImplements)
			if m := reBlockedBy.FindStringSubmatch(lineText); len(m) > 1 {
				val := strings.TrimSpace(m[1])
				if val != "—" && val != "-" && val != "none" && val != "" {
					// Extract potential task references
					reTaskRef := regexp.MustCompile(`(?:\[.*?\]\()?(?:tasks/)?([0-9a-zA-Z_\-]+)(?:\.md)?\)?`)
					matches := reTaskRef.FindAllStringSubmatch(val, -1)
					for _, match := range matches {
						if len(match) > 1 && match[1] != "" && match[1] != "none" {
							targetSlug := cleanSlug(match[1])
							if targetSlug != "" && targetSlug != slug {
								targetID := fmt.Sprintf("task:%s", targetSlug)
								rel := domain.Relationship{
									ID:       fmt.Sprintf("rel:%s-governs-%s", sanitizeID(targetID), sanitizeID(nodeID)),
									SourceID: targetID,
									TargetID: nodeID,
									Verb:     domain.VerbGoverns,
									Evidence: []domain.EvidenceRecord{
										{
											Locator: domain.Locator{
												Repo:      file.Repo,
												CommitSHA: file.CommitSHA,
												Path:      file.Path,
												StartLine: currentLine,
												EndLine:   currentLine,
											},
											Origin:            domain.OriginExtracted,
											DeploymentProfile: "task",
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
			}

			// Check implements
			if m := reImplements.FindStringSubmatch(lineText); len(m) > 1 {
				val := strings.TrimSpace(m[1])
				parts := strings.Split(val, ",")
				for _, p := range parts {
					target := cleanSlug(p)
					if target != "" {
						targetID := fmt.Sprintf("policy:%s", target)
						rel := domain.Relationship{
							ID:       fmt.Sprintf("rel:%s-implements-%s", sanitizeID(nodeID), sanitizeID(targetID)),
							SourceID: nodeID,
							TargetID: targetID,
							Verb:     domain.VerbImplements,
							Evidence: []domain.EvidenceRecord{
								{
									Locator: domain.Locator{
										Repo:      file.Repo,
										CommitSHA: file.CommitSHA,
										Path:      file.Path,
										StartLine: currentLine,
										EndLine:   currentLine,
									},
									Origin:            domain.OriginExtracted,
									DeploymentProfile: "task",
									Status:            domain.StatusCurrent,
									DiscoveredAt:      discoveredAt,
								},
							},
						}
						relationships = append(relationships, rel)
					}
				}
			}

			// Check verification table
			if m := reVerTable.FindStringSubmatch(lineText); len(m) > 3 {
				checkID := strings.TrimSpace(m[1])
				checkDesc := strings.TrimSpace(m[2])
				if checkID != "ID" && checkID != "---" {
					// Emits a verification relation if check references a target or verifies the task itself
					verTargetID := fmt.Sprintf("contract:check:%s", strings.ToLower(checkID))
					rel := domain.Relationship{
						ID:       fmt.Sprintf("rel:%s-verifies-%s", sanitizeID(nodeID), sanitizeID(verTargetID)),
						SourceID: nodeID,
						TargetID: verTargetID,
						Verb:     domain.VerbVerifies,
						Evidence: []domain.EvidenceRecord{
							{
								Locator: domain.Locator{
									Repo:      file.Repo,
									CommitSHA: file.CommitSHA,
									Path:      file.Path,
									StartLine: currentLine,
									EndLine:   currentLine,
								},
								Origin:            domain.OriginExtracted,
								DeploymentProfile: "task",
								Status:            domain.StatusCurrent,
								DiscoveredAt:      discoveredAt,
							},
						},
					}
					relationships = append(relationships, rel)
					_ = checkDesc
				}
			}
		}

	} else if isSkill {
		skillName := title
		if fm.hasFrontmatter {
			if n, ok := fm.data["name"].(string); ok && strings.TrimSpace(n) != "" {
				skillName = strings.TrimSpace(n)
			}
		}
		if skillName == "" || skillName == "skill" {
			// fallback to parent dir name
			dir := filepath.Base(filepath.Dir(file.Path))
			if dir != "." && dir != "/" {
				skillName = dir
			}
		}

		nodeID := fmt.Sprintf("skill:%s", cleanSlug(skillName))
		meta := make(map[string]any)
		meta["path"] = file.Path
		meta["name"] = skillName
		if fm.hasFrontmatter {
			for k, v := range fm.data {
				meta[k] = v
			}
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodeSkill,
			Name:        skillName,
			Description: fmt.Sprintf("Agent Skill: %s", skillName),
			Metadata:    meta,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: 1,
						EndLine:   totalLines,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "skill",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)

	} else if isPolicy {
		slug := cleanSlug(filepath.Base(file.Path))
		nodeID := fmt.Sprintf("policy:%s", slug)

		meta := make(map[string]any)
		meta["path"] = file.Path
		meta["title"] = title
		if fm.hasFrontmatter {
			for k, v := range fm.data {
				meta[k] = v
			}
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodePolicy,
			Name:        title,
			Description: fmt.Sprintf("Policy / ADR: %s", title),
			Metadata:    meta,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: 1,
						EndLine:   totalLines,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: "markdown",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)

		// Look for governing rules / paths mentioned in the ADR or Policy
		scanner := bufio.NewScanner(bytes.NewReader(file.Content))
		currentLine := 0
		reGoverns := regexp.MustCompile(`(?i)(?:governa|governs|applies to|scope):\s*([^\r\n]+)`)

		for scanner.Scan() {
			currentLine++
			lineText := scanner.Text()

			if m := reGoverns.FindStringSubmatch(lineText); len(m) > 1 {
				val := strings.TrimSpace(m[1])
				parts := strings.Split(val, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						targetSlug := cleanSlug(p)
						targetID := fmt.Sprintf("service:%s", targetSlug)
						rel := domain.Relationship{
							ID:       fmt.Sprintf("rel:%s-governs-%s", sanitizeID(nodeID), sanitizeID(targetID)),
							SourceID: nodeID,
							TargetID: targetID,
							Verb:     domain.VerbGoverns,
							Evidence: []domain.EvidenceRecord{
								{
									Locator: domain.Locator{
										Repo:      file.Repo,
										CommitSHA: file.CommitSHA,
										Path:      file.Path,
										StartLine: currentLine,
										EndLine:   currentLine,
									},
									Origin:            domain.OriginExtracted,
									DeploymentProfile: "markdown",
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
	}

	_ = titleLine
	return nodes, relationships, nil, nil
}
