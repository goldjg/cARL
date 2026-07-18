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

func (f *fakeArts) List() ([]string, error) {
	out := make([]string, 0, len(f.files))
	for k := range f.files {
		out = append(out, k)
	}
	return out, nil
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

	buf := make([]byte, 32*1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func newCommand(arts *fakeArts, bundledVersion string) *version.Command {
	return version.New(
		"1.2.0",
		bundledVersion,
		"goldjg/cARL",
		"v1.2.0",
		"98f680b3",
		arts,
	)
}

func writeRuntime(t *testing.T, dir string, rt *manifest.Runtime) {
	t.Helper()
	if err := manifest.Write(dir, rt); err != nil {
		t.Fatalf("manifest.Write: %v", err)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func bundledArtifacts() *fakeArts {
	return &fakeArts{
		files: map[string][]byte{
			".github/instructions/core/baseline.instructions.md": []byte("<!-- version: 1.1.0 -->\n"),
			".github/instructions/core/carl.instructions.md":     []byte("<!-- version: 2.0.0 -->\n"),
			".github/instructions/cloud/azure.instructions.md":   []byte("<!-- version: 1.0.1 -->\n"),
			".github/copilot-instructions.md":                    []byte("<!-- version: 2.1.0 -->\n"),
			"CLAUDE.md":                                          []byte("# Claude loader without version\n"),
			"AGENTS.md":                                          []byte("# Codex loader without version\n"),
			".cursor/rules/carl.mdc":                             []byte("# Cursor loader without version\n"),
			".agents/rules/carl.md":                              []byte("# Antigravity loader without version\n"),
		},
	}
}

func TestVersion_NoRuntimeStillShowsCLIAndBundled(t *testing.T) {
	dir := t.TempDir()
	cmd := newCommand(bundledArtifacts(), "1.2.0")

	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	for _, want := range []string{
		"cARL CLI:",
		"  Version:          1.2.0",
		"Bundled Runtime:",
		"  Version:          1.2.0",
		"  Source:           goldjg/cARL",
		"  Tag:              v1.2.0",
		"  Commit:           98f680b3",
		"Repository Runtime:",
		"  Not installed in the current repository.",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output:\n%s", want, output)
		}
	}
}

func TestVersion_InstalledRuntimeAndStatusCurrent(t *testing.T) {
	dir := t.TempDir()
	arts := bundledArtifacts()
	rt := &manifest.Runtime{
		RuntimeVersion: "1.2.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.2.0",
		SourceCommit:   "742ac661",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/instructions/core/baseline.instructions.md",
			".github/instructions/core/carl.instructions.md",
			".github/instructions/cloud/azure.instructions.md",
		},
	}
	writeRuntime(t, dir, rt)
	writeFile(t, dir, ".github/instructions/core/baseline.instructions.md", "<!-- version: 1.1.0 -->\n")
	writeFile(t, dir, ".github/instructions/core/carl.instructions.md", "<!-- version: 2.0.0 -->\n")
	writeFile(t, dir, ".github/instructions/cloud/azure.instructions.md", "<!-- version: 1.0.1 -->\n")

	// Harness detection files (installed + varied metadata quality).
	writeFile(t, dir, ".github/copilot-instructions.md", "<!-- version: 2.1.0 -->\n")
	writeFile(t, dir, "CLAUDE.md", "<!-- version: 1.0.0 -->\n")
	writeFile(t, dir, "AGENTS.md", "# no version header\n")

	cmd := newCommand(arts, "1.2.0")
	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	for _, want := range []string{
		"Repository Runtime:",
		"  Version:          1.2.0",
		"  Status:           Current",
		"Installed Packs:",
		"  cloud/azure                       1.0.1",
		"  core/baseline                     1.1.0",
		"  core/carl                         2.0.0",
		"Harness Shims:",
		"  copilot      .github/copilot-instructions.md",
		"  claude       CLAUDE.md",
		"  codex        AGENTS.md",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output:\n%s", want, output)
		}
	}
}

func TestVersion_RuntimeComparisonStates(t *testing.T) {
	tests := []struct {
		name           string
		repoVersion    string
		bundledVersion string
		wantStatus     string
	}{
		{
			name:           "bundled newer",
			repoVersion:    "1.0.0",
			bundledVersion: "1.2.0",
			wantStatus:     "Upgrade available",
		},
		{
			name:           "repository newer",
			repoVersion:    "1.3.0",
			bundledVersion: "1.2.0",
			wantStatus:     "Repository runtime is newer",
		},
		{
			name:           "malformed semver",
			repoVersion:    "release-42",
			bundledVersion: "1.2.0",
			wantStatus:     "Unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			rt := &manifest.Runtime{
				RuntimeVersion:   tc.repoVersion,
				Source:           "goldjg/cARL",
				SourceTag:        "vX",
				SourceCommit:     "abc",
				InstalledAt:      time.Now(),
				ManagedArtifacts: []string{},
			}
			writeRuntime(t, dir, rt)

			cmd := newCommand(bundledArtifacts(), tc.bundledVersion)
			output := captureStdout(t, func() {
				if err := cmd.RunInDir(dir); err != nil {
					t.Fatalf("RunInDir: %v", err)
				}
			})
			if !strings.Contains(output, "  Status:           "+tc.wantStatus) {
				t.Fatalf("expected status %q, got output:\n%s", tc.wantStatus, output)
			}
		})
	}
}

