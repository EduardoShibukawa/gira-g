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

	"github.com/EduardoShibukawa/gira-g/internal/domain"
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
	normalized := filepath.ToSlash(filepath.Clean(path))
	ext := strings.ToLower(filepath.Ext(normalized))
	if ext != ".md" && ext != ".markdown" {
		return false
	}

	lower := strings.ToLower(normalized)
	base := strings.ToLower(filepath.Base(normalized))

	// 1. .scratch patterns: .scratch/*/issues/*.md, .scratch/*/map.md, .scratch/*/brief.md
	if strings.HasPrefix(lower, ".scratch/") || strings.Contains(lower, "/.scratch/") {
		if strings.Contains(lower, "/issues/") || base == "map.md" || base == "brief.md" {
			return true
		}
	}

	// 2. Well-known documentation files (case-insensitive)
	if base == "context-map.md" || base == "readme.md" || base == "agents.md" ||
		base == "claude.md" || base == "gemini.md" || base == "context.md" || base == "skill.md" {
		return true
	}

	// 3. Explicit ADR / Decision paths
	if strings.HasPrefix(lower, "docs/decisions/") || strings.Contains(lower, "/docs/decisions/") ||
		strings.HasPrefix(lower, "docs/adr/") || strings.Contains(lower, "/docs/adr/") {
		return true
	}

	// 4. Task / Spec patterns
	if strings.HasPrefix(lower, "tasks/") || strings.Contains(lower, "/tasks/") ||
		strings.HasPrefix(lower, "specs/") || strings.Contains(lower, "/specs/") {
		return true
	}

	return false
}

// extractEffort extracts the effort name from a path containing .scratch/<effort>/...
func extractEffort(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(path))
	parts := strings.Split(normalized, "/")
	for i, p := range parts {
		if strings.ToLower(p) == ".scratch" && i+1 < len(parts) {
			if strings.ToLower(parts[i+1]) != "issues" {
				return parts[i+1]
			}
		}
	}
	return ""
}

