package pack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/cmdutil"
)

type fakeRegistryFetcher struct {
	responses map[string][]byte
	errs      map[string]error
	calls     []string
}

func (f *fakeRegistryFetcher) Fetch(_ context.Context, location string, _ int64) ([]byte, error) {
	f.calls = append(f.calls, location)
	if err := f.errs[location]; err != nil {
		return nil, err
	}
	data, ok := f.responses[location]
	if !ok {
		return nil, fmt.Errorf("unexpected fetch %s", location)
	}
	return append([]byte(nil), data...), nil
}

func registryPack(version, title string, requires ...string) []byte {
	var lines []string
	lines = append(lines, "<!-- version: "+version+" -->")
	if len(requires) > 0 {
		lines = append(lines, "<!-- requires: "+strings.Join(requires, ", ")+" -->")
	}
	lines = append(lines, "# "+title, "", title+" guidance.", "")
	return []byte(strings.Join(lines, "\n"))
}

func writeJSONFile(t *testing.T, root, relative string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, relative, append(data, '\n'))
}

func configureLocalRegistry(t *testing.T, root, id, indexLocation string, releases []RegistryRelease) {
	t.Helper()
	writeJSONFile(t, root, RegistryFileName, Registries{
		SchemaVersion: metadataSchemaVersion,
		Registries: []Registry{{
			ID:       id,
			Location: indexLocation,
		}},
	})
	writeJSONFile(t, root, indexLocation, RegistryIndex{
		SchemaVersion: metadataSchemaVersion,
		Packs:         releases,
	})
}

func localRelease(t *testing.T, root, indexLocation, id, version string, data []byte) RegistryRelease {
	t.Helper()
	artifact := "packs/" + strings.ReplaceAll(id, "/", "-") + "-" + version + ".instructions.md"
	writeFile(t, root, pathJoin(pathDir(indexLocation), artifact), data)
	return RegistryRelease{
		ID:       id,
		Version:  version,
		Artifact: artifact,
		SHA256:   digestSHA256(data),
		Title:    id,
	}
}

func pathDir(value string) string {
	i := strings.LastIndex(value, "/")
	if i < 0 {
		return "."
	}
	return value[:i]
}

func pathJoin(base, relative string) string {
	if base == "." || base == "" {
		return relative
	}
	return base + "/" + relative
}

func runCapturingStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	var runErr error
	out := captureStdout(t, func() {
		runErr = fn()
	})
	return out, runErr
}

func TestSemanticVersionComparison(t *testing.T) {
	ordered := []string{
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
		"1.0.0",
		"1.0.1",
		"1.1.0",
		"2.0.0",
	}
	for i := 0; i < len(ordered)-1; i++ {
		cmp, err := compareSemanticVersions(ordered[i], ordered[i+1])
		if err != nil {
			t.Fatalf("compare %s and %s: %v", ordered[i], ordered[i+1], err)
		}
		if cmp >= 0 {
			t.Fatalf("compare %s and %s = %d, want negative", ordered[i], ordered[i+1], cmp)
		}
	}
	for _, invalid := range []string{"v1.0.0", "1.0", "01.0.0", "1.0.0-01"} {
		if _, err := parseSemanticVersion(invalid); err == nil {
			t.Errorf("parseSemanticVersion(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestRegistryConfigurationStrictValidation(t *testing.T) {
	tests := []struct {
		name     string
		location string
	}{
		{name: "plaintext HTTP", location: "http://example.com/index.json"},
		{name: "credentials", location: "https://user:pass@example.com/index.json"},
		{name: "query", location: "https://example.com/index.json?token=secret"},
		{name: "fragment", location: "https://example.com/index.json#fragment"},
		{name: "local traversal", location: "../index.json"},
		{name: "backslashes", location: `.registry\index.json`},
		{name: "local colon", location: "https:index.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidateRegistries(&Registries{
				SchemaVersion: metadataSchemaVersion,
				Registries:    []Registry{{ID: "test", Location: tt.location}},
			})
			if len(issues) == 0 {
				t.Fatalf("ValidateRegistries(%q) returned no issues", tt.location)
			}
		})
	}

	dir := t.TempDir()
	writeFile(t, dir, RegistryFileName, []byte(`{"schemaVersion":1,"registries":[],"extra":true}`))
	if _, err := ReadRegistries(dir); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("ReadRegistries strict JSON error = %v", err)
	}
}

