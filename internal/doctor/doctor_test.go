package doctor_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goldjg/carl/internal/doctor"
	"github.com/goldjg/carl/internal/manifest"
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

// captureStdout redirects os.Stdout for the duration of fn and returns
// everything written to it.
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

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// writeRuntime writes a manifest and, for each managed artefact, creates an
// on-disk file whose content matches arts (so the runtime appears healthy by
// default). Pass nil arts to skip file creation.
func writeRuntime(t *testing.T, dir string, rt *manifest.Runtime, arts *fakeArts) {
	t.Helper()
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatalf("manifest.Write: %v", err)
	}
	if arts == nil {
		return
	}
	for _, f := range rt.ManagedArtifacts {
		target := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(target); os.IsNotExist(err) {
			var content []byte
			if data, err2 := arts.Open(f); err2 == nil {
				content = data
			}
			if err := os.WriteFile(target, content, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// Contract assertion 1: no runtime installed → ERROR finding with `carl init` action,
// return nil (success).
func TestDoctor_NoRuntime(t *testing.T) {
	dir := t.TempDir()
	cmd := doctor.New(nil)

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir returned error; want nil; got: %v", runErr)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected ERROR finding; got: %q", output)
	}
	if !strings.Contains(output, "carl init") {
		t.Errorf("expected `carl init` action; got: %q", output)
	}
}

// Contract assertion 2: healthy runtime → INFO finding, no errors or warnings,
// return nil.
func TestDoctor_Healthy(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml":                    []byte("invariants: []"),
		".github/instructions/core/carl.instructions.md": []byte("# carl pack"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion: "1.0.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.0.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/carl/invariants.yml",
			".github/instructions/core/carl.instructions.md",
		},
	}
	writeRuntime(t, dir, rt, arts)

	cmd := doctor.New(arts)
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir returned error; want nil; got: %v", runErr)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO finding for healthy runtime; got: %q", output)
	}
	if strings.Contains(output, "ERROR") {
		t.Errorf("unexpected ERROR in healthy output; got: %q", output)
	}
	if strings.Contains(output, "WARNING") {
		t.Errorf("unexpected WARNING in healthy output; got: %q", output)
	}
}

// Contract assertion 3: missing artefact → ERROR finding with `carl repair` action,
// return nil.
func TestDoctor_MissingArtefact(t *testing.T) {
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

	cmd := doctor.New(arts)
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir returned error; want nil; got: %v", runErr)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected ERROR finding; got: %q", output)
	}
	if !strings.Contains(output, ".github/carl/invariants.yml") {
		t.Errorf("expected artefact path in output; got: %q", output)
	}
	if !strings.Contains(output, "carl repair") {
		t.Errorf("expected `carl repair` action; got: %q", output)
	}
}

// Contract assertion 4: drifted artefact → WARNING finding with `carl repair` action,
// return nil.
func TestDoctor_DriftedArtefact(t *testing.T) {
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

	cmd := doctor.New(arts)
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir returned error; want nil; got: %v", runErr)
	}
	if !strings.Contains(output, "WARNING") {
		t.Errorf("expected WARNING finding; got: %q", output)
	}
	if !strings.Contains(output, ".github/carl/invariants.yml") {
		t.Errorf("expected artefact path in output; got: %q", output)
	}
	if !strings.Contains(output, "carl repair") {
		t.Errorf("expected `carl repair` action; got: %q", output)
	}
}

// Contract assertion 5: doctor never returns a non-nil error when findings exist —
// it always returns nil (diagnostics complete successfully even with issues).
func TestDoctor_AlwaysReturnsSuccess(t *testing.T) {
	dir := t.TempDir()
	arts := &fakeArts{files: map[string][]byte{
		".github/carl/invariants.yml": []byte("canonical content"),
	}}
	rt := &manifest.Runtime{
		RuntimeVersion:   "1.0.0",
		Source:           "goldjg/cARL",
		ManagedArtifacts: []string{".github/carl/invariants.yml"},
	}
	// Write manifest; leave artefact absent → doctor will find missing artefact.
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatal(err)
	}

	cmd := doctor.New(arts)
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Errorf("RunInDir returned non-nil error; want nil; got: %v", err)
		}
	})
}

// TestDoctor_Run exercises the Run method via the command interface.
func TestDoctor_Run(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := doctor.New(nil)
	_ = captureStdout(t, func() {
		if err := cmd.Run(context.Background(), nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})
}
