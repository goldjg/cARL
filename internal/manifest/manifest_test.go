package manifest_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goldjg/carl/internal/manifest"
)

func TestWriteRead(t *testing.T) {
	dir := t.TempDir()

	r := &manifest.Runtime{
		RuntimeVersion:   "1.0.0",
		Source:           "goldjg/cARL",
		SourceTag:        "v1.0.0",
		SourceCommit:     "abc123",
		InstalledAt:      time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC),
		ManagedArtifacts: []string{".github/carl/memory.md", ".github/carl/invariants.yml"},
	}

	if err := manifest.Write(dir, r); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := manifest.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.RuntimeVersion != r.RuntimeVersion {
		t.Errorf("RuntimeVersion = %q; want %q", got.RuntimeVersion, r.RuntimeVersion)
	}
	if got.Source != r.Source {
		t.Errorf("Source = %q; want %q", got.Source, r.Source)
	}
	if got.SourceTag != r.SourceTag {
		t.Errorf("SourceTag = %q; want %q", got.SourceTag, r.SourceTag)
	}
	if got.SourceCommit != r.SourceCommit {
		t.Errorf("SourceCommit = %q; want %q", got.SourceCommit, r.SourceCommit)
	}
	if len(got.ManagedArtifacts) != len(r.ManagedArtifacts) {
		t.Errorf("ManagedArtifacts len = %d; want %d", len(got.ManagedArtifacts), len(r.ManagedArtifacts))
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()

	if manifest.Exists(dir) {
		t.Error("Exists should be false before Write")
	}

	r := &manifest.Runtime{RuntimeVersion: "1.0.0"}
	if err := manifest.Write(dir, r); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !manifest.Exists(dir) {
		t.Error("Exists should be true after Write")
	}
}

func TestRead_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, manifest.FileName)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := manifest.Read(dir)
	if err == nil {
		t.Error("expected error reading invalid JSON")
	}
}

func TestRead_Missing(t *testing.T) {
	dir := t.TempDir()
	_, err := manifest.Read(dir)
	if err == nil {
		t.Error("expected error reading missing manifest")
	}
}