// resolveTargetNodeID determines the canonical target node ID from a relative markdown link.
func resolveTargetNodeID(currentPath string, targetRef string) (string, bool) {
	targetRef = strings.TrimSpace(targetRef)
	if idx := strings.Index(targetRef, "#"); idx != -1 {
		targetRef = targetRef[:idx]
	}
	if idx := strings.Index(targetRef, "?"); idx != -1 {
		targetRef = targetRef[:idx]
	}
	targetRef = strings.Trim(targetRef, "`\"'")
	targetRef = strings.TrimSpace(targetRef)
	if targetRef == "" {
		return "", false
	}

	// Ignore external URLs
	lowerRef := strings.ToLower(targetRef)
	if strings.HasPrefix(lowerRef, "http://") || strings.HasPrefix(lowerRef, "https://") ||
		strings.HasPrefix(lowerRef, "mailto:") || strings.HasPrefix(lowerRef, "ftp://") {
		return "", false
	}

	// Direct repo-relative paths
	if strings.HasPrefix(lowerRef, "tasks/") {
		slug := cleanSlug(filepath.Base(lowerRef))
		if slug == "" {
			return "", false
		}
		return fmt.Sprintf("task:%s", slug), true
	}
	if strings.HasPrefix(lowerRef, "docs/decisions/") || strings.HasPrefix(lowerRef, "docs/adr/") {
		slug := cleanSlug(filepath.Base(lowerRef))
		if slug == "" {
			return "", false
		}
		return fmt.Sprintf("policy:%s", slug), true
	}

	dir := filepath.Dir(filepath.ToSlash(currentPath))
	targetPath := filepath.Clean(filepath.Join(dir, targetRef))
	targetPath = filepath.ToSlash(targetPath)
	targetLower := strings.ToLower(targetPath)
	targetBase := strings.ToLower(filepath.Base(targetPath))

	ext := filepath.Ext(targetPath)
	if ext != ".md" && ext != ".markdown" {
		return "", false
	}

	// 1. .scratch: .scratch/<effort>/issues/<file>.md, .scratch/<effort>/map.md, .scratch/<effort>/brief.md
	if strings.Contains(targetLower, "/.scratch/") || strings.HasPrefix(targetLower, ".scratch/") {
		parts := strings.Split(targetPath, "/")
		for i, p := range parts {
			if strings.ToLower(p) == ".scratch" && i+2 < len(parts) {
				effort := parts[i+1]
				if strings.ToLower(parts[i+2]) == "issues" && i+3 == len(parts)-1 {
					slug := cleanSlug(parts[i+3])
					if slug == "" {
						return "", false
					}
					return fmt.Sprintf("task:%s:%s", effort, slug), true
				}
				if targetBase == "map.md" || targetBase == "map.markdown" {
					return fmt.Sprintf("policy:%s:map", effort), true
				}
				if targetBase == "brief.md" || targetBase == "brief.markdown" {
					return fmt.Sprintf("policy:%s:brief", effort), true
				}
			}
		}
	}

	// 2. CONTEXT-MAP.md
	if targetBase == "context-map.md" || targetBase == "context-map.markdown" {
		return "policy:context-map", true
	}

	// 3. README.md
	if targetBase == "readme.md" || targetBase == "readme.markdown" {
		return "policy:readme", true
	}

	// 4. docs/decisions/ or docs/adr/
	if strings.Contains(targetLower, "/docs/decisions/") || strings.HasPrefix(targetLower, "docs/decisions/") ||
		strings.Contains(targetLower, "/docs/adr/") || strings.HasPrefix(targetLower, "docs/adr/") {
		slug := cleanSlug(targetBase)
		if slug == "" {
			return "", false
		}
		return fmt.Sprintf("policy:%s", slug), true
	}

	// 5. tasks/
	if strings.Contains(targetLower, "/tasks/") || strings.HasPrefix(targetLower, "tasks/") {
		slug := cleanSlug(targetBase)
		if slug == "" {
			return "", false
		}
		return fmt.Sprintf("task:%s", slug), true
	}

	// 6. Generic well-known docs
	if targetBase == "agents.md" || targetBase == "claude.md" || targetBase == "gemini.md" ||
		targetBase == "context.md" {
		slug := cleanSlug(targetBase)
		if slug == "" {
			return "", false
		}
		return fmt.Sprintf("policy:%s", slug), true
	}

	// Default fallback
	slug := cleanSlug(targetBase)
	if slug == "" {
		return "", false
	}
	return fmt.Sprintf("policy:%s", slug), true
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

var slugNonAlnumReg = regexp.MustCompile(`[^a-z0-9_\-]+`)

// cleanSlug normalizes a filename or title into a kebab-case slug.
// Returns an empty string if the input consists only of punctuation, special characters, or spaces.
func cleanSlug(s string) string {
	s = strings.TrimSuffix(s, filepath.Ext(s))
	s = strings.ToLower(s)
	cleaned := slugNonAlnumReg.ReplaceAllString(s, "-")
	cleaned = strings.Trim(cleaned, "-_")
	hasAlphaNum := false
	for _, r := range cleaned {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			hasAlphaNum = true
			break
		}
	}
	if !hasAlphaNum {
		return ""
	}
	return cleaned
}

