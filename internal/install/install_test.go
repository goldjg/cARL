package install_test

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

func TestInit_AdoptPreservesExistingAndInstallsMissing(t *testing.T) {
	dir := t.TempDir()
	arts := newFakeArtifacts()
	existingPath := filepath.Join(dir, ".github", "carl", "memory.md")
	existingContent := []byte("# Repository memory\n")
	if err := os.MkdirAll(filepath.Dir(existingPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existingPath, existingContent, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	if err := cmd.RunInDirWithOptions(dir, true); err != nil {
		t.Fatalf("RunInDirWithOptions: %v", err)
	}
	got, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(existingContent) {
		t.Errorf("existing artefact overwritten: got %q; want %q", got, existingContent)
	}
	for path := range arts.files {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(path))); err != nil {
			t.Errorf("managed artefact %s not present after adoption: %v", path, err)
		}
	}
	rt, err := manifest.Read(dir)
	if err != nil {
		t.Fatalf("manifest.Read: %v", err)
	}
	if len(rt.ManagedArtifacts) != len(arts.files) {
		t.Errorf("managed artifact count = %d; want %d", len(rt.ManagedArtifacts), len(arts.files))
	}
}

func TestInit_AdoptDoesNotCreateManifestAfterInstallFailure(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArtifacts{files: map[string][]byte{
		".github/carl/memory.md":            []byte("# Memory"),
		".github/instructions/core/test.md": []byte("# Test"),
	}}
	blockingPath := filepath.Join(dir, ".github", "instructions")
	if err := os.MkdirAll(filepath.Dir(blockingPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blockingPath, []byte("not a directory"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	if err := cmd.RunInDirWithOptions(dir, true); err == nil {
		t.Fatal("expected adoption failure")
	}
	if manifest.Exists(dir) {
		t.Error("runtime.json created after an earlier installation failure")
	}
}

func TestInit_AdoptThenRepairPreservesProtectedMemory(t *testing.T) {
	dir := t.TempDir()
	arts := newFakeArtifacts()
	memoryPath := filepath.Join(dir, ".github", "carl", "memory.md")
	customMemory := []byte("# Custom memory")
	if err := os.MkdirAll(filepath.Dir(memoryPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(memoryPath, customMemory, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	if err := cmd.RunInDirWithOptions(dir, true); err != nil {
		t.Fatal(err)
	}

	repairable := filepath.Join(dir, ".github", "carl", "invariants.yml")
	if err := os.WriteFile(repairable, []byte("drifted"), 0644); err != nil {
		t.Fatal(err)
	}
	rc := repair.New(arts)
	if err := rc.RunInDir(dir); err != nil {
		t.Fatal(err)
	}
	gotMemory, err := os.ReadFile(memoryPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotMemory) != string(customMemory) {
		t.Errorf("memory overwritten: got %q; want %q", gotMemory, customMemory)
	}
	gotRepairable, err := os.ReadFile(repairable)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotRepairable) != string(arts.files[".github/carl/invariants.yml"]) {
		t.Errorf("repairable artefact not restored: %q", gotRepairable)
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

func TestInit_RunAcceptsAdoptFlag(t *testing.T) {
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
	existing := filepath.Join(dir, ".github", "carl", "memory.md")
	if err := os.MkdirAll(filepath.Dir(existing), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("preserve me"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := install.New(arts, "1.0.0", "goldjg/cARL", "v1.0.0", "dev")
	if err := cmd.Run(context.Background(), []string{"--adopt"}); err != nil {
		t.Fatalf("Run --adopt: %v", err)
	}
	if !manifest.Exists(dir) {
		t.Error("runtime.json not created by Run --adopt")
	}
}
