package main

// SC1/SC2 (release contract, F1/F2): GoReleaser packaging posture plus the
// hermetic syft SPDX SBOM and cosign blob signing round-trip.

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func releaseRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

type goreleaserFile struct {
	ProjectName string `yaml:"project_name"`
	Before      struct {
		Hooks []string `yaml:"hooks"`
	} `yaml:"before"`
	Release struct {
		Github struct {
			Owner string `yaml:"owner"`
			Name  string `yaml:"name"`
		} `yaml:"github"`
	} `yaml:"release"`
	Builds []struct {
		Env    []string            `yaml:"env"`
		Goos   []string            `yaml:"goos"`
		Goarch []string            `yaml:"goarch"`
		Ignore []map[string]string `yaml:"ignore"`
	} `yaml:"builds"`
	Archives []struct {
		Files []string `yaml:"files"`
	} `yaml:"archives"`
	Sboms []struct {
		Cmd       string   `yaml:"cmd"`
		Args      []string `yaml:"args"`
		Documents []string `yaml:"documents"`
	} `yaml:"sboms"`
	Signs []struct {
		Cmd       string   `yaml:"cmd"`
		Args      []string `yaml:"args"`
		Artifacts string   `yaml:"artifacts"`
	} `yaml:"signs"`
}

func loadGoReleaser(t *testing.T, root string) goreleaserFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("read .goreleaser.yaml: %v", err)
	}
	var cfg goreleaserFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse .goreleaser.yaml: %v", err)
	}
	return cfg
}

func containsFold(list []string, want string) bool {
	for _, s := range list {
		if strings.EqualFold(strings.TrimSpace(s), want) {
			return true
		}
	}
	return false
}

// TestSCGoReleaserConfig is SC1: the release config targets
// EduardoShibukawa/gira-g with CGO_ENABLED=0, wires syft SBOMs and cosign
// signs, ships LICENSE + notices + the skills manifest, and passes
// `goreleaser check`.
func TestSCGoReleaserConfig(t *testing.T) {
	root := releaseRepoRoot(t)
	cfg := loadGoReleaser(t, root)

	if cfg.Release.Github.Owner != "EduardoShibukawa" || cfg.Release.Github.Name != "gira-g" {
		t.Fatalf("release targets %s/%s, want EduardoShibukawa/gira-g",
			cfg.Release.Github.Owner, cfg.Release.Github.Name)
	}

	if len(cfg.Builds) == 0 {
		t.Fatal("no builds configured")
	}
	b := cfg.Builds[0]
	if !containsFold(b.Env, "CGO_ENABLED=0") {
		t.Fatalf("build env %v missing CGO_ENABLED=0", b.Env)
	}
	for _, e := range b.Env {
		if strings.TrimSpace(e) == "CGO_ENABLED=1" {
			t.Fatalf("build env %v must not set CGO_ENABLED=1", b.Env)
		}
	}
	for _, want := range []string{"linux", "darwin", "windows"} {
		if !containsFold(b.Goos, want) {
			t.Errorf("build goos %v missing %s", b.Goos, want)
		}
	}
	for _, want := range []string{"amd64", "arm64"} {
		if !containsFold(b.Goarch, want) {
			t.Errorf("build goarch %v missing %s", b.Goarch, want)
		}
	}
	ignoredWinArm := false
	for _, ig := range b.Ignore {
		if strings.EqualFold(ig["goos"], "windows") && strings.EqualFold(ig["goarch"], "arm64") {
			ignoredWinArm = true
		}
	}
	if !ignoredWinArm {
		t.Error("builds must ignore windows/arm64")
	}

	hookOK := false
	for _, h := range cfg.Before.Hooks {
		if strings.Contains(h, "skills") && strings.Contains(h, "manifest") {
			hookOK = true
		}
	}
	if !hookOK {
		t.Errorf("before.hooks %v must generate skills-manifest.json", cfg.Before.Hooks)
	}

	if len(cfg.Archives) == 0 {
		t.Fatal("no archives configured")
	}
	files := strings.Join(cfg.Archives[0].Files, "\n")
	for _, want := range []string{"README.md", "LICENSE*", "THIRD_PARTY_NOTICES.md", "skills-manifest.json"} {
		if !strings.Contains(files, want) {
			t.Errorf("archive files %v missing %s", cfg.Archives[0].Files, want)
		}
	}

	if len(cfg.Sboms) == 0 {
		t.Fatal("no sboms configured")
	}
	sbom := cfg.Sboms[0]
	if sbom.Cmd != "syft" {
		t.Errorf("sboms cmd = %q, want syft", sbom.Cmd)
	}
	sbomText := strings.Join(append(append([]string{}, sbom.Args...), sbom.Documents...), " ")
	if !strings.Contains(sbomText, "spdx") {
		t.Errorf("sboms must emit SPDX, got args=%v documents=%v", sbom.Args, sbom.Documents)
	}

	if len(cfg.Signs) == 0 {
		t.Fatal("no signs configured")
	}
	sign := cfg.Signs[0]
	if sign.Cmd != "cosign" {
		t.Errorf("signs cmd = %q, want cosign", sign.Cmd)
	}
	argText := strings.Join(sign.Args, " ")
	for _, want := range []string{"sign-blob", "--yes", "${signature}", "${certificate}", "${artifact}"} {
		if !strings.Contains(argText, want) {
			t.Errorf("signs args missing %s: %v", want, sign.Args)
		}
	}
	if sign.Artifacts != "checksum" {
		t.Errorf("signs artifacts = %q, want exactly checksum (GoReleaser enum, singular)", sign.Artifacts)
	}

	lic, err := os.ReadFile(filepath.Join(root, "LICENSE"))
	if err != nil || len(lic) == 0 {
		t.Fatalf("LICENSE missing or empty: %v", err)
	}
	if !strings.Contains(string(lic), "MIT") {
		t.Error("LICENSE does not look like MIT")
	}
	notices, err := os.ReadFile(filepath.Join(root, "THIRD_PARTY_NOTICES.md"))
	if err != nil || len(notices) == 0 {
		t.Fatalf("THIRD_PARTY_NOTICES.md missing or empty: %v", err)
	}
	if !strings.Contains(string(notices), "Apache-2.0") {
		t.Error("THIRD_PARTY_NOTICES.md must attribute Apache-2.0 works")
	}
	if !strings.Contains(string(notices), "OpenTelemetry") {
		t.Error("THIRD_PARTY_NOTICES.md must attribute the OpenTelemetry corpus subset")
	}

	goreleaser, err := exec.LookPath("goreleaser")
	if err != nil {
		t.Fatalf("goreleaser not on PATH: %v", err)
	}
	cmd := exec.Command(goreleaser, "check")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("goreleaser check failed: %v\n%s", err, out)
	}
}

