package pack

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	// RegistryFileName is the repository-owned registry configuration.
	RegistryFileName = ".github/carl/registries.json"
	// InstalledPacksFileName records provenance for registry-managed packs.
	InstalledPacksFileName = ".github/carl/installed-packs.json"

	registryIndexMaxBytes = int64(2 << 20)
	registryPackMaxBytes  = int64(1 << 20)
)

// Registries is the schema-versioned repository registry configuration.
type Registries struct {
	SchemaVersion int        `json:"schemaVersion"`
	Registries    []Registry `json:"registries"`
}

// Registry identifies an explicit pack-index location.
type Registry struct {
	ID       string `json:"id"`
	Location string `json:"location"`
}

// RegistryIndex is the schema-versioned document served by a registry.
type RegistryIndex struct {
	SchemaVersion int               `json:"schemaVersion"`
	Packs         []RegistryRelease `json:"packs"`
}

// RegistryRelease describes one immutable pack artifact.
type RegistryRelease struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Artifact    string `json:"artifact"`
	SHA256      string `json:"sha256"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// InstalledPacks is the schema-versioned registry provenance manifest.
type InstalledPacks struct {
	SchemaVersion int             `json:"schemaVersion"`
	Packs         []InstalledPack `json:"packs"`
}

// InstalledPack records the verified source of one registry-managed pack.
type InstalledPack struct {
	ID               string `json:"id"`
	Version          string `json:"version"`
	Registry         string `json:"registry"`
	RegistryLocation string `json:"registryLocation"`
	Artifact         string `json:"artifact"`
	SHA256           string `json:"sha256"`
	InstalledPath    string `json:"installedPath"`
}

// RegistryCandidate is a release plus the explicit registry that supplied it.
type RegistryCandidate struct {
	Registry Registry
	Release  RegistryRelease
}

// RegistryFetcher retrieves bounded HTTPS resources. Local repository
// registries are read directly and never cross this interface.
type RegistryFetcher interface {
	Fetch(ctx context.Context, location string, maxBytes int64) ([]byte, error)
}

type httpRegistryFetcher struct {
	client *http.Client
}

func newHTTPRegistryFetcher() RegistryFetcher {
	return &httpRegistryFetcher{client: &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if len(via) == 0 {
				return nil
			}
			first := via[0].URL
			if req.URL.Scheme != first.Scheme || !strings.EqualFold(req.URL.Host, first.Host) {
				return errors.New("cross-origin redirect rejected")
			}
			if req.URL.User != nil || req.URL.RawQuery != "" || req.URL.Fragment != "" {
				return errors.New("redirect containing credentials, query, or fragment rejected")
			}
			return nil
		},
	}}
}

func (f *httpRegistryFetcher) Fetch(ctx context.Context, location string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, fmt.Errorf("create registry request: %w", err)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", location, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("fetch %s: unexpected HTTP status %s", location, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", location, err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("fetch %s: response exceeds %d bytes", location, maxBytes)
	}
	return data, nil
}

// ReadRegistries loads strict registry configuration. Absence is represented
// as nil so registry commands can distinguish not-configured from empty.
func ReadRegistries(rootDir string) (*Registries, error) {
	data, err := readRepoFile(rootDir, RegistryFileName, registryIndexMaxBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", RegistryFileName, err)
	}
	var registries Registries
	if err := decodeStrictJSON(data, &registries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", RegistryFileName, err)
	}
	if issues := ValidateRegistries(&registries); len(issues) > 0 {
		return nil, fmt.Errorf("%s validation failed:\n- %s", RegistryFileName, joinIssues(issues))
	}
	sort.Slice(registries.Registries, func(i, j int) bool {
		return registries.Registries[i].ID < registries.Registries[j].ID
	})
	return &registries, nil
}

// ValidateRegistries returns deterministic validation issues.
func ValidateRegistries(registries *Registries) []string {
	if registries == nil {
		return []string{"registry configuration is nil"}
	}
	var issues []string
	if registries.SchemaVersion != metadataSchemaVersion {
		issues = append(issues, fmt.Sprintf("unsupported schema version %d", registries.SchemaVersion))
	}
	seen := map[string]bool{}
	for _, registry := range registries.Registries {
		if !contextIDRE.MatchString(registry.ID) {
			issues = append(issues, fmt.Sprintf("malformed registry id %q", registry.ID))
		} else if seen[registry.ID] {
			issues = append(issues, fmt.Sprintf("duplicate registry id %q", registry.ID))
		}
		seen[registry.ID] = true
		if err := validateRegistryLocation(registry.Location); err != nil {
			issues = append(issues, fmt.Sprintf("registry %s: %v", registry.ID, err))
		}
	}
	sort.Strings(issues)
	return issues
}

// ReadInstalledPacks loads registry provenance. Absence is an empty manifest.
func ReadInstalledPacks(rootDir string) (*InstalledPacks, error) {
	data, err := readRepoFile(rootDir, InstalledPacksFileName, registryIndexMaxBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return &InstalledPacks{SchemaVersion: metadataSchemaVersion, Packs: []InstalledPack{}}, nil
		}
		return nil, fmt.Errorf("read %s: %w", InstalledPacksFileName, err)
	}
	var installed InstalledPacks
	if err := decodeStrictJSON(data, &installed); err != nil {
		return nil, fmt.Errorf("parse %s: %w", InstalledPacksFileName, err)
	}
	if issues := ValidateInstalledPacks(&installed); len(issues) > 0 {
		return nil, fmt.Errorf("%s validation failed:\n- %s", InstalledPacksFileName, joinIssues(issues))
	}
	normalizeInstalledPacks(&installed)
	return &installed, nil
}

// ValidateInstalledPacks validates provenance without trusting its paths or
// source claims.
func ValidateInstalledPacks(installed *InstalledPacks) []string {
	if installed == nil {
		return []string{"installed-pack provenance is nil"}
	}
	var issues []string
	if installed.SchemaVersion != metadataSchemaVersion {
		issues = append(issues, fmt.Sprintf("unsupported schema version %d", installed.SchemaVersion))
	}
	seen := map[string]bool{}
	for _, entry := range installed.Packs {
		if !packIDRE.MatchString(entry.ID) {
			issues = append(issues, fmt.Sprintf("malformed pack id %q", entry.ID))
		} else if seen[entry.ID] {
			issues = append(issues, fmt.Sprintf("duplicate installed pack %q", entry.ID))
		}
		seen[entry.ID] = true
		if _, err := parseSemanticVersion(entry.Version); err != nil {
			issues = append(issues, fmt.Sprintf("%s: invalid version %q", entry.ID, entry.Version))
		}
		if !contextIDRE.MatchString(entry.Registry) {
			issues = append(issues, fmt.Sprintf("%s: malformed registry id %q", entry.ID, entry.Registry))
		}
		if err := validateRegistryLocation(entry.RegistryLocation); err != nil {
			issues = append(issues, fmt.Sprintf("%s: invalid registry location: %v", entry.ID, err))
		}
		if err := validateArtifactLocation(entry.Artifact); err != nil {
			issues = append(issues, fmt.Sprintf("%s: invalid artifact: %v", entry.ID, err))
		}
		if !isSHA256(entry.SHA256) {
			issues = append(issues, fmt.Sprintf("%s: invalid SHA-256 digest", entry.ID))
		}
		if entry.InstalledPath != expectedPackPath(entry.ID) {
			issues = append(issues, fmt.Sprintf("%s: contradictory installed path %q", entry.ID, entry.InstalledPath))
		}
	}
	sort.Strings(issues)
	return issues
}

func marshalInstalledPacks(installed *InstalledPacks) ([]byte, error) {
	copyValue := *installed
	copyValue.Packs = append([]InstalledPack(nil), installed.Packs...)
	normalizeInstalledPacks(&copyValue)
	if issues := ValidateInstalledPacks(&copyValue); len(issues) > 0 {
		return nil, fmt.Errorf("%s validation failed:\n- %s", InstalledPacksFileName, joinIssues(issues))
	}
	data, err := json.MarshalIndent(copyValue, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal %s: %w", InstalledPacksFileName, err)
	}
	return append(data, '\n'), nil
}

func normalizeInstalledPacks(installed *InstalledPacks) {
	if installed.Packs == nil {
		installed.Packs = []InstalledPack{}
	}
	sort.Slice(installed.Packs, func(i, j int) bool {
		return installed.Packs[i].ID < installed.Packs[j].ID
	})
}

func decodeStrictJSON(data []byte, target any) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	return ensureJSONEOF(dec)
}

func validateRegistryLocation(location string) error {
	if strings.TrimSpace(location) != location || location == "" {
		return errors.New("location must be non-empty without surrounding whitespace")
	}
	if containsControl(location) {
		return errors.New("location cannot contain control characters")
	}
	if strings.Contains(location, "\\") {
		return errors.New("location must use forward slashes")
	}
	if strings.Contains(location, "://") {
		u, err := url.Parse(location)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		if u.Scheme != "https" || u.Host == "" {
			return errors.New("remote location must use HTTPS with a host")
		}
		if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return errors.New("remote location cannot contain credentials, query, or fragment")
		}
		return nil
	}
	if strings.Contains(location, ":") {
		return errors.New("local location cannot contain a colon")
	}
	return validateRepoRelativePath(location, ".json")
}

func validateArtifactLocation(location string) error {
	if strings.TrimSpace(location) != location || location == "" {
		return errors.New("artifact must be non-empty without surrounding whitespace")
	}
	if containsControl(location) {
		return errors.New("artifact cannot contain control characters")
	}
	if strings.Contains(location, "\\") {
		return errors.New("artifact must use forward slashes")
	}
	u, err := url.Parse(location)
	if err != nil {
		return fmt.Errorf("invalid artifact location: %w", err)
	}
	if u.IsAbs() || u.Host != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("artifact must be a relative path without credentials, query, or fragment")
	}
	return validateRepoRelativePath(location, ".instructions.md")
}

func validateRepoRelativePath(value, suffix string) error {
	if value == "" || strings.ContainsRune(value, '\x00') || filepath.IsAbs(value) || path.IsAbs(value) {
		return errors.New("path must be repository-relative")
	}
	clean := path.Clean(value)
	if clean == "." || clean != value || clean == ".." || strings.HasPrefix(clean, "../") {
		return errors.New("path traversal or non-canonical path rejected")
	}
	if !strings.HasSuffix(clean, suffix) {
		return fmt.Errorf("path must end in %s", suffix)
	}
	return nil
}

func containsControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func validateRegistryIndex(index *RegistryIndex) []string {
	if index == nil {
		return []string{"registry index is nil"}
	}
	var issues []string
	if index.SchemaVersion != metadataSchemaVersion {
		issues = append(issues, fmt.Sprintf("unsupported schema version %d", index.SchemaVersion))
	}
	seen := map[string]bool{}
	versionsByID := map[string][]string{}
	for _, release := range index.Packs {
		label := release.ID + "@" + release.Version
		if !packIDRE.MatchString(release.ID) {
			issues = append(issues, fmt.Sprintf("malformed pack id %q", release.ID))
		}
		if _, err := parseSemanticVersion(release.Version); err != nil {
			issues = append(issues, fmt.Sprintf("%s: invalid version", label))
		}
		if err := validateArtifactLocation(release.Artifact); err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", label, err))
		}
		if !isSHA256(release.SHA256) {
			issues = append(issues, fmt.Sprintf("%s: invalid SHA-256 digest", label))
		}
		if seen[label] {
			issues = append(issues, fmt.Sprintf("duplicate release %s", label))
		}
		seen[label] = true
		for _, previous := range versionsByID[release.ID] {
			cmp, err := compareSemanticVersions(previous, release.Version)
			if err == nil && cmp == 0 && previous != release.Version {
				issues = append(
					issues,
					fmt.Sprintf(
						"%s: precedence-equivalent versions %s and %s are ambiguous",
						release.ID, previous, release.Version,
					),
				)
			}
		}
		versionsByID[release.ID] = append(versionsByID[release.ID], release.Version)
	}
	sort.Strings(issues)
	return issues
}

func isSHA256(value string) bool {
	if len(value) != sha256.Size*2 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func digestSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func readRepoFile(rootDir, relative string, maxBytes int64) ([]byte, error) {
	target, err := secureRepoPath(rootDir, relative)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(target)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("%s exceeds %d bytes", relative, maxBytes)
	}
	return data, nil
}

// secureRepoPath resolves a canonical forward-slash relative path and rejects
// every existing symlink component. This keeps both reads and writes inside
// the explicit repository root.
func secureRepoPath(rootDir, relative string) (string, error) {
	if err := validateRepoRelativePath(relative, path.Ext(relative)); err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	target := filepath.Join(rootAbs, filepath.FromSlash(relative))
	rel, err := filepath.Rel(rootAbs, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes repository root", relative)
	}
	current := rootAbs
	for _, component := range strings.Split(filepath.FromSlash(relative), string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return "", fmt.Errorf("inspect %s: %w", relative, statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("path %q traverses symlink %q", relative, current)
		}
	}
	return target, nil
}

func (c *Command) loadRegistries(ctx context.Context, rootDir, registryID string) ([]loadedRegistry, error) {
	config, err := ReadRegistries(rootDir)
	if err != nil {
		return nil, err
	}
	if config == nil || len(config.Registries) == 0 {
		return nil, fmt.Errorf("%s does not configure any registries", RegistryFileName)
	}
	var selected []Registry
	for _, registry := range config.Registries {
		if registryID == "" || registry.ID == registryID {
			selected = append(selected, registry)
		}
	}
	if registryID != "" && len(selected) == 0 {
		return nil, fmt.Errorf("unknown registry %q", registryID)
	}

	loaded := make([]loadedRegistry, 0, len(selected))
	for _, registry := range selected {
		data, err := c.readRegistryIndex(ctx, rootDir, registry)
		if err != nil {
			return nil, fmt.Errorf("registry %s: %w", registry.ID, err)
		}
		var index RegistryIndex
		if err := decodeStrictJSON(data, &index); err != nil {
			return nil, fmt.Errorf("registry %s: parse index: %w", registry.ID, err)
		}
		if issues := validateRegistryIndex(&index); len(issues) > 0 {
			return nil, fmt.Errorf("registry %s index validation failed:\n- %s", registry.ID, joinIssues(issues))
		}
		sort.Slice(index.Packs, func(i, j int) bool {
			if index.Packs[i].ID != index.Packs[j].ID {
				return index.Packs[i].ID < index.Packs[j].ID
			}
			cmp, _ := compareSemanticVersions(index.Packs[i].Version, index.Packs[j].Version)
			return cmp > 0
		})
		loaded = append(loaded, loadedRegistry{Registry: registry, Index: index})
	}
	return loaded, nil
}

type loadedRegistry struct {
	Registry Registry
	Index    RegistryIndex
}

func (c *Command) readRegistryIndex(ctx context.Context, rootDir string, registry Registry) ([]byte, error) {
	if strings.Contains(registry.Location, "://") {
		return c.fetcher.Fetch(ctx, registry.Location, registryIndexMaxBytes)
	}
	data, err := readRepoFile(rootDir, registry.Location, registryIndexMaxBytes)
	if err != nil {
		return nil, fmt.Errorf("read local index %s: %w", registry.Location, err)
	}
	return data, nil
}

func (c *Command) readRegistryArtifact(ctx context.Context, rootDir string, candidate RegistryCandidate) ([]byte, error) {
	if strings.Contains(candidate.Registry.Location, "://") {
		base, err := url.Parse(candidate.Registry.Location)
		if err != nil {
			return nil, fmt.Errorf("parse registry location: %w", err)
		}
		ref, err := url.Parse(candidate.Release.Artifact)
		if err != nil {
			return nil, fmt.Errorf("parse artifact location: %w", err)
		}
		resolved := base.ResolveReference(ref)
		if resolved.Scheme != "https" || !strings.EqualFold(resolved.Host, base.Host) {
			return nil, errors.New("artifact resolved outside registry HTTPS origin")
		}
		return c.fetcher.Fetch(ctx, resolved.String(), registryPackMaxBytes)
	}
	relative := path.Join(path.Dir(candidate.Registry.Location), candidate.Release.Artifact)
	data, err := readRepoFile(rootDir, relative, registryPackMaxBytes)
	if err != nil {
		return nil, fmt.Errorf("read local artifact %s: %w", relative, err)
	}
	return data, nil
}

func allCandidates(loaded []loadedRegistry) []RegistryCandidate {
	var out []RegistryCandidate
	for _, registry := range loaded {
		for _, release := range registry.Index.Packs {
			out = append(out, RegistryCandidate{Registry: registry.Registry, Release: release})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Release.ID != out[j].Release.ID {
			return out[i].Release.ID < out[j].Release.ID
		}
		cmp, _ := compareSemanticVersions(out[i].Release.Version, out[j].Release.Version)
		if cmp != 0 {
			return cmp > 0
		}
		return out[i].Registry.ID < out[j].Registry.ID
	})
	return out
}

func resolveCandidate(candidates []RegistryCandidate, packID, version string) (RegistryCandidate, error) {
	var matches []RegistryCandidate
	for _, candidate := range candidates {
		if candidate.Release.ID != packID {
			continue
		}
		if version != "" && candidate.Release.Version != version {
			continue
		}
		matches = append(matches, candidate)
	}
	if len(matches) == 0 {
		if version != "" {
			return RegistryCandidate{}, fmt.Errorf("pack %s version %s was not found", packID, version)
		}
		return RegistryCandidate{}, fmt.Errorf("pack %s was not found", packID)
	}
	sort.Slice(matches, func(i, j int) bool {
		cmp, _ := compareSemanticVersions(matches[i].Release.Version, matches[j].Release.Version)
		if cmp != 0 {
			return cmp > 0
		}
		return matches[i].Registry.ID < matches[j].Registry.ID
	})
	winner := matches[0]
	var ambiguous []string
	for _, candidate := range matches[1:] {
		cmp, _ := compareSemanticVersions(candidate.Release.Version, winner.Release.Version)
		if cmp != 0 {
			break
		}
		if candidate.Registry.ID != winner.Registry.ID {
			ambiguous = append(ambiguous, candidate.Registry.ID)
		}
	}
	if len(ambiguous) > 0 {
		registries := append([]string{winner.Registry.ID}, ambiguous...)
		sort.Strings(registries)
		return RegistryCandidate{}, fmt.Errorf(
			"pack %s version %s is ambiguous across registries %s; use --registry",
			packID, winner.Release.Version, strings.Join(registries, ", "),
		)
	}
	return winner, nil
}

type semanticVersion struct {
	major      uint64
	minor      uint64
	patch      uint64
	prerelease []string
}

func parseSemanticVersion(value string) (semanticVersion, error) {
	if !semverRE.MatchString(value) || strings.HasPrefix(value, "v") {
		return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
	}
	versionParts := strings.SplitN(value, "+", 2)
	if len(versionParts) == 2 {
		for _, identifier := range strings.Split(versionParts[1], ".") {
			if identifier == "" {
				return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
			}
		}
	}
	withoutBuild := versionParts[0]
	parts := strings.SplitN(withoutBuild, "-", 2)
	core := strings.Split(parts[0], ".")
	if len(core) != 3 {
		return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
	}
	numbers := make([]uint64, 3)
	for i, part := range core {
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
		}
		numbers[i] = n
	}
	out := semanticVersion{major: numbers[0], minor: numbers[1], patch: numbers[2]}
	if len(parts) == 2 {
		out.prerelease = strings.Split(parts[1], ".")
		for _, identifier := range out.prerelease {
			if identifier == "" {
				return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
			}
			if isNumeric(identifier) && len(identifier) > 1 && identifier[0] == '0' {
				return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
			}
		}
	}
	return out, nil
}

func compareSemanticVersions(a, b string) (int, error) {
	left, err := parseSemanticVersion(a)
	if err != nil {
		return 0, err
	}
	right, err := parseSemanticVersion(b)
	if err != nil {
		return 0, err
	}
	for _, pair := range [][2]uint64{{left.major, right.major}, {left.minor, right.minor}, {left.patch, right.patch}} {
		if pair[0] < pair[1] {
			return -1, nil
		}
		if pair[0] > pair[1] {
			return 1, nil
		}
	}
	if len(left.prerelease) == 0 && len(right.prerelease) == 0 {
		return 0, nil
	}
	if len(left.prerelease) == 0 {
		return 1, nil
	}
	if len(right.prerelease) == 0 {
		return -1, nil
	}
	for i := 0; i < len(left.prerelease) && i < len(right.prerelease); i++ {
		l, r := left.prerelease[i], right.prerelease[i]
		if l == r {
			continue
		}
		ln, rn := isNumeric(l), isNumeric(r)
		switch {
		case ln && rn:
			if len(l) < len(r) || (len(l) == len(r) && l < r) {
				return -1, nil
			}
			return 1, nil
		case ln:
			return -1, nil
		case rn:
			return 1, nil
		case l < r:
			return -1, nil
		default:
			return 1, nil
		}
	}
	if len(left.prerelease) < len(right.prerelease) {
		return -1, nil
	}
	if len(left.prerelease) > len(right.prerelease) {
		return 1, nil
	}
	return 0, nil
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

type plannedPack struct {
	Candidate RegistryCandidate
	Data      []byte
	Metadata  packFileMetadata
}

func (c *Command) planInstall(
	ctx context.Context,
	rootDir string,
	loaded []loadedRegistry,
	rootCandidate RegistryCandidate,
	existing []PackMetadata,
	provenance *InstalledPacks,
	updating map[string]bool,
) ([]plannedPack, error) {
	candidatesByRegistry := map[string][]RegistryCandidate{}
	for _, candidate := range allCandidates(loaded) {
		candidatesByRegistry[candidate.Registry.ID] = append(candidatesByRegistry[candidate.Registry.ID], candidate)
	}
	existingByID := map[string]PackMetadata{}
	for _, pack := range existing {
		existingByID[pack.ID] = pack
	}
	provenanceByID := map[string]InstalledPack{}
	for _, entry := range provenance.Packs {
		provenanceByID[entry.ID] = entry
	}

	var plan []plannedPack
	planned := map[string]RegistryCandidate{}
	visiting := map[string]bool{}
	var visit func(RegistryCandidate, bool) error
	visit = func(candidate RegistryCandidate, root bool) error {
		id := candidate.Release.ID
		if previous, ok := planned[id]; ok {
			if previous.Release.Version != candidate.Release.Version ||
				previous.Registry.ID != candidate.Registry.ID {
				return fmt.Errorf(
					"conflicting planned releases for %s: %s@%s and %s@%s",
					id,
					previous.Registry.ID,
					previous.Release.Version,
					candidate.Registry.ID,
					candidate.Release.Version,
				)
			}
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("registry dependency cycle includes %s", id)
		}
		if !root {
			if available, ok := existingByID[id]; ok && available.Version != "" {
				return nil
			}
		}

		currentProvenance, owned := provenanceByID[id]
		if available, ok := existingByID[id]; ok && root {
			if available.State.Bundled && !owned {
				return fmt.Errorf("pack %s is bundled and cannot be replaced by a registry install", id)
			}
			if available.State.Installed && !owned {
				return fmt.Errorf("pack %s already exists as an unowned repository-local pack", id)
			}
		}
		if owned && !updating[id] {
			return fmt.Errorf("pack %s is already registry-managed; use carl pack update %s", id, id)
		}
		if updating[id] {
			if !owned {
				return fmt.Errorf("pack %s is not registry-managed", id)
			}
			if currentProvenance.Registry != candidate.Registry.ID ||
				currentProvenance.RegistryLocation != candidate.Registry.Location {
				return fmt.Errorf("pack %s update source does not match recorded provenance", id)
			}
		}

		visiting[id] = true
		data, err := c.readRegistryArtifact(ctx, rootDir, candidate)
		if err != nil {
			return fmt.Errorf("fetch %s@%s: %w", id, candidate.Release.Version, err)
		}
		actualDigest := digestSHA256(data)
		if actualDigest != candidate.Release.SHA256 {
			return fmt.Errorf(
				"verify %s@%s: SHA-256 mismatch (expected %s, got %s)",
				id, candidate.Release.Version, candidate.Release.SHA256, actualDigest,
			)
		}
		metadata, err := parsePackFileMetadata(expectedPackPath(id), data)
		if err != nil {
			return fmt.Errorf("validate %s@%s: %w", id, candidate.Release.Version, err)
		}
		if metadata.Version != candidate.Release.Version {
			return fmt.Errorf(
				"validate %s@%s: pack declares version %q",
				id, candidate.Release.Version, metadata.Version,
			)
		}
		if containsControl(metadata.Title) || containsControl(metadata.Description) {
			return fmt.Errorf(
				"validate %s@%s: title or description contains control characters",
				id, candidate.Release.Version,
			)
		}
		for _, dependencyID := range metadata.Requires {
			if available, ok := existingByID[dependencyID]; ok && available.Version != "" {
				continue
			}
			dependency, err := resolveCandidate(candidatesByRegistry[candidate.Registry.ID], dependencyID, "")
			if err != nil {
				return fmt.Errorf("%s@%s dependency: %w", id, candidate.Release.Version, err)
			}
			if err := visit(dependency, false); err != nil {
				return err
			}
		}
		visiting[id] = false
		planned[id] = candidate
		plan = append(plan, plannedPack{Candidate: candidate, Data: data, Metadata: metadata})
		return nil
	}
	if err := visit(rootCandidate, true); err != nil {
		return nil, err
	}
	sort.Slice(plan, func(i, j int) bool {
		return plan[i].Candidate.Release.ID < plan[j].Candidate.Release.ID
	})
	return plan, nil
}

func validatePlannedSet(existing []PackMetadata, plan []plannedPack) error {
	index := map[string]PackMetadata{}
	for _, pack := range existing {
		index[pack.ID] = pack
	}
	for _, item := range plan {
		id := item.Candidate.Release.ID
		var dependencies []PackDependency
		for _, dependency := range item.Metadata.Requires {
			dependencies = append(dependencies, PackDependency{ID: dependency, Required: true})
		}
		var precedence *Precedence
		if item.Metadata.Mode != "" || item.Metadata.Priority != nil || len(item.Metadata.Overrides) > 0 {
			precedence = &Precedence{
				Mode:      item.Metadata.Mode,
				Priority:  item.Metadata.Priority,
				Overrides: item.Metadata.Overrides,
			}
		}
		previous := index[id]
		index[id] = PackMetadata{
			SchemaVersion:  metadataSchemaVersion,
			ID:             id,
			Version:        item.Candidate.Release.Version,
			Title:          item.Metadata.Title,
			Description:    item.Metadata.Description,
			Category:       categoryFromID(id),
			Source:         "registry:" + item.Candidate.Registry.ID,
			State:          previous.State,
			OwnedArtifacts: []string{expectedPackPath(id)},
			Dependencies:   dependencies,
			Precedence:     precedence,
		}
	}
	packs := make([]PackMetadata, 0, len(index))
	for _, pack := range index {
		packs = append(packs, pack)
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].ID < packs[j].ID })
	if issues := validatePackSet(packs); len(issues) > 0 {
		return fmt.Errorf("planned pack set validation failed:\n- %s", joinIssues(issues))
	}
	return nil
}

func applyInstallPlan(rootDir string, provenance *InstalledPacks, plan []plannedPack) error {
	entries := map[string]InstalledPack{}
	for _, entry := range provenance.Packs {
		entries[entry.ID] = entry
	}
	var writes []transactionWrite
	for _, item := range plan {
		release := item.Candidate.Release
		target := expectedPackPath(release.ID)
		writes = append(writes, transactionWrite{Relative: target, Data: item.Data})
		entries[release.ID] = InstalledPack{
			ID:               release.ID,
			Version:          release.Version,
			Registry:         item.Candidate.Registry.ID,
			RegistryLocation: item.Candidate.Registry.Location,
			Artifact:         release.Artifact,
			SHA256:           release.SHA256,
			InstalledPath:    target,
		}
	}
	updated := &InstalledPacks{SchemaVersion: metadataSchemaVersion}
	for _, entry := range entries {
		updated.Packs = append(updated.Packs, entry)
	}
	manifestData, err := marshalInstalledPacks(updated)
	if err != nil {
		return err
	}
	writes = append(writes, transactionWrite{Relative: InstalledPacksFileName, Data: manifestData})
	return writeTransaction(rootDir, writes)
}

type transactionWrite struct {
	Relative string
	Data     []byte
}

type transactionBackup struct {
	Path    string
	Existed bool
	Data    []byte
	Mode    os.FileMode
}

func writeTransaction(rootDir string, writes []transactionWrite) error {
	backups := make([]transactionBackup, 0, len(writes))
	targets := make([]string, len(writes))
	for i, write := range writes {
		target, err := secureRepoPath(rootDir, write.Relative)
		if err != nil {
			return fmt.Errorf("validate write target %s: %w", write.Relative, err)
		}
		targets[i] = target
		backup := transactionBackup{Path: target, Mode: 0644}
		info, err := os.Stat(target)
		switch {
		case err == nil:
			if !info.Mode().IsRegular() {
				return fmt.Errorf("write target %s is not a regular file", write.Relative)
			}
			backup.Existed = true
			backup.Mode = info.Mode().Perm()
			backup.Data, err = os.ReadFile(target)
			if err != nil {
				return fmt.Errorf("backup %s: %w", write.Relative, err)
			}
		case os.IsNotExist(err):
		default:
			return fmt.Errorf("inspect %s: %w", write.Relative, err)
		}
		backups = append(backups, backup)
	}

	for i, write := range writes {
		if err := os.MkdirAll(filepath.Dir(targets[i]), 0755); err != nil {
			rollbackErr := rollbackTransaction(backups[:i])
			return transactionError(fmt.Errorf("create directory for %s: %w", write.Relative, err), rollbackErr)
		}
		if err := os.WriteFile(targets[i], write.Data, backups[i].Mode); err != nil {
			rollbackErr := rollbackTransaction(backups[:i+1])
			return transactionError(fmt.Errorf("write %s: %w", write.Relative, err), rollbackErr)
		}
	}
	return nil
}

func rollbackTransaction(backups []transactionBackup) error {
	var rollbackErrors []error
	for i := len(backups) - 1; i >= 0; i-- {
		backup := backups[i]
		if backup.Existed {
			if err := os.WriteFile(backup.Path, backup.Data, backup.Mode); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restore %s: %w", backup.Path, err))
			}
		} else {
			if err := os.Remove(backup.Path); err != nil && !os.IsNotExist(err) {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("remove %s: %w", backup.Path, err))
			}
		}
	}
	return errors.Join(rollbackErrors...)
}

func transactionError(operationErr, rollbackErr error) error {
	if rollbackErr == nil {
		return operationErr
	}
	return fmt.Errorf("%w; rollback also failed: %v", operationErr, rollbackErr)
}

func installedPackByID(installed *InstalledPacks, id string) (InstalledPack, bool) {
	for _, entry := range installed.Packs {
		if entry.ID == id {
			return entry, true
		}
	}
	return InstalledPack{}, false
}

func validateInstalledDigest(rootDir string, entry InstalledPack) error {
	data, err := readRepoFile(rootDir, entry.InstalledPath, registryPackMaxBytes)
	if err != nil {
		return fmt.Errorf("read installed pack %s: %w", entry.ID, err)
	}
	actual := digestSHA256(data)
	if actual != entry.SHA256 {
		return fmt.Errorf(
			"installed pack %s has drifted (expected SHA-256 %s, got %s)",
			entry.ID, entry.SHA256, actual,
		)
	}
	return nil
}
