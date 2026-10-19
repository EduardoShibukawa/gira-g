package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/EduardoShibukawa/gira-g/internal/domain"
	"github.com/EduardoShibukawa/gira-g/internal/gitcache"
	"github.com/EduardoShibukawa/gira-g/internal/manifest"
)

// Domain tags for cryptographic namespace separation.
const (
	TagSourceKey = "gira.index.source/v1"
	TagToolID    = "gira.index.tool/v1"
	TagVersionID = "gira.index.version/v1"
)

// ErrNonDeterministicBuild indicates that re-indexing produced divergent canonical
// bytes for the same version_id. Publication fails closed without overwrite.
var ErrNonDeterministicBuild = errors.New("nondeterministic_build: canonical bytes differ for existing version_id")

// InjectiveEncoder encodes variable-length records with explicit big-endian lengths,
// guaranteeing provable injectivity: no two distinct input tuples can produce the same bytes.
type InjectiveEncoder struct {
	buf bytes.Buffer
}

// NewInjectiveEncoder initializes an encoder with a domain separation tag.
func NewInjectiveEncoder(domainTag string) *InjectiveEncoder {
	enc := &InjectiveEncoder{}
	enc.WriteString(domainTag)
	return enc
}

// WriteBytes writes an 8-byte big-endian length prefix followed by b.
func (e *InjectiveEncoder) WriteBytes(b []byte) {
	var lenBuf [8]byte
	binary.BigEndian.PutUint64(lenBuf[:], uint64(len(b)))
	e.buf.Write(lenBuf[:])
	e.buf.Write(b)
}

// WriteString writes an 8-byte big-endian length prefix followed by string s.
func (e *InjectiveEncoder) WriteString(s string) {
	e.WriteBytes([]byte(s))
}

// WriteUint64 writes an 8-byte big-endian unsigned integer.
func (e *InjectiveEncoder) WriteUint64(v uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	e.buf.Write(buf[:])
}

// Bytes returns a copy of the encoded buffer bytes.
func (e *InjectiveEncoder) Bytes() []byte {
	return e.buf.Bytes()
}

// Sum256 computes the SHA-256 digest of the encoded buffer.
func (e *InjectiveEncoder) Sum256() [32]byte {
	return sha256.Sum256(e.buf.Bytes())
}

// SumHex returns the 64-character lowercase hexadecimal SHA-256 digest.
func (e *InjectiveEncoder) SumHex() string {
	sum := e.Sum256()
	return hex.EncodeToString(sum[:])
}

// InjectiveDecoder decodes fields serialized by InjectiveEncoder.
type InjectiveDecoder struct {
	reader *bytes.Reader
}

// NewInjectiveDecoder initializes a decoder over data.
func NewInjectiveDecoder(data []byte) *InjectiveDecoder {
	return &InjectiveDecoder{reader: bytes.NewReader(data)}
}

