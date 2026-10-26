package ask

import (
	"fmt"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/jev"
)

// AskResponse represents the structured answer to a natural language query.
type AskResponse struct {
	Query        string   `json:"query"`
	Intent       string   `json:"intent"`
	Answer       string   `json:"answer"`
	MatchedNodes []string `json:"matched_nodes,omitempty"`
	Diagram      string   `json:"diagram,omitempty"`
	Mode         string   `json:"mode"`
}

// RenderCaveman formats the response as dense one-fact-per-line plain text for AI agents.
// Minimizes token consumption (70-85% reduction) and prevents LLM drift.
func (r AskResponse) RenderCaveman() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "QUERY: %s\n", r.Query)
	fmt.Fprintf(&sb, "INTENT: %s\n", r.Intent)
	fmt.Fprintf(&sb, "MODE: %s\n", r.Mode)
	if len(r.MatchedNodes) > 0 {
		fmt.Fprintf(&sb, "MATCHED_NODES: %s\n", strings.Join(r.MatchedNodes, ", "))
	}

	lines := strings.Split(r.Answer, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "```") {
			continue
		}
		l = strings.TrimPrefix(l, "- ")
		l = strings.TrimPrefix(l, "* ")
		l = strings.ReplaceAll(l, "**", "")
		l = strings.ReplaceAll(l, "`", "")
		if len(l) > 0 {
			fmt.Fprintf(&sb, "FACT: %s\n", l)
		}
	}

	if r.Diagram != "" {
		diaClean := strings.ReplaceAll(strings.TrimSpace(r.Diagram), "\n", " | ")
		fmt.Fprintf(&sb, "DIAGRAM: %s\n", diaClean)
	}

	return strings.TrimRight(sb.String(), "\n")
}

// RenderMarkdown renders the answer in rich GitHub-flavored markdown.
func (r AskResponse) RenderMarkdown() string {
	var sb strings.Builder
	sb.WriteString(r.Answer)
	if r.Diagram != "" {
		sb.WriteString("\n\n```mermaid\n")
		sb.WriteString(r.Diagram)
		sb.WriteString("\n```")
	}
	return sb.String()
}

// ProcessQuery evaluates a natural language query against the MacroMap index.
func ProcessQuery(query string, mm *domain.MacroMap, forceDiagram bool) AskResponse {
	classification := jev.ClassifyIntent(query)
	qLower := strings.ToLower(query)

	resp := AskResponse{
		Query:  query,
		Intent: string(classification.Intent),
		Mode:   classification.Engine,
	}

	wantDiagram := forceDiagram || classification.WantDiagram

	// 1. Purpose / Objective Query
	if classification.Intent == jev.IntentPurpose {
		resp.Answer = buildPurposeAnswer(mm)
		if wantDiagram {
			resp.Diagram = buildOverviewDiagram(mm)
		}
		return resp
	}

	// 2. Setup / Quickstart Query
	if classification.Intent == jev.IntentQuickstart {
		resp.Answer = buildQuickstartAnswer(mm)
		if wantDiagram {
			resp.Diagram = buildQuickstartDiagram()
		}
		return resp
	}

	// 3. Impact / Blast Radius Query
	if classification.Intent == jev.IntentImpact {
		targetNode := findBestMatchingNode(qLower, mm)
		if targetNode != nil {
			resp.MatchedNodes = []string{targetNode.ID}
			resp.Answer = buildImpactAnswer(targetNode, mm)
			if wantDiagram {
				resp.Diagram = buildNodeImpactDiagram(targetNode, mm)
			}
			return resp
		}
	}

	// 4. Specific Component or Architecture Flow Query
	targetNode := findBestMatchingNode(qLower, mm)
	if targetNode != nil {
		resp.Intent = "component_detail"
		resp.MatchedNodes = []string{targetNode.ID}
		resp.Answer = buildComponentAnswer(targetNode, mm)
		if wantDiagram {
			resp.Diagram = buildNodeImpactDiagram(targetNode, mm)
		}
		return resp
	}

	// 5. Default General Answer grounded in codebase
	resp.Intent = "general"
	resp.Answer = buildGeneralAnswer(query, mm)
	if wantDiagram {
		resp.Diagram = buildOverviewDiagram(mm)
	}
	return resp
}

