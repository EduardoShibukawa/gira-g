package extractors

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// ProtobufExtractor extracts Contract nodes and RPC methods from .proto files.
type ProtobufExtractor struct{}

// NewProtobufExtractor creates a new Layer 3 Protobuf extractor.
func NewProtobufExtractor() *ProtobufExtractor {
	return &ProtobufExtractor{}
}

func (e *ProtobufExtractor) Name() string {
	return "Layer3Protobuf"
}

func (e *ProtobufExtractor) Layer() Layer {
	return LayerLanguageAST
}

// CanExtract matches protobuf (.proto) files.
func (e *ProtobufExtractor) CanExtract(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".proto"
}

var (
	protoPackageRegex = regexp.MustCompile(`^\s*package\s+([A-Za-z0-9_.]+)\s*;`)
	protoSyntaxRegex  = regexp.MustCompile(`^\s*syntax\s*=\s*["']([^"']+)["']\s*;`)
	protoServiceRegex = regexp.MustCompile(`\bservice\s+([A-Za-z0-9_]+)\s*\{?`)
	protoRpcRegex     = regexp.MustCompile(`\brpc\s+([A-Za-z0-9_]+)\s*\(\s*(?:stream\s+)?([A-Za-z0-9_.]+)\s*\)\s*returns\s*\(\s*(?:stream\s+)?([A-Za-z0-9_.]+)\s*\)`)
	protoMessageRegex = regexp.MustCompile(`\bmessage\s+([A-Za-z0-9_]+)\s*\{?`)
)

type protoMethodInfo struct {
	Name     string `json:"name"`
	Request  string `json:"request"`
	Response string `json:"response"`
}

type protoServiceInfo struct {
	Name      string
	StartLine int
	EndLine   int
	Methods   []protoMethodInfo
}

// Extract parses a .proto file and produces Contract domain nodes.
func (e *ProtobufExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	lines := strings.Split(string(file.Content), "\n")
	discoveredAt := time.Now().UTC()

	var packageName string
	var syntax string
	var messages []string
	var services []protoServiceInfo

	var currentService *protoServiceInfo
	var serviceBraceDepth int
	inBlockComment := false

	for idx, rawLine := range lines {
		lineNum := idx + 1
		line := rawLine

		// Handle block comments /* ... */
		if inBlockComment {
			if endIdx := strings.Index(line, "*/"); endIdx != -1 {
				inBlockComment = false
				line = line[endIdx+2:]
			} else {
				continue
			}
		}

		for {
			startIdx := strings.Index(line, "/*")
			if startIdx == -1 {
				break
			}
			endIdx := strings.Index(line[startIdx+2:], "*/")
			if endIdx == -1 {
				inBlockComment = true
				line = line[:startIdx]
				break
			} else {
				line = line[:startIdx] + line[startIdx+2+endIdx+2:]
			}
		}

		// Strip single line comment //
		if commentIdx := strings.Index(line, "//"); commentIdx != -1 {
			line = line[:commentIdx]
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Detect syntax
		if syntax == "" {
			if m := protoSyntaxRegex.FindStringSubmatch(trimmed); len(m) > 1 {
				syntax = m[1]
			}
		}

		// Detect package
		if packageName == "" {
			if m := protoPackageRegex.FindStringSubmatch(trimmed); len(m) > 1 {
				packageName = m[1]
			}
		}

		// Detect message definitions outside service
		if currentService == nil {
			if m := protoMessageRegex.FindStringSubmatch(trimmed); len(m) > 1 {
				messages = append(messages, m[1])
			}
		}

		// Check service start
		if currentService == nil {
			if m := protoServiceRegex.FindStringSubmatch(trimmed); len(m) > 1 {
				svcName := m[1]
				currentService = &protoServiceInfo{
					Name:      svcName,
					StartLine: lineNum,
					Methods:   make([]protoMethodInfo, 0),
				}
				serviceBraceDepth = 0
				for _, ch := range trimmed {
					if ch == '{' {
						serviceBraceDepth++
					}
				}
				continue
			}
		} else {
			// Inside service block
			// Extract RPC methods
			if rpcMatches := protoRpcRegex.FindStringSubmatch(trimmed); len(rpcMatches) > 3 {
				currentService.Methods = append(currentService.Methods, protoMethodInfo{
					Name:     rpcMatches[1],
					Request:  rpcMatches[2],
					Response: rpcMatches[3],
				})
			}

			// Update brace depth
			for _, ch := range trimmed {
				if ch == '{' {
					serviceBraceDepth++
				} else if ch == '}' {
					serviceBraceDepth--
					if serviceBraceDepth == 0 {
						currentService.EndLine = lineNum
						services = append(services, *currentService)
						currentService = nil
						break
					}
				}
			}
		}
	}

	if currentService != nil {
		return nil, nil, nil, fmt.Errorf("malformed protobuf in %s: unclosed service %s at EOF", file.Path, currentService.Name)
	}

	var nodes []domain.Node

	if len(services) > 0 {
		for _, svc := range services {
			methodNames := make([]string, len(svc.Methods))
			rpcDefs := make([]map[string]string, len(svc.Methods))
			for i, m := range svc.Methods {
				methodNames[i] = m.Name
				rpcDefs[i] = map[string]string{
					"name":     m.Name,
					"request":  m.Request,
					"response": m.Response,
				}
			}

			meta := map[string]any{
				"syntax":          syntax,
				"methods":         methodNames,
				"rpc_definitions": rpcDefs,
			}
			if packageName != "" {
				meta["package"] = packageName
				meta["full_name"] = fmt.Sprintf("%s.%s", packageName, svc.Name)
			}
			if len(messages) > 0 {
				meta["messages"] = messages
			}

			node := domain.Node{
				ID:       fmt.Sprintf("contract:%s", svc.Name),
				Kind:     domain.NodeContract,
				Name:     svc.Name,
				Metadata: meta,
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      file.Repo,
							CommitSHA: file.CommitSHA,
							Path:      file.Path,
							StartLine: svc.StartLine,
							EndLine:   svc.EndLine,
						},
						Origin:            domain.OriginDeclared,
						DeploymentProfile: "protobuf",
						Status:            domain.StatusCurrent,
						DiscoveredAt:      discoveredAt,
					},
				},
			}
			nodes = append(nodes, node)
		}
	} else if len(messages) > 0 || packageName != "" {
		// Schema contract for pure message / package definitions without RPC service
		baseName := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
		contractName := baseName
		if packageName != "" {
			contractName = packageName
		}

		meta := map[string]any{
			"syntax":   syntax,
			"messages": messages,
		}
		if packageName != "" {
			meta["package"] = packageName
		}

		endLine := len(lines)
		if endLine < 1 {
			endLine = 1
		}

		node := domain.Node{
			ID:       fmt.Sprintf("contract:%s", contractName),
			Kind:     domain.NodeContract,
			Name:     contractName,
			Metadata: meta,
			Evidence: []domain.EvidenceRecord{
				{
					Locator: domain.Locator{
						Repo:      file.Repo,
						CommitSHA: file.CommitSHA,
						Path:      file.Path,
						StartLine: 1,
						EndLine:   endLine,
					},
					Origin:            domain.OriginDeclared,
					DeploymentProfile: "protobuf",
					Status:            domain.StatusCurrent,
					DiscoveredAt:      discoveredAt,
				},
			},
		}
		nodes = append(nodes, node)
	}

	return nodes, nil, nil, nil
}