// isValidTargetSlug validates that a target slug has at least 2 alphanumeric characters
// and is not purely punctuation, trivial symbols, or empty.
func isValidTargetSlug(s string) bool {
	if len(s) < 2 {
		return false
	}
	alnumCount := 0
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			alnumCount++
		}
	}
	return alnumCount >= 2
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

	// 1. Specialized Classification for .scratch issues, maps, briefs, and system docs
	isScratchIssue := (strings.HasPrefix(lower, ".scratch/") || strings.Contains(lower, "/.scratch/")) && strings.Contains(lower, "/issues/")
	isScratchMap := (strings.HasPrefix(lower, ".scratch/") || strings.Contains(lower, "/.scratch/")) && (base == "map.md" || base == "map.markdown")
	isScratchBrief := (strings.HasPrefix(lower, ".scratch/") || strings.Contains(lower, "/.scratch/")) && (base == "brief.md" || base == "brief.markdown")
	isContextMap := base == "context-map.md" || base == "context-map.markdown"
	isReadme := base == "readme.md" || base == "readme.markdown"

	if isScratchIssue {
		effort := extractEffort(file.Path)
		slug := cleanSlug(filepath.Base(file.Path))
		var nodeID string
		if effort != "" {
			nodeID = fmt.Sprintf("task:%s:%s", effort, slug)
		} else {
			nodeID = fmt.Sprintf("task:%s", slug)
		}

		var (
			issueType     string
			issueStatus   string
			blockedByVal  string
			mapVal        string
			blockedByLine int
			mapLine       int
		)

		scanner := bufio.NewScanner(bytes.NewReader(file.Content))
		lineNum := 0
		reType := regexp.MustCompile(`(?i)^Type:\s*(.+)$`)
		reStatus := regexp.MustCompile(`(?i)^Status:\s*(.+)$`)
		reBlockedBy := regexp.MustCompile(`(?i)^Blocked\s+by:\s*(.+)$`)
		reMap := regexp.MustCompile(`(?i)^Map:\s*(.+)$`)

		for scanner.Scan() {
			lineNum++
			lineText := scanner.Text()
			trimmed := strings.TrimSpace(lineText)

			if m := reType.FindStringSubmatch(trimmed); len(m) > 1 && issueType == "" {
				issueType = strings.TrimSpace(m[1])
			}
			if m := reStatus.FindStringSubmatch(trimmed); len(m) > 1 && issueStatus == "" {
				issueStatus = strings.TrimSpace(m[1])
			}
			if m := reBlockedBy.FindStringSubmatch(trimmed); len(m) > 1 && blockedByVal == "" {
				blockedByVal = strings.TrimSpace(m[1])
				blockedByLine = lineNum
			}
			if m := reMap.FindStringSubmatch(trimmed); len(m) > 1 && mapVal == "" {
				mapVal = strings.TrimSpace(m[1])
				mapLine = lineNum
			}
		}

		if fm.hasFrontmatter {
			if t, ok := fm.data["type"].(string); ok && issueType == "" {
				issueType = strings.TrimSpace(t)
			}
			if s, ok := fm.data["status"].(string); ok && issueStatus == "" {
				issueStatus = strings.TrimSpace(s)
			}
			if b, ok := fm.data["blocked_by"].(string); ok && blockedByVal == "" {
				blockedByVal = strings.TrimSpace(b)
			}
			if m, ok := fm.data["map"].(string); ok && mapVal == "" {
				mapVal = strings.TrimSpace(m)
			}
		}

		nodeKind := domain.NodeTask
		lowerType := strings.ToLower(issueType)
		if lowerType == "decision" || lowerType == "grilling" {
			nodeKind = domain.NodePolicy
		}

		deployProfile := "task"
		if nodeKind == domain.NodePolicy {
			deployProfile = "markdown"
		}

		meta := make(map[string]any)
		meta["path"] = file.Path
		meta["title"] = title
		if effort != "" {
			meta["effort"] = effort
		}
		meta["slug"] = slug
		if issueType != "" {
			meta["type"] = issueType
		}
		if issueStatus != "" {
			meta["status"] = issueStatus
		}
		if blockedByVal != "" {
			meta["blocked_by"] = blockedByVal
		}
		if mapVal != "" {
			meta["map"] = mapVal
		}
		if fm.hasFrontmatter {
			for k, v := range fm.data {
				meta[k] = v
			}
		}

		descPrefix := "Task"
		if nodeKind == domain.NodePolicy {
			descPrefix = "Policy issue"
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        nodeKind,
			Name:        title,
			Description: fmt.Sprintf("%s: %s", descPrefix, title),
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
					DeploymentProfile: deployProfile,
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)

		// 1. Blocked by: -> blocker VerbGoverns this issue
		if blockedByVal != "" && blockedByVal != "—" && blockedByVal != "-" && strings.ToLower(blockedByVal) != "none" {
			reMarkdownLink := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
			blockers := strings.Split(blockedByVal, ",")
			for _, b := range blockers {
				b = strings.TrimSpace(b)
				if b == "" || b == "—" || b == "-" || strings.ToLower(b) == "none" {
					continue
				}
				var targetID string
				if m := reMarkdownLink.FindStringSubmatch(b); len(m) > 2 {
					linkTarget := m[2]
					if tid, ok := resolveTargetNodeID(file.Path, linkTarget); ok {
						targetID = tid
					} else {
						targetSlug := cleanSlug(linkTarget)
						if effort != "" {
							targetID = fmt.Sprintf("task:%s:%s", effort, targetSlug)
						} else {
							targetID = fmt.Sprintf("task:%s", targetSlug)
						}
					}
				} else {
					targetSlug := cleanSlug(b)
					if targetSlug != "" && targetSlug != slug {
						if effort != "" {
							targetID = fmt.Sprintf("task:%s:%s", effort, targetSlug)
						} else {
							targetID = fmt.Sprintf("task:%s", targetSlug)
						}
					}
				}

				if targetID != "" && targetID != nodeID {
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
									StartLine: blockedByLine,
									EndLine:   blockedByLine,
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

		// 2. Map: -> map VerbGoverns this issue
		if mapVal != "" {
			reMarkdownLink := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
			var mapNodeID string
			if m := reMarkdownLink.FindStringSubmatch(mapVal); len(m) > 2 {
				linkTarget := m[2]
				if tid, ok := resolveTargetNodeID(file.Path, linkTarget); ok {
					mapNodeID = tid
				}
			} else {
				if tid, ok := resolveTargetNodeID(file.Path, mapVal); ok {
					mapNodeID = tid
				}
			}
			if mapNodeID == "" && effort != "" {
				mapNodeID = fmt.Sprintf("policy:%s:map", effort)
			}

			if mapNodeID != "" && mapNodeID != nodeID {
				rel := domain.Relationship{
					ID:       fmt.Sprintf("rel:%s-governs-%s", sanitizeID(mapNodeID), sanitizeID(nodeID)),
					SourceID: mapNodeID,
					TargetID: nodeID,
					Verb:     domain.VerbGoverns,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: mapLine,
								EndLine:   mapLine,
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

		return nodes, relationships, nil, nil
	}

	if isScratchMap || isScratchBrief || isContextMap || isReadme {
		effort := extractEffort(file.Path)
		var nodeID string
		docType := "policy"

		if isScratchMap {
			if effort != "" {
				nodeID = fmt.Sprintf("policy:%s:map", effort)
			} else {
				nodeID = "policy:map"
			}
			docType = "map"
		} else if isScratchBrief {
			if effort != "" {
				nodeID = fmt.Sprintf("policy:%s:brief", effort)
			} else {
				nodeID = "policy:brief"
			}
			docType = "brief"
		} else if isContextMap {
			nodeID = "policy:context-map"
			docType = "context-map"
		} else if isReadme {
			nodeID = "policy:readme"
			docType = "readme"
		}

		meta := make(map[string]any)
		meta["path"] = file.Path
		meta["title"] = title
		meta["type"] = docType
		if effort != "" {
			meta["effort"] = effort
		}
		if fm.hasFrontmatter {
			for k, v := range fm.data {
				meta[k] = v
			}
		}

		node := domain.Node{
			ID:          nodeID,
			Kind:        domain.NodePolicy,
			Name:        title,
			Description: fmt.Sprintf("Policy (%s): %s", docType, title),
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

		// Scan lines for internal markdown links: [Label](target)
		scanner := bufio.NewScanner(bytes.NewReader(file.Content))
		currentLine := 0
		reMarkdownLink := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
		relIndexMap := make(map[string]int)

		for scanner.Scan() {
			currentLine++
			select {
			case <-ctx.Done():
				return nil, nil, nil, ctx.Err()
			default:
			}

			lineText := scanner.Text()
			matches := reMarkdownLink.FindAllStringSubmatch(lineText, -1)
			for _, m := range matches {
				if len(m) > 2 {
					targetRef := m[2]
					targetID, ok := resolveTargetNodeID(file.Path, targetRef)
					if ok && targetID != "" && targetID != nodeID {
						ev := domain.EvidenceRecord{
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
						}
						if idx, exists := relIndexMap[targetID]; exists {
							relationships[idx].Evidence = append(relationships[idx].Evidence, ev)
						} else {
							rel := domain.Relationship{
								ID:       fmt.Sprintf("rel:%s-governs-%s", sanitizeID(nodeID), sanitizeID(targetID)),
								SourceID: nodeID,
								TargetID: targetID,
								Verb:     domain.VerbGoverns,
								Evidence: []domain.EvidenceRecord{ev},
							}
							relIndexMap[targetID] = len(relationships)
							relationships = append(relationships, rel)
						}
					}
				}
			}
		}

		return nodes, relationships, nil, nil
	}

	// 2. Standard Classification
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
		inCodeBlock := false

		for scanner.Scan() {
			currentLine++
			select {
			case <-ctx.Done():
				return nil, nil, nil, ctx.Err()
			default:
			}

			lineText := scanner.Text()

			if strings.HasPrefix(strings.TrimSpace(lineText), "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			if inCodeBlock {
				continue
			}

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
							if isValidTargetSlug(targetSlug) && targetSlug != slug {
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
					if isValidTargetSlug(target) {
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
		inCodeBlock := false

		for scanner.Scan() {
			currentLine++
			select {
			case <-ctx.Done():
				return nil, nil, nil, ctx.Err()
			default:
			}

			lineText := scanner.Text()

			if strings.HasPrefix(strings.TrimSpace(lineText), "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			if inCodeBlock {
				continue
			}

			if m := reGoverns.FindStringSubmatch(lineText); len(m) > 1 {
				val := strings.TrimSpace(m[1])
				parts := strings.Split(val, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					targetSlug := cleanSlug(p)
					if !isValidTargetSlug(targetSlug) {
						continue
					}
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

	_ = titleLine
	return nodes, relationships, nil, nil
}