// ReadBytes reads a length-prefixed byte slice.
func (d *InjectiveDecoder) ReadBytes() ([]byte, error) {
	var lenBuf [8]byte
	if _, err := io.ReadFull(d.reader, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint64(lenBuf[:])
	buf := make([]byte, length)
	if _, err := io.ReadFull(d.reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// ReadString reads a length-prefixed string.
func (d *InjectiveDecoder) ReadString() (string, error) {
	b, err := d.ReadBytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ReadUint64 reads an 8-byte big-endian unsigned integer.
func (d *InjectiveDecoder) ReadUint64() (uint64, error) {
	var buf [8]byte
	if _, err := io.ReadFull(d.reader, buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf[:]), nil
}

// Remaining returns the number of unread bytes remaining in the decoder.
func (d *InjectiveDecoder) Remaining() int {
	return d.reader.Len()
}

// SourceKeyOptions configures the inputs for source_key derivation.
type SourceKeyOptions struct {
	SchemaVersion   any
	ManifestLocator string
	ManifestBytes   []byte
	Sources         []domain.SourceRepoIdentity
}

// ComputeSourceKey computes the deterministic SHA-256 hex digest for a manifest invocation.
// It binds the raw manifest bytes, normalized locator, and sorted source provenance.
func ComputeSourceKey(opts SourceKeyOptions) string {
	enc := NewInjectiveEncoder(TagSourceKey)

	schemaVer := "2"
	if opts.SchemaVersion != nil {
		schemaVer = fmt.Sprint(opts.SchemaVersion)
	}
	enc.WriteString(schemaVer)
	enc.WriteString(opts.ManifestLocator)

	manifestSum := sha256.Sum256(opts.ManifestBytes)
	manifestHex := hex.EncodeToString(manifestSum[:])
	enc.WriteString(manifestHex)
	enc.WriteBytes(opts.ManifestBytes)

	// Sort sources deterministically by repo_id
	sortedSources := make([]domain.SourceRepoIdentity, len(opts.Sources))
	copy(sortedSources, opts.Sources)
	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].RepoID < sortedSources[j].RepoID
	})

	enc.WriteUint64(uint64(len(sortedSources)))
	for _, src := range sortedSources {
		enc.WriteString(src.RepoID)
		enc.WriteString(src.URLSHA256)
		commitSHA := src.CommitSHA
		if commitSHA == "" {
			commitSHA = src.SHA
		}
		enc.WriteString(commitSHA)
		enc.WriteString(src.PathScope)
	}

	return enc.SumHex()
}

// ToolIDOptions configures the inputs for tool_id derivation.
type ToolIDOptions struct {
	Version        string
	Commit         string
	BuildInfo      *debug.BuildInfo
	ExecutableHash string
	ExecutablePath string
}

// ComputeToolID derives the binary identity hash implementing Option C (Gate 3).
// The entire -ldflags setting is excluded, ensuring main.date variations do not
// change tool_id or invalidate published release indexes. Untrusted development
// builds include the SHA-256 hash of the executable binary.
func ComputeToolID(opts ToolIDOptions) (string, error) {
	enc := NewInjectiveEncoder(TagToolID)

	v := opts.Version
	if v == "" {
		v = "0.1.0-dev"
	}
	c := opts.Commit
	if c == "" {
		c = "none"
	}
	enc.WriteString(v)
	enc.WriteString(c)

	bi := opts.BuildInfo
	if bi == nil {
		bi, _ = debug.ReadBuildInfo()
	}

	var goVersion, mainPath, mainVersion, mainSum string
	var filteredSettings []debug.BuildSetting
	var sortedDeps []*debug.Module

	if bi != nil {
		goVersion = bi.GoVersion
		mainPath = bi.Path
		mainVersion = bi.Main.Version
		mainSum = bi.Main.Sum

		// Option C: exclude the entire "-ldflags" setting.
		for _, s := range bi.Settings {
			if s.Key == "-ldflags" {
				continue
			}
			filteredSettings = append(filteredSettings, s)
		}
		sort.Slice(filteredSettings, func(i, j int) bool {
			if filteredSettings[i].Key != filteredSettings[j].Key {
				return filteredSettings[i].Key < filteredSettings[j].Key
			}
			return filteredSettings[i].Value < filteredSettings[j].Value
		})

		sortedDeps = append([]*debug.Module(nil), bi.Deps...)
		sort.Slice(sortedDeps, func(i, j int) bool {
			return sortedDeps[i].Path < sortedDeps[j].Path
		})
	}

	enc.WriteString(goVersion)
	enc.WriteString(mainPath)
	enc.WriteString(mainVersion)
	enc.WriteString(mainSum)

	enc.WriteUint64(uint64(len(filteredSettings)))
	for _, s := range filteredSettings {
		enc.WriteString(s.Key)
		enc.WriteString(s.Value)
	}

	enc.WriteUint64(uint64(len(sortedDeps)))
	for _, dep := range sortedDeps {
		if dep == nil {
			continue
		}
		enc.WriteString(dep.Path)
		enc.WriteString(dep.Version)
		enc.WriteString(dep.Sum)
		if dep.Replace != nil {
			enc.WriteString(dep.Replace.Path)
			enc.WriteString(dep.Replace.Version)
			enc.WriteString(dep.Replace.Sum)
		} else {
			enc.WriteString("")
			enc.WriteString("")
			enc.WriteString("")
		}
	}

	// Untrusted dev build evaluation:
	// If commit is "none", "0.1.0-dev", vcs.modified is true, or vcs.revision is absent/empty.
	isUntrusted := c == "none" || c == "0.1.0-dev"
	if bi != nil {
		hasRevision := false
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && strings.TrimSpace(s.Value) != "" {
				hasRevision = true
			}
			if s.Key == "vcs.modified" && s.Value == "true" {
				isUntrusted = true
			}
		}
		if !hasRevision {
			isUntrusted = true
		}
	} else {
		isUntrusted = true
	}

	if isUntrusted {
		exeHash := opts.ExecutableHash
		if exeHash == "" {
			var err error
			exeHash, err = ComputeExecutableHash(opts.ExecutablePath)
			if err != nil {
				return "", fmt.Errorf("failed computing executable hash for untrusted build: %w", err)
			}
		}
		enc.WriteString(exeHash)
	} else {
		enc.WriteString("")
	}

	return enc.SumHex(), nil
}

