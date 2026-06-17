package version_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goldjg/carl/internal/manifest"
	"github.com/goldjg/carl/internal/version"
)

// fakeArts is a minimal in-memory Artifacts implementation for testing.
type fakeArts struct {
	files map[string][]byte
}

func (f *fakeArts) Open(p string) ([]byte, error) {
	data, ok := f.files[p]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

// writeRuntime writes a manifest and creates on-disk files whose content
// matches what arts.Open returns, so the health check reports "Healthy".
// Protected paths (memory.md) are written as empty stubs.
func writeRuntime(t *testing.T, dir string, rt *manifest.Runtime, arts *fakeArts) {
	t.Helper()
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatalf("manifest.Write: %v", err)
	}
	for _, f := range rt.ManagedArtifacts {
		target := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(target); os.IsNotExist(err) {
			var content []byte
			if arts != nil {
				if data, err := arts.Open(f); err == nil {
					content = data
				}
			}
			if err := os.WriteFile(target, content, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// TestVersion_NoRuntime verifies the "not installed" message.
func TestVersion_NoRuntime(t *testing.T) {
	dir := t.TempDir()
	// Redirect stdout to capture output.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0", nil)
	if err := cmd.RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	_ = w.Close()
	os.Stdout = old

	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	if !strings.Contains(output, "No cARL runtime installed") {
		t.Errorf("expected 'No cARL runtime installed'; got: %q", output)
	}
}

// TestVersion_Installed verifies that version output includes all expected fields.
func TestVersion_Installed(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml":                         []byte("invariants: []"),
		".github/instructions/core/carl.instructions.md":      []byte("# carl pack"),
		".github/instructions/core/memory-cache.instructions.md": []byte("# memory-cache pack"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion: "1.0.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.0.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/carl/invariants.yml",
			".github/carl/memory.md",
			".github/instructions/core/carl.instructions.md",
			".github/instructions/core/memory-cache.instructions.md",
		},
	}
	writeRuntime(t, dir, rt, arts)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0", arts)
	if err := cmd.RunInDir(dir); err != nil {
		_ = w.Close()
		os.Stdout = old
		t.Fatalf("RunInDir: %v", err)
	}

	_ = w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	checks := []string{
		"CLI Version:      1.0.0",
		"Runtime Version:  1.0.0",
		"Source:           goldjg/cARL",
		"Tag:              v1.0.0",
		"Commit:           abc123",
		"Installed Packs:",
		"core/carl",
		"core/memory-cache",
		"Runtime Status:",
		"Healthy",
	}
	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}

// TestVersion_DriftStatus verifies that a missing artefact triggers drift status.
func TestVersion_DriftStatus(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml": []byte("invariants: []"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion: "1.0.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.0.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/carl/invariants.yml",
		},
	}
	// Write manifest but do NOT create the artefact file.
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0", arts)
	if err := cmd.RunInDir(dir); err != nil {
		_ = w.Close()
		os.Stdout = old
		t.Fatalf("RunInDir: %v", err)
	}

	_ = w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	if !strings.Contains(output, "Drift detected") {
		t.Errorf("expected 'Drift detected'; got: %q", output)
	}
}

// TestVersion_StateFromManifest verifies that version output is driven entirely
// by runtime.json content, not filesystem scans.
func TestVersion_StateFromManifest(t *testing.T) {
	dir := t.TempDir()
	rt := &manifest.Runtime{
		RuntimeVersion:   "2.3.4",
		Source:           "acme/cARL-fork",
		SourceTag:        "v2.3.4",
		SourceCommit:     "deadbeef",
		InstalledAt:      time.Now(),
		ManagedArtifacts: []string{},
	}
	writeRuntime(t, dir, rt, nil)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0", nil)
	if err := cmd.RunInDir(dir); err != nil {
		_ = w.Close()
		os.Stdout = old
		t.Fatalf("RunInDir: %v", err)
	}

	_ = w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	for _, want := range []string{"2.3.4", "acme/cARL-fork", "v2.3.4", "deadbeef"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q — version state must come from runtime.json\nfull output: %s", want, output)
		}
	}
}

// TestVersion_Run exercises the Run method via the command interface.
func TestVersion_Run(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := version.New("1.0.0", nil)
	if err := cmd.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

// TestVersion_DriftOnModifiedContent verifies that a present but corrupted
// artefact is reported as drift, not healthy.
func TestVersion_DriftOnModifiedContent(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml": []byte("canonical content"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion:   "1.0.0",
		Source:           "goldjg/cARL",
		SourceTag:        "v1.0.0",
		SourceCommit:     "abc123",
		InstalledAt:      time.Now(),
		ManagedArtifacts: []string{".github/carl/invariants.yml"},
	}
	// Write the manifest and create the file with WRONG content.
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, ".github", "carl", "invariants.yml")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("corrupted content"), 0644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0", arts)
	if err := cmd.RunInDir(dir); err != nil {
		_ = w.Close()
		os.Stdout = old
		t.Fatalf("RunInDir: %v", err)
	}

	_ = w.Close()
	os.Stdout = old

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	if !strings.Contains(output, "Drift detected") {
		t.Errorf("expected 'Drift detected' for modified content; got: %q", output)
	}
}