func TestRegistryIndexRejectsPrecedenceEquivalentVersions(t *testing.T) {
	index := &RegistryIndex{
		SchemaVersion: metadataSchemaVersion,
		Packs: []RegistryRelease{
			{
				ID:       "languages/rust",
				Version:  "1.0.0+alpha",
				Artifact: "rust-alpha.instructions.md",
				SHA256:   strings.Repeat("a", 64),
			},
			{
				ID:       "languages/rust",
				Version:  "1.0.0+beta",
				Artifact: "rust-beta.instructions.md",
				SHA256:   strings.Repeat("b", 64),
			},
		},
	}
	issues := validateRegistryIndex(index)
	if len(issues) == 0 || !strings.Contains(strings.Join(issues, "\n"), "precedence-equivalent") {
		t.Fatalf("validateRegistryIndex issues = %v", issues)
	}
}

func TestRegistryListAndExistingPackCommandsDoNotFetch(t *testing.T) {
	dir := t.TempDir()
	writeJSONFile(t, dir, RegistryFileName, Registries{
		SchemaVersion: metadataSchemaVersion,
		Registries: []Registry{{
			ID:       "remote",
			Location: "https://registry.example/index.json",
		}},
	})
	fetcher := &fakeRegistryFetcher{}
	cmd := New(bundledArtifacts())
	cmd.fetcher = fetcher

	if _, err := runCapturingStdout(t, func() error {
		return cmd.RunRegistryListInDir(dir, false)
	}); err != nil {
		t.Fatalf("RunRegistryListInDir: %v", err)
	}
	if _, err := runCapturingStdout(t, func() error {
		return cmd.RunListInDir(dir, false)
	}); err != nil {
		t.Fatalf("RunListInDir: %v", err)
	}
	if len(fetcher.calls) != 0 {
		t.Fatalf("read-only local commands fetched %v", fetcher.calls)
	}
}

