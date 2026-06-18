package repomap_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/repomap"
)

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

// Contract assertion 1: running `carl map` creates .github/carl/repo-map.json.
func TestMap_CreatesOutputFile(t *testing.T) {
	dir := t.TempDir()
	cmd := repomap.New()

	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	path := filepath.Join(dir, ".github", "carl", "repo-map.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

// Contract assertion 2: output JSON contains generated_by, last_updated, and _note.
func TestMap_ValidJSONWithRequiredFields(t *testing.T) {
	dir := t.TempDir()
	cmd := repomap.New()

	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	data, err := os.ReadFile(filepath.Join(dir, ".github", "carl", "repo-map.json"))
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	var m repomap.Map
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m.GeneratedBy != "carl map" {
		t.Errorf("generated_by: got %q, want %q", m.GeneratedBy, "carl map")
	}
	if m.LastUpdated == "" {
		t.Error("last_updated must not be empty")
	}
	if m.Note == "" {
		t.Error("_note must not be empty")
	}
}

// Contract assertion 3: Go source files cause "Go" to appear in languages.
func TestMap_DetectsGoLanguage(t *testing.T) {
	dir := t.TempDir()
	// Write a minimal Go file.
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, l := range m.Languages {
		if l == "Go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Go' in languages; got %v", m.Languages)
	}
}

// Contract assertion 3 (extended): multiple language files produce sorted language list.
func TestMap_DetectsMultipleLanguages(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"main.go":   "package main\n",
		"script.py": "print('hello')\n",
		"app.ts":    "const x = 1;\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{"Go": true, "Python": true, "TypeScript": true}
	for _, l := range m.Languages {
		delete(want, l)
	}
	if len(want) > 0 {
		var missing []string
		for k := range want {
			missing = append(missing, k)
		}
		t.Errorf("missing languages %v; got %v", missing, m.Languages)
	}

	// Languages must be sorted.
	for i := 1; i < len(m.Languages); i++ {
		if m.Languages[i] < m.Languages[i-1] {
			t.Errorf("languages not sorted: %v", m.Languages)
			break
		}
	}
}

// Contract assertion 4: .github/workflows/*.yml files appear in workflows.
func TestMap_DetectsWorkflows(t *testing.T) {
	dir := t.TempDir()
	wfDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wfDir, "release.yml"), []byte("on: push\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Workflows) == 0 {
		t.Fatal("expected at least one workflow entry")
	}
	found := false
	for _, wf := range m.Workflows {
		if strings.Contains(wf.Path, "release.yml") {
			found = true
			if !strings.Contains(wf.Purpose, "release") {
				t.Errorf("workflow purpose should reference name; got %q", wf.Purpose)
			}
		}
	}
	if !found {
		t.Errorf("release.yml not found in workflows: %v", m.Workflows)
	}
}