func runTool(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("%s not on PATH: %v", name, err)
	}
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
	return string(out)
}

// TestSCSBOMAndCosign is SC2: syft emits valid SPDX JSON for a built gira
// binary, and cosign signs + verifies a checksums blob with an ephemeral
// keypair inside t.TempDir().
func TestSCSBOMAndCosign(t *testing.T) {
	root := releaseRepoRoot(t)
	dir := t.TempDir()

	binPath := filepath.Join(dir, "gira")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	build := exec.Command("go", "build", "-o", binPath, "./cmd/gira")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/gira failed: %v\n%s", err, out)
	}

	sbomPath := filepath.Join(dir, "gira.spdx.sbom.json")
	runTool(t, root, nil, "syft", binPath, "-o", "spdx-json="+sbomPath, "-q")
	raw, err := os.ReadFile(sbomPath)
	if err != nil {
		t.Fatalf("read SBOM: %v", err)
	}
	var doc struct {
		SpdxVersion string `json:"spdxVersion"`
		Packages    []struct {
			Name string `json:"name"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("SBOM is not valid JSON: %v", err)
	}
	if !strings.HasPrefix(doc.SpdxVersion, "SPDX-") {
		t.Fatalf("SBOM spdxVersion = %q, want SPDX-*", doc.SpdxVersion)
	}
	if len(doc.Packages) == 0 {
		t.Fatal("SBOM has no packages")
	}

	keyPrefix := filepath.Join(dir, "cosign-test")
	runTool(t, dir, []string{"COSIGN_PASSWORD="}, "cosign", "generate-key-pair", "--output-key-prefix", keyPrefix)
	for _, p := range []string{keyPrefix + ".key", keyPrefix + ".pub"} {
		if st, err := os.Stat(p); err != nil || st.Size() == 0 {
			t.Fatalf("keypair file %s missing or empty: %v", p, err)
		}
	}

	blob := filepath.Join(dir, "checksums.txt")
	binBytes, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("read built gira binary: %v", err)
	}
	sum := sha256.Sum256(binBytes)
	if err := os.WriteFile(blob, []byte(fmt.Sprintf("%x  gira\n", sum)), 0o644); err != nil {
		t.Fatalf("write checksums blob: %v", err)
	}

	bundle := filepath.Join(dir, "checksums.bundle.json")
	runTool(t, dir, []string{"COSIGN_PASSWORD="}, "cosign",
		"sign-blob", "--yes", "--key", keyPrefix+".key", "--bundle", bundle, blob)
	if st, err := os.Stat(bundle); err != nil || st.Size() == 0 {
		t.Fatalf("bundle %s missing or empty: %v", bundle, err)
	}
	runTool(t, dir, nil, "cosign",
		"verify-blob", "--key", keyPrefix+".pub", "--bundle", bundle, blob)

	if err := os.WriteFile(blob, []byte("tampered checksums"), 0o644); err != nil {
		t.Fatal(err)
	}
	cosignPath, _ := exec.LookPath("cosign")
	verify := exec.Command(cosignPath, "verify-blob", "--key", keyPrefix+".pub", "--bundle", bundle, blob)
	verify.Dir = dir
	if out, err := verify.CombinedOutput(); err == nil {
		t.Fatalf("verify-blob accepted a tampered blob:\n%s", out)
	}

	if _, err := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(err) {
		t.Fatal("test left dist/ behind in the repo root")
	}
}