func TestRegistrySearchRemoteDeterministic(t *testing.T) {
	dir := t.TempDir()
	location := "https://registry.example/index.json"
	writeJSONFile(t, dir, RegistryFileName, Registries{
		SchemaVersion: metadataSchemaVersion,
		Registries:    []Registry{{ID: "remote", Location: location}},
	})
	index, err := json.Marshal(RegistryIndex{
		SchemaVersion: metadataSchemaVersion,
		Packs: []RegistryRelease{
			{
				ID:       "languages/rust",
				Version:  "1.0.0",
				Artifact: "packs/rust-1.0.0.instructions.md",
				SHA256:   strings.Repeat("a", 64),
				Title:    "Rust",
			},
			{
				ID:       "languages/rust",
				Version:  "2.0.0",
				Artifact: "packs/rust-2.0.0.instructions.md",
				SHA256:   strings.Repeat("b", 64),
				Title:    "Rust",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	fetcher := &fakeRegistryFetcher{responses: map[string][]byte{location: index}}
	cmd := New(nil)
	cmd.fetcher = fetcher

	out, err := runCapturingStdout(t, func() error {
		return cmd.RunRegistrySearchInDir(context.Background(), dir, "rust", "", true)
	})
	if err != nil {
		t.Fatalf("RunRegistrySearchInDir: %v", err)
	}
	var payload registrySearchPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Packs) != 2 ||
		payload.Packs[0].Version != "2.0.0" ||
		payload.Packs[1].Version != "1.0.0" {
		t.Fatalf("search order = %#v", payload.Packs)
	}
	if len(fetcher.calls) != 1 || fetcher.calls[0] != location {
		t.Fatalf("fetch calls = %v", fetcher.calls)
	}
}

func TestHTTPRegistryFetcherBoundaries(t *testing.T) {
	t.Run("same-origin redirect allowed", func(t *testing.T) {
		var server *httptest.Server
		server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/index.json" {
				http.Redirect(w, r, server.URL+"/final.json", http.StatusFound)
				return
			}
			_, _ = w.Write([]byte(`{"schemaVersion":1,"packs":[]}`))
		}))
		defer server.Close()
		fetcher := newHTTPRegistryFetcher().(*httpRegistryFetcher)
		fetcher.client.Transport = server.Client().Transport
		data, err := fetcher.Fetch(context.Background(), server.URL+"/index.json", 1024)
		if err != nil {
			t.Fatalf("Fetch same-origin redirect: %v", err)
		}
		if !strings.Contains(string(data), `"schemaVersion":1`) {
			t.Fatalf("Fetch data = %q", data)
		}
	})

	t.Run("cross-origin redirect rejected", func(t *testing.T) {
		target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("unexpected"))
		}))
		defer target.Close()
		source := httptest.NewTLSServer(http.RedirectHandler(target.URL, http.StatusFound))
		defer source.Close()
		fetcher := newHTTPRegistryFetcher().(*httpRegistryFetcher)
		fetcher.client.Transport = source.Client().Transport
		_, err := fetcher.Fetch(context.Background(), source.URL, 1024)
		if err == nil || !strings.Contains(err.Error(), "cross-origin redirect rejected") {
			t.Fatalf("Fetch cross-origin redirect error = %v", err)
		}
	})

	t.Run("response size bounded", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("12345"))
		}))
		defer server.Close()
		fetcher := newHTTPRegistryFetcher().(*httpRegistryFetcher)
		fetcher.client.Transport = server.Client().Transport
		_, err := fetcher.Fetch(context.Background(), server.URL, 4)
		if err == nil || !strings.Contains(err.Error(), "exceeds 4 bytes") {
			t.Fatalf("Fetch oversized response error = %v", err)
		}
	})
}

func TestResolveCandidateRejectsCrossRegistryAmbiguity(t *testing.T) {
	release := RegistryRelease{
		ID:       "languages/rust",
		Version:  "1.0.0",
		Artifact: "rust.instructions.md",
		SHA256:   strings.Repeat("a", 64),
	}
	_, err := resolveCandidate([]RegistryCandidate{
		{Registry: Registry{ID: "alpha"}, Release: release},
		{Registry: Registry{ID: "beta"}, Release: release},
	}, release.ID, "")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("resolveCandidate error = %v", err)
	}
}

func TestPackInstallLocalRegistryHighestVersionAndProvenance(t *testing.T) {
	dir := t.TempDir()
	indexLocation := ".registry/index.json"
	v1 := registryPack("1.0.0", "Rust")
	v2 := registryPack("2.0.0", "Rust")
	releases := []RegistryRelease{
		localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", v1),
		localRelease(t, dir, indexLocation, "languages/rust", "2.0.0", v2),
	}
	configureLocalRegistry(t, dir, "local", indexLocation, releases)
	cmd := New(nil)

	out, err := runCapturingStdout(t, func() error {
		return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", true)
	})
	if err != nil {
		t.Fatalf("RunInstallInDir: %v", err)
	}
	var payload registryChangePayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Packs) != 1 || payload.Packs[0].Version != "2.0.0" {
		t.Fatalf("install payload = %#v", payload)
	}
	installedData, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(expectedPackPath("languages/rust"))))
	if err != nil {
		t.Fatal(err)
	}
	if string(installedData) != string(v2) {
		t.Fatalf("installed data = %q, want v2", installedData)
	}
	provenance, err := ReadInstalledPacks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(provenance.Packs) != 1 ||
		provenance.Packs[0].Registry != "local" ||
		provenance.Packs[0].SHA256 != digestSHA256(v2) {
		t.Fatalf("provenance = %#v", provenance.Packs)
	}
	packs, err := cmd.discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 1 ||
		packs[0].Source != "registry:local" ||
		packs[0].Provenance == nil ||
		packs[0].Provenance.SHA256 != digestSHA256(v2) {
		t.Fatalf("discovered registry provenance = %#v", packs)
	}
	for _, forbidden := range []string{SelectionFileName, ProfileFileName, ".github/carl/runtime.json"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Fatalf("%s was unexpectedly written (err=%v)", forbidden, err)
		}
	}
}

