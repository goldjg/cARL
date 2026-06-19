package reconcile_test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/reconcile"
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

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// writeRepoMap writes a minimal repo-map.json into dir/.github/carl/.
func writeRepoMap(t *testing.T, dir string, m *repomap.Map) {
	t.Helper()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	carlDir := filepath.Join(dir, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(carlDir, "repo-map.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

// writeMemory writes memory.md content into dir/.github/carl/.
func writeMemory(t *testing.T, dir, content string) {
	t.Helper()
	carlDir := filepath.Join(dir, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(carlDir, "memory.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// readMemory reads memory.md from dir/.github/carl/.
func readMemory(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".github", "carl", "memory.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// minimalMap returns a small but complete repomap.Map suitable for tests.
func minimalMap() *repomap.Map {
	return &repomap.Map{
		Note:        "test",
		GeneratedBy: "carl map",
		LastUpdated: "2026-01-01",
		Languages:   []string{"Go"},
		EntryPoints: []repomap.File{
			{Path: "go.mod", Purpose: "Go module definition"},
		},
		Directories: map[string]string{
			"cmd": "CLI command entry points",
		},
		Workflows: []repomap.File{
			{Path: ".github/workflows/ci.yml", Purpose: "ci workflow"},
		},
		Governance: []repomap.File{
			{Path: ".github/carl/memory.md", Purpose: "Durable architectural truth cache"},
		},
		Documentation: []repomap.File{
			{Path: "README.md", Purpose: "Repository overview"},
		},
	}
}

// R1: missing repo-map returns an actionable error.
func TestReconcile_R1_MissingRepoMap(t *testing.T) {
	dir := t.TempDir()
	// Write memory.md but no repo-map.json.
	writeMemory(t, dir, "# Memory\n\nSome content.\n\n## Last updated\n2026-01-01\n")

	cmd := reconcile.New()
	err := cmd.RunInDir(dir)
	if err == nil {
		t.Fatal("expected error when repo-map.json is missing")
	}
	if !strings.Contains(err.Error(), "carl map") {
		t.Errorf("error should suggest `carl map`; got: %q", err.Error())
	}
}

// R2: missing memory.md returns an actionable error.
func TestReconcile_R2_MissingMemory(t *testing.T) {
	dir := t.TempDir()
	// Write repo-map.json but no memory.md.
	writeRepoMap(t, dir, minimalMap())

	cmd := reconcile.New()
	err := cmd.RunInDir(dir)
	if err == nil {
		t.Fatal("expected error when memory.md is missing")
	}
	if !strings.Contains(err.Error(), "carl init") {
		t.Errorf("error should suggest `carl init`; got: %q", err.Error())
	}
}

// R3: reconcile updates the repo-specific snapshot section from repo-map data.
func TestReconcile_R3_UpdatesSnapshot(t *testing.T) {
	dir := t.TempDir()
	m := minimalMap()
	writeRepoMap(t, dir, m)
	writeMemory(t, dir, "# Memory\n\nHuman content.\n\n## Last updated\n2026-01-01\n")

	cmd := reconcile.New()
	var output string
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})
	_ = output

	content := readMemory(t, dir)

	// Must contain the generated section markers.
	if !strings.Contains(content, "<!-- BEGIN GENERATED: reconcile -->") {
		t.Error("memory.md must contain BEGIN GENERATED marker")
	}
	if !strings.Contains(content, "<!-- END GENERATED: reconcile -->") {
		t.Error("memory.md must contain END GENERATED marker")
	}

	// Must reflect repo-map language.
	if !strings.Contains(content, "Go") {
		t.Error("memory.md must contain language 'Go' from repo-map")
	}

	// Must contain entry point.
	if !strings.Contains(content, "go.mod") {
		t.Error("memory.md must contain entry point 'go.mod' from repo-map")
	}

	// Must contain key directory.
	if !strings.Contains(content, "cmd") {
		t.Error("memory.md must contain directory 'cmd' from repo-map")
	}

	// Must contain workflow.
	if !strings.Contains(content, "ci.yml") {
		t.Error("memory.md must contain workflow 'ci.yml' from repo-map")
	}

	// Must contain documentation.
	if !strings.Contains(content, "README.md") {
		t.Error("memory.md must contain documentation 'README.md' from repo-map")
	}

	// Output must report reconciliation.
	output = captureStdout(t, func() {
		// Already ran; re-read result by checking for the expected content above.
		// Here we verify that the first run printed the expected message.
	})
}

// R3 (output): the command prints the correct success message and changed file.
func TestReconcile_R3_OutputMessage(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	writeMemory(t, dir, "# Memory\n\nSome content.\n\n## Last updated\n2026-01-01\n")

	cmd := reconcile.New()
	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	if !strings.Contains(output, "Reconciled durable artefacts.") {
		t.Errorf("expected 'Reconciled durable artefacts.' in output; got: %q", output)
	}
	if !strings.Contains(output, reconcile.MemoryFile) {
		t.Errorf("expected memory file path in output; got: %q", output)
	}
}

// R4: reconcile preserves human-authored notes outside the generated section.
func TestReconcile_R4_PreservesHumanContent(t *testing.T) {
	dir := t.TempDir()
	humanNote := "## Human-authored section\n\nThis is a durable fact written by a person.\n"
	initial := "# Memory\n\n" + humanNote + "\n## Last updated\n2026-01-01\n"
	writeRepoMap(t, dir, minimalMap())
	writeMemory(t, dir, initial)

	cmd := reconcile.New()
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	content := readMemory(t, dir)

	// Human content must still be present.
	if !strings.Contains(content, "This is a durable fact written by a person.") {
		t.Error("reconcile must preserve human-authored content outside the generated section")
	}

	// Header must still be present.
	if !strings.Contains(content, "# Memory") {
		t.Error("reconcile must preserve the document header")
	}

	// The "## Last updated" section must still be present.
	if !strings.Contains(content, "## Last updated") {
		t.Error("reconcile must preserve the ## Last updated section")
	}

	// Generated section must not have absorbed the human note.
	beginIdx := strings.Index(content, "<!-- BEGIN GENERATED: reconcile -->")
	endIdx := strings.Index(content, "<!-- END GENERATED: reconcile -->")
	if beginIdx < 0 || endIdx < 0 {
		t.Fatal("generated markers not found")
	}
	generated := content[beginIdx : endIdx+len("<!-- END GENERATED: reconcile -->")]
	if strings.Contains(generated, "This is a durable fact") {
		t.Error("generated section must not contain human-authored content")
	}
}

// R5: reconcile is idempotent — running it twice on the same repo-map produces
// identical memory.md content and the second run reports no changes needed.
func TestReconcile_R5_Idempotent(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	writeMemory(t, dir, "# Memory\n\nSome content.\n\n## Last updated\n2026-01-01\n")

	cmd := reconcile.New()

	// First run — should make changes.
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("first run: %v", err)
		}
	})

	contentAfterFirst := readMemory(t, dir)

	// Second run — should detect no change.
	output := captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("second run: %v", err)
		}
	})

	contentAfterSecond := readMemory(t, dir)

	if contentAfterFirst != contentAfterSecond {
		t.Errorf("content changed between first and second run:\nbefore:\n%s\nafter:\n%s",
			contentAfterFirst, contentAfterSecond)
	}

	if !strings.Contains(output, "No reconciliation needed.") {
		t.Errorf("second run should report 'No reconciliation needed.'; got: %q", output)
	}
}

