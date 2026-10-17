package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

func TestInstallDocParity(t *testing.T) {
	root := releaseRepoRoot(t)

	t.Run("READMEDocumentsCommandsAndPublicInstall", func(t *testing.T) {
		raw, err := os.ReadFile(filepath.Join(root, "README.md"))
		if err != nil {
			t.Fatalf("read README.md: %v", err)
		}
		readme := string(raw)

		required := []string{
			"gira serve mcp",
			"gira init",
			"gira index",
			"gira impact",
			"gira slice",
			"gira explain",
			"gira tour",
			"gira flows",
			"gira doc",
			"gira diff",
			"gira locate",
			"gira export",
			"gira view",
			"gira skills",
			"gira eval",
			"gira verify",
			"gira review",
			"gira gate",
			"gira gate-lint",
		}
		for _, want := range required {
			if !strings.Contains(readme, want) {
				t.Errorf("README.md missing documented command %q", want)
			}
		}

		forbiddenPrivateClaims := []string{
			"O repositório é privado",
			"Because the repository is private",
			"repo privado exige token",
		}
		for _, bad := range forbiddenPrivateClaims {
			if strings.Contains(readme, bad) {
				t.Errorf("README.md still claims repository is private-only (%q)", bad)
			}
		}
	})

	t.Run("CLIReferenceCoversAllPrintUsageCommands", func(t *testing.T) {
		var buf bytes.Buffer
		printUsage(&buf)
		usage := buf.String()

		var usageCmds []string
		seen := map[string]bool{}
		inCommands := false
		for _, line := range strings.Split(usage, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "Commands:" {
				inCommands = true
				continue
			}
			if inCommands && strings.HasSuffix(trimmed, "Flags:") {
				break
			}
			if !inCommands || trimmed == "" {
				continue
			}
			fields := strings.Fields(trimmed)
			if len(fields) == 0 {
				continue
			}
			cmd := fields[0]
			if cmd == "version" {
				continue
			}
			if !seen[cmd] {
				seen[cmd] = true
				usageCmds = append(usageCmds, cmd)
			}
		}
		if len(usageCmds) == 0 {
			t.Fatal("failed to parse commands from printUsage")
		}

		for _, cliDirRel := range []string{
			filepath.Join("website", "docs", "cli"),
			filepath.Join("website", "versioned_docs", "version-0.1.0", "cli"),
		} {
			cliDir := filepath.Join(root, cliDirRel)
			entries, err := os.ReadDir(cliDir)
			if err != nil {
				t.Fatalf("read %s: %v", cliDirRel, err)
			}
			var combined strings.Builder
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				b, err := os.ReadFile(filepath.Join(cliDir, e.Name()))
				if err != nil {
					t.Fatalf("read %s/%s: %v", cliDirRel, e.Name(), err)
				}
				combined.Write(b)
				combined.WriteByte('\n')
			}
			allDocs := combined.String()
			for _, cmd := range usageCmds {
				needle := "gira " + cmd
				if !strings.Contains(allDocs, needle) {
					t.Errorf("%s does not document command %q (from printUsage)", cliDirRel, needle)
				}
			}
		}
	})

	t.Run("ServeDocsCoverAllMCPToolsAndScorecardExists", func(t *testing.T) {
		serveDoc, err := os.ReadFile(filepath.Join(root, "website", "docs", "cli", "serve.md"))
		if err != nil {
			t.Fatalf("read serve.md: %v", err)
		}
		serveText := string(serveDoc)
		mcpTools := []string{
			"get_impact_briefing",
			"get_pr_impact",
			"get_tour",
			"get_doc",
			"diff_maps",
			"get_context_slice",
			"explain_evidence",
			"locate_boundary",
			"verify_artifact",
			"get_control_total",
			"get_findings_register",
			"get_token_accounting",
			"gate_status",
		}
		for _, tool := range mcpTools {
			if !strings.Contains(serveText, tool) {
				t.Errorf("website/docs/cli/serve.md missing MCP tool %q", tool)
			}
		}

		scorecardDoc, err := os.ReadFile(filepath.Join(root, "website", "docs", "concepts", "scorecard.md"))
		if err != nil {
			t.Fatalf("read scorecard.md: %v", err)
		}
		scText := string(scorecardDoc)
		for _, want := range []string{"otel-demo-3.1.0-mini", "100.0%", "82.4%", "81.2%"} {
			if !strings.Contains(scText, want) {
				t.Errorf("website/docs/concepts/scorecard.md missing metric %q", want)
			}
		}
	})

	t.Run("InstallShZeroTokenDownloadAndChecksumVerification", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("install.sh POSIX shell test runs on Linux/macOS")
		}

		assetName := fmt.Sprintf("gira_0.1.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		fakeBin := []byte("#!/bin/sh\necho 'gira 0.1.0-test'\n")
		tarball := buildTarGzWithBinary(t, "gira", fakeBin)
		goodSum := fmt.Sprintf("%x", sha256.Sum256(tarball))
		badSum := strings.Repeat("0", 64)

		var serveTampered atomic.Bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				t.Errorf("unexpected Authorization header on zero-token public path: %q", r.Header.Get("Authorization"))
			}
			switch r.URL.Path {
			case "/latest":
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintln(w, `{"tag_name": "v0.1.0"}`)
			case "/v0.1.0/" + assetName:
				w.Header().Set("Content-Type", "application/octet-stream")
				_, _ = w.Write(tarball)
			case "/v0.1.0/checksums.txt":
				w.Header().Set("Content-Type", "text/plain")
				sum := goodSum
				if serveTampered.Load() {
					sum = badSum
				}
				fmt.Fprintf(w, "%s  %s\n", sum, assetName)
			default:
				http.NotFound(w, r)
			}
		}))
		defer srv.Close()

		tmp := t.TempDir()
		// Place an unauthenticated `gh` shim first in PATH so even if `gh` is
		// installed on the runner, `gh auth status` returns 1.
		fakeBinDir := filepath.Join(tmp, "fakebin")
		if err := os.MkdirAll(fakeBinDir, 0o755); err != nil {
			t.Fatal(err)
		}
		ghShim := filepath.Join(fakeBinDir, "gh")
		if err := os.WriteFile(ghShim, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		installDir := filepath.Join(tmp, "bin")
		scriptPath := filepath.Join(root, "install.sh")

		runInstall := func() (string, error) {
			cmd := exec.Command("sh", scriptPath, "--dir", installDir)
			var env []string
			for _, kv := range os.Environ() {
				if strings.HasPrefix(kv, "GITHUB_TOKEN=") || strings.HasPrefix(kv, "PATH=") {
					continue
				}
				env = append(env, kv)
			}
			env = append(env,
				"PATH="+fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GIRA_RELEASE_BASE_URL="+srv.URL,
			)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			return string(out), err
		}

		// 1. Happy path: resolves "latest" -> v0.1.0, downloads tarball + checksums.txt, verifies SHA-256, extracts binary.
		out, err := runInstall()
		if err != nil {
			t.Fatalf("install.sh failed on valid checksum: %v\n%s", err, out)
		}
		installedBin := filepath.Join(installDir, "gira")
		st, err := os.Stat(installedBin)
		if err != nil || st.Size() == 0 {
			t.Fatalf("installed binary missing at %s: %v", installedBin, err)
		}
		if !strings.Contains(out, "gira 0.1.0-test") {
			t.Errorf("expected installed binary version output in %q", out)
		}

		// 2. Tampered SHA-256 checksum must fail closed without overwriting/installing.
		serveTampered.Store(true)
		tamperedDir := filepath.Join(tmp, "bin-tampered")
		cmd := exec.Command("sh", scriptPath, "--version", "v0.1.0", "--dir", tamperedDir)
		cmd.Env = []string{
			"PATH=" + fakeBinDir + string(os.PathListSeparator) + os.Getenv("PATH"),
			"GIRA_RELEASE_BASE_URL=" + srv.URL,
			"HOME=" + tmp,
		}
		tamperedOut, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("install.sh succeeded on tampered SHA-256 checksum:\n%s", tamperedOut)
		}
		if !strings.Contains(string(tamperedOut), "checksum divergente") {
			t.Errorf("expected checksum mismatch error, got:\n%s", tamperedOut)
		}
		if _, err := os.Stat(filepath.Join(tamperedDir, "gira")); !os.IsNotExist(err) {
			t.Errorf("binary must not be installed when checksum verification fails")
		}
	})
}

func buildTarGzWithBinary(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
