package extractors

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
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

type dockerfilePort struct {
	port      string
	startLine int
	endLine   int
}

type dockerfileStage struct {
	exposed       []dockerfilePort
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
	dir := filepath.ToSlash(filepath.Dir(normalized))
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
				for _, p := range ports {
					currentStage.exposed = append(currentStage.exposed, dockerfilePort{port: p, startLine: inst.startLine, endLine: inst.endLine})
				}
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

	// One Dockerfile builds one Service. Build stages and base images are how it is
	// built, not Components: they are recorded as metadata on the Service, each
	// entry with the locator of the line it was read from.
	// Entries cite path#Lstart[-Lend]: the repo and commit are the Service's own
	// Dockerfile evidence, so repeating them only spends the token budget.
	locate := func(start, end int) string {
		if end > start {
			return fmt.Sprintf("%s#L%d-L%d", file.Path, start, end)
		}
		return fmt.Sprintf("%s#L%d", file.Path, start)
	}
	stageNames := make(map[string]bool)
	var stageMeta, baseImages []map[string]any
	seenImage := make(map[string]bool)
	addImage := func(img string, start, end int) {
		if img == "" || stageNames[img] || seenImage[img] {
			return
		}
		seenImage[img] = true
		baseImages = append(baseImages, map[string]any{"image": img, "locator": locate(start, end)})
	}
	for i, st := range stages {
		img := cleanImageName(st.baseImage)
		name := st.stageName
		if name == "" {
			name = fmt.Sprintf("stage-%d", i)
		}
		if len(stages) > 1 {
			stageMeta = append(stageMeta, map[string]any{"name": name, "locator": locate(st.fromStartLine, st.fromEndLine)})
		}
		addImage(img, st.fromStartLine, st.fromEndLine)
		for _, c := range st.copiedFrom {
			if _, numeric := strconv.Atoi(c.target); numeric != nil {
				addImage(c.target, c.startLine, c.endLine)
			}
		}
		stageNames[st.stageName] = st.stageName != ""
		stageNames[fmt.Sprintf("%d", i)] = true
	}

	final := stages[len(stages)-1]
	metadata := make(map[string]any)
	if final.baseImage != "" {
		metadata["base_image"] = final.baseImage
	}
	if len(baseImages) > 0 {
		metadata["base_images"] = baseImages
	}
	if len(stageMeta) > 0 {
		metadata["stages"] = stageMeta
	}
	if len(final.ports) > 0 {
		metadata["ports"] = final.ports
		var exposed []map[string]any
		for _, p := range final.exposed {
			exposed = append(exposed, map[string]any{"port": p.port, "locator": locate(p.startLine, p.endLine)})
		}
		metadata["exposed_ports"] = exposed
	}
	if final.entrypoint != "" {
		metadata["entrypoint"] = final.entrypoint
	}
	if final.cmd != "" {
		metadata["cmd"] = final.cmd
	}

	nodes := []domain.Node{{
		ID:       fmt.Sprintf("service:%s", defaultSvcName),
		Kind:     domain.NodeService,
		Name:     defaultSvcName,
		Metadata: metadata,
		Evidence: []domain.EvidenceRecord{{
			Locator: domain.Locator{
				Repo:      file.Repo,
				CommitSHA: file.CommitSHA,
				Path:      file.Path,
				StartLine: final.fromStartLine,
				EndLine:   final.fromEndLine,
			},
			Origin:            domain.OriginExtracted,
			DeploymentProfile: deploymentProfile,
			Status:            domain.StatusCurrent,
			DiscoveredAt:      discoveredAt,
		}},
	}}
	return nodes, nil, nil, nil
}

// Hints reports which Service node the Dockerfile produced, so the resolution
// step can rename it to the compose service that builds this file (ADR 0013).
func (e *DockerfileExtractor) Hints(ctx context.Context, file TargetFile) ([]Hint, error) {
	if len(parseDockerfileInstructions(file.Content)) == 0 {
		return nil, nil
	}
	return []Hint{{
		Kind:    HintDockerfileService,
		Locator: domain.Locator{Repo: file.Repo, CommitSHA: file.CommitSHA, Path: file.Path},
		Service: "service:" + deriveDefaultServiceName(file.Path),
	}}, nil
}
