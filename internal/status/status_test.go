package status_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goldjg/carl/internal/manifest"
	"github.com/goldjg/carl/internal/status"
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
// matches arts, so status reports "Healthy" by default.
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

// captureStdout redirects os.Stdout for the duration of f and returns
// everything written to it.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// Contract assertion 1: when no runtime is installed, print "No cARL runtime installed."
func TestStatus_NoRuntime(t *testing.T) {
	dir := t.TempDir()
	cmd := status.New("1.0.0", nil)

	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if !strings.Contains(output, "No cARL runtime installed") {
		t.Errorf("expected 'No cARL runtime installed'; got: %q", output)
	}
}

// Contract assertion 2: healthy runtime outputs all expected fields and "Healthy" status.
func TestStatus_Healthy(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml":                              []byte("invariants: []"),
		".github/instructions/core/carl.instructions.md":           []byte("# carl pack"),
		".github/instructions/core/memory-cache.instructions.md":   []byte("# memory-cache pack"),
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

	output := captureStdout(t, func() {
		cmd := status.New("1.0.0", arts)
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	for _, want := range []string{
		"CLI Version:      1.0.0",
		"Runtime Version:  1.0.0",
		"Source:           goldjg/cARL",
		"Tag:              v1.0.0",
		"Commit:           abc123",
		"Installed Packs:",
		"core/carl",
		"core/memory-cache",
		"Missing Artefacts:",
		"Drifted Artefacts:",
		"Status:           Healthy",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}

// Contract assertion 3: a missing artefact is listed under "Missing Artefacts:"
// and overall status is "Incomplete".
func TestStatus_MissingArtefact(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml": []byte("invariants: []"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion:   "1.0.0",
		Source:           "goldjg/cARL",
		SourceTag:        "v1.0.0",
		SourceCommit:     "abc123",
		InstalledAt:      time.Now(),
		ManagedArtifacts: []string{".github/carl/invariants.yml"},
	}
	// Write manifest but do NOT create the artefact file on disk.
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		cmd := status.New("1.0.0", arts)
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if !strings.Contains(output, ".github/carl/invariants.yml") {
		t.Errorf("expected missing artefact in output; got: %q", output)
	}
	if !strings.Contains(output, "Status:           Incomplete") {
		t.Errorf("expected 'Status: Incomplete'; got: %q", output)
	}
}

// Contract assertion 4: a drifted (content-modified) artefact is listed under
// "Drifted Artefacts:" and overall status is "Drifted".
func TestStatus_DriftedArtefact(t *testing.T) {
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
	// Write manifest and the file with WRONG content.
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, ".github", "carl", "invariants.yml")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("modified content"), 0644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		cmd := status.New("1.0.0", arts)
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if !strings.Contains(output, ".github/carl/invariants.yml") {
		t.Errorf("expected drifted artefact in output; got: %q", output)
	}
	if !strings.Contains(output, "Status:           Drifted") {
		t.Errorf("expected 'Status: Drifted'; got: %q", output)
	}
}

// Contract assertion 5: memory.md and runtime.json are never reported as
// missing or drifted.
func TestStatus_ProtectedArtefactsIgnored(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/memory.md": []byte("canonical memory"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion: "1.0.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.0.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/carl/memory.md",
			".github/carl/runtime.json",
		},
	}
	// Write manifest; modify memory.md to diverge from canonical.
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatal(err)
	}
	memPath := filepath.Join(dir, ".github", "carl", "memory.md")
	if err := os.MkdirAll(filepath.Dir(memPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(memPath, []byte("user-modified content"), 0644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		cmd := status.New("1.0.0", arts)
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	// Protected files must not appear in the artefact lists.
	if strings.Contains(output, "memory.md") {
		t.Errorf("memory.md must not be reported; got: %q", output)
	}
	if strings.Contains(output, "runtime.json") {
		t.Errorf("runtime.json must not be reported; got: %q", output)
	}
	// With no repairable artefacts to check, the runtime is Healthy.
	if !strings.Contains(output, "Status:           Healthy") {
		t.Errorf("expected 'Status: Healthy' when only protected artefacts exist; got: %q", output)
	}
}

// TestStatus_Run exercises the Run method via the command interface.
func TestStatus_Run(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := status.New("1.0.0", nil)
	if err := cmd.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
