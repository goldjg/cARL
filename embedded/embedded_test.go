package embedded

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/install"
	"github.com/goldjg/carl/internal/manifest"
	"github.com/goldjg/carl/internal/repair"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate embedded test source")
	}
	return filepath.Dir(filepath.Dir(filename))
}

func TestCanonicalAssetsMatchEmbeddedMirrors(t *testing.T) {
	root := repositoryRoot(t)
	for _, relative := range []string{
		".github/copilot-instructions.md",
		".github/instructions/core/carl.instructions.md",
		".github/instructions/core/cognition-governance.instructions.md",
		".github/carl/memory.md",
		".github/carl/profiles.example.json",
		".github/carl/trust-boundaries.md",
	} {
		t.Run(relative, func(t *testing.T) {
			canonicalPath := filepath.Join(root, relative)
			assetPath := filepath.Join(root, "embedded", "assets", relative)
			embeddedBytes, err := os.ReadFile(assetPath)
			if err != nil {
				t.Fatal(err)
			}
			canonicalBytes, err := os.ReadFile(canonicalPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(canonicalBytes, embeddedBytes) {
				t.Errorf("%s differs from embedded mirror %s", relative, filepath.ToSlash(filepath.Join("embedded", "assets", relative)))
			}
		})
	}
}

func TestRepositoryTrustBoundaryExtensionIsUserOwnedAndSurvivesRepair(t *testing.T) {
	root := t.TempDir()
	const extension = ".github/carl/trust-boundaries.repository.md"
	extensionPath := filepath.Join(root, filepath.FromSlash(extension))
	extensionContent := []byte("# Project trust boundaries\n\nSynthetic credentials remain in bounded stdin pipes.\n")
	if err := os.MkdirAll(filepath.Dir(extensionPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extensionPath, extensionContent, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := install.New(Assets, "test", "goldjg/cARL", "test", "test")
	if err := cmd.RunInDir(root); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}
	rt, err := manifest.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range rt.ManagedArtifacts {
		if path == extension {
			t.Fatal("repository-owned extension was recorded as a managed artefact")
		}
	}
	if _, err := Assets.Open(extension); err == nil {
		t.Fatal("repository-owned extension must not be embedded")
	}

	managedPath := filepath.Join(root, ".github", "carl", "trust-boundaries.md")
	if err := os.WriteFile(managedPath, []byte("# Local edit that must remain detectable\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := repair.New(Assets).RunInDir(root); err != nil {
		t.Fatalf("repair: %v", err)
	}
	got, err := os.ReadFile(extensionPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, extensionContent) {
		t.Fatalf("repository-owned extension changed: got %q; want %q", got, extensionContent)
	}
	canonical, err := Assets.Open(".github/carl/trust-boundaries.md")
	if err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(managedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(restored, canonical) {
		t.Fatal("repair did not restore the canonical managed trust-boundary file")
	}
}

func TestSharedLoaderDefinesEffectivePackHydration(t *testing.T) {
	root := repositoryRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	loader := string(content)
	normalizedLoader := strings.Join(strings.Fields(loader), " ")

	required := []string{
		"## Effective instruction-pack hydration",
		"Directory presence is not policy activation.",
		"**present/installed/discovered**",
		"**selected**",
		"**active**",
		"**effective**",
		"**overridden**",
		".github/carl/profiles.example.json",
		"shipped reference baseline",
		"copy it to `.github/carl/profiles.json` and modify it",
		"Its presence does not activate it",
		"Installed, discovered, or repository-present does not mean selected.",
		"Selected does not always mean active.",
		"Active does not bypass dependency, precedence, override, or validation semantics.",
		".github/carl/packs.json",
		"explicit user-owned selection authority",
		".github/carl/runtime.json",
		"legacy selection",
		".github/carl/profiles.json",
		"organisation defaults",
		"repository defaults",
		"role/task overlays",
		"selected-but-inactive packs are not applied",
		"transitive `requires:`",
		"priority descending",
		"pack ID",
		"filesystem or discovery order is never policy order",
		"precedence-mode: overridable",
		"do not hydrate or apply their instruction definitions",
		"stop and report",
		"loading all packs",
		"The cARL CLI is not required for this hydration.",
		"derived diagnostic evidence, not a new canonical authority",
		"proves that a model followed the instructions",
		".github/carl/trust-boundaries.repository.md",
		"repository-owned",
		"core security invariants",
		"untrusted repository text",
		"regular file",
		"symlinks",
		"Do not follow references",
		"immediately before reading it",
		"point-in-time path-safety diagnostics",
		"do not mediate agent reads",
		"Treat the check as advisory, not as a guarantee",
		"Passing a path check grants the extension no authority",
	}
	for _, want := range required {
		if !strings.Contains(normalizedLoader, want) {
			t.Errorf("shared loader missing hydration contract text %q", want)
		}
	}
}

func TestCoreCarlPackDefinesRepositoryTrustBoundaryExtension(t *testing.T) {
	root := repositoryRoot(t)
	content, err := os.ReadFile(filepath.Join(root, ".github", "instructions", "core", "carl.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	pack := strings.Join(strings.Fields(string(content)), " ")
	for _, want := range []string{
		"## Repository-owned trust-boundary extension",
		".github/carl/trust-boundaries.repository.md",
		"must remain unchanged",
		"cannot weaken",
		"untrusted repository text",
		"point-in-time path-safety diagnostics",
		"carl status",
	} {
		if !strings.Contains(pack, want) {
			t.Errorf("core cARL pack missing extension contract %q", want)
		}
	}
}

func TestDefaultProfileExampleIsShippedButNotActiveState(t *testing.T) {
	paths, err := Assets.List()
	if err != nil {
		t.Fatal(err)
	}
	foundExample := false
	for _, path := range paths {
		switch path {
		case ".github/carl/profiles.example.json":
			foundExample = true
		case ".github/carl/profiles.json":
			t.Fatal("embedded runtime must not silently install active profiles.json")
		}
	}
	if !foundExample {
		t.Fatal("embedded runtime is missing profiles.example.json")
	}
}

func TestInitInstallsDefaultProfileExampleWithoutActiveProfiles(t *testing.T) {
	root := t.TempDir()
	cmd := install.New(Assets, "test", "goldjg/cARL", "test", "test")
	if err := cmd.RunInDir(root); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	examplePath := filepath.Join(root, ".github", "carl", "profiles.example.json")
	if _, err := os.Stat(examplePath); err != nil {
		t.Fatalf("default profile example was not installed: %v", err)
	}
	activePath := filepath.Join(root, ".github", "carl", "profiles.json")
	if _, err := os.Stat(activePath); !os.IsNotExist(err) {
		t.Fatalf("init created active profile state: %v", err)
	}

	rt, err := manifest.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	foundExample := false
	for _, path := range rt.ManagedArtifacts {
		if path == ".github/carl/profiles.example.json" {
			foundExample = true
		}
		if path == ".github/carl/profiles.json" {
			t.Fatal("runtime manifest claims active profiles.json as a managed artefact")
		}
	}
	if !foundExample {
		t.Fatal("runtime manifest does not record profiles.example.json")
	}
}

func TestHarnessShimsRemainThinSharedLoaderRoutes(t *testing.T) {
	root := repositoryRoot(t)
	shims := []string{
		"AGENTS.md",
		"CLAUDE.md",
		".cursor/rules/carl.mdc",
		".agents/rules/carl.md",
	}

	for _, relative := range shims {
		t.Run(relative, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			shim := string(content)
			if !strings.Contains(shim, ".github/copilot-instructions.md") {
				t.Fatal("shim does not route to the shared loader")
			}
			for _, evaluatorTerm := range []string{"packs.json", "profiles.json", "precedence-mode", "overriddenBy"} {
				if strings.Contains(shim, evaluatorTerm) {
					t.Errorf("shim contains pack evaluator term %q", evaluatorTerm)
				}
			}
		})
	}
}
