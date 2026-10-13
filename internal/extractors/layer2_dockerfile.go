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

// DockerfileExtractor extracts services, base image libraries/infrastructures, and build stages from Dockerfiles.
type DockerfileExtractor struct{}

// NewDockerfileExtractor creates a new Layer 2 Dockerfile extractor.
func NewDockerfileExtractor() *DockerfileExtractor {
	return &DockerfileExtractor{}
}

func (e *DockerfileExtractor) Name() string {
	return "Layer2Dockerfile"
}

func (e *DockerfileExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches Dockerfiles by name or extension patterns:
// base name "dockerfile", begins with "dockerfile.", ends with ".dockerfile" (case-insensitive).
func (e *DockerfileExtractor) CanExtract(path string) bool {
	normalized := filepath.ToSlash(path)
	base := strings.ToLower(filepath.Base(normalized))
	if base == "dockerfile" {
		return true
	}
	if strings.HasPrefix(base, "dockerfile.") {
		return true
	}
	if strings.HasSuffix(base, ".dockerfile") {
		return true
	}
	return false
}

type dockerfileCopy struct {
	target    string
	startLine int
	endLine   int
}

type dockerfileStage struct {
	baseImage     string
	stageName     string
	fromStartLine int
	fromEndLine   int
	ports         []string
	cmd           string
	entrypoint    string
	copiedFrom    []dockerfileCopy
}

// parseDockerfileLines handles line continuation (backslash `\`), comments `#`, and blank lines.
type parsedInstruction struct {
	directive string
	args      string
	startLine int
	endLine   int
}

func parseDockerfileInstructions(content []byte) []parsedInstruction {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var instructions []parsedInstruction

	var currentDirective string
	var currentArgs strings.Builder
	var currentStartLine int
	lineNum := 0
	inContinuation := false

	for scanner.Scan() {
		lineNum++
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		// Comments outside continuation
		if !inContinuation && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			continue
		}

		// Check if line ends with continuation character `\`
		hasContinuation := strings.HasSuffix(trimmed, "\\")
		contentLine := trimmed
		if hasContinuation {
			contentLine = strings.TrimSpace(strings.TrimSuffix(trimmed, "\\"))
		}

		if !inContinuation {
			// First line of an instruction
			currentStartLine = lineNum
			fields := strings.Fields(contentLine)
			if len(fields) == 0 {
				continue
			}
			currentDirective = strings.ToUpper(fields[0])
			rest := strings.TrimSpace(strings.TrimPrefix(contentLine, fields[0]))
			currentArgs.Reset()
			currentArgs.WriteString(rest)

			if hasContinuation {
				inContinuation = true
			} else {
				instructions = append(instructions, parsedInstruction{
					directive: currentDirective,
					args:      strings.TrimSpace(currentArgs.String()),
					startLine: currentStartLine,
					endLine:   lineNum,
				})
			}
		} else {
			// Continuing previous instruction
			if currentArgs.Len() > 0 {
				currentArgs.WriteString(" ")
			}
			currentArgs.WriteString(contentLine)

			if !hasContinuation {
				inContinuation = false
				instructions = append(instructions, parsedInstruction{
					directive: currentDirective,
					args:      strings.TrimSpace(currentArgs.String()),
					startLine: currentStartLine,
					endLine:   lineNum,
				})
			}
		}
	}

	// If file ended while still in continuation
	if inContinuation {
		instructions = append(instructions, parsedInstruction{
			directive: currentDirective,
			args:      strings.TrimSpace(currentArgs.String()),
			startLine: currentStartLine,
			endLine:   lineNum,
		})
	}

	return instructions
}

