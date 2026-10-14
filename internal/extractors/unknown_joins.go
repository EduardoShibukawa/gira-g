package extractors

import (
	"bytes"
	"regexp"
	"sort"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

// contractNameRegex admits the gRPC-style Contract names a file can mention
// (`contract:<Name>`); topic contracts (`contract:kafka:<topic>`) never match.
var contractNameRegex = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)

// annotateUnknownJoins is the J3 posture (spec gira-recall-joins, SC3): a file
// no Extractor parses that names a known Contract (or its generated client)
// keeps its boundary, now listing those Contracts as unknown joins and citing
// the first mention's line. Nothing is extracted from it: no node, no edge,
// so the provider or consumer stays unknown rather than named.
func annotateUnknownJoins(unsupported map[string]domain.UnsupportedBoundary, content map[string][]byte, nodes map[string]*domain.Node) {
	type contract struct {
		id string
		re *regexp.Regexp
	}
	var contracts []contract
	for id, n := range nodes {
		if n.Kind != domain.NodeContract || id != "contract:"+n.Name || !contractNameRegex.MatchString(n.Name) {
			continue
		}
		contracts = append(contracts, contract{id, regexp.MustCompile(`\b` + regexp.QuoteMeta(n.Name) + `(?:Client)?\b`)})
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].id < contracts[j].id })

	for p, data := range content {
		ub := unsupported[p]
		first := -1
		for _, c := range contracts {
			loc := c.re.FindIndex(data)
			if loc == nil {
				continue
			}
			ub.UnknownJoins = append(ub.UnknownJoins, c.id)
			if first < 0 || loc[0] < first {
				first = loc[0]
			}
		}
		if first < 0 {
			continue
		}
		line := bytes.Count(data[:first], []byte("\n")) + 1
		ub.Locator.StartLine, ub.Locator.EndLine = line, line
		unsupported[p] = ub
	}
}
