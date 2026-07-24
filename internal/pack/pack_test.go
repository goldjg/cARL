package pack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goldjg/carl/internal/cmdutil"
	"github.com/goldjg/carl/internal/manifest"
)

type fakeArtifacts struct {
	files map[string][]byte
}

func (f *fakeArtifacts) Open(targetPath string) ([]byte, error) {
	data, ok := f.files[targetPath]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeArtifacts) List() ([]string, error) {
	out := make([]string, 0, len(f.files))
	for p := range f.files {
		out = append(out, p)
	}
	return out, nil
}

func bundledArtifacts() *fakeArtifacts {
	return &fakeArtifacts{
		files: map[string][]byte{
			".github/instructions/core/baseline.instructions.md": []byte("<!-- version: 1.2.0 -->\n# Baseline\nBaseline guidance.\n"),
			".github/instructions/core/carl.instructions.md":     []byte("<!-- version: 1.3.0 -->\n# cARL\nCognition governance.\n"),
			".github/instructions/cloud/azure.instructions.md":   []byte("<!-- version: 1.0.1 -->\n# Azure\nAzure guidance.\n"),
		},
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func writeFile(t *testing.T, root, rel string, content []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, content, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPackListHumanDeterministicOrder(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())

	out := captureStdout(t, func() {
		if err := cmd.RunListInDir(dir, false); err != nil {
			t.Fatalf("RunListInDir: %v", err)
		}
	})

	order := []string{"cloud/azure", "core/baseline", "core/carl"}
	prev := -1
	for _, id := range order {
		i := strings.Index(out, id)
		if i < 0 {
			t.Fatalf("missing %q in output:\n%s", id, out)
		}
		if i <= prev {
			t.Fatalf("nondeterministic order in output:\n%s", out)
		}
		prev = i
	}
}

func TestPackListJSONAndShowJSON(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())

	listOut := captureStdout(t, func() {
		if err := cmd.RunListInDir(dir, true); err != nil {
			t.Fatalf("RunListInDir --json: %v", err)
		}
	})

	var listDoc listPayload
	if err := json.Unmarshal([]byte(listOut), &listDoc); err != nil {
		t.Fatalf("unmarshal list json: %v\n%s", err, listOut)
	}
	if listDoc.SchemaVersion != metadataSchemaVersion {
		t.Fatalf("unexpected schema version: %d", listDoc.SchemaVersion)
	}
	if len(listDoc.Packs) == 0 {
		t.Fatal("expected non-empty pack list")
	}

	showOut := captureStdout(t, func() {
		if err := cmd.RunShowInDir(dir, "core/carl", true); err != nil {
			t.Fatalf("RunShowInDir --json: %v", err)
		}
	})
	var showDoc showPayload
	if err := json.Unmarshal([]byte(showOut), &showDoc); err != nil {
		t.Fatalf("unmarshal show json: %v\n%s", err, showOut)
	}
	if showDoc.Pack.ID != "core/carl" {
		t.Fatalf("unexpected pack ID: %s", showDoc.Pack.ID)
	}
}

func TestPackShowUnknownJSONError(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())

	err := cmd.RunShowInDir(dir, "core/unknown", true)
	if err == nil {
		t.Fatal("expected error")
	}
	exitErr := &cmdutil.ExitError{}
	if !strings.Contains(err.Error(), "pack_not_found") {
		t.Fatalf("expected structured json error, got: %v", err)
	}
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.Code != 1 {
		t.Fatalf("unexpected exit code: %d", exitErr.Code)
	}
	var payload errorPayload
	if unmarshalErr := json.Unmarshal([]byte(exitErr.Message), &payload); unmarshalErr != nil {
		t.Fatalf("error payload is not valid JSON: %v", unmarshalErr)
	}
	if payload.Error.Code != "pack_not_found" {
		t.Fatalf("unexpected payload code: %s", payload.Error.Code)
	}
}

func TestPackDiscoveryInsideInitializedRepoSetsSelected(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())
	writeFile(t, dir, ".github/instructions/core/carl.instructions.md", []byte("<!-- version: 1.3.0 -->\n# cARL\nRepo copy.\n"))
	rt := &manifest.Runtime{
		RuntimeVersion: "1.0.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.0.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now().UTC(),
		ManagedArtifacts: []string{
			".github/instructions/core/carl.instructions.md",
		},
	}
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatalf("manifest write: %v", err)
	}

	packs, err := cmd.discover(dir)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	var found *PackMetadata
	for i := range packs {
		if packs[i].ID == "core/carl" {
			found = &packs[i]
			break
		}
	}
	if found == nil {
		t.Fatal("expected core/carl pack")
	}
	if !found.State.Selected || !found.State.Active {
		t.Fatalf("expected selected+active, got %+v", found.State)
	}
}

func TestPackValidationFailures(t *testing.T) {
	base := PackMetadata{
		SchemaVersion: metadataSchemaVersion,
		ID:            "core/carl",
		Version:       "1.0.0",
		Category:      "core",
		Source:        "bundled",
		State:         PackState{Bundled: true},
		OwnedArtifacts: []string{
			".github/instructions/core/carl.instructions.md",
		},
	}

	packs := []PackMetadata{
		base,
		{
			SchemaVersion: metadataSchemaVersion,
			ID:            "core/dep",
			Version:       "1.0.0",
			Category:      "core",
			Source:        "bundled",
			Dependencies:  []PackDependency{{ID: "core/missing", Required: true}},
			OwnedArtifacts: []string{
				".github/instructions/core/dep.instructions.md",
			},
		},
		{
			SchemaVersion: metadataSchemaVersion,
			ID:            "core/one",
			Version:       "1.0.0",
			Category:      "core",
			Source:        "bundled",
			Dependencies:  []PackDependency{{ID: "core/two", Required: true}},
			OwnedArtifacts: []string{
				".github/instructions/core/one.instructions.md",
			},
		},
		{
			SchemaVersion: metadataSchemaVersion,
			ID:            "core/two",
			Version:       "1.0.0",
			Category:      "core",
			Source:        "bundled",
			Dependencies:  []PackDependency{{ID: "core/one", Required: true}},
			OwnedArtifacts: []string{
				".github/instructions/core/two.instructions.md",
			},
		},
		{
			SchemaVersion: metadataSchemaVersion,
			ID:            "bad-id",
			Version:       "not-semver",
			Category:      "core",
			Source:        "bundled",
			OwnedArtifacts: []string{
				"outside.md",
			},
		},
	}

	issues := validatePackSet(packs)
	joined := strings.Join(issues, "\n")
	for _, want := range []string{
		"missing dependency",
		"dependency cycle detected",
		"malformed pack id",
		"invalid version",
		"invalid owned artefact",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected validation issue %q, got:\n%s", want, joined)
		}
	}
}

func TestPackRunParsesSubcommands(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	out := captureStdout(t, func() {
		if runErr := cmd.Run(context.Background(), []string{"list", "--json"}); runErr != nil {
			t.Fatalf("Run list --json: %v", runErr)
		}
	})
	if !strings.Contains(out, `"packs"`) {
		t.Fatalf("expected JSON list output, got:\n%s", out)
	}
}
