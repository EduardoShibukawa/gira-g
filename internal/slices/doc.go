package slices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// DocRunTask is one runnable manifest task a doc cites for "how to run/test":
// a `service:task:*` node extracted from a repo manifest (mise), with the
// command it runs and the locator of its declaration. Tasks are repo-level:
// the doc labels them as such and never invents a service scoping.
type DocRunTask struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Run     string         `json:"run,omitempty"`
	Depends []string       `json:"depends,omitempty"`
	Locator domain.Locator `json:"locator"`
}

// Doc is the F3 per-service onboarding document: what the service is, its
// entrypoints, symbols, provided/consumed contracts, dependencies, how to
// run/test it from manifests, where to change it, a guided tour, its flows
// and its known unknowns. It assembles child-02's IntraServiceDoc with the
// existing tour/flows/explain data; no new pipeline, no model client
// (ADR 0001). Markdown and JSON render the same struct, so both formats
// cover the same facts (same node/edge/evidence IDs); `html` stays a
// follow-up.
type Doc struct {
	ServiceID    string                    `json:"service_id"`
	ServiceName  string                    `json:"service_name"`
	Kind         domain.NodeType           `json:"kind"`
	Description  string                    `json:"description,omitempty"`
	Ports        []string                  `json:"ports,omitempty"`
	PortsLocator string                    `json:"ports_locator,omitempty"`
	Profiles     []string                  `json:"profiles,omitempty"`
	Origins      []domain.DerivationOrigin `json:"origins,omitempty"`
	Statuses     []domain.FreshnessStatus  `json:"statuses,omitempty"`
	Intra        IntraServiceDoc           `json:"intra"`
	RunTasks     []DocRunTask              `json:"run_tasks,omitempty"`
	ChangeDirs   []string                  `json:"change_dirs,omitempty"`
	Tour         *Tour                     `json:"tour,omitempty"`
	Flows        *Flows                    `json:"flows,omitempty"`
	Budget       *Budget                   `json:"budget,omitempty"`
}

// GetDoc assembles the onboarding document of one service: the F2
// intra-service view plus the existing tour, flows and explain data, the
// runnable manifest tasks and the change dirs derived from the service's own
// packages and entrypoints. Only an unknown service ID errors; a service
// with no extractor detail still gets contracts, run/test manifests and
// unknowns, never guesses.
func (e *Engine) GetDoc(serviceID string) (*Doc, error) {
	node, ok := e.nodesByID[serviceID]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	intra, err := e.GetIntraServiceDoc(serviceID)
	if err != nil {
		return nil, err
	}

	tour, err := e.GetTour(serviceID, DefaultTourSteps)
	if err != nil {
		return nil, err
	}

	flows, err := e.GetFlows(serviceID)
	if err != nil {
		return nil, err
	}

	report, err := e.ExplainEvidence(serviceID, "")
	if err != nil {
		return nil, err
	}

	doc := &Doc{
		ServiceID:   node.ID,
		ServiceName: node.Name,
		Kind:        node.Kind,
		Description: node.Description,
		Ports:       servicePorts(node),
		Intra:       *intra,
		RunTasks:    e.docRunTasks(),
		ChangeDirs:  changeDirs(intra),
		Tour:        tour,
		Flows:       flows,
		Profiles:    report.DeploymentProfiles,
		Origins:     report.Origins,
		Statuses:    report.Statuses,
	}
	if loc := portsLocator(node); loc != "" {
		doc.PortsLocator = loc
	}
	return doc, nil
}

// servicePorts reads the ports a service declares in its extracted manifests
// (Dockerfile EXPOSE / compose ports), or nil when none was extracted.
func servicePorts(n domain.Node) []string {
	raw, ok := n.Metadata["ports"]
	if !ok {
		return nil
	}
	var out []string
	switch v := raw.(type) {
	case []string:
		out = append(out, v...)
	case []any:
		for _, p := range v {
			out = append(out, fmt.Sprint(p))
		}
	}
	return out
}

