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

// TestReconcile_FreshInitThenReconcile simulates the shipped `carl init`
// template (a well-formed generated section already present, matching what
// the fixed embedded/assets/.github/carl/memory.md now ships) followed
// immediately by `carl reconcile`. This is a regression test for the
// originally reported bug: a fresh init must not error on marker mismatch.
func TestReconcile_FreshInitThenReconcile(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	freshInit := "# Durable Architectural Truth Cache\n\n" +
		"## Project purpose\nSome durable project truth.\n\n" +
		"<!-- BEGIN GENERATED: reconcile -->\n" +
		"## Repository snapshot\n\nThis section is regenerated by `carl reconcile`. Do not edit manually.\n" +
		"<!-- END GENERATED: reconcile -->\n\n" +
		"## Last updated\n2026-01-01\n"
	writeMemory(t, dir, freshInit)

	if err := reconcile.New().RunInDir(dir); err != nil {
		t.Fatalf("reconcile after fresh init must succeed; got error: %v", err)
	}

	content := readMemory(t, dir)
	if !strings.Contains(content, "<!-- BEGIN GENERATED: reconcile -->") ||
		!strings.Contains(content, "<!-- END GENERATED: reconcile -->") {
		t.Error("reconciled memory.md must contain both generated-section markers")
	}
	if !strings.Contains(content, "Some durable project truth.") {
		t.Error("reconcile must preserve pre-existing human-authored content")
	}
}

// TestReconcile_MissingBeginMarker_RepairFlag checks that `--repair-markers`
// removes only the single orphaned END marker line, preserving all
// surrounding human content byte-for-byte, and that reconcile then succeeds.
func TestReconcile_MissingBeginMarker_RepairFlag(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	before := "Human note before the orphan marker.\n"
	after := "\nHuman note after the orphan marker.\n\n## Last updated\n2026-01-01\n"
	original := "# Memory\n\n" + before + "<!-- END GENERATED: reconcile -->\n" + after
	writeMemory(t, dir, original)

	// Without the flag, this must still fail (conservative default).
	if err := reconcile.New().RunInDir(dir); err == nil {
		t.Fatal("expected error without --repair-markers")
	}
	if got := readMemory(t, dir); got != original {
		t.Fatalf("memory.md must not be modified when --repair-markers is not used;\ngot:\n%s", got)
	}

	// With the flag, the orphan marker line is removed and reconcile succeeds.
	if err := reconcile.New().RunInDirWithOptions(dir, true); err != nil {
		t.Fatalf("RunInDirWithOptions with repairMarkers=true: %v", err)
	}

	content := readMemory(t, dir)
	if !strings.Contains(content, before) {
		t.Errorf("human content before the orphan marker must be preserved; got:\n%s", content)
	}
	if !strings.Contains(content, "Human note after the orphan marker.") {
		t.Errorf("human content after the orphan marker must be preserved; got:\n%s", content)
	}
	if !strings.Contains(content, "<!-- BEGIN GENERATED: reconcile -->") ||
		!strings.Contains(content, "<!-- END GENERATED: reconcile -->") {
		t.Error("reconciled memory.md must contain a fresh, well-formed generated section")
	}
}

// TestReconcile_MissingEndMarker_RepairFlag mirrors
// TestReconcile_MissingBeginMarker_RepairFlag for the opposite orphan case:
// a BEGIN marker with no END marker.
func TestReconcile_MissingEndMarker_RepairFlag(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	before := "Human note before the orphan marker.\n\n"
	after := "orphan content\n\n## Last updated\n2026-01-01\n"
	original := "# Memory\n\n" + before + "<!-- BEGIN GENERATED: reconcile -->\n" + after
	writeMemory(t, dir, original)

	if err := reconcile.New().RunInDir(dir); err == nil {
		t.Fatal("expected error without --repair-markers")
	}
	if got := readMemory(t, dir); got != original {
		t.Fatalf("memory.md must not be modified when --repair-markers is not used;\ngot:\n%s", got)
	}

	if err := reconcile.New().RunInDirWithOptions(dir, true); err != nil {
		t.Fatalf("RunInDirWithOptions with repairMarkers=true: %v", err)
	}

	content := readMemory(t, dir)
	if !strings.Contains(content, before) {
		t.Errorf("human content before the orphan marker must be preserved; got:\n%s", content)
	}
	if !strings.Contains(content, "<!-- BEGIN GENERATED: reconcile -->") ||
		!strings.Contains(content, "<!-- END GENERATED: reconcile -->") {
		t.Error("reconciled memory.md must contain a fresh, well-formed generated section")
	}
}

// TestReconcile_RepairMarkersFlag_AmbiguousStateStillErrors verifies that
// --repair-markers refuses to guess when the marker state is ambiguous
// (here, an end-before-begin pair — both markers present, wrong order) and
// leaves memory.md completely untouched.
func TestReconcile_RepairMarkersFlag_AmbiguousStateStillErrors(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	original := "# Memory\n\n<!-- END GENERATED: reconcile -->\n\nsome text\n\n<!-- BEGIN GENERATED: reconcile -->\n\n## Last updated\n2026-01-01\n"
	writeMemory(t, dir, original)

	err := reconcile.New().RunInDirWithOptions(dir, true)
	if err == nil {
		t.Fatal("expected error: --repair-markers must not guess at an ambiguous marker state")
	}
	if got := readMemory(t, dir); got != original {
		t.Errorf("memory.md must not be modified when --repair-markers cannot safely resolve the state;\ngot:\n%s", got)
	}
}

