package repair_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/install"
	"github.com/goldjg/carl/internal/manifest"
	"github.com/goldjg/carl/internal/repair"
)

// fakeArtifacts is a shared in-memory Artifacts implementation.
type fakeArtifacts struct {
	files map[string][]byte
}

func (f *fakeArtifacts) List() ([]string, error) {
	keys := make([]string, 0, len(f.files))
	for k := range f.files {
		keys = append(keys, k)
	}
	return keys, nil
}

func (f *fakeArtifacts) Open(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func newFakeArtifacts() *fakeArtifacts {
	return &fakeArtifacts{
		files: map[string][]byte{
			".github/copilot-instructions.md":                []byte("# Instructions"),
			".github/carl/memory.md":                         []byte("# Memory"),
			".github/carl/invariants.yml":                    []byte("invariants: []"),
			".github/instructions/core/carl.instructions.md": []byte("# carl pack"),
		},
	}
}

// installedDir creates a temp dir, runs carl init, and returns the dir path.
func installedDir(t *testing.T, arts *fakeArtifacts) string {
	t.Helper()
	dir := t.TempDir()
	ic := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	if err := ic.RunInDir(dir); err != nil {
		t.Fatalf("init: %v", err)
	}
	return dir
}

// TestRepair_NoDrift verifies that repair reports no drift when files match.
func TestRepair_NoDrift(t *testing.T) {
	arts := newFakeArtifacts()
	dir := installedDir(t, arts)

	rc := repair.New(arts)
	if err := rc.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}
}

// TestRepair_RestoresDriftedFile verifies that a modified managed artefact is
// restored to its embedded canonical version.
func TestRepair_RestoresDriftedFile(t *testing.T) {
	arts := newFakeArtifacts()
	dir := installedDir(t, arts)

	// Corrupt a managed artefact.
	target := filepath.Join(dir, ".github", "copilot-instructions.md")
	if err := os.WriteFile(target, []byte("corrupted"), 0644); err != nil {
		t.Fatal(err)
	}

	rc := repair.New(arts)
	if err := rc.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	// File must be restored to canonical content.
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# Instructions" {
		t.Errorf("restored content = %q; want %q", got, "# Instructions")
	}
}

// TestRepair_RestoresMissingFile verifies that a deleted managed artefact is
// re-created from the embedded canonical version.
func TestRepair_RestoresMissingFile(t *testing.T) {
	arts := newFakeArtifacts()
	dir := installedDir(t, arts)

	// Delete a managed artefact.
	target := filepath.Join(dir, ".github", "copilot-instructions.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}

	rc := repair.New(arts)
	if err := rc.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	// File must be restored.
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Error("artefact not restored after repair")
	}
}

// TestRepair_MemoryNotOverwritten verifies that memory.md is never overwritten
// by repair, even if it has been modified.
func TestRepair_MemoryNotOverwritten(t *testing.T) {
	arts := newFakeArtifacts()
	dir := installedDir(t, arts)

	// Modify memory.md.
	memPath := filepath.Join(dir, ".github", "carl", "memory.md")
	userContent := []byte("# My custom memory notes\n\nThis must survive repair.")
	if err := os.WriteFile(memPath, userContent, 0644); err != nil {
		t.Fatal(err)
	}

	rc := repair.New(arts)
	if err := rc.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	// memory.md must retain user content.
	got, err := os.ReadFile(memPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(userContent) {
		t.Errorf("memory.md modified by repair; got %q; want %q", got, userContent)
	}
}

// TestRepair_RuntimeJSONNotOverwritten verifies that runtime.json is never
// touched by repair.
func TestRepair_RuntimeJSONNotOverwritten(t *testing.T) {
	arts := newFakeArtifacts()
	dir := installedDir(t, arts)

	// Read the original runtime.json content.
	original, err := manifest.Read(dir)
	if err != nil {
		t.Fatal(err)
	}

	rc := repair.New(arts)
	if err := rc.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	// runtime.json must be unchanged.
	after, err := manifest.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !original.InstalledAt.Equal(after.InstalledAt) {
		t.Error("runtime.json InstalledAt changed after repair")
	}
}

// TestRepair_NoRuntime verifies that repair fails gracefully when no runtime
// is installed.
func TestRepair_NoRuntime(t *testing.T) {
	arts := newFakeArtifacts()
	dir := t.TempDir()

	rc := repair.New(arts)
	err := rc.RunInDir(dir)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if !strings.Contains(err.Error(), "carl init") {
		t.Errorf("error should mention carl init; got: %v", err)
	}
}

// TestRepair_Run exercises the Run method via the command interface.
func TestRepair_Run(t *testing.T) {
	arts := newFakeArtifacts()
	dir := installedDir(t, arts)

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	rc := repair.New(arts)
	if err := rc.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