func TestPackDiscoveryRejectsRegistryManagedDrift(t *testing.T) {
	dir := t.TempDir()
	indexLocation := ".registry/index.json"
	data := registryPack("1.0.0", "Rust")
	configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
		localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", data),
	})
	cmd := New(nil)
	if _, err := runCapturingStdout(t, func() error {
		return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", false)
	}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, expectedPackPath("languages/rust"), registryPack("1.0.0", "Locally Edited"))
	if _, err := cmd.discover(dir); err == nil || !strings.Contains(err.Error(), "has drifted") {
		t.Fatalf("discover drift error = %v", err)
	}
}

func TestPackInstallExactVersion(t *testing.T) {
	dir := t.TempDir()
	indexLocation := ".registry/index.json"
	v1 := registryPack("1.0.0", "Rust")
	v2 := registryPack("2.0.0", "Rust")
	configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
		localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", v1),
		localRelease(t, dir, indexLocation, "languages/rust", "2.0.0", v2),
	})
	cmd := New(nil)

	if _, err := runCapturingStdout(t, func() error {
		return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "1.0.0", "", false)
	}); err != nil {
		t.Fatalf("RunInstallInDir: %v", err)
	}
	provenance, err := ReadInstalledPacks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := provenance.Packs[0].Version; got != "1.0.0" {
		t.Fatalf("installed version = %s, want 1.0.0", got)
	}
}

func TestPackInstallResolvesDependenciesBeforeWriting(t *testing.T) {
	dir := t.TempDir()
	indexLocation := ".registry/index.json"
	dependency := registryPack("1.0.0", "Dependency")
	rootPack := registryPack("1.0.0", "Root", "core/dependency")
	configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
		localRelease(t, dir, indexLocation, "languages/root", "1.0.0", rootPack),
		localRelease(t, dir, indexLocation, "core/dependency", "1.0.0", dependency),
	})
	cmd := New(nil)

	out, err := runCapturingStdout(t, func() error {
		return cmd.RunInstallInDir(context.Background(), dir, "languages/root", "", "", true)
	})
	if err != nil {
		t.Fatalf("RunInstallInDir: %v", err)
	}
	var payload registryChangePayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Packs) != 2 ||
		payload.Packs[0].ID != "core/dependency" ||
		payload.Packs[1].ID != "languages/root" {
		t.Fatalf("installed packs = %#v", payload.Packs)
	}
	provenance, err := ReadInstalledPacks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(provenance.Packs) != 2 ||
		provenance.Packs[0].ID != "core/dependency" ||
		provenance.Packs[1].ID != "languages/root" {
		t.Fatalf("provenance order = %#v", provenance.Packs)
	}
}

