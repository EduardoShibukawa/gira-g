package eval_test

import (
	"context"
	"path"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/eval"
	"github.com/EduardoShibukawa/gira-g/internal/slices"
)

// variant is one hermetic variant indexed through the real pipeline, with its
// ground truth (spec gira-impact-precision).
type variant struct {
	name string
	gt   *eval.GroundTruth
	mm   *domain.MacroMap
}

func hermeticVariants(t *testing.T) []variant {
	t.Helper()
	gt := loadPublished(t)
	pub, err := publishedCorpus(gt).Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	ps := perturbedSuite(t)
	per, err := ps.Corpus.Index(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return []variant{{"published", gt, pub.MacroMap}, {"perturbed", ps.GroundTruth, per.MacroMap}}
}

func isDockerfile(p string) bool {
	b := strings.ToLower(path.Base(p))
	return b == "dockerfile" || strings.HasPrefix(b, "dockerfile.") || strings.HasSuffix(b, ".dockerfile")
}

// SC1: Dockerfile stages and base images are not nodes, briefings list only
// Services and Infrastructure as upstream and downstream, and Libraries reached
// through imports are listed in upstream_libraries.
func TestSC1ImpactSetHoldsOnlyServices(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		for _, n := range v.mm.Nodes {
			if strings.HasPrefix(n.ID, "service:stage:") {
				t.Errorf("%s: Dockerfile stage emitted as node %s", v.name, n.ID)
			}
			if n.Kind == domain.NodeService {
				continue
			}
			for _, ev := range n.Evidence {
				if isDockerfile(ev.Locator.Path) {
					t.Errorf("%s: Dockerfile base image emitted as %s node %s", v.name, n.Kind, n.ID)
					break
				}
			}
		}

		engine := slices.NewEngine(v.mm)
		libraries := 0
		for _, task := range v.gt.Tasks {
			b, err := engine.GetImpactBriefing(task.TargetID, task.Depth)
			if err != nil {
				continue // a target the index does not know is cause 2's (SC3)
			}
			for field, group := range map[string][]domain.Node{"upstream_services": b.UpstreamServices, "downstream_services": b.DownstreamServices} {
				for _, n := range group {
					if n.Kind != domain.NodeService && n.Kind != domain.NodeInfrastructure {
						t.Errorf("%s %s: %s holds %s node %s", v.name, task.ID, field, n.Kind, n.ID)
					}
				}
			}
			for _, n := range b.UpstreamLibraries {
				if n.Kind != domain.NodeLibrary {
					t.Errorf("%s %s: upstream_libraries holds %s node %s", v.name, task.ID, n.Kind, n.ID)
				}
			}
			libraries += len(b.UpstreamLibraries)
		}
		if libraries == 0 {
			t.Errorf("%s: no briefing lists upstream_libraries; the Go imports must still be reported", v.name)
		}
	}
}

// SC2: a Dockerfile a compose service builds is attributed to that compose service.
// In the hermetic corpus every Dockerfile is compose-built, so every Service with
// Dockerfile evidence must also be declared in a compose file, and cart's
// Dockerfile (src/cart/src/) is service:cart, never service:src.
func TestSC2DockerfilesTakeTheirComposeName(t *testing.T) {
	for _, v := range hermeticVariants(t) {
		cart := "service:cart"
		if v.name == "perturbed" {
			cart = "service:basket"
		}
		found := false
		for _, n := range v.mm.Nodes {
			if n.ID == "service:src" {
				t.Errorf("%s: cart's Dockerfile named by its directory: service:src", v.name)
			}
			var dockerfile, compose bool
			for _, ev := range n.Evidence {
				dockerfile = dockerfile || isDockerfile(ev.Locator.Path)
				compose = compose || strings.HasPrefix(path.Base(ev.Locator.Path), "compose")
			}
			if dockerfile && !compose {
				t.Errorf("%s: %s comes from a Dockerfile but no compose service of that name builds it", v.name, n.ID)
			}
			if n.ID == cart && dockerfile {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: %s does not cite its Dockerfile", v.name, cart)
		}
	}
}