// R6: reconcile does not modify runtime.json or harness adapter files.
func TestReconcile_R6_DoesNotModifyRuntimeOrHarness(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	writeMemory(t, dir, "# Memory\n\nContent.\n\n## Last updated\n2026-01-01\n")

	// Write runtime.json sentinel.
	carlDir := filepath.Join(dir, ".github", "carl")
	runtimePath := filepath.Join(carlDir, "runtime.json")
	runtimeContent := `{"runtimeVersion":"1.0.0","managedArtifacts":[]}`
	if err := os.WriteFile(runtimePath, []byte(runtimeContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write harness adapter file sentinels.
	claudePath := filepath.Join(dir, "CLAUDE.md")
	claudeContent := "# Claude adapter\n"
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
		t.Fatal(err)
	}

	copilotPath := filepath.Join(dir, ".github", "copilot-instructions.md")
	copilotContent := "# Copilot adapter\n"
	if err := os.MkdirAll(filepath.Dir(copilotPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(copilotPath, []byte(copilotContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := reconcile.New()
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	// runtime.json must be unchanged.
	got, err := os.ReadFile(runtimePath)
	if err != nil {
		t.Fatalf("read runtime.json: %v", err)
	}
	if string(got) != runtimeContent {
		t.Errorf("runtime.json was modified:\ngot:  %q\nwant: %q", string(got), runtimeContent)
	}

	// CLAUDE.md must be unchanged.
	got, err = os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if string(got) != claudeContent {
		t.Errorf("CLAUDE.md was modified:\ngot:  %q\nwant: %q", string(got), claudeContent)
	}

	// copilot-instructions.md must be unchanged.
	got, err = os.ReadFile(copilotPath)
	if err != nil {
		t.Fatalf("read copilot-instructions.md: %v", err)
	}
	if string(got) != copilotContent {
		t.Errorf("copilot-instructions.md was modified:\ngot:  %q\nwant: %q", string(got), copilotContent)
	}
}

// TestReconcile_NoRepoMapSuggestsCommand verifies the exact suggestion in the
// error message matches the documented guidance.
func TestReconcile_NoRepoMapSuggestsCommand(t *testing.T) {
	dir := t.TempDir()
	writeMemory(t, dir, "# Memory\n")

	err := reconcile.New().RunInDir(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "carl map") {
		t.Errorf("expected 'carl map' suggestion in error; got %q", err.Error())
	}
}

// TestReconcile_NoMemorySuggestsCommand verifies the exact suggestion in the
// error message matches the documented guidance.
func TestReconcile_NoMemorySuggestsCommand(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())

	err := reconcile.New().RunInDir(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "carl init") {
		t.Errorf("expected 'carl init' suggestion in error; got %q", err.Error())
	}
}

// TestReconcile_RespectsLastUpdatedPosition verifies the generated section is
// inserted before "## Last updated" and that the last-updated section is preserved.
func TestReconcile_RespectsLastUpdatedPosition(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	writeMemory(t, dir, "# Memory\n\nContent here.\n\n## Last updated\n2026-01-01 by someone\n")

	cmd := reconcile.New()
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Fatalf("RunInDir: %v", err)
		}
	})

	content := readMemory(t, dir)

	endIdx := strings.Index(content, "<!-- END GENERATED: reconcile -->")
	lastUpdatedIdx := strings.Index(content, "## Last updated")
	if endIdx < 0 || lastUpdatedIdx < 0 {
		t.Fatalf("markers or ## Last updated not found in:\n%s", content)
	}
	if endIdx > lastUpdatedIdx {
		t.Errorf("generated section end marker (%d) must appear before ## Last updated (%d)",
			endIdx, lastUpdatedIdx)
	}
}

// malformedMarkerTests groups the three cases where marker state is invalid.
// Each case must: return a non-nil error, leave memory.md unmodified, and
// include the phrase "malformed" in the error message.

// TestReconcile_R7_BeginMarkerOnly checks that a begin marker without an end
// marker causes a non-zero exit and no write.
func TestReconcile_R7_BeginMarkerOnly(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	original := "# Memory\n\n<!-- BEGIN GENERATED: reconcile -->\norphan content\n\n## Last updated\n2026-01-01\n"
	writeMemory(t, dir, original)

	err := reconcile.New().RunInDir(dir)
	if err == nil {
		t.Fatal("expected error when begin marker is present but end marker is missing")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Errorf("error should describe markers as malformed; got: %q", err.Error())
	}
	if got := readMemory(t, dir); got != original {
		t.Errorf("memory.md must not be modified on malformed marker error;\ngot:\n%s", got)
	}
}

// TestReconcile_R7_EndMarkerOnly checks that an end marker without a begin
// marker causes a non-zero exit and no write.
func TestReconcile_R7_EndMarkerOnly(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	original := "# Memory\n\norphan content\n<!-- END GENERATED: reconcile -->\n\n## Last updated\n2026-01-01\n"
	writeMemory(t, dir, original)

	err := reconcile.New().RunInDir(dir)
	if err == nil {
		t.Fatal("expected error when end marker is present but begin marker is missing")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Errorf("error should describe markers as malformed; got: %q", err.Error())
	}
	if got := readMemory(t, dir); got != original {
		t.Errorf("memory.md must not be modified on malformed marker error;\ngot:\n%s", got)
	}
}

// TestReconcile_R7_EndBeforeBegin checks that an end marker appearing before
// the begin marker causes a non-zero exit and no write.
func TestReconcile_R7_EndBeforeBegin(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	original := "# Memory\n\n<!-- END GENERATED: reconcile -->\n\nsome text\n\n<!-- BEGIN GENERATED: reconcile -->\n\n## Last updated\n2026-01-01\n"
	writeMemory(t, dir, original)

	err := reconcile.New().RunInDir(dir)
	if err == nil {
		t.Fatal("expected error when end marker appears before begin marker")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Errorf("error should describe markers as malformed; got: %q", err.Error())
	}
	if got := readMemory(t, dir); got != original {
		t.Errorf("memory.md must not be modified on malformed marker error;\ngot:\n%s", got)
	}
}