func TestPackInstallValidationFailuresWriteNothing(t *testing.T) {
	tests := []struct {
		name          string
		releaseDigest func([]byte) string
		indexVersion  string
		packVersion   string
		requires      []string
		want          string
	}{
		{
			name:          "digest mismatch",
			releaseDigest: func([]byte) string { return strings.Repeat("0", 64) },
			indexVersion:  "1.0.0",
			packVersion:   "1.0.0",
			want:          "SHA-256 mismatch",
		},
		{
			name:          "declared version mismatch",
			releaseDigest: digestSHA256,
			indexVersion:  "2.0.0",
			packVersion:   "1.0.0",
			want:          "declares version",
		},
		{
			name:          "missing dependency",
			releaseDigest: digestSHA256,
			indexVersion:  "1.0.0",
			packVersion:   "1.0.0",
			requires:      []string{"core/missing"},
			want:          "dependency",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			indexLocation := ".registry/index.json"
			data := registryPack(tt.packVersion, "Rust", tt.requires...)
			release := localRelease(t, dir, indexLocation, "languages/rust", tt.indexVersion, data)
			release.SHA256 = tt.releaseDigest(data)
			configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{release})
			cmd := New(nil)

			_, err := runCapturingStdout(t, func() error {
				return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", false)
			})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RunInstallInDir error = %v, want %q", err, tt.want)
			}
			for _, relative := range []string{
				expectedPackPath("languages/rust"),
				InstalledPacksFileName,
			} {
				if _, statErr := os.Stat(filepath.Join(dir, filepath.FromSlash(relative))); !os.IsNotExist(statErr) {
					t.Fatalf("%s was written after failed validation (err=%v)", relative, statErr)
				}
			}
		})
	}
}

func TestPackInstallRejectsUnownedAndBundledPacks(t *testing.T) {
	tests := []struct {
		name string
		arts Artifacts
	}{
		{name: "unowned local", arts: nil},
		{name: "bundled", arts: bundledArtifacts()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			indexLocation := ".registry/index.json"
			id := "languages/rust"
			if tt.name == "bundled" {
				id = "core/carl"
			} else {
				writeFile(t, dir, expectedPackPath(id), registryPack("0.9.0", "Local Rust"))
			}
			data := registryPack("1.0.0", "Registry Pack")
			configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
				localRelease(t, dir, indexLocation, id, "1.0.0", data),
			})
			cmd := New(tt.arts)
			_, err := runCapturingStdout(t, func() error {
				return cmd.RunInstallInDir(context.Background(), dir, id, "", "", false)
			})
			if err == nil {
				t.Fatal("RunInstallInDir unexpectedly succeeded")
			}
			if !strings.Contains(err.Error(), "cannot be replaced") &&
				!strings.Contains(err.Error(), "unowned") {
				t.Fatalf("RunInstallInDir error = %v", err)
			}
		})
	}
}

func TestPackInstallJSONErrorIsStructured(t *testing.T) {
	dir := t.TempDir()
	cmd := New(nil)
	err := cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", true)
	var exitErr *cmdutil.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error type = %T, want *cmdutil.ExitError", err)
	}
	var payload errorPayload
	if jsonErr := json.Unmarshal([]byte(exitErr.Message), &payload); jsonErr != nil {
		t.Fatalf("JSON error payload: %v", jsonErr)
	}
	if payload.Error.Code != "registry_unavailable" {
		t.Fatalf("error code = %q", payload.Error.Code)
	}
}