// ComputeExecutableHash computes the SHA-256 hex digest of the running binary (or path).
func ComputeExecutableHash(customPath string) (string, error) {
	target := customPath
	if target == "" {
		var err error
		target, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	f, err := os.Open(target)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeVersionID calculates the deterministic version_id:
// version_id = H(TagVersionID, source_key, tool_id).
func ComputeVersionID(sourceKey, toolID string) string {
	enc := NewInjectiveEncoder(TagVersionID)
	enc.WriteString(sourceKey)
	enc.WriteString(toolID)
	return enc.SumHex()
}

// BuildSourceIdentityFromManifest constructs a canonical SourceIdentity from raw manifest
// bytes, repository declarations, and extractor definitions.
func BuildSourceIdentityFromManifest(
	manifestBytes []byte,
	repos []manifest.RepositoryDeclaration,
	extractors []domain.ExtractorIdentity,
) *domain.SourceIdentity {
	manifestSum := sha256.Sum256(manifestBytes)
	manifestHex := hex.EncodeToString(manifestSum[:])

	sortedExtractors := make([]domain.ExtractorIdentity, len(extractors))
	copy(sortedExtractors, extractors)
	sort.Slice(sortedExtractors, func(i, j int) bool {
		return sortedExtractors[i].ID < sortedExtractors[j].ID
	})

	sources := make([]domain.SourceRepoIdentity, len(repos))
	for i, r := range repos {
		urlHash := gitcache.RepoHash(r.URL)
		sources[i] = domain.SourceRepoIdentity{
			RepoID:    r.ID,
			CommitSHA: r.SHA,
			SHA:       r.SHA,
			PathScope: r.PathScope,
			URLSHA256: urlHash,
		}
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].RepoID < sources[j].RepoID
	})

	return &domain.SourceIdentity{
		SchemaVersion:  2,
		ManifestSHA256: manifestHex,
		Extractors:     sortedExtractors,
		Sources:        sources,
	}
}

// VerifyDeterministicBuild checks if an existing published index at existingPath matches
// canonicalBytes. If existingPath exists and its content differs, ErrNonDeterministicBuild
// is returned without modifying or corrupting the existing file. If content matches or the
// file does not exist, nil is returned.
func VerifyDeterministicBuild(existingPath string, canonicalBytes []byte) error {
	data, err := os.ReadFile(existingPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read existing index at %s: %w", existingPath, err)
	}
	if !bytes.Equal(data, canonicalBytes) {
		return fmt.Errorf("%w: existing index at %s differs from newly built index", ErrNonDeterministicBuild, existingPath)
	}
	return nil
}