// deriveDefaultServiceName determines a default service name from path (e.g. Dockerfile -> "app", frontend/Dockerfile -> "frontend", Dockerfile.dev -> "app-dev").
func deriveDefaultServiceName(path string) string {
	normalized := filepath.ToSlash(path)
	dir := filepath.Dir(normalized)
	base := filepath.Base(normalized)
	baseLower := strings.ToLower(base)

	dirName := ""
	if dir != "" && dir != "." && dir != "/" {
		parts := strings.Split(dir, "/")
		dirName = parts[len(parts)-1]
	}

	if strings.HasPrefix(baseLower, "dockerfile.") {
		suffix := base[len("dockerfile."):]
		if dirName != "" {
			return fmt.Sprintf("%s-%s", dirName, suffix)
		}
		return suffix
	}
	if strings.HasSuffix(baseLower, ".dockerfile") {
		prefix := base[:len(base)-len(".dockerfile")]
		if prefix != "" {
			return prefix
		}
	}

	if dirName != "" {
		return dirName
	}
	return "app"
}

// parseFromArgs parses `FROM [--platform=...] <image> [AS <stage>]`
func parseFromArgs(args string) (baseImage string, stageName string) {
	fields := strings.Fields(args)
	var nonFlags []string
	for _, f := range fields {
		if strings.HasPrefix(f, "--") {
			continue
		}
		nonFlags = append(nonFlags, f)
	}

	if len(nonFlags) == 0 {
		return "", ""
	}

	baseImage = nonFlags[0]

	for i := 1; i < len(nonFlags); i++ {
		if strings.EqualFold(nonFlags[i], "as") && i+1 < len(nonFlags) {
			stageName = nonFlags[i+1]
			break
		}
	}

	return baseImage, stageName
}

// extractCopyFrom extracts the stage/image from `COPY --from=<stage|image>`
func extractCopyFrom(args string) string {
	fields := strings.Fields(args)
	for _, f := range fields {
		if strings.HasPrefix(f, "--from=") {
			return strings.TrimPrefix(f, "--from=")
		}
	}
	return ""
}

// cleanImageName removes flags, tags or digests if needed or keeps the canonical repository/image tag.
// For node IDs, library:<image> or infra:image:<image>.
func cleanImageName(img string) string {
	return strings.TrimSpace(img)
}

// Known infrastructure base image patterns
var knownDockerfileInfraImages = []string{
	"postgres", "postgresql", "redis", "valkey", "kafka", "zookeeper",
	"jaeger", "otelcol", "otel-collector", "opentelemetry-collector",
	"prometheus", "grafana", "flagd", "elasticsearch", "rabbitmq",
	"mysql", "mariadb", "mongo", "mongodb", "cassandra", "memcached",
	"nginx", "caddy", "haproxy", "traefik", "vault", "consul",
}

func isDockerfileInfraImage(img string) bool {
	lower := strings.ToLower(img)
	// Strip repo prefix (e.g. library/postgres:16 -> postgres:16)
	parts := strings.Split(lower, "/")
	base := parts[len(parts)-1]
	// Strip tag or digest
	base = strings.Split(base, ":")[0]
	base = strings.Split(base, "@")[0]

	for _, kw := range knownDockerfileInfraImages {
		if base == kw || strings.Contains(base, kw) {
			return true
		}
	}
	return false
}

