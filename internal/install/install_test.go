package install_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/install"
	"github.com/goldjg/carl/internal/manifest"
)

// fakeArtifacts is a simple in-memory Artifacts implementation for testing.
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
	return f.files[path], nil
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

// TestInit_Success verifies that init installs all artefacts and creates runtime.json.
func TestInit_Success(t *testing.T) {
	dir := t.TempDir()
	arts := newFakeArtifacts()
	cmd := install.New(arts, "1.2.0", "goldjg/cARL", "v1.2.0", "deadbeef")

	if err := cmd.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	// runtime.json must exist.
	if !manifest.Exists(dir) {
		t.Error("runtime.json not created")
	}

	// All artefacts must exist.
	files, _ := arts.List()
	for _, f := range files {
		target := filepath.Join(dir, filepath.FromSlash(f))
		if _, err := os.Stat(target); os.IsNotExist(err) {
			t.Errorf("artefact not installed: %s", f)
		}
	}

	// runtime.json must record managedArtifacts.
	rt, err := manifest.Read(dir)
	if err != nil {
		t.Fatalf("manifest.Read: %v", err)
	}
	if len(rt.ManagedArtifacts) != len(files) {
		t.Errorf("ManagedArtifacts count = %d; want %d", len(rt.ManagedArtifacts), len(files))
	}
	if rt.RuntimeVersion != "1.2.0" {
		t.Errorf("RuntimeVersion = %q; want %q", rt.RuntimeVersion, "1.2.0")
	}
	if rt.Source != "goldjg/cARL" || rt.SourceTag != "v1.2.0" || rt.SourceCommit != "deadbeef" {
		t.Errorf("unexpected bundled runtime metadata in manifest: %+v", rt)
	}
}

// TestInit_AlreadyInstalled verifies that re-running init fails safely when
// runtime.json already exists.
func TestInit_AlreadyInstalled(t *testing.T) {
	dir := t.TempDir()
	arts := newFakeArtifacts()
	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")

	// First init succeeds.
	if err := cmd.RunInDir(dir); err != nil {
		t.Fatalf("first RunInDir: %v", err)
	}

	// Second init must fail.
	err := cmd.RunInDir(dir)
	if err == nil {
		t.Fatal("expected error on second init; got nil")
	}
	if !strings.Contains(err.Error(), "already installed") {
		t.Errorf("error message should mention already installed; got: %v", err)
	}
}

// TestInit_ConflictingFiles verifies that init fails when artefacts already
// exist even without runtime.json.
func TestInit_ConflictingFiles(t *testing.T) {
	dir := t.TempDir()
	arts := newFakeArtifacts()

	// Pre-create one of the artefact files.
	target := filepath.Join(dir, ".github", "carl", "memory.md")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("existing content"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	err := cmd.RunInDir(dir)
	if err == nil {
		t.Fatal("expected conflict error; got nil")
	}
	if !strings.Contains(err.Error(), ".github/carl/memory.md") {
		t.Errorf("error should list conflicting file; got: %v", err)
	}
}

// TestInit_Run exercises the Run method via the command interface (uses cwd).
func TestInit_Run(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	arts := newFakeArtifacts()
	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	if err := cmd.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !manifest.Exists(dir) {
		t.Error("runtime.json not created")
	}
}
