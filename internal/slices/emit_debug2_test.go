package slices

import (
	"fmt"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
)

func TestEmitDebug2(t *testing.T) {
	er := &EvidenceReport{
		TargetID: "test",
		Evidence: []domain.EvidenceRecord{
			{Locator: domain.Locator{Repo: "r", CommitSHA: "c", Path: "p", StartLine: 1, EndLine: 10}, Origin: domain.OriginExtracted, Status: domain.StatusCurrent},
		},
		DeploymentProfiles: []string{"dev"},
		Origins:            []domain.DerivationOrigin{domain.OriginExtracted},
		Statuses:           []domain.FreshnessStatus{domain.StatusCurrent},
	}

	render := func(r *EvidenceReport) ([]byte, error) {
		return r.CanonicalJSON()
	}

	// Build drop order
	type evKey struct {
		target string
		origin domain.DerivationOrigin
		status domain.FreshnessStatus
	}
	firstIdx := make(map[evKey]int)
	var drops []string
	for i, ev := range er.Evidence {
		key := evKey{target: er.TargetID, origin: ev.Origin, status: ev.Status}
		if _, ok := firstIdx[key]; !ok {
			firstIdx[key] = i
		} else {
			drops = append(drops, fmt.Sprintf("evidence:%d", i))
		}
	}
	fmt.Printf("Drops: %v\n", drops)

	out, err := fit(500, drops, func(trimEvidence bool, drop map[string]bool) (func(*Budget) ([]byte, error), Dropped) {
		var d Dropped
		r := *er
		r.Evidence = nil
		for i, ev := range er.Evidence {
			if drop[fmt.Sprintf("evidence:%d", i)] {
				d.Evidence++
				continue
			}
			r.Evidence = append(r.Evidence, ev)
		}
		return func(b *Budget) ([]byte, error) {
			r.Budget = b
			return render(&r)
		}, d
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Output: %s\n", string(out))
}
