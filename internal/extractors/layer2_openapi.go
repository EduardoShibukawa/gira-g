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

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9_-]+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// OpenAPIExtractor extracts REST API contracts, endpoints, schemas, and relationships from OpenAPI and Swagger specs.
type OpenAPIExtractor struct{}

// NewOpenAPIExtractor creates a new Layer 2 OpenAPI extractor.
func NewOpenAPIExtractor() *OpenAPIExtractor {
	return &OpenAPIExtractor{}
}

func (e *OpenAPIExtractor) Name() string {
	return "Layer2OpenAPI"
}

func (e *OpenAPIExtractor) Layer() Layer {
	return LayerToolMeaning
}

// CanExtract matches OpenAPI / Swagger files by filename or path.
func (e *OpenAPIExtractor) CanExtract(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(base))
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		return false
	}

	return strings.Contains(base, "openapi") ||
		strings.Contains(base, "swagger") ||
		base == "api.yaml" || base == "api.yml" || base == "api.json" ||
		base == "api-spec.yaml" || base == "api-spec.json"
}

var httpMethods = map[string]bool{
	"get":     true,
	"post":    true,
	"put":     true,
	"delete":  true,
	"patch":   true,
	"options": true,
	"head":    true,
	"trace":   true,
}

// Extract parses the OpenAPI or Swagger specification in YAML or JSON format.
func (e *OpenAPIExtractor) Extract(ctx context.Context, file TargetFile) ([]domain.Node, []domain.Relationship, []domain.UnsupportedBoundary, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(file.Content, &root); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse OpenAPI spec in %s: %w", file.Path, err)
	}

	if len(root.Content) == 0 {
		return nil, nil, nil, nil
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, nil, nil, nil
	}

	var openapiVersion string
	var swaggerVersion string
	var infoNode *yaml.Node
	var pathsNode *yaml.Node
	var componentsNode *yaml.Node
	var definitionsNode *yaml.Node

	for i := 0; i < len(doc.Content); i += 2 {
		k := doc.Content[i]
		v := doc.Content[i+1]

		switch k.Value {
		case "openapi":
			openapiVersion = v.Value
		case "swagger":
			swaggerVersion = v.Value
		case "info":
			infoNode = v
		case "paths":
			pathsNode = v
		case "components":
			componentsNode = v
		case "definitions":
			definitionsNode = v
		}
	}

	// Must declare openapi (3.*) or swagger (2.0)
	if openapiVersion == "" && swaggerVersion == "" {
		return nil, nil, nil, nil
	}

	specProfile := "openapi"
	specVersion := openapiVersion
	if swaggerVersion != "" {
		specProfile = "swagger"
		specVersion = swaggerVersion
	}

	var title string
	var apiDescription string
	var apiVersion string

	if infoNode != nil && infoNode.Kind == yaml.MappingNode {
		for i := 0; i < len(infoNode.Content); i += 2 {
			k := infoNode.Content[i]
			v := infoNode.Content[i+1]
			switch k.Value {
			case "title":
				title = v.Value
			case "description":
				apiDescription = v.Value
			case "version":
				apiVersion = v.Value
			}
		}
	}

	if title == "" {
		base := filepath.Base(file.Path)
		ext := filepath.Ext(base)
		title = strings.TrimSuffix(base, ext)
	}

	discoveredAt := time.Now().UTC()
	var nodes []domain.Node
	var relationships []domain.Relationship

	// Primary API Service node providing the contracts
	apiSlug := slugify(title)
	if apiSlug == "" {
		apiSlug = "api"
	}
	apiNodeID := fmt.Sprintf("service:%s", apiSlug)
	infoStart, infoEnd := 1, 1
	if infoNode != nil {
		infoStart, infoEnd = getYamlNodeLineRange(infoNode)
	}

	apiMeta := map[string]any{
		"title":        title,
		"spec_format":  specProfile,
		"spec_version": specVersion,
		"file":         file.Path,
	}
	if apiVersion != "" {
		apiMeta["api_version"] = apiVersion
	}

	apiNode := domain.Node{
		ID:          apiNodeID,
		Kind:        domain.NodeService,
		Name:        title,
		Description: apiDescription,
		Metadata:    apiMeta,
		Evidence: []domain.EvidenceRecord{
			{
				Locator: domain.Locator{
					Repo:      file.Repo,
					CommitSHA: file.CommitSHA,
					Path:      file.Path,
					StartLine: infoStart,
					EndLine:   infoEnd,
				},
				Origin:            domain.OriginExtracted,
				DeploymentProfile: specProfile,
				Status:            domain.StatusCurrent,
				DiscoveredAt:      discoveredAt,
			},
		},
	}
	nodes = append(nodes, apiNode)

	// Process Paths & Endpoints
	if pathsNode != nil && pathsNode.Kind == yaml.MappingNode {
		for i := 0; i < len(pathsNode.Content); i += 2 {
			select {
			case <-ctx.Done():
				return nil, nil, nil, ctx.Err()
			default:
			}

			pathKey := pathsNode.Content[i]
			pathItem := pathsNode.Content[i+1]
			pathStr := pathKey.Value

			if pathItem.Kind != yaml.MappingNode {
				continue
			}

			for j := 0; j < len(pathItem.Content); j += 2 {
				verbKey := pathItem.Content[j]
				opNode := pathItem.Content[j+1]

				verb := strings.ToLower(verbKey.Value)
				if !httpMethods[verb] {
					continue
				}

				var summary string
				var opDesc string
				var operationID string
				var tags []string

				if opNode.Kind == yaml.MappingNode {
					for k := 0; k < len(opNode.Content); k += 2 {
						okKey := opNode.Content[k]
						okVal := opNode.Content[k+1]
						switch okKey.Value {
						case "summary":
							summary = okVal.Value
						case "description":
							opDesc = okVal.Value
						case "operationId":
							operationID = okVal.Value
						case "tags":
							if okVal.Kind == yaml.SequenceNode {
								for _, tagItem := range okVal.Content {
									tags = append(tags, tagItem.Value)
								}
							}
						}
					}
				}

				opStart, opEnd := getYamlNodeLineRange(opNode)
				if verbKey.Line > 0 && verbKey.Line < opStart {
					opStart = verbKey.Line
				}

				contractID := fmt.Sprintf("contract:%s:%s", pathStr, verb)
				contractName := fmt.Sprintf("%s %s", strings.ToUpper(verb), pathStr)
				desc := summary
				if desc == "" {
					desc = opDesc
				}

				contractMeta := map[string]any{
					"path":         pathStr,
					"method":       strings.ToUpper(verb),
					"api":          title,
					"spec_version": specVersion,
				}
				if operationID != "" {
					contractMeta["operation_id"] = operationID
				}
				if len(tags) > 0 {
					contractMeta["tags"] = tags
				}

				contractNode := domain.Node{
					ID:          contractID,
					Kind:        domain.NodeContract,
					Name:        contractName,
					Description: desc,
					Metadata:    contractMeta,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: opStart,
								EndLine:   opEnd,
							},
							Origin:            domain.OriginExtracted,
							DeploymentProfile: specProfile,
							Status:            domain.StatusCurrent,
							DiscoveredAt:      discoveredAt,
						},
					},
				}
				nodes = append(nodes, contractNode)

				// Relationship: Service -> Contract (Provides)
				relID := fmt.Sprintf("rel:%s-provides-%s", sanitizeID(apiNode.ID), sanitizeID(contractID))
				rel := domain.Relationship{
					ID:       relID,
					SourceID: apiNode.ID,
					TargetID: contractID,
					Verb:     domain.VerbProvides,
					Evidence: []domain.EvidenceRecord{
						{
							Locator: domain.Locator{
								Repo:      file.Repo,
								CommitSHA: file.CommitSHA,
								Path:      file.Path,
								StartLine: opStart,
								EndLine:   opEnd,
							},
							Origin:            domain.OriginExtracted,
							DeploymentProfile: specProfile,
							Status:            domain.StatusCurrent,
							DiscoveredAt:      discoveredAt,
						},
					},
				}
				relationships = append(relationships, rel)
			}
		}
	}

	// Process Schemas (OpenAPI 3 components.schemas or Swagger 2 definitions)
	var schemasNode *yaml.Node
	if componentsNode != nil && componentsNode.Kind == yaml.MappingNode {
		for i := 0; i < len(componentsNode.Content); i += 2 {
			if componentsNode.Content[i].Value == "schemas" {
				schemasNode = componentsNode.Content[i+1]
				break
			}
		}
	} else if definitionsNode != nil && definitionsNode.Kind == yaml.MappingNode {
		schemasNode = definitionsNode
	}

	if schemasNode != nil && schemasNode.Kind == yaml.MappingNode {
		for i := 0; i < len(schemasNode.Content); i += 2 {
			schemaKey := schemasNode.Content[i]
			schemaVal := schemasNode.Content[i+1]

			schemaName := schemaKey.Value
			schemaStart, schemaEnd := getYamlNodeLineRange(schemaVal)
			if schemaKey.Line > 0 && schemaKey.Line < schemaStart {
				schemaStart = schemaKey.Line
			}

			schemaID := fmt.Sprintf("contract:schema:%s", schemaName)
			schemaMeta := map[string]any{
				"schema": schemaName,
				"api":    title,
			}

			schemaNode := domain.Node{
				ID:          schemaID,
				Kind:        domain.NodeContract,
				Name:        schemaName,
				Description: fmt.Sprintf("Data schema definition %s", schemaName),
				Metadata:    schemaMeta,
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      file.Repo,
							CommitSHA: file.CommitSHA,
							Path:      file.Path,
							StartLine: schemaStart,
							EndLine:   schemaEnd,
						},
						Origin:            domain.OriginExtracted,
						DeploymentProfile: specProfile,
						Status:            domain.StatusCurrent,
						DiscoveredAt:      discoveredAt,
					},
				},
			}
			nodes = append(nodes, schemaNode)

			// Relationship: Service -> Schema (Provides)
			relID := fmt.Sprintf("rel:%s-provides-%s", sanitizeID(apiNode.ID), sanitizeID(schemaID))
			rel := domain.Relationship{
				ID:       relID,
				SourceID: apiNode.ID,
				TargetID: schemaID,
				Verb:     domain.VerbProvides,
				Evidence: []domain.EvidenceRecord{
					{
						Locator: domain.Locator{
							Repo:      file.Repo,
							CommitSHA: file.CommitSHA,
							Path:      file.Path,
							StartLine: schemaStart,
							EndLine:   schemaEnd,
						},
						Origin:            domain.OriginExtracted,
						DeploymentProfile: specProfile,
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
