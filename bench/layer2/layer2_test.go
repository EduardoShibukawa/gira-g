package layer2

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EduardoShibukawa/gira-g/bench/internal/benchtest"
)

func readFrom(root string) FileReader {
	return func(p string) ([]byte, error) { return os.ReadFile(filepath.Join(root, filepath.FromSlash(p))) }
}

func orderFlow(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(benchtest.RepoRoot, "fixtures", "evaluation", "reference-corpus", "ground-truth.v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gt struct {
		Flows []struct {
			ID    string   `json:"id"`
			Steps []string `json:"steps"`
		} `json:"flows"`
	}
	if err := json.Unmarshal(data, &gt); err != nil {
		t.Fatal(err)
	}
	for _, f := range gt.Flows {
		if f.ID == "order-flow" {
			return f.Steps
		}
	}
	t.Fatal("no order-flow in the ground truth")
	return nil
}

// SC15: the automated proxies run on GIRA tours produced by the real binary over
// the hermetic corpus, and they are deterministic.
func TestSC15ProxiesOnGIRATours(t *testing.T) {
	c := benchtest.Hermetic(t)
	flow := orderFlow(t)
	for _, args := range [][]string{{"tour", "--index", c.Index, "--focus", "service:frontend"}, {"tour", "--index", c.Index}} {
		out := benchtest.Run(t, c.Repo, c.Gira, args...)
		tour, err := FromGIRA([]byte(out))
		if err != nil {
			t.Fatal(err)
		}
		p := Measure(tour, flow, readFrom(c.Repo), c.SHA)
		again := Measure(tour, flow, readFrom(c.Repo), c.SHA)
		if p.Assertiveness != again.Assertiveness || p.Coverage != again.Coverage {
			t.Fatal("proxies are not deterministic")
		}
		if p.Steps == 0 || p.Claims != p.Steps || p.Assertiveness < 0 || p.Assertiveness > 1 || p.Coverage < 0 || p.Coverage > 1 || p.WordsPerStep <= 0 {
			t.Fatalf("%v: implausible proxies %+v", args, p)
		}
		t.Logf("%s: steps %d, assertiveness %.1f%% (%d/%d), order-flow coverage %.1f%% (missed %v), %.1f words/step",
			strings.Join(args[3:], " "), p.Steps, p.Assertiveness*100, p.Verified, p.Claims, p.Coverage*100, p.Missed, p.WordsPerStep)
		if len(args) > 3 && p.Coverage < 0.8 {
			t.Fatalf("focused tour coverage %.2f fell below SC8's 80%%", p.Coverage)
		}
	}
}

// The proxies on a hand-checked tour: 3 claims, 2 verifiable.
func TestMeasureHandChecked(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "src", "cart"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "compose.yaml"), []byte("services:\n  cart:\n    image: cart\n  checkout:\n    depends_on: [cart]\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "src", "cart", "main.go"), []byte("package cart\n"), 0o644)
	tour := Tour{Steps: []Step{
		{Title: "cart", Text: "cart is the focus", NodeIDs: []string{"service:cart"}, Read: []string{"repo@abc:compose.yaml#L2-L3"}, Claims: []Claim{{Anchor: "cart"}}},
		{Title: "checkout", Text: "checkout depends on cart", NodeIDs: []string{"service:checkout"}, Read: []string{"compose.yaml#L4-L5"}, Claims: []Claim{{Anchor: "checkout"}}},
		{Title: "ghost", Text: "ghost is out of bounds", NodeIDs: []string{"service:ghost"}, Read: []string{"compose.yaml#L9-L12"}, Claims: []Claim{{Anchor: "ghost"}}},
	}}
	p := Measure(tour, []string{"service:cart", "service:checkout", "service:payment", "service:shipping"}, readFrom(root), "abc")
	if p.Claims != 3 || p.Verified != 2 || math.Abs(p.Assertiveness-2.0/3) > 1e-9 || p.Coverage != 0.5 || p.Steps != 3 {
		t.Fatalf("got %+v", p)
	}
	if strings.Join(p.Missed, ",") != "service:payment,service:shipping" {
		t.Fatalf("missed %v", p.Missed)
	}
	wrongCommit := Measure(Tour{Steps: tour.Steps[:1]}, nil, readFrom(root), "def")
	if wrongCommit.Verified != 0 {
		t.Fatal("a locator citing another commit must not verify a claim")
	}
}

// The UA adapter reads the documented TourStep shape; only claims naming a symbol
// that exists in the cited file verify.
func TestUATourAdapter(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "ua-tour.json"))
	if err != nil {
		t.Fatal(err)
	}
	lift := func(p string) string {
		for _, c := range []string{"frontend", "checkout"} {
			if strings.HasPrefix(p, "src/"+c+"/") {
				return "service:" + c
			}
		}
		return ""
	}
	tour, err := FromUA(data, lift)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "src", "checkout"), 0o755)
	_ = os.MkdirAll(filepath.Join(root, "src", "frontend", "gateways", "rpc"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "src", "checkout", "main.go"), []byte("func (cs *checkout) PlaceOrder() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "src", "frontend", "gateways", "rpc", "Checkout.gateway.ts"), []byte("export {}\n"), 0o644)
	p := Measure(tour, []string{"service:frontend", "service:checkout", "service:cart"}, readFrom(root), "")
	// Step 1 names only a file (unverifiable), step 2's PlaceOrder exists, step 3's ChargeTwice does not.
	if p.Claims != 3 || p.Verified != 1 || math.Abs(p.Coverage-2.0/3) > 1e-9 {
		t.Fatalf("got %+v", p)
	}
}

// SC15: the moderated-study protocol is committed with tasks, counterbalancing,
// rubric and instruments.
func TestStudyProtocolDocument(t *testing.T) {
	data, err := os.ReadFile("study-protocol.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	for _, section := range []string{"## Design", "## Counterbalancing", "## Tasks", "## Rubric", "## Instruments", "## Procedure", "## Analysis", "NASA-TLX", "Latin square", "time to first useful understanding", "how does an order flow from the cart to shipping?", "who must be notified if PaymentService changes?"} {
		if !strings.Contains(doc, section) {
			t.Errorf("protocol is missing %q", section)
		}
	}
}

// Anchors match case-insensitively with '-' and '_' alike, on both sides; a
// different name still does not match.
func TestMeasureAnchorHyphenUnderscore(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "compose.yaml"), []byte("environment:\n  - IMAGE_PROVIDER_HOST\n  - IMAGE_PROXY\n  - frontend-proxy_addr\n"), 0o644)
	step := func(anchor, loc string) Step {
		return Step{Title: anchor, Text: anchor, Claims: []Claim{{Anchor: anchor, Locators: []string{loc}}}}
	}
	for _, c := range []struct {
		anchor, loc string
		want        int
	}{
		{"image-provider", "compose.yaml#L2", 1}, // '-' anchor vs '_' text
		{"FRONTEND_PROXY", "compose.yaml#L4", 1}, // '_' anchor vs '-' text
		{"Image-Provider", "compose.yaml#L2", 1}, // case folds
		{"image-provider", "compose.yaml#L3", 0}, // IMAGE_PROXY is another name
		{"image-proxy-db", "compose.yaml#L3", 0}, // longer anchor never matches a shorter name
		{"imageprovider", "compose.yaml#L2", 0},  // separators are not dropped
	} {
		if p := Measure(Tour{Steps: []Step{step(c.anchor, c.loc)}}, nil, readFrom(root), ""); p.Verified != c.want {
			t.Errorf("anchor %q in %s: verified %d, want %d", c.anchor, c.loc, p.Verified, c.want)
		}
	}
}
