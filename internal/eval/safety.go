package eval

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// nonWave1Sources are source-file extensions no Wave 1 Extractor parses (ADR 0005).
// Extracted evidence citing one of them can only be a hallucinated AST fact.
var nonWave1Sources = map[string]bool{
	".c": true, ".cc": true, ".cpp": true, ".cs": true, ".ex": true, ".exs": true,
	".h": true, ".hpp": true, ".java": true, ".js": true, ".jsx": true, ".kt": true,
	".mjs": true, ".php": true, ".py": true, ".rb": true, ".rs": true, ".scala": true,
	".swift": true, ".ts": true, ".tsx": true,
}

// UnsupportedSafety scores the index-level zero-hallucination rule. For every
// expected unsupported path there must be an UnsupportedBoundary under it (one
// check), and no extracted evidence may cite a non-Wave-1 source file under it (one
// check per evidence record found there). Score = 100 × (1 − violations / checks).
func UnsupportedSafety(mm *domain.MacroMap, paths []string) (float64, []string) {
	if len(paths) == 0 {
		return 100, nil
	}
	checks := 0
	var violations []string

	under := func(p, prefix string) bool {
		p = strings.TrimPrefix(path.Clean(p), "./")
		prefix = strings.TrimSuffix(strings.TrimPrefix(path.Clean(prefix), "./"), "/")
		return p == prefix || strings.HasPrefix(p, prefix+"/")
	}

	for _, prefix := range paths {
		checks++
		reported := false
		for _, ub := range mm.Unsupported {
			if under(ub.Path, prefix) {
				reported = true
				break
			}
		}
		if !reported {
			violations = append(violations, fmt.Sprintf("%s: no unsupported boundary reported", prefix))
		}

		scan := func(owner string, evs []domain.EvidenceRecord) {
			for _, ev := range evs {
				p := ev.Locator.Path
				if ev.Origin != domain.OriginExtracted || !under(p, prefix) {
					continue
				}
				if !nonWave1Sources[strings.ToLower(path.Ext(p))] {
					continue
				}
				checks++
				violations = append(violations, fmt.Sprintf("%s: %s cites extracted evidence in unsupported source %s", prefix, owner, p))
			}
		}
		for _, n := range mm.Nodes {
			scan(n.ID, n.Evidence)
		}
		for _, r := range mm.Relationships {
			scan(r.ID, r.Evidence)
		}
	}

	sort.Strings(violations)
	return 100 * (1 - float64(len(violations))/float64(checks)), violations
}