// Contract assertion 5: files directly under .github/carl/ appear in governance.
func TestMap_DetectsGovernance(t *testing.T) {
	dir := t.TempDir()
	carlDir := filepath.Join(dir, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(carlDir, "invariants.yml"), []byte("invariants: []"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(carlDir, "memory.md"), []byte("# memory\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	governancePaths := map[string]bool{}
	for _, g := range m.Governance {
		governancePaths[g.Path] = true
	}

	if !governancePaths[".github/carl/invariants.yml"] {
		t.Errorf("invariants.yml missing from governance: %v", m.Governance)
	}
	if !governancePaths[".github/carl/memory.md"] {
		t.Errorf("memory.md missing from governance: %v", m.Governance)
	}
}

// Governance files with known names must carry the known purpose description.
func TestMap_GovernanceKnownPurposes(t *testing.T) {
	dir := t.TempDir()
	carlDir := filepath.Join(dir, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(carlDir, "invariants.yml"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, g := range m.Governance {
		if g.Path == ".github/carl/invariants.yml" {
			if g.Purpose == "" {
				t.Errorf("invariants.yml should have a non-empty purpose; got empty")
			}
			return
		}
	}
	t.Error("invariants.yml not found in governance")
}

// Contract assertion 6: root-level *.md files appear in documentation.
func TestMap_DetectsDocumentation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Readme\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte("# Log\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	docPaths := map[string]bool{}
	for _, d := range m.Documentation {
		docPaths[d.Path] = true
	}
	if !docPaths["README.md"] {
		t.Errorf("README.md missing from documentation: %v", m.Documentation)
	}
	if !docPaths["CHANGELOG.md"] {
		t.Errorf("CHANGELOG.md missing from documentation: %v", m.Documentation)
	}
}

// README.md must carry the known purpose description.
func TestMap_DocumentationKnownPurpose(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Readme\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range m.Documentation {
		if d.Path == "README.md" {
			if d.Purpose == "" {
				t.Errorf("README.md should have a non-empty purpose")
			}
			return
		}
	}
	t.Error("README.md not found in documentation")
}

// Contract assertion 7: repeated invocation (idempotent) still produces valid JSON.
func TestMap_Idempotent(t *testing.T) {
	dir := t.TempDir()
	cmd := repomap.New()

	for i := range 2 {
		_ = captureStdout(t, func() {
			if err := cmd.RunInDir(dir); err != nil {
				t.Fatalf("run %d: %v", i, err)
			}
		})
	}

	data, err := os.ReadFile(filepath.Join(dir, ".github", "carl", "repo-map.json"))
	if err != nil {
		t.Fatalf("read output file after idempotent run: %v", err)
	}
	var m repomap.Map
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON after idempotent run: %v", err)
	}
	if m.GeneratedBy != "carl map" {
		t.Errorf("generated_by: got %q, want %q", m.GeneratedBy, "carl map")
	}
}

// Contract assertion 8: .git directory is never included in directories or languages.
func TestMap_SkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	// Create a .git directory with a Go file inside.
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "COMMIT_EDITMSG"), []byte("chore: test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// .git-internal Go file must not be detected as a language.
	if err := os.WriteFile(filepath.Join(gitDir, "hook.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	for rel := range m.Directories {
		if strings.HasPrefix(rel, ".git") {
			t.Errorf(".git must not appear in directories; got %q", rel)
		}
	}

	// No languages should be detected (only the .git Go file exists).
	if len(m.Languages) != 0 {
		t.Errorf("languages should be empty (only .git Go file exists); got %v", m.Languages)
	}
}

// go.mod entry point includes the module name in its purpose.
func TestMap_EntryPoint_GoMod(t *testing.T) {
	dir := t.TempDir()
	modContent := "module github.com/example/myapp\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modContent), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, ep := range m.EntryPoints {
		if ep.Path == "go.mod" && strings.Contains(ep.Purpose, "github.com/example/myapp") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("go.mod entry with module name not found; got: %v", m.EntryPoints)
	}
}

// cmd/*/main.go files appear in entry points with a derived purpose.
func TestMap_EntryPoint_CmdMain(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "cmd", "myapp")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, ep := range m.EntryPoints {
		if ep.Path == "cmd/myapp/main.go" {
			found = true
			if !strings.Contains(ep.Purpose, "myapp") {
				t.Errorf("entry point purpose should reference cmd name; got %q", ep.Purpose)
			}
			break
		}
	}
	if !found {
		t.Errorf("cmd/myapp/main.go not found in entry points; got: %v", m.EntryPoints)
	}
}

// Directories derived from Go package doc comments carry the doc text.
func TestMap_DirectoryFromGoPackageDoc(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "mylib")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	src := "// Package mylib implements the widget factory.\npackage mylib\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "mylib.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	purpose, ok := m.Directories["mylib"]
	if !ok {
		t.Fatalf("mylib not found in directories; got %v", m.Directories)
	}
	if !strings.Contains(purpose, "widget factory") {
		t.Errorf("expected package doc in purpose; got %q", purpose)
	}
}

// TestMap_Run exercises the Run method via the process working directory.
func TestMap_Run(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := repomap.New()
	_ = captureStdout(t, func() {
		if err := cmd.Run(context.Background(), nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})

	// Verify the output file was created.
	if _, err := os.Stat(filepath.Join(dir, ".github", "carl", "repo-map.json")); err != nil {
		t.Fatalf("repo-map.json not created by Run: %v", err)
	}
}

// RunInDir prints a summary to stdout.
func TestMap_PrintsSummary(t *testing.T) {
	dir := t.TempDir()
	cmd := repomap.New()

	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if !strings.Contains(output, "Repo map updated") {
		t.Errorf("expected 'Repo map updated' in output; got: %q", output)
	}
	if !strings.Contains(output, repomap.OutputFile) {
		t.Errorf("expected output file path in summary; got: %q", output)
	}
}
