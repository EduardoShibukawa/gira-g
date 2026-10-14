package domain

import "sort"

// IndexedCommits returns the distinct "repo@commit" pairs the locators cite,
// sorted: the commits the index was actually built from. It walks node and
// relationship evidence (ADR 0001: locators only).
func (m *MacroMap) IndexedCommits() []string {
	if m == nil {
		return nil
	}
	seen := map[string]bool{}
	add := func(evs []EvidenceRecord) {
		for _, ev := range evs {
			if ev.Locator.CommitSHA == "" {
				continue
			}
			seen[ev.Locator.Repo+"@"+ev.Locator.CommitSHA] = true
		}
	}
	for _, n := range m.Nodes {
		add(n.Evidence)
	}
	for _, r := range m.Relationships {
		add(r.Evidence)
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// HasCommit reports whether sha is one of the commits the index cites.
func (m *MacroMap) HasCommit(sha string) bool {
	if m == nil || sha == "" {
		return false
	}
	has := func(evs []EvidenceRecord) bool {
		for _, ev := range evs {
			if ev.Locator.CommitSHA == sha {
				return true
			}
		}
		return false
	}
	for _, n := range m.Nodes {
		if has(n.Evidence) {
			return true
		}
	}
	for _, r := range m.Relationships {
		if has(r.Evidence) {
			return true
		}
	}
	return false
}
