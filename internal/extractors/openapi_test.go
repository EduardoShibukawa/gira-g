package extractors

import (
	"context"
	"strings"
	"testing"

	"labs/gira-g/internal/domain"
)

func TestOpenAPIExtractor_MetadataAndCanExtract(t *testing.T) {
	ext := NewOpenAPIExtractor()

	if ext.Name() != "Layer2OpenAPI" {
		t.Fatalf("expected name 'Layer2OpenAPI', got '%s'", ext.Name())
	}

	if ext.Layer() != LayerToolMeaning {
		t.Fatalf("expected layer %d, got %d", LayerToolMeaning, ext.Layer())
	}

	testCases := []struct {
		path     string
		expected bool
	}{
		{"api/openapi.yaml", true},
		{"specs/openapi.json", true},
		{"docs/swagger.yaml", true},
		{"swagger.json", true},
		{"api.yaml", true},
		{"api.json", true},
		{"service-openapi-v3.yml", true},
		{"other.yaml", false},
		{"main.go", false},
		{"openapi.txt", false},
	}

	for _, tc := range testCases {
		res := ext.CanExtract(tc.path)
		if res != tc.expected {
			t.Errorf("CanExtract(%q) = %v, expected %v", tc.path, res, tc.expected)
		}
	}
}

func TestOpenAPIExtractor_ExtractOpenAPI3YAML(t *testing.T) {
	ext := NewOpenAPIExtractor()

	yamlSpec := `openapi: 3.0.3
info:
  title: PetStore Service
  version: 1.0.0
  description: Sample Petstore API

paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      responses:
        '200':
          description: A paged array of pets
    post:
      summary: Create a pet
      operationId: createPet
      tags:
        - pets
  /pets/{petId}:
    delete:
      summary: Delete pet by ID
      operationId: deletePet

components:
  schemas:
    Pet:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "sha987",
		Path:      "api/openapi.yaml",
		Content:   []byte(yamlSpec),
	}

	nodes, rels, boundaries, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}

	if len(boundaries) != 0 {
		t.Fatalf("expected 0 boundaries, got %d", len(boundaries))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// 1. Verify Service Node
	svcNode, ok := nodeMap["service:petstore-service"]
	if !ok {
		t.Fatalf("missing service:petstore-service node")
	}
	if svcNode.Kind != domain.NodeService {
		t.Errorf("expected Service kind, got %s", svcNode.Kind)
	}
	if svcNode.Metadata["spec_format"] != "openapi" {
		t.Errorf("expected spec_format 'openapi', got %v", svcNode.Metadata["spec_format"])
	}

	// 2. Verify Endpoints as Contract Nodes
	getPetNode, ok := nodeMap["contract:/pets:get"]
	if !ok {
		t.Fatalf("missing contract:/pets:get node")
	}
	if getPetNode.Kind != domain.NodeContract {
		t.Errorf("expected Contract kind for /pets:get, got %s", getPetNode.Kind)
	}
	if getPetNode.Name != "GET /pets" {
		t.Errorf("expected name 'GET /pets', got %s", getPetNode.Name)
	}
	if getPetNode.Description != "List all pets" {
		t.Errorf("expected summary 'List all pets', got %s", getPetNode.Description)
	}
	if len(getPetNode.Evidence) == 0 {
		t.Fatalf("missing evidence for /pets:get")
	}
	ev := getPetNode.Evidence[0]
	if ev.DeploymentProfile != "openapi" {
		t.Errorf("expected deployment_profile 'openapi', got %s", ev.DeploymentProfile)
	}
	if ev.Locator.StartLine <= 0 {
		t.Errorf("invalid start line: %d", ev.Locator.StartLine)
	}

	postPetNode, ok := nodeMap["contract:/pets:post"]
	if !ok {
		t.Fatalf("missing contract:/pets:post node")
	}
	if postPetNode.Name != "POST /pets" {
		t.Errorf("expected name 'POST /pets', got %s", postPetNode.Name)
	}

	deletePetNode, ok := nodeMap["contract:/pets/{petId}:delete"]
	if !ok {
		t.Fatalf("missing contract:/pets/{petId}:delete node")
	}
	if deletePetNode.Name != "DELETE /pets/{petId}" {
		t.Errorf("expected name 'DELETE /pets/{petId}', got %s", deletePetNode.Name)
	}

	// 3. Verify Schemas
	petSchema, ok := nodeMap["contract:schema:Pet"]
	if !ok {
		t.Fatalf("missing contract:schema:Pet node")
	}
	if petSchema.Kind != domain.NodeContract {
		t.Errorf("expected schema kind Contract, got %s", petSchema.Kind)
	}

	// 4. Verify Relationships (Provides)
	relMap := make(map[string]domain.Relationship)
	for _, r := range rels {
		relMap[r.ID] = r
	}

	getRelID := "rel:service_petstore-service-provides-contract__pets_get"
	getRel, ok := relMap[getRelID]
	if !ok {
		// Look by source and target
		for _, r := range rels {
			if r.SourceID == "service:petstore-service" && r.TargetID == "contract:/pets:get" {
				getRel = r
				ok = true
				break
			}
		}
	}
	if !ok {
		t.Fatalf("missing provides relationship for /pets:get")
	}
	if getRel.Verb != domain.VerbProvides {
		t.Errorf("expected VerbProvides, got %s", getRel.Verb)
	}
}

func TestOpenAPIExtractor_ExtractSwagger2JSON(t *testing.T) {
	ext := NewOpenAPIExtractor()

	jsonSpec := `{
  "swagger": "2.0",
  "info": {
    "title": "Legacy Swagger API",
    "version": "1.0"
  },
  "paths": {
    "/users": {
      "get": {
        "summary": "Fetch users",
        "operationId": "getUsers"
      }
    }
  },
  "definitions": {
    "User": {
      "type": "object"
    }
  }
}`

	file := TargetFile{
		Repo:      "duh-homelab",
		CommitSHA: "abc",
		Path:      "swagger.json",
		Content:   []byte(jsonSpec),
	}

	nodes, rels, bounds, err := ext.Extract(context.Background(), file)
	if err != nil {
		t.Fatalf("unexpected extract error on JSON spec: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("expected 0 boundaries, got %d", len(bounds))
	}

	nodeMap := make(map[string]domain.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	userNode, ok := nodeMap["contract:/users:get"]
	if !ok {
		t.Fatalf("missing contract:/users:get from Swagger JSON")
	}
	if len(userNode.Evidence) == 0 || userNode.Evidence[0].DeploymentProfile != "swagger" {
		t.Errorf("expected deployment_profile 'swagger', got %+v", userNode.Evidence)
	}

	userDef, ok := nodeMap["contract:schema:User"]
	if !ok {
		t.Fatalf("missing contract:schema:User from Swagger JSON definitions")
	}
	if userDef.Name != "User" {
		t.Errorf("expected schema name User, got %s", userDef.Name)
	}

	var foundProvides bool
	for _, r := range rels {
		if r.TargetID == "contract:/users:get" && r.Verb == domain.VerbProvides {
			foundProvides = true
			break
		}
	}
	if !foundProvides {
		t.Errorf("expected relationship Provides for /users:get")
	}
}

func TestOpenAPIExtractor_EdgeCases(t *testing.T) {
	ext := NewOpenAPIExtractor()

	// Non-OpenAPI YAML (no openapi/swagger key)
	fileNotOpenAPI := TargetFile{
		Repo:    "duh-homelab",
		Path:    "api.yaml",
		Content: []byte("foo: bar\ncount: 42\n"),
	}
	nodes, rels, bounds, err := ext.Extract(context.Background(), fileNotOpenAPI)
	if err != nil || len(nodes) != 0 || len(rels) != 0 || len(bounds) != 0 {
		t.Fatalf("expected nil result on non-OpenAPI YAML, got nodes: %v, err: %v", nodes, err)
	}

	// Empty file
	fileEmpty := TargetFile{
		Repo:    "duh-homelab",
		Path:    "openapi.yaml",
		Content: []byte(""),
	}
	nodes, rels, bounds, err = ext.Extract(context.Background(), fileEmpty)
	if err != nil || len(nodes) != 0 {
		t.Fatalf("expected nil result on empty file, got %v", err)
	}

	// Malformed YAML
	fileMalformed := TargetFile{
		Repo:    "duh-homelab",
		Path:    "openapi.yaml",
		Content: []byte("openapi: 3.0.0\npaths: [unterminated"),
	}
	_, _, _, err = ext.Extract(context.Background(), fileMalformed)
	if err == nil || !strings.Contains(err.Error(), "failed to parse OpenAPI spec") {
		t.Fatalf("expected parse error on malformed YAML, got %v", err)
	}
}
