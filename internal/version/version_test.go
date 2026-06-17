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

// writeRuntime is a helper that writes a runtime manifest into dir.
func writeRuntime(t *testing.T, dir string, rt *manifest.Runtime) {
	t.Helper()
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatalf("manifest.Write: %v", err)
	}
	// Create stub files for any managed artefacts so status is Healthy.
	for _, f := range rt.ManagedArtifacts {
		target := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(target); os.IsNotExist(err) {
			if err := os.WriteFile(target, []byte("stub"), 0644); err != nil {
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

	cmd := version.New("1.0.0")
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
	writeRuntime(t, dir, rt)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0")
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

	cmd := version.New("1.0.0")
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
	writeRuntime(t, dir, rt)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := version.New("1.0.0")
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

	cmd := version.New("1.0.0")
	if err := cmd.Run(context.Background(), nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