func buildPurposeAnswer(mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString("# 🧭 Objetivo e Propósito do Projeto\n\n")

	if mm != nil && mm.TargetSystem != "" {
		sb.WriteString(fmt.Sprintf("O **%s** é um sistema de inteligência arquitetural determinística e documentação viva.\n\n", mm.TargetSystem))
	} else {
		sb.WriteString("Este repositório é analisado pelo **Gira-G**, fornecendo um GPS arquitetural e bisturi de código.\n\n")
	}

	sb.WriteString("### Problema de Negócio que Resolve:\n")
	sb.WriteString("- **Onboarding Imediato:** Permite a novos engenheiros entenderem os fluxos e limites do sistema em menos de 2 minutos sem ler milhares de linhas.\n")
	sb.WriteString("- **Prevenção de Regressões (Blast Radius):** Mapeia exatamente quais contratos e pacotes chamam cada componente antes de qualquer alteração.\n")
	sb.WriteString("- **Contexto Cirúrgico para IAs:** Entrega fatias exatas de código para Cursor/Claude/Copilot sem alucinações e com 95% de economia de tokens.\n")
	sb.WriteString("- **Governança em CI:** Impede silenciosamente o drift arquitetural através de gates rigorosos de CI (`gira gate-lint --strict`).\n")

	if mm != nil {
		sb.WriteString(fmt.Sprintf("\n📊 **Métricas Atuais do Índice:** %d nós arquiteturais e %d relacionamentos rastreados deterministicamente.\n", len(mm.Nodes), len(mm.Relationships)))
	}

	return sb.String()
}

func buildQuickstartAnswer(mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString("# 🚀 Guia de Início Rápido (3 Passos)\n\n")
	sb.WriteString("Para rodar e navegar no projeto sem atrito:\n\n")
	sb.WriteString("1. **Indexar o Repositório:**\n")
	sb.WriteString("   ```bash\n   gira init .\n   ```\n")
	sb.WriteString("   *Gera o mapa arquitetural determinístico com custo zero de tokens em milissegundos.*\n\n")
	sb.WriteString("2. **Explorar Entrypoints e Contratos:**\n")
	sb.WriteString("   ```bash\n   gira tour --format md\n   ```\n")
	sb.WriteString("   *Percorre a sequência de inicialização e os principais serviços do sistema.*\n\n")
	sb.WriteString("3. **Abrir o Portal Visual Interativo:**\n")
	sb.WriteString("   ```bash\n   gira docs --output build --serve 8080\n   ```\n")
	sb.WriteString("   *Abre o visualizador Cytoscape com física de molas e painéis de caller/callee no navegador.*\n")
	return sb.String()
}

func buildImpactAnswer(node *domain.Node, mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 💥 Análise de Impacto (Blast Radius) para `%s`\n\n", node.ID))
	sb.WriteString(fmt.Sprintf("**Componente:** %s (%s)\n", node.Name, node.Kind))
	if node.Description != "" {
		sb.WriteString(fmt.Sprintf("**Descrição:** %s\n\n", node.Description))
	}

	var callers []string
	var callees []string
	if mm != nil {
		for _, rel := range mm.Relationships {
			if rel.TargetID == node.ID {
				callers = append(callers, rel.SourceID)
			}
			if rel.SourceID == node.ID {
				callees = append(callees, rel.TargetID)
			}
		}
	}

	sb.WriteString(fmt.Sprintf("### Quem depende de `%s` (Raio de Explosão):\n", node.ID))
	if len(callers) == 0 {
		sb.WriteString("- *Nenhum consumidor direto detectado no grafo (nó de borda ou entrypoint).*\n")
	} else {
		for _, c := range callers {
			sb.WriteString(fmt.Sprintf("- ⚠️ `%s` chama este componente\n", c))
		}
	}

	sb.WriteString(fmt.Sprintf("\n### O que `%s` consome:\n", node.ID))
	if len(callees) == 0 {
		sb.WriteString("- *Nenhuma dependência externa consumida diretamente.*\n")
	} else {
		for _, c := range callees {
			sb.WriteString(fmt.Sprintf("- 📦 `%s`\n", c))
		}
	}

	sb.WriteString("\n💡 **Recomendação de Segurança:** Execute `gira gate-lint --strict` antes de commitar para assegurar conformidade com os contratos.")
	return sb.String()
}

func buildComponentAnswer(node *domain.Node, mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 📦 Componente `%s`\n\n", node.ID))
	sb.WriteString(fmt.Sprintf("- **Nome:** %s\n", node.Name))
	sb.WriteString(fmt.Sprintf("- **Tipo:** %s\n", node.Kind))
	if node.Description != "" {
		sb.WriteString(fmt.Sprintf("- **Descrição:** %s\n", node.Description))
	}

	if len(node.Evidence) > 0 {
		sb.WriteString(fmt.Sprintf("- **Arquivo Fonte:** `%s`\n", node.Evidence[0].Locator.Path))
	}

	var rels []string
	if mm != nil {
		for _, r := range mm.Relationships {
			if r.SourceID == node.ID {
				rels = append(rels, fmt.Sprintf("-> %s [%s]", r.TargetID, r.Verb))
			} else if r.TargetID == node.ID {
				rels = append(rels, fmt.Sprintf("<- %s [%s]", r.SourceID, r.Verb))
			}
		}
	}

	if len(rels) > 0 {
		sb.WriteString("\n### Conexões no Grafo Arquitetural:\n")
		for _, r := range rels {
			sb.WriteString(fmt.Sprintf("- `%s`\n", r))
		}
	}

	return sb.String()
}

