package docsite

import (
	"encoding/json"
	"fmt"

	"github.com/EduardoShibukawa/gira-g/internal/architecture"
	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

type GraphNodeData struct {
	ID         string  `json:"id"`
	Label      string  `json:"label"`
	Kind       string  `json:"kind"`
	Community  string  `json:"community"`
	Centrality float64 `json:"centrality"`
	InDegree   int     `json:"in_degree"`
	OutDegree  int     `json:"out_degree"`
	Loc        int     `json:"loc,omitempty"`
	File       string  `json:"file,omitempty"`
	DocURL     string  `json:"doc_url,omitempty"`
	Color      string  `json:"color"`
}

type GraphEdgeData struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Label       string `json:"label"`
	CrossDomain bool   `json:"cross_domain,omitempty"`
}

type CytoscapeElement struct {
	Data interface{} `json:"data"`
}

// GenerateInteractiveGraphHTML builds an Understand-Anything / Graphify style interactive
// visual exploration page powered by Cytoscape.js.
func GenerateInteractiveGraphHTML(config *SiteConfig, mm *domain.MacroMap, arch *architecture.ClusteredArchitecture, pkgs []architecture.PackageInfo, navGroups []NavGroup) (string, error) {
	elements := buildGraphElements(mm, arch, pkgs)
	elementsJSON, err := json.Marshal(elements)
	if err != nil {
		return "", err
	}

	navHTML := renderSidebarNav(navGroups, "architecture/graph.html")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Interactive Architecture Explorer — %s</title>
  <link rel="stylesheet" href="../assets/style.css">
  <script src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.28.1/cytoscape.min.js"></script>
  <style>
    body { overflow: hidden; background: #0b0f19; }
    .graph-view-container {
      display: flex;
      margin-top: var(--header-h);
      height: calc(100vh - var(--header-h));
      position: relative;
    }
    .graph-canvas-wrap {
      flex: 1;
      position: relative;
      background: radial-gradient(circle at 50%% 50%%, #111827 0%%, #0b0f19 100%%);
      height: 100%%;
    }
    #cy { width: 100%%; height: 100%%; display: block; }
    .graph-toolbar {
      position: absolute;
      top: 1rem;
      left: 1rem;
      z-index: 10;
      background: rgba(17, 24, 39, 0.9);
      backdrop-filter: blur(8px);
      border: 1px solid var(--border-color);
      border-radius: 8px;
      padding: 0.6rem 0.9rem;
      display: flex;
      align-items: center;
      gap: 0.75rem;
      box-shadow: 0 4px 12px rgba(0,0,0,0.4);
    }
    .graph-input {
      background: #030712;
      border: 1px solid var(--border-color);
      color: #fff;
      padding: 0.35rem 0.6rem;
      border-radius: 6px;
      font-size: 0.825rem;
      outline: none;
      width: 180px;
    }
    .graph-select {
      background: #030712;
      border: 1px solid var(--border-color);
      color: #fff;
      padding: 0.35rem 0.6rem;
      border-radius: 6px;
      font-size: 0.825rem;
      outline: none;
    }
    .graph-btn {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      color: var(--text-primary);
      padding: 0.35rem 0.7rem;
      border-radius: 6px;
      font-size: 0.825rem;
      cursor: pointer;
      transition: all 0.15s;
    }
    .graph-btn:hover { background: var(--accent-bg); color: var(--accent); border-color: var(--accent); }
    .graph-legend {
      position: absolute;
      bottom: 1rem;
      left: 1rem;
      z-index: 10;
      background: rgba(17, 24, 39, 0.9);
      backdrop-filter: blur(8px);
      border: 1px solid var(--border-color);
      border-radius: 8px;
      padding: 0.6rem 0.9rem;
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      font-size: 0.75rem;
    }
    .legend-item { display: flex; align-items: center; gap: 0.35rem; color: var(--text-secondary); }
    .legend-dot { width: 10px; height: 10px; border-radius: 50%%; }
    .node-drawer {
      position: absolute;
      top: 0;
      right: 0;
      bottom: 0;
      width: 380px;
      background: rgba(17, 24, 39, 0.96);
      backdrop-filter: blur(12px);
      border-left: 1px solid var(--border-color);
      padding: 1.75rem;
      z-index: 20;
      overflow-y: auto;
      transform: translateX(100%%);
      transition: transform 0.25s ease;
      box-shadow: -10px 0 25px rgba(0,0,0,0.5);
    }
    .node-drawer.open { transform: translateX(0); }
    .drawer-close {
      position: absolute;
      top: 1rem;
      right: 1rem;
      background: transparent;
      border: none;
      color: var(--text-muted);
      font-size: 1.25rem;
      cursor: pointer;
    }
    .drawer-close:hover { color: #fff; }
    .drawer-title { font-size: 1.2rem; font-weight: 700; color: #fff; margin-bottom: 0.25rem; word-break: break-all; }
    .drawer-kind { display: inline-block; font-size: 0.75rem; font-weight: 600; padding: 0.15rem 0.5rem; border-radius: 9999px; background: var(--accent-bg); color: var(--accent); margin-bottom: 1rem; }
    .drawer-prop { margin-bottom: 0.85rem; }
    .drawer-prop-label { font-size: 0.75rem; text-transform: uppercase; color: var(--text-muted); font-weight: 600; }
    .drawer-prop-val { font-size: 0.9rem; color: #e5e7eb; margin-top: 0.15rem; }
    .drawer-ref-item {
      display: inline-block;
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: 4px;
      padding: 0.2rem 0.45rem;
      margin: 0.2rem 0.2rem 0.2rem 0;
      font-size: 0.775rem;
      color: var(--accent);
      cursor: pointer;
      transition: all 0.15s;
    }
    .drawer-ref-item:hover {
      background: var(--accent-bg);
      border-color: var(--accent);
    }
    .drawer-actions {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      margin-top: 1.25rem;
      padding-top: 1.25rem;
      border-top: 1px solid var(--border-color);
    }
    .drawer-btn {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      color: var(--text-primary);
      padding: 0.5rem 0.75rem;
      border-radius: 6px;
      font-size: 0.825rem;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.4rem;
      text-decoration: none;
      transition: all 0.15s;
    }
    .drawer-btn:hover {
      background: var(--accent-bg);
      color: var(--accent);
      border-color: var(--accent);
      text-decoration: none;
    }
  </style>
</head>
<body>
  <header class="site-header">
    <div style="display: flex; align-items: center; gap: 1rem;">
      <a href="../index.html" class="brand">
        <div class="brand-logo">G</div>
        <span>%s</span>
      </a>
      <span class="version-badge">%s</span>
    </div>
    <div class="header-center">
      <button id="search-btn" class="search-btn">
        <span>Quick search docs...</span>
        <span class="search-kbd">⌘K</span>
      </button>
    </div>
    <nav class="header-nav">
      <a href="../intro.html" class="header-link">Docs</a>
      <a href="../tour.html" class="header-link">Guided Tour</a>
      <a href="index.html" class="header-link">Architecture</a>
      <a href="graph.html" class="header-link active">Visual Explorer</a>
      <a href="../cli/index.html" class="header-link">CLI</a>
      <a href="%s" target="_blank" rel="noopener" class="header-link">GitHub ↗</a>
    </nav>
  </header>

  <div class="graph-view-container">
    <aside class="site-sidebar">
      %s
    </aside>

    <div class="graph-canvas-wrap" style="margin-left: var(--sidebar-w);">
      <div class="graph-toolbar">
        <input type="text" id="graph-search" class="graph-input" placeholder="Find node / symbol...">
        <select id="community-filter" class="graph-select">
          <option value="all">All Communities</option>
          <option value="Core Engine">Core Engine</option>
          <option value="AST Manipulator">AST Manipulator</option>
          <option value="Extractors Pipeline">Extractors</option>
          <option value="CLI & MCP Entrypoints">CLI & MCP</option>
          <option value="Governance & Gate">Governance</option>
          <option value="Storage & Persistence">Storage</option>
        </select>
        <select id="layout-select" class="graph-select">
          <option value="cose">Force CoSE</option>
          <option value="concentric">Concentric (Centrality)</option>
          <option value="breadthfirst">Hierarchical</option>
          <option value="circle">Circle</option>
        </select>
        <button id="fit-btn" class="graph-btn">Fit View</button>
        <button id="reset-btn" class="graph-btn">Reset</button>
      </div>

      <div class="graph-legend">
        <div class="legend-item"><div class="legend-dot" style="background: #38bdf8;"></div> Core Engine</div>
        <div class="legend-item"><div class="legend-dot" style="background: #c084fc;"></div> AST Manipulator</div>
        <div class="legend-item"><div class="legend-dot" style="background: #34d399;"></div> Extractors</div>
        <div class="legend-item"><div class="legend-dot" style="background: #fbbf24;"></div> CLI & MCP</div>
        <div class="legend-item"><div class="legend-dot" style="background: #f87171;"></div> Governance</div>
        <div class="legend-item"><div class="legend-dot" style="background: #94a3b8;"></div> Storage</div>
      </div>

      <div id="cy"></div>

      <div id="node-drawer" class="node-drawer">
        <button id="drawer-close" class="drawer-close">&times;</button>
        <div id="drawer-content"></div>
      </div>
    </div>
  </div>

  <div id="search-modal" class="search-modal">
    <div class="search-box">
      <div class="search-input-wrap">
        <input type="text" id="search-input" class="search-input" placeholder="Search documentation, symbols, commands...">
      </div>
      <div id="search-results" class="search-results"></div>
    </div>
  </div>

  <script>
    var elements = %s;
    var cy = cytoscape({
      container: document.getElementById('cy'),
      elements: elements,
      style: [
        {
          selector: 'node',
          style: {
            'label': 'data(label)',
            'background-color': 'data(color)',
            'color': '#f9fafb',
            'font-size': '11px',
            'font-family': 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
            'text-valign': 'bottom',
            'text-margin-y': 6,
            'text-background-color': '#030712',
            'text-background-opacity': 0.85,
            'text-background-padding': 3,
            'text-background-shape': 'roundrectangle',
            'width': 'mapData(centrality, 0, 80, 26, 60)',
            'height': 'mapData(centrality, 0, 80, 26, 60)',
            'border-width': 2,
            'border-color': '#1f2937'
          }
        },
        {
          selector: 'node:selected',
          style: {
            'border-width': 4,
            'border-color': '#ffffff',
            'text-background-color': '#38bdf8',
            'text-background-opacity': 1,
            'color': '#030712'
          }
        },
        {
          selector: 'edge',
          style: {
            'width': 1.5,
            'line-color': '#374151',
            'target-arrow-color': '#374151',
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier',
            'arrow-scale': 0.8,
            'opacity': 0.65
          }
        },
        {
          selector: 'edge[?cross_domain]',
          style: {
            'line-color': '#60a5fa',
            'target-arrow-color': '#60a5fa',
            'line-style': 'dashed',
            'opacity': 0.8
          }
        },
        {
          selector: 'edge:selected',
          style: {
            'width': 3,
            'line-color': '#38bdf8',
            'target-arrow-color': '#38bdf8',
            'opacity': 1
          }
        }
      ],
      layout: {
        name: 'cose',
        idealEdgeLength: 110,
        nodeOverlap: 25,
        refresh: 20,
        fit: true,
        padding: 40,
        randomize: false,
        componentSpacing: 120,
        nodeRepulsion: 450000,
        edgeElasticity: 100,
        nestingFactor: 5
      }
    });

    var drawer = document.getElementById('node-drawer');
    var drawerContent = document.getElementById('drawer-content');
    var drawerClose = document.getElementById('drawer-close');

    window.selectNodeById = function(nodeId) {
      var n = cy.$('#' + nodeId);
      if (n.length > 0) {
        cy.nodes().unselect();
        n.select();
        cy.center(n);
        openDrawer(n.data());
      }
    };

    window.copyCliCmd = function(cmd) {
      navigator.clipboard.writeText(cmd).then(function() {
        var btn = document.getElementById('drawer-copy-btn');
        if (btn) {
          var orig = btn.innerText;
          btn.innerText = '✓ Copied!';
          setTimeout(function() { btn.innerText = orig; }, 2000);
        }
      });
    };

    function openDrawer(nodeData) {
      var inbound = cy.edges('[target = "' + nodeData.id + '"]').map(function(e) { return e.data('source'); });
      var outbound = cy.edges('[source = "' + nodeData.id + '"]').map(function(e) { return e.data('target'); });

      var cleanName = nodeData.label;
      var filePath = nodeData.file || '';

      var html = '<div class="drawer-title">' + cleanName + '</div>' +
        '<div class="drawer-kind">' + (nodeData.kind || 'package') + '</div>' +
        '<div class="drawer-prop"><div class="drawer-prop-label">Community</div><div class="drawer-prop-val" style="color: ' + nodeData.color + '; font-weight: 600;">' + (nodeData.community || 'General') + '</div></div>' +
        '<div class="drawer-prop"><div class="drawer-prop-label">Centrality Score</div><div class="drawer-prop-val">' + (nodeData.centrality ? nodeData.centrality.toFixed(1) : 'N/A') + '</div></div>' +
        '<div class="drawer-prop"><div class="drawer-prop-label">In-Degree / Out-Degree</div><div class="drawer-prop-val">In: ' + nodeData.in_degree + ' | Out: ' + nodeData.out_degree + '</div></div>';

      if (filePath) {
        html += '<div class="drawer-prop"><div class="drawer-prop-label">File / Locator</div><div class="drawer-prop-val"><code style="font-size:0.8rem;">' + filePath + '</code></div></div>';
      }

      if (inbound.length > 0) {
        html += '<div class="drawer-prop"><div class="drawer-prop-label">Inbound Callers / Consumers (' + inbound.length + ')</div><div class="drawer-prop-val">';
        inbound.forEach(function(srcId) {
          var srcNode = cy.$('#' + srcId);
          var lbl = srcNode.length > 0 ? srcNode.data('label') : srcId;
          html += '<span class="drawer-ref-item" onclick="selectNodeById(\'' + srcId + '\')">' + lbl + '</span>';
        });
        html += '</div></div>';
      }

      if (outbound.length > 0) {
        html += '<div class="drawer-prop"><div class="drawer-prop-label">Outbound Dependencies (' + outbound.length + ')</div><div class="drawer-prop-val">';
        outbound.forEach(function(tgtId) {
          var tgtNode = cy.$('#' + tgtId);
          var lbl = tgtNode.length > 0 ? tgtNode.data('label') : tgtId;
          html += '<span class="drawer-ref-item" onclick="selectNodeById(\'' + tgtId + '\')">' + lbl + '</span>';
        });
        html += '</div></div>';
      }

      var cliCmd = 'gira read --file ' + (filePath || cleanName);
      html += '<div class="drawer-actions">' +
        '<button id="drawer-copy-btn" class="drawer-btn" onclick="copyCliCmd(\'' + cliCmd + '\')">📋 Copy Read Command</button>' +
        '<a href="index.html" class="drawer-btn">🏛️ View Living Architecture Docs</a>' +
        '</div>';

      drawerContent.innerHTML = html;
      drawer.classList.add('open');
    }

    cy.on('tap', 'node', function(evt) {
      openDrawer(evt.target.data());
    });

    cy.on('tap', function(evt) {
      if (evt.target === cy) {
        drawer.classList.remove('open');
      }
    });

    drawerClose.addEventListener('click', function() {
      drawer.classList.remove('open');
    });

    document.getElementById('fit-btn').addEventListener('click', function() {
      cy.fit(40);
    });

    document.getElementById('reset-btn').addEventListener('click', function() {
      cy.nodes().show();
      cy.edges().show();
      cy.layout({ name: 'cose', fit: true, padding: 40 }).run();
    });

    document.getElementById('layout-select').addEventListener('change', function(e) {
      var val = e.target.value;
      cy.layout({ name: val, fit: true, padding: 40 }).run();
    });

    document.getElementById('community-filter').addEventListener('change', function(e) {
      var val = e.target.value;
      if (val === 'all') {
        cy.nodes().show();
        cy.edges().show();
      } else {
        cy.nodes().forEach(function(n) {
          if (n.data('community') === val) {
            n.show();
          } else {
            n.hide();
          }
        });
      }
      cy.fit(40);
    });

    document.getElementById('graph-search').addEventListener('input', function(e) {
      var term = e.target.value.toLowerCase().trim();
      if (!term) {
        cy.nodes().style('opacity', 1);
        cy.edges().style('opacity', 0.65);
        return;
      }
      cy.nodes().forEach(function(n) {
        if (n.data('label').toLowerCase().indexOf(term) !== -1 || n.data('id').toLowerCase().indexOf(term) !== -1) {
          n.style('opacity', 1);
          n.select();
        } else {
          n.style('opacity', 0.15);
          n.unselect();
        }
      });
    });
  </script>
  <script src="../assets/app.js"></script>
</body>
</html>`,
		config.Title,
		config.Title,
		config.Version,
		config.GitHubURL,
		navHTML,
		string(elementsJSON),
	)

	return html, nil
}

func buildGraphElements(mm *domain.MacroMap, arch *architecture.ClusteredArchitecture, pkgs []architecture.PackageInfo) []CytoscapeElement {
	var elements []CytoscapeElement
	nodeIDs := make(map[string]bool)

	// Community palette
	communityColors := map[architecture.DomainID]string{
		architecture.DomainCore:        "#38bdf8",
		architecture.DomainManipulator: "#c084fc",
		architecture.DomainExtractors:  "#34d399",
		architecture.DomainEntrypoints: "#fbbf24",
		architecture.DomainGovernance:  "#f87171",
		architecture.DomainStorage:     "#94a3b8",
		architecture.DomainExternal:    "#818cf8",
	}

	// 1. Ingest Nodes from Clustered Architecture
	if arch != nil && len(arch.Communities) > 0 {
		for domainID, comm := range arch.Communities {
			color := communityColors[domainID]
			if color == "" {
				color = "#60a5fa"
			}
			nodes := append(comm.CoreAbstractions, comm.Utilities...)
			for _, n := range nodes {
				if nodeIDs[n.ID] {
					continue
				}
				elements = append(elements, CytoscapeElement{
					Data: GraphNodeData{
						ID:         n.ID,
						Label:      n.Name,
						Kind:       string(n.Kind),
						Community:  comm.Title,
						Centrality: float64(n.Centrality),
						InDegree:   n.InDegree,
						OutDegree:  n.OutDegree,
						File:       n.FileLocator.Path,
						Color:      color,
					},
				})
				nodeIDs[n.ID] = true
			}
		}
	} else if mm != nil {
		for _, n := range mm.Nodes {
			filePath := ""
			if len(n.Evidence) > 0 {
				filePath = n.Evidence[0].Locator.Path
			}
			elements = append(elements, CytoscapeElement{
				Data: GraphNodeData{
					ID:        n.ID,
					Label:     n.Name,
					Kind:      string(n.Kind),
					Community: "Core Engine",
					Color:     "#38bdf8",
					File:      filePath,
				},
			})
			nodeIDs[n.ID] = true
		}
	}

	// 2. Add package nodes if scanned and not yet present
	for _, p := range pkgs {
		pkgNodeID := "package:" + p.Path
		if !nodeIDs[pkgNodeID] {
			color := communityColors[p.Domain]
			if color == "" {
				color = "#60a5fa"
			}
			elements = append(elements, CytoscapeElement{
				Data: GraphNodeData{
					ID:         pkgNodeID,
					Label:      p.Name,
					Kind:       "package",
					Community:  string(p.Domain),
					Centrality: float64(len(p.Imports) + 2),
					InDegree:   0,
					OutDegree:  len(p.Imports),
					File:       p.Path,
					Color:      color,
				},
			})
			nodeIDs[pkgNodeID] = true
		}
	}

	// 3. Connect Genuine Package Import Edges!
	addedEdges := make(map[string]bool)
	for _, p := range pkgs {
		srcID := "package:" + p.Path
		for _, imp := range p.Imports {
			tgtID := "package:" + imp
			if nodeIDs[srcID] && nodeIDs[tgtID] && srcID != tgtID {
				edgeKey := fmt.Sprintf("%s->%s", srcID, tgtID)
				if !addedEdges[edgeKey] {
					isCrossDomain := false
					if arch != nil {
						srcComm := ""
						tgtComm := ""
						if srcNode, ok := arch.NodesByID[srcID]; ok {
							srcComm = string(srcNode.Domain)
						}
						if tgtNode, ok := arch.NodesByID[tgtID]; ok {
							tgtComm = string(tgtNode.Domain)
						}
						if srcComm != "" && tgtComm != "" && srcComm != tgtComm {
							isCrossDomain = true
						}
					}

					elements = append(elements, CytoscapeElement{
						Data: GraphEdgeData{
							ID:          "e:" + edgeKey,
							Source:      srcID,
							Target:      tgtID,
							Label:       "imports",
							CrossDomain: isCrossDomain,
						},
					})
					addedEdges[edgeKey] = true
				}
			}
		}
	}

	// 4. Ingest MacroMap Relationships if present
	if mm != nil {
		for idx, rel := range mm.Relationships {
			if nodeIDs[rel.SourceID] && nodeIDs[rel.TargetID] {
				edgeKey := fmt.Sprintf("%s->%s", rel.SourceID, rel.TargetID)
				if !addedEdges[edgeKey] {
					elements = append(elements, CytoscapeElement{
						Data: GraphEdgeData{
							ID:     fmt.Sprintf("e:rel:%d", idx),
							Source: rel.SourceID,
							Target: rel.TargetID,
							Label:  string(rel.Verb),
						},
					})
					addedEdges[edgeKey] = true
				}
			}
		}
	}

	return elements
}