// docRunTasks lists every runnable manifest task in the index
// (`service:task:*` nodes with a `task` key), sorted by ID. Tasks are
// repo-level manifests, shown as such; the doc never claims they belong to
// one service.
func (e *Engine) docRunTasks() []DocRunTask {
	var out []DocRunTask
	for _, n := range e.macroMap.Nodes {
		if n.Kind != domain.NodeService || n.Metadata == nil {
			continue
		}
		taskName, ok := n.Metadata["task"].(string)
		if !ok || taskName == "" {
			continue
		}
		t := DocRunTask{ID: n.ID, Name: n.Name}
		if run, ok := n.Metadata["run"].(string); ok {
			t.Run = run
		}
		switch deps := n.Metadata["depends"].(type) {
		case []string:
			t.Depends = append(t.Depends, deps...)
		case []any:
			for _, d := range deps {
				t.Depends = append(t.Depends, fmt.Sprint(d))
			}
		}
		sort.Strings(t.Depends)
		if len(n.Evidence) > 0 {
			t.Locator = n.Evidence[0].Locator
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// changeDirs derives "where to change" from the service's own packages and
// entrypoints: the distinct source dirs, sorted. No guessing beyond them.
func changeDirs(intra *IntraServiceDoc) []string {
	seen := map[string]bool{}
	var out []string
	for _, ep := range intra.Entrypoints {
		if ep.Dir != "" && !seen[ep.Dir] {
			seen[ep.Dir] = true
			out = append(out, ep.Dir)
		}
	}
	for _, p := range intra.Packages {
		if p.Dir != "" && !seen[p.Dir] {
			seen[p.Dir] = true
			out = append(out, p.Dir)
		}
	}
	sort.Strings(out)
	return out
}

// normalized returns a copy with every list in canonical order so the JSON
// envelope and the Markdown render agree byte for byte across runs.
func (d *Doc) normalized() Doc {
	c := *d
	c.Intra.Entrypoints = append([]domain.GoEntrypoint(nil), d.Intra.Entrypoints...)
	c.Intra.Packages = append([]domain.GoPackageSymbol(nil), d.Intra.Packages...)
	c.Intra.Types = append([]domain.GoTypeSymbol(nil), d.Intra.Types...)
	c.Intra.Funcs = append([]domain.GoFuncSymbol(nil), d.Intra.Funcs...)
	c.Intra.Calls = append([]domain.GoCallRecord(nil), d.Intra.Calls...)
	c.Intra.IntraModuleImports = append([]domain.GoImportRecord(nil), d.Intra.IntraModuleImports...)
	c.Intra.ExternalImports = append([]domain.GoImportRecord(nil), d.Intra.ExternalImports...)
	c.Intra.StdlibImports = append([]domain.GoImportRecord(nil), d.Intra.StdlibImports...)
	sort.Slice(c.Intra.Entrypoints, func(i, j int) bool { return c.Intra.Entrypoints[i].ID < c.Intra.Entrypoints[j].ID })
	sort.Slice(c.Intra.Packages, func(i, j int) bool { return c.Intra.Packages[i].ID < c.Intra.Packages[j].ID })
	sort.Slice(c.Intra.Types, func(i, j int) bool { return c.Intra.Types[i].ID < c.Intra.Types[j].ID })
	sort.Slice(c.Intra.Funcs, func(i, j int) bool { return c.Intra.Funcs[i].ID < c.Intra.Funcs[j].ID })
	sort.Slice(c.Intra.Calls, func(i, j int) bool { return c.Intra.Calls[i].ID < c.Intra.Calls[j].ID })
	sortImports := func(ims []domain.GoImportRecord) {
		sort.Slice(ims, func(i, j int) bool { return ims[i].ID < ims[j].ID })
	}
	sortImports(c.Intra.IntraModuleImports)
	sortImports(c.Intra.ExternalImports)
	sortImports(c.Intra.StdlibImports)
	return c
}

// CanonicalJSON serializes a doc deterministically (sorted records, sorted
// keys) so the agent envelope carries exactly the facts the Markdown shows.
func (d *Doc) CanonicalJSON() ([]byte, error) {
	c := d.normalized()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Emit renders a doc in format exactly as the CLI writes it, fitted to limit
// when limit > 0. Truncation drops call sites first (most numerous, least
// structural), then whole flows from the end, then trailing tour steps (the
// focal first step is never dropped); evidence beyond the first locator per
// contract and the extra read locators per tour step go first. Service
// identity, entrypoints, symbols, contracts, imports, run/test manifests,
// change dirs and known unknowns are never dropped: if they alone exceed
// limit, Emit returns ErrBudgetExceeded.
func (d *Doc) Emit(format Format, limit int) ([]byte, error) {
	render := func(doc *Doc) ([]byte, error) {
		switch format {
		case FormatMarkdown:
			return []byte(doc.RenderMarkdown() + "\n"), nil
		case FormatHTML:
			return []byte(doc.RenderHTML() + "\n"), nil
		default:
			return doc.CanonicalJSON()
		}
	}
	if limit <= 0 {
		return render(d)
	}
	base := d.normalized()
	var drops []string
	for i := len(base.Intra.Calls) - 1; i >= 0; i-- {
		drops = append(drops, "call:"+base.Intra.Calls[i].ID)
	}
	if base.Flows != nil {
		for i := len(base.Flows.Flows) - 1; i >= 0; i-- {
			drops = append(drops, fmt.Sprintf("flow:%d", i))
		}
	}
	if base.Tour != nil {
		for i := len(base.Tour.Steps) - 1; i >= 1; i-- {
			drops = append(drops, fmt.Sprintf("step:%d", i))
		}
	}
	return fit(limit, drops, func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped) {
		var dd Dropped
		c := base
		if trimEvidence {
			trimContractLocators := func(refs []IntraContractRef) []IntraContractRef {
				for i := range refs {
					if len(refs[i].Locators) > 1 {
						dd.Evidence += len(refs[i].Locators) - 1
						refs[i].Locators = refs[i].Locators[:1]
					}
				}
				return refs
			}
			c.Intra.ProvidedContracts = trimContractLocators(append([]IntraContractRef(nil), c.Intra.ProvidedContracts...))
			c.Intra.ConsumedContracts = trimContractLocators(append([]IntraContractRef(nil), c.Intra.ConsumedContracts...))
			if c.Tour != nil {
				t := *c.Tour
				t.Steps = append([]TourStep(nil), c.Tour.Steps...)
				for i := range t.Steps {
					if len(t.Steps[i].Read) > 1 {
						dd.Evidence += len(t.Steps[i].Read) - 1
						t.Steps[i].Read = t.Steps[i].Read[:1]
					}
				}
				c.Tour = &t
			}
		}
		if len(drop) > 0 {
			kept := make([]domain.GoCallRecord, 0, len(c.Intra.Calls))
			for _, call := range c.Intra.Calls {
				if drop["call:"+call.ID] {
					dd.Relationships++
					continue
				}
				kept = append(kept, call)
			}
			c.Intra.Calls = kept
			if c.Flows != nil {
				f := *c.Flows
				f.Flows = nil
				for i, fl := range c.Flows.Flows {
					if drop[fmt.Sprintf("flow:%d", i)] {
						dd.Nodes++
						continue
					}
					f.Flows = append(f.Flows, fl)
				}
				c.Flows = &f
			}
			if c.Tour != nil {
				t := *c.Tour
				t.Steps = nil
				for i, s := range c.Tour.Steps {
					if drop[fmt.Sprintf("step:%d", i)] {
						dd.Nodes++
						continue
					}
					t.Steps = append(t.Steps, s)
				}
				c.Tour = &t
			}
		}
		return func(b *Budget) ([]byte, error) {
			b.Estimator = "bytes/4"
			b.Measured = false
			c.Budget = b
			return render(&c)
		}, dd
	})
}

// RenderMarkdown renders a doc as a human onboarding page. Every fact cites
// its `[repo@commit:path#Lstart-Lend]` locator and its stable ID, so the page
// covers exactly the facts the JSON envelope carries. What no extractor
// covers reads as an explicit unknown, never a guess.
func (d *Doc) RenderMarkdown() string {
	c := d.normalized()
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Service Doc: %s (`%s`, %s)\n", c.ServiceName, c.ServiceID, c.Kind))
	if c.Intra.ServiceLocator.Path != "" {
		sb.WriteString(fmt.Sprintf("- Locator: [%s]\n", c.Intra.ServiceLocator.String()))
	}
	if c.Description != "" {
		sb.WriteString(fmt.Sprintf("- What it is: %s\n", c.Description))
	}
	if len(c.Ports) > 0 {
		ports := "`" + strings.Join(c.Ports, "`, `") + "`"
		if c.PortsLocator != "" {
			ports += fmt.Sprintf(" [%s]", c.PortsLocator)
		}
		sb.WriteString(fmt.Sprintf("- Ports: %s\n", ports))
	} else {
		sb.WriteString("- Ports: (unknown: no ports declared in the extracted manifests)\n")
	}
	provenance := []string{}
	if len(c.Profiles) > 0 {
		provenance = append(provenance, "profiles "+strings.Join(c.Profiles, ", "))
	}
	if len(c.Origins) > 0 {
		strs := make([]string, len(c.Origins))
		for i, o := range c.Origins {
			strs[i] = string(o)
		}
		provenance = append(provenance, "origins "+strings.Join(strs, ", "))
	}
	if len(c.Statuses) > 0 {
		strs := make([]string, len(c.Statuses))
		for i, s := range c.Statuses {
			strs[i] = string(s)
		}
		provenance = append(provenance, "statuses "+strings.Join(strs, ", "))
	}
	if len(provenance) > 0 {
		sb.WriteString(fmt.Sprintf("- Provenance: %s\n", strings.Join(provenance, "; ")))
	}
	for _, s := range c.Statuses {
		if s == domain.StatusStale {
			sb.WriteString(fmt.Sprintf("> NOTICE: Stale evidence detected for `%s`\n", c.ServiceID))
			break
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Entrypoints (%d)\n", len(c.Intra.Entrypoints)))
	if len(c.Intra.Entrypoints) == 0 {
		sb.WriteString("- (unknown: no Go entrypoint extracted)\n")
	} else {
		for _, ep := range c.Intra.Entrypoints {
			sb.WriteString(fmt.Sprintf("- `%s` %s (package %s, dir `%s`) [%s] (%s)\n",
				ep.ID, ep.Func, ep.Package, ep.Dir, ep.Locator.String(), ep.Origin))
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Symbols (%d packages, %d types, %d funcs)\n",
		len(c.Intra.Packages), len(c.Intra.Types), len(c.Intra.Funcs)))
	if len(c.Intra.Packages) == 0 && len(c.Intra.Types) == 0 && len(c.Intra.Funcs) == 0 {
		sb.WriteString("- (unknown: no Go symbols extracted)\n")
	} else {
		for _, p := range c.Intra.Packages {
			sb.WriteString(fmt.Sprintf("- `%s` package %s (dir `%s`) [%s] (%s)\n",
				p.ID, p.Name, p.Dir, p.Locator.String(), p.Origin))
		}
		for _, t := range c.Intra.Types {
			sb.WriteString(fmt.Sprintf("- `%s` type %s (%s) [%s] (%s)\n",
				t.ID, t.Name, t.Kind, t.Locator.String(), t.Origin))
		}
		for _, f := range c.Intra.Funcs {
			what := "func " + f.Name
			if f.Receiver != "" {
				what = fmt.Sprintf("method %s of %s", f.Name, f.Receiver)
			} else if !f.Exported {
				what = "unexported func " + f.Name
			}
			sb.WriteString(fmt.Sprintf("- `%s` %s [%s] (%s)\n",
				f.ID, what, f.Locator.String(), f.Origin))
		}
	}

	renderContracts := func(title string, refs []IntraContractRef) {
		sb.WriteString(fmt.Sprintf("\n## %s (%d)\n", title, len(refs)))
		if len(refs) == 0 {
			sb.WriteString("- (none)\n")
			return
		}
		for _, r := range refs {
			locs := make([]string, len(r.Locators))
			for i, l := range r.Locators {
				locs[i] = l.String()
			}
			sb.WriteString(fmt.Sprintf("- `%s` %s via `%s` (%s) [%s]\n",
				r.ContractID, r.ContractName, r.EdgeID, r.Verb, strings.Join(locs, ", ")))
		}
	}
	renderContracts("Contracts Provided", c.Intra.ProvidedContracts)
	renderContracts("Contracts Consumed", c.Intra.ConsumedContracts)

	renderImports := func(title string, ims []domain.GoImportRecord) {
		sb.WriteString(fmt.Sprintf("\n### %s (%d)\n", title, len(ims)))
		if len(ims) == 0 {
			sb.WriteString("- (none)\n")
			return
		}
		for _, im := range ims {
			alias := ""
			if im.Alias != "" {
				alias = fmt.Sprintf(" as %s", im.Alias)
			}
			sb.WriteString(fmt.Sprintf("- `%s` %s%s [%s] (%s)\n",
				im.ID, im.Path, alias, im.Locator.String(), im.Origin))
		}
	}
	totalImports := len(c.Intra.IntraModuleImports) + len(c.Intra.ExternalImports) + len(c.Intra.StdlibImports)
	sb.WriteString(fmt.Sprintf("\n## Dependencies: Imports (%d)\n", totalImports))
	if totalImports == 0 {
		sb.WriteString("- (unknown: no Go imports extracted)\n")
	} else {
		renderImports("Intra-module", c.Intra.IntraModuleImports)
		renderImports("External", c.Intra.ExternalImports)
		renderImports("Stdlib", c.Intra.StdlibImports)
	}

	sb.WriteString(fmt.Sprintf("\n## Calls (%d)\n", len(c.Intra.Calls)))
	if len(c.Intra.Calls) == 0 {
		sb.WriteString("- (unknown: no Go calls extracted)\n")
	} else {
		for _, call := range c.Intra.Calls {
			target := "(unresolved: no target)"
			switch call.Resolution {
			case domain.GoCallResolved:
				target = "`" + call.TargetID + "`"
			case domain.GoCallExternal:
				if call.External != nil {
					target = fmt.Sprintf("stub %s.%s", call.External.ImportPath, call.External.Name)
				}
			}
			caller := "file scope"
			if call.Caller != "" {
				caller = "`" + call.Caller + "`"
			}
			sb.WriteString(fmt.Sprintf("- `%s` %s --%s--> %s (%s) [%s]\n",
				call.ID, caller, call.Callee, target, call.Resolution, call.Locator.String()))
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Run & Test (repo manifests) (%d)\n", len(c.RunTasks)))
	if len(c.RunTasks) == 0 {
		sb.WriteString("- (unknown: no runnable manifest tasks extracted)\n")
	} else {
		for _, t := range c.RunTasks {
			run := t.Run
			if run == "" {
				run = "(unknown: task declares no run command)"
			} else {
				run = "`" + run + "`"
			}
			loc := ""
			if t.Locator.Path != "" {
				loc = fmt.Sprintf(" [%s]", t.Locator.String())
			}
			sb.WriteString(fmt.Sprintf("- `%s` %s: %s%s\n", t.ID, t.Name, run, loc))
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Where to Change (%d)\n", len(c.ChangeDirs)))
	if len(c.ChangeDirs) == 0 {
		sb.WriteString("- (unknown: no source dirs extracted)\n")
	} else {
		for _, dir := range c.ChangeDirs {
			sb.WriteString(fmt.Sprintf("- `%s`\n", dir))
		}
	}

	if c.Tour != nil {
		steps := "steps"
		if len(c.Tour.Steps) == 1 {
			steps = "step"
		}
		sb.WriteString(fmt.Sprintf("\n## Guided Tour (%d %s)\n", len(c.Tour.Steps), steps))
		for _, s := range c.Tour.Steps {
			sb.WriteString(fmt.Sprintf("%d. %s (`%s`)\n", s.Order, s.Title, strings.Join(s.NodeIDs, "`, `")))
			sb.WriteString(fmt.Sprintf("   - Why: %s\n", s.Why))
			if s.WhyLocator != "" {
				sb.WriteString(fmt.Sprintf("   - Why, evidence: %s\n", s.WhyLocator))
			}
			for _, l := range s.Read {
				sb.WriteString(fmt.Sprintf("   - Read: %s\n", l))
			}
		}
	}

	if c.Flows != nil {
		sb.WriteString(fmt.Sprintf("\n## Flows from %s (%d)\n", c.Flows.Entry, len(c.Flows.Flows)))
		if len(c.Flows.Flows) == 0 {
			sb.WriteString("- (none: no consumes/accesses chain leaves this service)\n")
		}
		for i, fl := range c.Flows.Flows {
			ids := make([]string, len(fl.Steps))
			for j, h := range fl.Steps {
				ids[j] = h.NodeID
			}
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.Join(ids, " -> ")))
			for _, h := range fl.Steps[1:] {
				if h.Locator != "" {
					sb.WriteString(fmt.Sprintf("   - %s `%s` via `%s` [%s]\n", h.Verb, h.NodeID, h.Via, h.Locator))
				} else {
					sb.WriteString(fmt.Sprintf("   - %s `%s` via `%s` (unknown: no locator)\n", h.Verb, h.NodeID, h.Via))
				}
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Known Unknowns (%d)\n", len(c.Intra.KnownUnknowns)))
	if len(c.Intra.KnownUnknowns) == 0 {
		sb.WriteString("- (none)\n")
	} else {
		for _, u := range c.Intra.KnownUnknowns {
			loc := ""
			if u.Locator.Path != "" {
				loc = fmt.Sprintf(" [%s]", u.Locator.String())
			}
			joins := ""
			if len(u.UnknownJoins) > 0 {
				joins = fmt.Sprintf(" (unknown joins: `%s`)", strings.Join(u.UnknownJoins, "`, `"))
			}
			sb.WriteString(fmt.Sprintf("- `%s` - %s%s%s\n", u.Path, u.Reason, loc, joins))
		}
	}

	renderBudgetMarkdown(&sb, c.Budget)
	return sb.String()
}

// RenderHTML renders a self-contained, zero-dependency HTML single-service
// onboarding document covering the exact same facts, IDs, and locators as
// RenderMarkdown and CanonicalJSON.
func (d *Doc) RenderHTML() string {
	c := d.normalized()
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Service Doc: ` + html.EscapeString(c.ServiceName) + ` (` + html.EscapeString(c.ServiceID) + `)</title>
<style>
:root {
  --bg: #0f141c;
  --surface: #18202c;
  --surface-border: #283548;
  --text: #e2e8f0;
  --text-dim: #94a3b8;
  --accent: #38bdf8;
  --accent-contract: #c084fc;
  --accent-warn: #f59e0b;
  --tag-stale: #ef4444;
  --code-bg: #0b0f17;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background-color: var(--bg);
  color: var(--text);
  line-height: 1.55;
  padding: 2rem;
}
.container { max-width: 1100px; margin: 0 auto; }
header { border-bottom: 2px solid var(--surface-border); padding-bottom: 1.25rem; margin-bottom: 1.5rem; }
h1 { font-size: 1.85rem; color: #f8fafc; margin-bottom: 0.5rem; }
h2 { font-size: 1.25rem; margin: 1.75rem 0 0.75rem 0; color: #f1f5f9; border-bottom: 1px solid var(--surface-border); padding-bottom: 0.35rem; }
h3 { font-size: 1rem; margin: 1rem 0 0.5rem 0; color: var(--accent); }
.panel { background: var(--surface); border: 1px solid var(--surface-border); border-radius: 6px; padding: 1rem 1.25rem; margin-bottom: 1rem; }
ul, ol { padding-left: 1.4rem; }
li { margin-bottom: 0.35rem; }
code { font-family: monospace; background: var(--code-bg); padding: 0.15rem 0.4rem; border-radius: 4px; color: var(--accent); font-size: 0.85em; }
.locator { font-family: monospace; background: var(--code-bg); padding: 0.15rem 0.45rem; border-radius: 4px; color: #38bdf8; font-size: 0.8em; }
.notice-stale { background: rgba(239, 68, 68, 0.15); border-left: 4px solid var(--tag-stale); padding: 0.6rem 0.9rem; margin-top: 0.75rem; border-radius: 0 4px 4px 0; }
.toolbar { display: flex; gap: 0.75rem; margin-bottom: 1.25rem; flex-wrap: wrap; }
.toolbar input { background: var(--surface); border: 1px solid var(--surface-border); color: var(--text); padding: 0.45rem 0.75rem; border-radius: 4px; min-width: 260px; }
</style>
</head>
<body>
<div class="container">
<header>
  <h1>Service Doc: ` + html.EscapeString(c.ServiceName) + ` (<code>` + html.EscapeString(c.ServiceID) + `</code>, ` + html.EscapeString(string(c.Kind)) + `)</h1>
  <div class="panel">
    <ul>
`)
	if c.Intra.ServiceLocator.Path != "" {
		sb.WriteString(fmt.Sprintf("      <li>Locator: <span class=\"locator\">[%s]</span></li>\n", html.EscapeString(c.Intra.ServiceLocator.String())))
	}
	if c.Description != "" {
		sb.WriteString(fmt.Sprintf("      <li>What it is: %s</li>\n", html.EscapeString(c.Description)))
	}
	if len(c.Ports) > 0 {
		ports := make([]string, len(c.Ports))
		for i, p := range c.Ports {
			ports[i] = "<code>" + html.EscapeString(p) + "</code>"
		}
		loc := ""
		if c.PortsLocator != "" {
			loc = fmt.Sprintf(" <span class=\"locator\">[%s]</span>", html.EscapeString(c.PortsLocator))
		}
		sb.WriteString(fmt.Sprintf("      <li>Ports: %s%s</li>\n", strings.Join(ports, ", "), loc))
	} else {
		sb.WriteString("      <li>Ports: (unknown: no ports declared in the extracted manifests)</li>\n")
	}

	provenance := []string{}
	if len(c.Profiles) > 0 {
		provenance = append(provenance, "profiles "+strings.Join(c.Profiles, ", "))
	}
	if len(c.Origins) > 0 {
		strs := make([]string, len(c.Origins))
		for i, o := range c.Origins {
			strs[i] = string(o)
		}
		provenance = append(provenance, "origins "+strings.Join(strs, ", "))
	}
	if len(c.Statuses) > 0 {
		strs := make([]string, len(c.Statuses))
		for i, s := range c.Statuses {
			strs[i] = string(s)
		}
		provenance = append(provenance, "statuses "+strings.Join(strs, ", "))
	}
	if len(provenance) > 0 {
		sb.WriteString(fmt.Sprintf("      <li>Provenance: %s</li>\n", html.EscapeString(strings.Join(provenance, "; "))))
	}
	sb.WriteString("    </ul>\n")
	for _, s := range c.Statuses {
		if s == domain.StatusStale {
			sb.WriteString(fmt.Sprintf("    <div class=\"notice-stale\">⚠️ NOTICE: Stale evidence detected for <code>%s</code></div>\n", html.EscapeString(c.ServiceID)))
			break
		}
	}
	sb.WriteString(`  </div>
</header>
<div class="toolbar">
  <input id="gira-doc-search" type="search" placeholder="Filter symbols, contracts, calls, or locators..." oninput="(function(q){q=q.toLowerCase();document.querySelectorAll('li[data-searchable]').forEach(function(el){el.style.display=!q||el.textContent.toLowerCase().indexOf(q)!==-1?'':'none';});})(this.value)">
</div>
`)

	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Entrypoints (%d)</h2>\n<ul>\n", len(c.Intra.Entrypoints)))
	if len(c.Intra.Entrypoints) == 0 {
		sb.WriteString("  <li>(unknown: no Go entrypoint extracted)</li>\n")
	} else {
		for _, ep := range c.Intra.Entrypoints {
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> %s (package %s, dir <code>%s</code>) <span class=\"locator\">[%s]</span> (%s)</li>\n",
				html.EscapeString(ep.ID), html.EscapeString(ep.Func), html.EscapeString(ep.Package),
				html.EscapeString(ep.Dir), html.EscapeString(ep.Locator.String()), html.EscapeString(string(ep.Origin))))
		}
	}
	sb.WriteString("</ul></section>\n")

	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Symbols (%d packages, %d types, %d funcs)</h2>\n<ul>\n",
		len(c.Intra.Packages), len(c.Intra.Types), len(c.Intra.Funcs)))
	if len(c.Intra.Packages) == 0 && len(c.Intra.Types) == 0 && len(c.Intra.Funcs) == 0 {
		sb.WriteString("  <li>(unknown: no Go symbols extracted)</li>\n")
	} else {
		for _, p := range c.Intra.Packages {
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> package %s (dir <code>%s</code>) <span class=\"locator\">[%s]</span> (%s)</li>\n",
				html.EscapeString(p.ID), html.EscapeString(p.Name), html.EscapeString(p.Dir),
				html.EscapeString(p.Locator.String()), html.EscapeString(string(p.Origin))))
		}
		for _, t := range c.Intra.Types {
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> type %s (%s) <span class=\"locator\">[%s]</span> (%s)</li>\n",
				html.EscapeString(t.ID), html.EscapeString(t.Name), html.EscapeString(string(t.Kind)),
				html.EscapeString(t.Locator.String()), html.EscapeString(string(t.Origin))))
		}
		for _, f := range c.Intra.Funcs {
			what := "func " + f.Name
			if f.Receiver != "" {
				what = fmt.Sprintf("method %s of %s", f.Name, f.Receiver)
			} else if !f.Exported {
				what = "unexported func " + f.Name
			}
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> %s <span class=\"locator\">[%s]</span> (%s)</li>\n",
				html.EscapeString(f.ID), html.EscapeString(what),
				html.EscapeString(f.Locator.String()), html.EscapeString(string(f.Origin))))
		}
	}
	sb.WriteString("</ul></section>\n")

	renderContractsHTML := func(title string, refs []IntraContractRef) {
		sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>%s (%d)</h2>\n<ul>\n", html.EscapeString(title), len(refs)))
		if len(refs) == 0 {
			sb.WriteString("  <li>(none)</li>\n")
		} else {
			for _, r := range refs {
				locs := make([]string, len(r.Locators))
				for i, l := range r.Locators {
					locs[i] = html.EscapeString(l.String())
				}
				sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> %s via <code>%s</code> (%s) <span class=\"locator\">[%s]</span></li>\n",
					html.EscapeString(r.ContractID), html.EscapeString(r.ContractName),
					html.EscapeString(r.EdgeID), html.EscapeString(string(r.Verb)), strings.Join(locs, ", ")))
			}
		}
		sb.WriteString("</ul></section>\n")
	}
	renderContractsHTML("Contracts Provided", c.Intra.ProvidedContracts)
	renderContractsHTML("Contracts Consumed", c.Intra.ConsumedContracts)

	renderImportsHTML := func(title string, ims []domain.GoImportRecord) {
		sb.WriteString(fmt.Sprintf("<h3>%s (%d)</h3>\n<ul>\n", html.EscapeString(title), len(ims)))
		if len(ims) == 0 {
			sb.WriteString("  <li>(none)</li>\n")
		} else {
			for _, im := range ims {
				alias := ""
				if im.Alias != "" {
					alias = fmt.Sprintf(" as %s", html.EscapeString(im.Alias))
				}
				sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> %s%s <span class=\"locator\">[%s]</span> (%s)</li>\n",
					html.EscapeString(im.ID), html.EscapeString(im.Path), alias,
					html.EscapeString(im.Locator.String()), html.EscapeString(string(im.Origin))))
			}
		}
		sb.WriteString("</ul>\n")
	}
	totalImports := len(c.Intra.IntraModuleImports) + len(c.Intra.ExternalImports) + len(c.Intra.StdlibImports)
	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Dependencies: Imports (%d)</h2>\n", totalImports))
	if totalImports == 0 {
		sb.WriteString("<ul><li>(unknown: no Go imports extracted)</li></ul>\n")
	} else {
		renderImportsHTML("Intra-module", c.Intra.IntraModuleImports)
		renderImportsHTML("External", c.Intra.ExternalImports)
		renderImportsHTML("Stdlib", c.Intra.StdlibImports)
	}
	sb.WriteString("</section>\n")

	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Calls (%d)</h2>\n<ul>\n", len(c.Intra.Calls)))
	if len(c.Intra.Calls) == 0 {
		sb.WriteString("  <li>(unknown: no Go calls extracted)</li>\n")
	} else {
		for _, call := range c.Intra.Calls {
			target := "(unresolved: no target)"
			switch call.Resolution {
			case domain.GoCallResolved:
				target = "<code>" + html.EscapeString(call.TargetID) + "</code>"
			case domain.GoCallExternal:
				if call.External != nil {
					target = fmt.Sprintf("stub %s.%s", html.EscapeString(call.External.ImportPath), html.EscapeString(call.External.Name))
				}
			}
			caller := "file scope"
			if call.Caller != "" {
				caller = "<code>" + html.EscapeString(call.Caller) + "</code>"
			}
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> %s --%s--&gt; %s (%s) <span class=\"locator\">[%s]</span></li>\n",
				html.EscapeString(call.ID), caller, html.EscapeString(call.Callee), target,
				html.EscapeString(string(call.Resolution)), html.EscapeString(call.Locator.String())))
		}
	}
	sb.WriteString("</ul></section>\n")

	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Run &amp; Test (repo manifests) (%d)</h2>\n<ul>\n", len(c.RunTasks)))
	if len(c.RunTasks) == 0 {
		sb.WriteString("  <li>(unknown: no runnable manifest tasks extracted)</li>\n")
	} else {
		for _, t := range c.RunTasks {
			run := "(unknown: task declares no run command)"
			if t.Run != "" {
				run = "<code>" + html.EscapeString(t.Run) + "</code>"
			}
			loc := ""
			if t.Locator.Path != "" {
				loc = fmt.Sprintf(" <span class=\"locator\">[%s]</span>", html.EscapeString(t.Locator.String()))
			}
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> %s: %s%s</li>\n",
				html.EscapeString(t.ID), html.EscapeString(t.Name), run, loc))
		}
	}
	sb.WriteString("</ul></section>\n")

	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Where to Change (%d)</h2>\n<ul>\n", len(c.ChangeDirs)))
	if len(c.ChangeDirs) == 0 {
		sb.WriteString("  <li>(unknown: no source dirs extracted)</li>\n")
	} else {
		for _, dir := range c.ChangeDirs {
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code></li>\n", html.EscapeString(dir)))
		}
	}
	sb.WriteString("</ul></section>\n")

	if c.Tour != nil {
		steps := "steps"
		if len(c.Tour.Steps) == 1 {
			steps = "step"
		}
		sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Guided Tour (%d %s)</h2>\n<ol>\n", len(c.Tour.Steps), steps))
		for _, s := range c.Tour.Steps {
			nids := make([]string, len(s.NodeIDs))
			for i, id := range s.NodeIDs {
				nids[i] = "<code>" + html.EscapeString(id) + "</code>"
			}
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><strong>%d. %s</strong> (%s)\n    <ul>\n",
				s.Order, html.EscapeString(s.Title), strings.Join(nids, ", ")))
			sb.WriteString(fmt.Sprintf("      <li>Why: %s</li>\n", html.EscapeString(s.Why)))
			if s.WhyLocator != "" {
				sb.WriteString(fmt.Sprintf("      <li>Why, evidence: <span class=\"locator\">%s</span></li>\n", html.EscapeString(s.WhyLocator)))
			}
			for _, l := range s.Read {
				sb.WriteString(fmt.Sprintf("      <li>Read: <span class=\"locator\">%s</span></li>\n", html.EscapeString(l)))
			}
			sb.WriteString("    </ul>\n  </li>\n")
		}
		sb.WriteString("</ol></section>\n")
	}

	if c.Flows != nil {
		sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Flows from %s (%d)</h2>\n", html.EscapeString(c.Flows.Entry), len(c.Flows.Flows)))
		if len(c.Flows.Flows) == 0 {
			sb.WriteString("<ul><li>(none: no consumes/accesses chain leaves this service)</li></ul>\n")
		} else {
			sb.WriteString("<ol>\n")
			for i, fl := range c.Flows.Flows {
				ids := make([]string, len(fl.Steps))
				for j, h := range fl.Steps {
					ids[j] = html.EscapeString(h.NodeID)
				}
				sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\">%d. %s\n    <ul>\n", i+1, strings.Join(ids, " -&gt; ")))
				for _, h := range fl.Steps[1:] {
					if h.Locator != "" {
						sb.WriteString(fmt.Sprintf("      <li>%s <code>%s</code> via <code>%s</code> <span class=\"locator\">[%s]</span></li>\n",
							html.EscapeString(h.Verb), html.EscapeString(h.NodeID), html.EscapeString(h.Via), html.EscapeString(h.Locator)))
					} else {
						sb.WriteString(fmt.Sprintf("      <li>%s <code>%s</code> via <code>%s</code> (unknown: no locator)</li>\n",
							html.EscapeString(h.Verb), html.EscapeString(h.NodeID), html.EscapeString(h.Via)))
					}
				}
				sb.WriteString("    </ul>\n  </li>\n")
			}
			sb.WriteString("</ol>\n")
		}
		sb.WriteString("</section>\n")
	}

	sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Known Unknowns (%d)</h2>\n<ul>\n", len(c.Intra.KnownUnknowns)))
	if len(c.Intra.KnownUnknowns) == 0 {
		sb.WriteString("  <li>(none)</li>\n")
	} else {
		for _, u := range c.Intra.KnownUnknowns {
			loc := ""
			if u.Locator.Path != "" {
				loc = fmt.Sprintf(" <span class=\"locator\">[%s]</span>", html.EscapeString(u.Locator.String()))
			}
			joins := ""
			if len(u.UnknownJoins) > 0 {
				escJoins := make([]string, len(u.UnknownJoins))
				for i, j := range u.UnknownJoins {
					escJoins[i] = "<code>" + html.EscapeString(j) + "</code>"
				}
				joins = fmt.Sprintf(" (unknown joins: %s)", strings.Join(escJoins, ", "))
			}
			sb.WriteString(fmt.Sprintf("  <li data-searchable=\"1\"><code>%s</code> - %s%s%s</li>\n",
				html.EscapeString(u.Path), html.EscapeString(u.Reason), loc, joins))
		}
	}
	sb.WriteString("</ul></section>\n")

	if c.Budget != nil {
		sb.WriteString(fmt.Sprintf("<section class=\"panel\"><h2>Budget</h2><ul><li>limit: %d, estimated: %d, truncated: %t</li><li>dropped: nodes %d, relationships %d, evidence %d</li>",
			c.Budget.Limit, c.Budget.Estimated, c.Budget.Truncated,
			c.Budget.Dropped.Nodes, c.Budget.Dropped.Relationships, c.Budget.Dropped.Evidence))
		if c.Budget.EscalateHint != "" {
			sb.WriteString(fmt.Sprintf("<li>escalate: %s</li>", html.EscapeString(c.Budget.EscalateHint)))
		}
		sb.WriteString("</ul></section>\n")
	}

	sb.WriteString("</div>\n</body>\n</html>")
	return sb.String()
}