// Extract parses the Dockerfile and generates nodes and relationships.
func (e *DockerfileExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	instructions := parseDockerfileInstructions(file.Content)
	if len(instructions) == 0 {
		return nil, nil, nil, nil
	}

	defaultSvcName := deriveDefaultServiceName(file.Path)
	discoveredAt := time.Now().UTC()
	deploymentProfile := "dockerfile"

	var stages []*dockerfileStage
	var currentStage *dockerfileStage

	for _, inst := range instructions {
		select {
		case <-ctx.Done():
			return nil, nil, nil, ctx.Err()
		default:
		}

		switch inst.directive {
		case "FROM":
			baseImg, stName := parseFromArgs(inst.args)
			currentStage = &dockerfileStage{
				baseImage:     baseImg,
				stageName:     stName,
				fromStartLine: inst.startLine,
				fromEndLine:   inst.endLine,
			}
			stages = append(stages, currentStage)

		case "EXPOSE":
			if currentStage != nil {
				ports := strings.Fields(inst.args)
				currentStage.ports = append(currentStage.ports, ports...)
			}

		case "CMD":
			if currentStage != nil {
				currentStage.cmd = inst.args
			}

		case "ENTRYPOINT":
			if currentStage != nil {
				currentStage.entrypoint = inst.args
			}

		case "COPY":
			if currentStage != nil {
				fromTarget := extractCopyFrom(inst.args)
				if fromTarget != "" {
					currentStage.copiedFrom = append(currentStage.copiedFrom, dockerfileCopy{
						target:    fromTarget,
						startLine: inst.startLine,
						endLine:   inst.endLine,
					})
				}
			}
		}
	}

	if len(stages) == 0 {
		return nil, nil, nil, nil
	}

	var nodes []domain.Node
	var relationships []domain.Relationship
	nodeMap := make(map[string]int)
	relMap := make(map[string]int)

	addNode := func(n domain.Node) {
		if idx, ok := nodeMap[n.ID]; ok {
			nodes[idx].Evidence = append(nodes[idx].Evidence, n.Evidence...)
			if nodes[idx].Metadata == nil && len(n.Metadata) > 0 {
				nodes[idx].Metadata = make(map[string]any)
			}
			for k, v := range n.Metadata {
				if _, has := nodes[idx].Metadata[k]; !has {
					nodes[idx].Metadata[k] = v
				}
			}
		} else {
			nodes = append(nodes, n)
			nodeMap[n.ID] = len(nodes) - 1
		}
	}

	addRel := func(r domain.Relationship) {
		if idx, ok := relMap[r.ID]; ok {
			relationships[idx].Evidence = append(relationships[idx].Evidence, r.Evidence...)
		} else {
			relationships = append(relationships, r)
			relMap[r.ID] = len(relationships) - 1
		}
	}

	// Map of stage names defined so far: stageName -> nodeID
	definedStages := make(map[string]string)

	numStages := len(stages)

	for i, st := range stages {
		isFinalStage := (i == numStages-1)

		var nodeID string
		var nodeKind domain.NodeType
		var nodeName string

		if isFinalStage {
			nodeKind = domain.NodeService
			if st.stageName != "" && numStages > 1 {
				// Final stage has an explicit AS name
				nodeName = st.stageName
				nodeID = fmt.Sprintf("service:%s", st.stageName)
			} else {
				nodeName = defaultSvcName
				nodeID = fmt.Sprintf("service:%s", defaultSvcName)
			}
		} else {
			// Intermediate stage
			if st.stageName != "" {
				nodeName = st.stageName
				nodeID = fmt.Sprintf("service:stage:%s", st.stageName)
			} else {
				nodeName = fmt.Sprintf("stage-%d", i)
				nodeID = fmt.Sprintf("service:stage:stage-%d", i)
			}
			nodeKind = domain.NodeService
		}

		if st.stageName != "" {
			definedStages[st.stageName] = nodeID
		}
		// Also register index as stage reference
		definedStages[fmt.Sprintf("%d", i)] = nodeID

		metadata := make(map[string]any)
		if st.baseImage != "" {
			metadata["base_image"] = st.baseImage
		}
		if len(st.ports) > 0 {
			metadata["ports"] = st.ports
		}
		if st.entrypoint != "" {
			metadata["entrypoint"] = st.entrypoint
		}
		if st.cmd != "" {
			metadata["cmd"] = st.cmd
		}
		if st.stageName != "" {
			metadata["stage"] = st.stageName
		}

		stageNode := domain.Node{
			ID:       nodeID,
			Kind:     nodeKind,
			Name:     nodeName,
			Metadata: metadata,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: st.fromStartLine,
						EndLine:   st.fromEndLine,
					},
					Origin:            domain.OriginExtracted,
					DeploymentProfile: deploymentProfile,
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		addNode(stageNode)

		// Base image dependency / relation
		if st.baseImage != "" {
			cleanImg := cleanImageName(st.baseImage)
			priorStageNodeID, isPriorStage := definedStages[cleanImg]

			if isPriorStage {
				// FROM references a previous stage
				rel := domain.Relationship{
					ID:       fmt.Sprintf("rel:%s-%s-%s", nodeID, string(domain.VerbConsumes), priorStageNodeID),
					SourceID: nodeID,
					TargetID: priorStageNodeID,
					Verb:     domain.VerbConsumes,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: st.fromStartLine,
								EndLine:   st.fromEndLine,
							},
							Origin:            domain.OriginExtracted,
							DeploymentProfile: deploymentProfile,
							Status:            domain.StatusCurrent,
							DiscoveredAt:      discoveredAt,
						},
					},
				}
				addRel(rel)
			} else {
				// External base image (Library or Infrastructure)
				var baseNodeID string
				var baseKind domain.NodeType

				if isDockerfileInfraImage(cleanImg) {
					baseNodeID = fmt.Sprintf("infra:image:%s", cleanImg)
					baseKind = domain.NodeInfrastructure
				} else {
					baseNodeID = fmt.Sprintf("library:%s", cleanImg)
					baseKind = domain.NodeLibrary
				}

				baseNode := domain.Node{
					ID:   baseNodeID,
					Kind: baseKind,
					Name: cleanImg,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: st.fromStartLine,
								EndLine:   st.fromEndLine,
							},
							Origin:            domain.OriginExtracted,
							DeploymentProfile: deploymentProfile,
							Status:            domain.StatusCurrent,
							DiscoveredAt:      discoveredAt,
						},
					},
				}
				addNode(baseNode)

				rel := domain.Relationship{
					ID:       fmt.Sprintf("rel:%s-%s-%s", nodeID, string(domain.VerbConsumes), baseNodeID),
					SourceID: nodeID,
					TargetID: baseNodeID,
					Verb:     domain.VerbConsumes,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: st.fromStartLine,
								EndLine:   st.fromEndLine,
							},
							Origin:            domain.OriginExtracted,
							DeploymentProfile: deploymentProfile,
							Status:            domain.StatusCurrent,
							DiscoveredAt:      discoveredAt,
						},
					},
				}
				addRel(rel)
			}
		}

		// Process COPY --from=...
		for _, copyItem := range st.copiedFrom {
			targetID, ok := definedStages[copyItem.target]
			if !ok {
				// Could be an external image in COPY --from=image:tag
				if isDockerfileInfraImage(copyItem.target) {
					targetID = fmt.Sprintf("infra:image:%s", copyItem.target)
					addNode(domain.Node{
						ID:   targetID,
						Kind: domain.NodeInfrastructure,
						Name: copyItem.target,
						Evidence: []domain.EvidenceRecord{
							{
								Locator: domain.Locator{
									Repo:      file.Repo,
									CommitSHA: file.CommitSHA,
									Path:      file.Path,
									StartLine: copyItem.startLine,
									EndLine:   copyItem.endLine,
								},
								Origin:            domain.OriginExtracted,
								DeploymentProfile: deploymentProfile,
								Status:            domain.StatusCurrent,
								DiscoveredAt:      discoveredAt,
							},
						},
					})
				} else {
					targetID = fmt.Sprintf("library:%s", copyItem.target)
					addNode(domain.Node{
						ID:   targetID,
						Kind: domain.NodeLibrary,
						Name: copyItem.target,
						Evidence: []domain.EvidenceRecord{
							{
								Locator: domain.Locator{
									Repo:      file.Repo,
									CommitSHA: file.CommitSHA,
									Path:      file.Path,
									StartLine: copyItem.startLine,
									EndLine:   copyItem.endLine,
								},
								Origin:            domain.OriginExtracted,
								DeploymentProfile: deploymentProfile,
								Status:            domain.StatusCurrent,
								DiscoveredAt:      discoveredAt,
							},
						},
					})
				}
			}

			rel := domain.Relationship{
				ID:       fmt.Sprintf("rel:%s-%s-%s", nodeID, string(domain.VerbConsumes), targetID),
				SourceID: nodeID,
				TargetID: targetID,
				Verb:     domain.VerbConsumes,
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      file.Repo,
							CommitSHA: file.CommitSHA,
							Path:      file.Path,
							StartLine: copyItem.startLine,
							EndLine:   copyItem.endLine,
						},
						Origin:            domain.OriginExtracted,
						DeploymentProfile: deploymentProfile,
						Status:            domain.StatusCurrent,
						DiscoveredAt:      discoveredAt,
					},
				},
			}
			addRel(rel)
		}
	}

	return nodes, relationships, nil, nil
}

