// L1 artifact-verified closing line (spec gira-evidence-gate F1, SC1).
//
// COPY = idea only, reimplemented in Go: the contract (success-only marker +
// canonical digest of CHECK/EXPECT/CWD + exit + frozen fingerprint + L3
// taxonomy, 1 MiB cap, out-of-scope writes fail closed, 5-point dispatcher
// self-check) follows the unlazy gate and Gallon L1 methodology, with the
// autoresearch rule that one canonical evaluation preimage serves every entry
// point so any digest re-derives from the run output alone. No JS is pasted.
package eval

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ArtifactClosingPrefix is the mandatory closing-line prefix. Every eval/verify
// run emits Prefix+<hex> as its final line.
const ArtifactClosingPrefix = "ARTIFACT-VERIFIED sha256:"

// ArtifactFrozenCap caps how many frozen input bytes enter the fingerprint
// (unlazy contract: 1 MiB).
const ArtifactFrozenCap = 1 << 20

// ArtifactRun is one eval/verify dispatch: what ran, on what frozen inputs,
// under which claim taxonomy, and the closing line it emitted.
type ArtifactRun struct {
	Check        string
	Expect       string
	CWD          string
	ExitCode     int
	Taxonomy     string
	Inputs       map[string][]byte
	ClosingLine  string
	AllowedScope []string
}

// FingerprintInputs hashes the frozen inputs deterministically: sorted paths,
// each as path + 0x00 + content, truncated so at most ArtifactFrozenCap
// content bytes enter the hash. Bytes beyond the cap do not affect the digest.
func FingerprintInputs(inputs map[string][]byte) string {
	paths := make([]string, 0, len(inputs))
	for p := range inputs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	h := sha256.New()
	remaining := ArtifactFrozenCap
	for _, p := range paths {
		h.Write([]byte(filepath.ToSlash(p)))
		h.Write([]byte{0})
		content := inputs[p]
		if remaining <= 0 {
			continue
		}
		if len(content) > remaining {
			content = content[:remaining]
		}
		h.Write(content)
		remaining -= len(content)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ArtifactDigest covers command, success marker, working directory, exit
// code, the frozen input fingerprint and the L3 claim taxonomy (I2). It is
// the single canonical constructor: every entry point (CLI eval/verify, MCP
// verify_artifact) shares it, so one digest re-derives from the run output
// alone. Any field flip changes the digest.
func ArtifactDigest(check, expect, cwd string, exit int, fingerprintHex, taxonomy string) string {
	h := sha256.New()
	fmt.Fprintf(h, "CHECK:%s\nEXPECT:%s\nCWD:%s\nEXIT:%d\nFROZEN:%s\nTAXONOMY:%s\n",
		check, expect, cwd, exit, fingerprintHex, taxonomy)
	return hex.EncodeToString(h.Sum(nil))
}

// FormatClosingLine renders the mandatory final line for a digest.
func FormatClosingLine(digestHex string) string {
	return ArtifactClosingPrefix + digestHex
}

// ParseClosingLine accepts the closing line with surrounding whitespace/newline
// and returns the digest hex.
func ParseClosingLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ArtifactClosingPrefix) {
		return "", false
	}
	digest := strings.TrimPrefix(trimmed, ArtifactClosingPrefix)
	digest = strings.TrimSpace(digest)
	if len(digest) != 64 {
		return "", false
	}
	for _, c := range digest {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return "", false
		}
	}
	return strings.ToLower(digest), true
}

// VerifyClosingLine recomputes the expected digest and compares it with the
// closing line. A tampered digest, or the wrong taxonomy, never verifies.
func VerifyClosingLine(line, check, expect, cwd string, exit int, fingerprintHex, taxonomy string) bool {
	digest, ok := ParseClosingLine(line)
	if !ok {
		return false
	}
	return digest == ArtifactDigest(check, expect, cwd, exit, fingerprintHex, taxonomy)
}

// EnforceWriteScope fails closed when any attempted write falls outside the
// frozen scope. An empty scope allows nothing. Matching is by slash-cleaned
// prefix; a scope entry without trailing slash matches the exact file and
// everything beneath it.
func EnforceWriteScope(writes, allowedScope []string) error {
	if len(writes) == 0 {
		return nil
	}
	clean := func(p string) string {
		return filepath.ToSlash(filepath.Clean("/" + p))
	}
	scopes := make([]string, 0, len(allowedScope))
	for _, s := range allowedScope {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		scopes = append(scopes, clean(s))
	}
	for _, w := range writes {
		cw := clean(w)
		allowed := false
		for _, s := range scopes {
			if cw == s || strings.HasPrefix(cw, s+"/") {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("write outside frozen scope: %q", w)
		}
	}
	return nil
}

// DispatcherSelfCheck runs the 5 dispatcher points and returns the names of
// the failing points (empty = pass):
//  1. existence: every declared input is present on read-back.
//  2. harness-digest: the closing line parses.
//  3. own-run: the closing line matches this run's recomputed digest.
//  4. hash-recompute: the read-back fingerprint matches the run's fingerprint.
//  5. read-for-substitute: read-back contents equal the run's inputs.
func DispatcherSelfCheck(run ArtifactRun, readBack map[string][]byte) []string {
	var failures []string
	for p, content := range run.Inputs {
		rb, ok := readBack[p]
		if !ok {
			failures = append(failures, "existence:"+p)
			continue
		}
		_ = content
		_ = rb
	}
	for p := range readBack {
		if _, ok := run.Inputs[p]; !ok {
			failures = append(failures, "existence:extra:"+p)
		}
	}
	digest, ok := ParseClosingLine(run.ClosingLine)
	if !ok {
		failures = append(failures, "harness-digest")
		return failures
	}
	wantFP := FingerprintInputs(run.Inputs)
	wantDigest := ArtifactDigest(run.Check, run.Expect, run.CWD, run.ExitCode, wantFP, run.Taxonomy)
	if digest != wantDigest {
		failures = append(failures, "own-run")
	}
	if gotFP := FingerprintInputs(readBack); gotFP != wantFP {
		failures = append(failures, "hash-recompute")
	}
	mismatch := false
	if len(readBack) != len(run.Inputs) {
		mismatch = true
	} else {
		for p, content := range run.Inputs {
			rb, ok := readBack[p]
			if !ok || string(rb) != string(content) {
				mismatch = true
				break
			}
		}
	}
	if mismatch {
		failures = append(failures, "read-for-substitute")
	}
	return failures
}

// EmitWithArtifact appends the run's closing line as the final output line.
// Callers that disable the artifact line (debug-only --no-artifact-verified)
// skip this function; production runs always end with it.
func EmitWithArtifact(body string, run ArtifactRun) string {
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return body + run.ClosingLine + "\n"
}