func TestVersion_PackVersionUnknownCases(t *testing.T) {
	dir := t.TempDir()
	rt := &manifest.Runtime{
		RuntimeVersion: "1.2.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.2.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/instructions/core/baseline.instructions.md",
			".github/instructions/core/carl.instructions.md",
			".github/instructions/cloud/azure.instructions.md",
		},
	}
	writeRuntime(t, dir, rt)
	// valid
	writeFile(t, dir, ".github/instructions/core/baseline.instructions.md", "<!-- version: 1.1.0 -->\n")
	// malformed
	writeFile(t, dir, ".github/instructions/core/carl.instructions.md", "<!-- version: not-semver -->\n")
	// azure file intentionally missing

	cmd := newCommand(bundledArtifacts(), "1.2.0")
	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	for _, want := range []string{
		"  core/baseline                     1.1.0",
		"  core/carl                         unknown",
		"  cloud/azure                       unknown",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
	if !strings.Contains(output, "not installed") || !strings.Contains(output, "unknown") {
		t.Fatalf("expected missing and malformed shim metadata states in output:\n%s", output)
	}
}

func TestVersion_HarnessShimMissingMalformedAndLoaderDedup(t *testing.T) {
	dir := t.TempDir()
	cmd := newCommand(bundledArtifacts(), "1.2.0")

	writeFile(t, dir, ".github/copilot-instructions.md", "<!-- version: malformed -->\n")
	writeFile(t, dir, "CLAUDE.md", "<!-- version: 1.0.0 -->\n")
	// AGENTS.md missing

	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if strings.Count(output, ".github/copilot-instructions.md") != 1 {
		t.Fatalf("shared loader path should appear once in Harness Shims section:\n%s", output)
	}
	for _, want := range []string{
		"  copilot      .github/copilot-instructions.md",
		"  claude       CLAUDE.md",
		"  codex        AGENTS.md",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
}

func TestVersion_DeterministicOutput(t *testing.T) {
	dir := t.TempDir()
	rt := &manifest.Runtime{
		RuntimeVersion: "1.2.0",
		Source:         "goldjg/cARL",
		SourceTag:      "v1.2.0",
		SourceCommit:   "abc123",
		InstalledAt:    time.Now(),
		ManagedArtifacts: []string{
			".github/instructions/core/carl.instructions.md",
			".github/instructions/core/baseline.instructions.md",
		},
	}
	writeRuntime(t, dir, rt)
	writeFile(t, dir, ".github/instructions/core/carl.instructions.md", "<!-- version: 2.0.0 -->\n")
	writeFile(t, dir, ".github/instructions/core/baseline.instructions.md", "<!-- version: 1.1.0 -->\n")

	cmd := newCommand(bundledArtifacts(), "1.2.0")
	first := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})
	second := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if first != second {
		t.Fatalf("output must be deterministic\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestVersion_ComponentsComparisonOutput(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".github/instructions/core/baseline.instructions.md", "<!-- version: 1.0.0 -->\n") // older
	writeFile(t, dir, ".github/instructions/core/carl.instructions.md", "<!-- version: 2.0.0 -->\n")     // current
	// cloud/azure missing
	writeFile(t, dir, ".github/copilot-instructions.md", "<!-- version: 2.0.0 -->\n") // older
	writeFile(t, dir, "CLAUDE.md", "<!-- version: 0.9.0 -->\n")                       // older vs unknown bundled => unknown

	cmd := newCommand(bundledArtifacts(), "1.2.0")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	output := captureStdout(t, func() {
		if err := cmd.Run(context.Background(), []string{"--components"}); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	for _, want := range []string{
		"Instruction Packs:",
		"Pack                              Bundled   Installed  State",
		"core/baseline                     1.1.0     1.0.0      older",
		"core/carl                         2.0.0     2.0.0      current",
		"cloud/azure                       1.0.1     missing    missing",
		"Harness Shims:",
		"Harness       File                              Bundled   Installed  State",
		"copilot       .github/copilot-instructions.md   2.1.0     2.0.0      older",
		"codex         AGENTS.md",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output:\n%s", want, output)
		}
	}
	if !strings.Contains(output, "unknown   missing    missing") {
		t.Fatalf("expected missing shim comparison state in output:\n%s", output)
	}
}