func TestPackUpdateUpgradeNoopDriftAndMutation(t *testing.T) {
	t.Run("upgrade then no-op", func(t *testing.T) {
		dir := t.TempDir()
		indexLocation := ".registry/index.json"
		v1 := registryPack("1.0.0", "Rust")
		releaseV1 := localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", v1)
		configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{releaseV1})
		cmd := New(nil)
		if _, err := runCapturingStdout(t, func() error {
			return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", false)
		}); err != nil {
			t.Fatal(err)
		}

		v2 := registryPack("2.0.0", "Rust")
		releaseV2 := localRelease(t, dir, indexLocation, "languages/rust", "2.0.0", v2)
		writeJSONFile(t, dir, indexLocation, RegistryIndex{
			SchemaVersion: metadataSchemaVersion,
			Packs:         []RegistryRelease{releaseV1, releaseV2},
		})
		out, err := runCapturingStdout(t, func() error {
			return cmd.RunUpdateInDir(context.Background(), dir, nil, true)
		})
		if err != nil {
			t.Fatalf("RunUpdateInDir upgrade: %v", err)
		}
		var payload registryChangePayload
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Packs) != 1 ||
			payload.Packs[0].Action != "updated" ||
			payload.Packs[0].FromVersion != "1.0.0" ||
			payload.Packs[0].Version != "2.0.0" {
			t.Fatalf("update payload = %#v", payload.Packs)
		}
		provenance, err := ReadInstalledPacks(dir)
		if err != nil {
			t.Fatal(err)
		}
		if provenance.Packs[0].Version != "2.0.0" {
			t.Fatalf("updated provenance = %#v", provenance.Packs[0])
		}

		out, err = runCapturingStdout(t, func() error {
			return cmd.RunUpdateInDir(context.Background(), dir, nil, true)
		})
		if err != nil {
			t.Fatalf("RunUpdateInDir no-op: %v", err)
		}
		payload = registryChangePayload{}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Packs) != 1 || payload.Packs[0].Action != "unchanged" {
			t.Fatalf("no-op payload = %#v", payload.Packs)
		}
	})

	t.Run("drift rejected before fetch", func(t *testing.T) {
		dir := t.TempDir()
		indexLocation := ".registry/index.json"
		v1 := registryPack("1.0.0", "Rust")
		configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
			localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", v1),
		})
		cmd := New(nil)
		if _, err := runCapturingStdout(t, func() error {
			return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", false)
		}); err != nil {
			t.Fatal(err)
		}
		writeFile(t, dir, expectedPackPath("languages/rust"), []byte("locally edited"))
		_, err := runCapturingStdout(t, func() error {
			return cmd.RunUpdateInDir(context.Background(), dir, nil, false)
		})
		if err == nil || !strings.Contains(err.Error(), "drifted") {
			t.Fatalf("RunUpdateInDir drift error = %v", err)
		}
	})

	t.Run("same-version digest mutation rejected", func(t *testing.T) {
		dir := t.TempDir()
		indexLocation := ".registry/index.json"
		v1 := registryPack("1.0.0", "Rust")
		configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
			localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", v1),
		})
		cmd := New(nil)
		if _, err := runCapturingStdout(t, func() error {
			return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", false)
		}); err != nil {
			t.Fatal(err)
		}
		mutated := registryPack("1.0.0", "Mutated Rust")
		writeJSONFile(t, dir, indexLocation, RegistryIndex{
			SchemaVersion: metadataSchemaVersion,
			Packs: []RegistryRelease{
				localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", mutated),
			},
		})
		_, err := runCapturingStdout(t, func() error {
			return cmd.RunUpdateInDir(context.Background(), dir, nil, false)
		})
		if err == nil || !strings.Contains(err.Error(), "changed the SHA-256") {
			t.Fatalf("RunUpdateInDir mutation error = %v", err)
		}
	})
}

func TestPackInstallRejectsSymlinkTargetPath(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	instructions := filepath.Join(dir, ".github", "instructions")
	if err := os.MkdirAll(filepath.Dir(instructions), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, instructions); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	indexLocation := ".registry/index.json"
	data := registryPack("1.0.0", "Rust")
	configureLocalRegistry(t, dir, "local", indexLocation, []RegistryRelease{
		localRelease(t, dir, indexLocation, "languages/rust", "1.0.0", data),
	})
	cmd := New(nil)
	_, err := runCapturingStdout(t, func() error {
		return cmd.RunInstallInDir(context.Background(), dir, "languages/rust", "", "", false)
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "symlink") {
		t.Fatalf("RunInstallInDir symlink error = %v", err)
	}
	if entries, readErr := os.ReadDir(outside); readErr != nil || len(entries) != 0 {
		t.Fatalf("outside directory changed: entries=%v err=%v", entries, readErr)
	}
}