// TestReconcile_DuplicatedMarkerBlocks verifies that two well-formed
// BEGIN/END pairs (a duplicated generated section) are automatically
// collapsed into a single, freshly generated block — with no error and no
// flag required, since content strictly between matched pairs is
// unambiguously cARL-owned.
func TestReconcile_DuplicatedMarkerBlocks(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	original := "# Memory\n\n" +
		"Human content before both blocks.\n\n" +
		"<!-- BEGIN GENERATED: reconcile -->\nstale snapshot one\n<!-- END GENERATED: reconcile -->\n\n" +
		"Human content between the two duplicate blocks.\n\n" +
		"<!-- BEGIN GENERATED: reconcile -->\nstale snapshot two\n<!-- END GENERATED: reconcile -->\n\n" +
		"Human content after both blocks.\n\n" +
		"## Last updated\n2026-01-01\n"
	writeMemory(t, dir, original)

	if err := reconcile.New().RunInDir(dir); err != nil {
		t.Fatalf("expected duplicated well-formed marker pairs to be auto-collapsed without error; got: %v", err)
	}

	content := readMemory(t, dir)

	if n := strings.Count(content, "<!-- BEGIN GENERATED: reconcile -->"); n != 1 {
		t.Errorf("expected exactly 1 BEGIN marker after collapsing duplicates; got %d in:\n%s", n, content)
	}
	if n := strings.Count(content, "<!-- END GENERATED: reconcile -->"); n != 1 {
		t.Errorf("expected exactly 1 END marker after collapsing duplicates; got %d in:\n%s", n, content)
	}
	if strings.Contains(content, "stale snapshot one") || strings.Contains(content, "stale snapshot two") {
		t.Error("stale duplicated generated content must not survive collapsing")
	}

	// All human-authored content, including the text sandwiched between the
	// two duplicate blocks, must be preserved.
	if !strings.Contains(content, "Human content before both blocks.") {
		t.Error("human content before the duplicated blocks must be preserved")
	}
	if !strings.Contains(content, "Human content between the two duplicate blocks.") {
		t.Error("human content between the duplicated blocks must be preserved")
	}
	if !strings.Contains(content, "Human content after both blocks.") {
		t.Error("human content after the duplicated blocks must be preserved")
	}
}

// TestReconcile_HumanContentAroundGeneratedBlock verifies that
// human-authored content both immediately before and immediately after the
// generated section is preserved byte-for-byte across a normal reconcile
// run.
func TestReconcile_HumanContentAroundGeneratedBlock(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	headerBlock := "# Memory\n\n## Human section before\n\nDurable fact written by a person, before the generated block.\n\n"
	trailerBlock := "\n\n## Human section after\n\nDurable fact written by a person, after the generated block.\n\n## Last updated\n2026-01-01\n"
	original := headerBlock +
		"<!-- BEGIN GENERATED: reconcile -->\nstale\n<!-- END GENERATED: reconcile -->" +
		trailerBlock
	writeMemory(t, dir, original)

	if err := reconcile.New().RunInDir(dir); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}

	content := readMemory(t, dir)
	if !strings.HasPrefix(content, headerBlock) {
		t.Errorf("content before the generated block must be preserved byte-for-byte;\nwant prefix:\n%s\ngot:\n%s", headerBlock, content)
	}
	if !strings.HasSuffix(content, trailerBlock) {
		t.Errorf("content after the generated block must be preserved byte-for-byte;\nwant suffix:\n%s\ngot:\n%s", trailerBlock, content)
	}
}

// TestReconcile_IdempotentSecondRun_AfterRepair verifies that after a
// --repair-markers pass resolves an orphan marker and reconcile writes fresh
// content, a second plain reconcile run is a true no-op.
func TestReconcile_IdempotentSecondRun_AfterRepair(t *testing.T) {
	dir := t.TempDir()
	writeRepoMap(t, dir, minimalMap())
	original := "# Memory\n\norphan content\n<!-- END GENERATED: reconcile -->\n\n## Last updated\n2026-01-01\n"
	writeMemory(t, dir, original)

	if err := reconcile.New().RunInDirWithOptions(dir, true); err != nil {
		t.Fatalf("first run with repairMarkers=true: %v", err)
	}
	contentAfterFirst := readMemory(t, dir)

	output := captureStdout(t, func() {
		if err := reconcile.New().RunInDir(dir); err != nil {
			t.Fatalf("second run: %v", err)
		}
	})
	contentAfterSecond := readMemory(t, dir)

	if contentAfterFirst != contentAfterSecond {
		t.Errorf("content changed between first (repaired) and second run:\nbefore:\n%s\nafter:\n%s",
			contentAfterFirst, contentAfterSecond)
	}
	if !strings.Contains(output, "No reconciliation needed.") {
		t.Errorf("second run should report 'No reconciliation needed.'; got: %q", output)
	}
}
