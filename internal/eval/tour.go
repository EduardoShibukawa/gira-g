package eval

import "github.com/EduardoShibukawa/gira-g/internal/slices"

// TourCoverage is the critical-path coverage proxy (spec B, Layer 2; SC8): the
// share of a ground-truth flow's steps that some tour step visits. It returns the
// score in [0, 1] and the flow steps the tour missed, in flow order.
func TourCoverage(t *slices.Tour, flow FlowTruth) (float64, []string) {
	if len(flow.Steps) == 0 {
		return 0, nil
	}
	visited := map[string]bool{}
	for _, s := range t.Steps {
		for _, id := range s.NodeIDs {
			visited[id] = true
		}
	}
	var missed []string
	for _, id := range flow.Steps {
		if !visited[id] {
			missed = append(missed, id)
		}
	}
	return float64(len(flow.Steps)-len(missed)) / float64(len(flow.Steps)), missed
}