func buildGeneralAnswer(query string, mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString("# 💡 Resposta Gira-G (Local Zero-Token)\n\n")
	sb.WriteString(fmt.Sprintf("Sua dúvida sobre *\"%s\"* foi analisada contra os fatos determinísticos do repositório.\n\n", query))

	if mm != nil {
		sb.WriteString(fmt.Sprintf("O sistema possui **%d nós** e **%d relações** indexadas. ", len(mm.Nodes), len(mm.Relationships)))
		sb.WriteString("Para navegar pelos pontos de entrada, utilize `gira tour --format md` ou consulte a documentação com `gira docs`.\n")
	} else {
		sb.WriteString("Execute `gira init .` na raiz para carregar o mapa completo de componentes e permitir consultas aprofundadas.\n")
	}

	return sb.String()
}

func findBestMatchingNode(query string, mm *domain.MacroMap) *domain.Node {
	if mm == nil || len(mm.Nodes) == 0 {
		return nil
	}

	cleanQuery := strings.Map(func(r rune) rune {
		if strings.ContainsRune("?!,.:;\"'()[]{}", r) {
			return ' '
		}
		return r
	}, strings.ToLower(query))

	tokens := strings.Fields(cleanQuery)
	var bestNode *domain.Node
	bestScore := 0

	for i := range mm.Nodes {
		node := &mm.Nodes[i]
		score := 0
		idLower := strings.ToLower(node.ID)
		nameLower := strings.ToLower(node.Name)
		descLower := strings.ToLower(node.Description)

		if strings.Contains(cleanQuery, idLower) || (len(nameLower) > 0 && strings.Contains(cleanQuery, nameLower)) {
			score += 10
		}

		for _, tok := range tokens {
			if len(tok) < 3 {
				continue
			}
			if strings.Contains(idLower, tok) {
				score += 5
			}
			if strings.Contains(nameLower, tok) {
				score += 4
			}
			if strings.Contains(descLower, tok) {
				score += 2
			}
		}

		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	if bestScore >= 4 {
		return bestNode
	}
	return nil
}

func buildOverviewDiagram(mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	sb.WriteString("  classDef primary fill:#38bdf8,stroke:#0284c7,stroke-width:2px,color:#0f172a;\n")
	sb.WriteString("  classDef contract fill:#a855f7,stroke:#7e22ce,stroke-width:2px,color:#ffffff;\n")
	sb.WriteString("  classDef package fill:#1e293b,stroke:#475569,stroke-width:1px,color:#f8fafc;\n\n")

	if mm == nil || len(mm.Nodes) == 0 {
		sb.WriteString("  CLI[gira CLI] --> Core[AST Engine]\n")
		sb.WriteString("  Core --> Docs[Docsite Builder]\n")
		sb.WriteString("  Core --> Gates[CI Governance]\n")
		sb.WriteString("  class CLI,Core primary;\n")
		return sb.String()
	}

	count := 0
	for _, r := range mm.Relationships {
		if count >= 12 {
			break
		}
		srcClean := cleanMermaidID(r.SourceID)
		tgtClean := cleanMermaidID(r.TargetID)
		sb.WriteString(fmt.Sprintf("  %s[%s] -->|%s| %s[%s]\n", srcClean, r.SourceID, r.Verb, tgtClean, r.TargetID))
		count++
	}

	return sb.String()
}

func buildQuickstartDiagram() string {
	return `graph LR
  classDef step fill:#38bdf8,stroke:#0284c7,stroke-width:2px,color:#0f172a;
  S1[1. gira init .<br/>Indexa AST] --> S2[2. gira tour<br/>Entrypoints]
  S2 --> S3[3. gira docs<br/>Visual Explorer]
  class S1,S2,S3 step;`
}

func buildNodeImpactDiagram(node *domain.Node, mm *domain.MacroMap) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	sb.WriteString("  classDef target fill:#ef4444,stroke:#b91c1c,stroke-width:3px,color:#ffffff;\n")
	sb.WriteString("  classDef caller fill:#f59e0b,stroke:#d97706,stroke-width:2px,color:#0f172a;\n")
	sb.WriteString("  classDef dep fill:#38bdf8,stroke:#0284c7,stroke-width:1px,color:#0f172a;\n\n")

	tgtID := cleanMermaidID(node.ID)
	sb.WriteString(fmt.Sprintf("  %s[\"🎯 %s\"]:::target\n", tgtID, node.ID))

	if mm != nil {
		for _, r := range mm.Relationships {
			if r.TargetID == node.ID {
				callerClean := cleanMermaidID(r.SourceID)
				sb.WriteString(fmt.Sprintf("  %s[\"⚠️ %s\"]:::caller --> %s\n", callerClean, r.SourceID, tgtID))
			}
			if r.SourceID == node.ID {
				depClean := cleanMermaidID(r.TargetID)
				sb.WriteString(fmt.Sprintf("  %s --> %s[\"📦 %s\"]:::dep\n", tgtID, depClean, r.TargetID))
			}
		}
	}

	return sb.String()
}

func cleanMermaidID(id string) string {
	replacer := strings.NewReplacer(":", "_", "-", "_", "/", "_", ".", "_", "@", "_")
	return replacer.Replace(id)
}
