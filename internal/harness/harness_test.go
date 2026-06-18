package harness_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/harness"
)

// testArts is a minimal Artifacts implementation for testing.
// It returns a fixed canonical content string for any path it recognises.
type testArts struct {
	files map[string][]byte
}

func (a *testArts) Open(path string) ([]byte, error) {
	if a.files != nil {
		if content, ok := a.files[path]; ok {
			return content, nil
		}
	}
	return nil, fmt.Errorf("testArts: file not found: %s", path)
}

// newTestArts returns a testArts pre-loaded with the canonical copilot
// instructions placeholder used by sync tests.
func newTestArts() *testArts {
	return &testArts{
		files: map[string][]byte{
			".github/copilot-instructions.md": []byte("# cARL Governance Instructions\n\nTest canonical content.\n"),
		},
	}
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

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// createFile creates a file at path with empty content,
// creating parent directories as needed.
func createFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
}

// Contract assertion 1: carl harness list lists all known adapters.
func TestHarness_List_ShowsAllAdapters(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	output := captureStdout(t, func() {
		if err := cmd.RunListInDir(dir); err != nil {
			t.Fatalf("RunListInDir: %v", err)
		}
	})

	for _, want := range []string{
		"copilot", "GitHub Copilot",
		"claude", "Claude Code",
		"codex", "Codex",
		"cursor", "Cursor",
		"antigravity", "Antigravity",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("list output missing %q\nfull output:\n%s", want, output)
		}
	}
}

// Contract assertion 2: copilot is "supported"; all known adapters are supported.
// Also verifies count summary line.
func TestHarness_List_SupportStatus(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	output := captureStdout(t, func() {
		if err := cmd.RunListInDir(dir); err != nil {
			t.Fatalf("RunListInDir: %v", err)
		}
	})

	if !strings.Contains(output, "supported") {
		t.Errorf("expected 'supported' in list output; got:\n%s", output)
	}

	// Derive expected counts from the registry itself for robustness.
	adapters := harness.Adapters()
	total := len(adapters)
	supported := 0
	for _, a := range adapters {
		if a.Support == "supported" {
			supported++
		}
	}
	wantLine := fmt.Sprintf("%d of %d adapter(s) supported.", supported, total)
	if !strings.Contains(output, wantLine) {
		t.Errorf("expected %q in list output; got:\n%s", wantLine, output)
	}
}

// Contract assertion 3: harness status shows "active" for copilot when
// the detection file is present.
func TestHarness_Status_CopilotDetected(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, ".github", "copilot-instructions.md"))

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "active") {
		t.Errorf("expected 'active' for copilot with detection file present; got:\n%s", output)
	}

	adapters := harness.Adapters()
	wantLine := fmt.Sprintf("1 of %d harness(es) active.", len(adapters))
	if !strings.Contains(output, wantLine) {
		t.Errorf("expected %q in status output; got:\n%s", wantLine, output)
	}
}

// Contract assertion 4: harness status shows "not active" for copilot when
// the detection file is absent.
func TestHarness_Status_CopilotNotDetected(t *testing.T) {
	dir := t.TempDir()
	// Do NOT create the detection file.

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "not active") {
		t.Errorf("expected 'not active' for copilot without detection file; got:\n%s", output)
	}

	adapters := harness.Adapters()
	wantLine := fmt.Sprintf("0 of %d harness(es) active.", len(adapters))
	if !strings.Contains(output, wantLine) {
		t.Errorf("expected %q in status output; got:\n%s", wantLine, output)
	}
}

// Contract assertion 5: carl harness with no args prints usage and returns nil.
func TestHarness_NoArgs_PrintsUsage(t *testing.T) {
	cmd := harness.New(newTestArts())
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Run(context.Background(), nil)
	})

	if runErr != nil {
		t.Fatalf("Run (no args): unexpected error: %v", runErr)
	}
	for _, want := range []string{"Usage:", "list", "status"} {
		if !strings.Contains(output, want) {
			t.Errorf("usage output missing %q\nfull output:\n%s", want, output)
		}
	}
}

// Contract assertion 5 (variant): --help flag prints usage and returns nil.
func TestHarness_HelpFlag_PrintsUsage(t *testing.T) {
	cmd := harness.New(newTestArts())
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Run(context.Background(), []string{"--help"})
	})

	if runErr != nil {
		t.Fatalf("Run (--help): unexpected error: %v", runErr)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("--help output missing 'Usage:'; got:\n%s", output)
	}
}

// Contract assertion 6: unknown subcommand returns a non-nil error containing
// "unknown subcommand".
func TestHarness_UnknownSubcommand_ReturnsError(t *testing.T) {
	cmd := harness.New(newTestArts())
	err := cmd.Run(context.Background(), []string{"frobble"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand; got nil")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("expected 'unknown subcommand' in error; got: %v", err)
	}
}

// Contract assertion 7: list and status are read-only and always return nil.
func TestHarness_ListAndStatus_AlwaysReturnNil(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	_ = captureStdout(t, func() {
		if err := cmd.RunListInDir(dir); err != nil {
			t.Errorf("RunListInDir: unexpected error: %v", err)
		}
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Errorf("RunStatusInDir: unexpected error: %v", err)
		}
	})
}

// TestHarness_Run_List exercises the Run method with "list" subcommand via cwd.
func TestHarness_Run_List(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := harness.New(newTestArts())
	_ = captureStdout(t, func() {
		if err := cmd.Run(context.Background(), []string{"list"}); err != nil {
			t.Fatalf("Run list: %v", err)
		}
	})
}

// TestHarness_Run_Status exercises the Run method with "status" subcommand via cwd.
func TestHarness_Run_Status(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := harness.New(newTestArts())
	_ = captureStdout(t, func() {
		if err := cmd.Run(context.Background(), []string{"status"}); err != nil {
			t.Fatalf("Run status: %v", err)
		}
	})
}

// TestHarness_Adapters_CopilotIsSupported verifies the exported registry.
func TestHarness_Adapters_CopilotIsSupported(t *testing.T) {
	adapters := harness.Adapters()

	var copilot *harness.Adapter
	for i := range adapters {
		if adapters[i].ID == "copilot" {
			copilot = &adapters[i]
			break
		}
	}
	if copilot == nil {
		t.Fatal("copilot adapter not found in registry")
	}
	if copilot.Support != "supported" {
		t.Errorf("copilot.Support = %q; want 'supported'", copilot.Support)
	}
	if copilot.DetectionFile == "" {
		t.Error("copilot.DetectionFile must not be empty for a supported adapter")
	}
}

// TestHarness_Adapters_SupportedHaveDetectionFile verifies all supported adapters
// have a non-empty DetectionFile defined.
func TestHarness_Adapters_SupportedHaveDetectionFile(t *testing.T) {
	for _, a := range harness.Adapters() {
		if a.Support == "supported" && a.DetectionFile == "" {
			t.Errorf("supported adapter %q has no DetectionFile; all supported adapters must define one", a.ID)
		}
	}
}

// TestHarness_Adapters_PlannedHaveNoDetectionFile verifies planned adapters
// have no detection file defined (they cannot be detected).
func TestHarness_Adapters_PlannedHaveNoDetectionFile(t *testing.T) {
	for _, a := range harness.Adapters() {
		if a.Support == "planned" && a.DetectionFile != "" {
			t.Errorf("planned adapter %q has a DetectionFile %q; planned adapters should have empty DetectionFile",
				a.ID, a.DetectionFile)
		}
	}
}

// TestHarness_Name verifies command metadata.
func TestHarness_Name(t *testing.T) {
	cmd := harness.New(newTestArts())
	if cmd.Name() != "harness" {
		t.Errorf("Name() = %q; want 'harness'", cmd.Name())
	}
	if cmd.Synopsis() == "" {
		t.Error("Synopsis() must not be empty")
	}
}

// TestHarness_Status_ClaudeDetected verifies Claude Code detection via CLAUDE.md.
func TestHarness_Status_ClaudeDetected(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "CLAUDE.md"))

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "claude") {
		t.Errorf("expected 'claude' in status output; got:\n%s", output)
	}
	// The active-count line must reflect at least 1 active harness.
	if !strings.Contains(output, "active") {
		t.Errorf("expected 'active' for claude with CLAUDE.md present; got:\n%s", output)
	}
}

// TestHarness_Status_CodexDetected verifies Codex detection via AGENTS.md.
func TestHarness_Status_CodexDetected(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "AGENTS.md"))

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "codex") {
		t.Errorf("expected 'codex' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "active") {
		t.Errorf("expected 'active' for codex with AGENTS.md present; got:\n%s", output)
	}
}

// TestHarness_Status_CursorDetected verifies Cursor detection via .cursorrules.
func TestHarness_Status_CursorDetected(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, ".cursorrules"))

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "cursor") {
		t.Errorf("expected 'cursor' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "active") {
		t.Errorf("expected 'active' for cursor with .cursorrules present; got:\n%s", output)
	}
}

// TestHarness_Status_AntigravityDetected verifies Antigravity detection via ANTIGRAVITY.md.
func TestHarness_Status_AntigravityDetected(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "ANTIGRAVITY.md"))

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "antigravity") {
		t.Errorf("expected 'antigravity' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "active") {
		t.Errorf("expected 'active' for antigravity with ANTIGRAVITY.md present; got:\n%s", output)
	}
}

// TestHarness_Status_AllDetected verifies all 5 adapters report active when
// all detection files are present.
func TestHarness_Status_AllDetected(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, ".github", "copilot-instructions.md"))
	createFile(t, filepath.Join(dir, "CLAUDE.md"))
	createFile(t, filepath.Join(dir, "AGENTS.md"))
	createFile(t, filepath.Join(dir, ".cursorrules"))
	createFile(t, filepath.Join(dir, "ANTIGRAVITY.md"))

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	adapters := harness.Adapters()
	wantLine := fmt.Sprintf("%d of %d harness(es) active.", len(adapters), len(adapters))
	if !strings.Contains(output, wantLine) {
		t.Errorf("expected %q in status output; got:\n%s", wantLine, output)
	}
}

// TestHarness_Adapters_DetectionFiles verifies the expected detection files
// for each known adapter.
func TestHarness_Adapters_DetectionFiles(t *testing.T) {
	wantDetectionFiles := map[string]string{
		"copilot":     ".github/copilot-instructions.md",
		"claude":      "CLAUDE.md",
		"codex":       "AGENTS.md",
		"cursor":      ".cursorrules",
		"antigravity": "ANTIGRAVITY.md",
	}

	for _, a := range harness.Adapters() {
		want, ok := wantDetectionFiles[a.ID]
		if !ok {
			t.Errorf("adapter %q has no expected detection file entry in test table; update the test", a.ID)
			continue
		}
		if a.DetectionFile != want {
			t.Errorf("adapter %q DetectionFile = %q; want %q", a.ID, a.DetectionFile, want)
		}
	}
}

// --- sync contract assertions ---

// Contract assertion S1: sync with no harness IDs writes adapter files for
// all supported harnesses.
func TestHarness_Sync_AllHarnesses(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	var syncErr error
	output := captureStdout(t, func() {
		syncErr = cmd.RunSyncInDir(dir, nil)
	})
	if syncErr != nil {
		t.Fatalf("RunSyncInDir: %v", syncErr)
	}

	wantFiles := map[string]string{
		"copilot":     filepath.Join(dir, ".github", "copilot-instructions.md"),
		"claude":      filepath.Join(dir, "CLAUDE.md"),
		"codex":       filepath.Join(dir, "AGENTS.md"),
		"cursor":      filepath.Join(dir, ".cursorrules"),
		"antigravity": filepath.Join(dir, "ANTIGRAVITY.md"),
	}

	for id, path := range wantFiles {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("harness %q: expected adapter file %q to exist after sync; stat: %v", id, path, err)
		}
	}

	// Summary line must report all supported adapters.
	adapters := harness.Adapters()
	supported := 0
	for _, a := range adapters {
		if a.Support == "supported" {
			supported++
		}
	}
	wantSummary := fmt.Sprintf("%d adapter file(s) synced.", supported)
	if !strings.Contains(output, wantSummary) {
		t.Errorf("expected %q in sync output; got:\n%s", wantSummary, output)
	}
}

// Contract assertion S2: sync with a specific harness ID writes only that
// adapter file and leaves others absent.
func TestHarness_Sync_SpecificHarness(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	_ = captureStdout(t, func() {
		if err := cmd.RunSyncInDir(dir, []string{"claude"}); err != nil {
			t.Fatalf("RunSyncInDir: %v", err)
		}
	})

	// CLAUDE.md should exist.
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md should exist after sync claude; stat: %v", err)
	}
	// Copilot adapter should NOT exist (was not requested).
	if _, err := os.Stat(filepath.Join(dir, ".github", "copilot-instructions.md")); err == nil {
		t.Errorf(".github/copilot-instructions.md should not exist when only claude was synced")
	}
}

// Contract assertion S3: sync with an unknown harness ID returns a non-nil
// error containing "unknown harness".
func TestHarness_Sync_UnknownHarness_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	err := cmd.RunSyncInDir(dir, []string{"doesnotexist"})
	if err == nil {
		t.Fatal("expected error for unknown harness; got nil")
	}
	if !strings.Contains(err.Error(), "unknown harness") {
		t.Errorf("expected 'unknown harness' in error; got: %v", err)
	}
}

// Contract assertion S4: sync writes the canonical content from embedded
// artefacts to the adapter file.
func TestHarness_Sync_WritesCanonicalContent(t *testing.T) {
	const wantContent = "# Test canonical content for sync\n"
	arts := &testArts{
		files: map[string][]byte{
			".github/copilot-instructions.md": []byte(wantContent),
		},
	}
	dir := t.TempDir()
	cmd := harness.New(arts)

	_ = captureStdout(t, func() {
		if err := cmd.RunSyncInDir(dir, []string{"copilot"}); err != nil {
			t.Fatalf("RunSyncInDir: %v", err)
		}
	})

	got, err := os.ReadFile(filepath.Join(dir, ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatalf("read copilot adapter file: %v", err)
	}
	if string(got) != wantContent {
		t.Errorf("adapter file content = %q; want %q", string(got), wantContent)
	}
}

// Contract assertion S5: sync is idempotent — running it twice produces the
// same file state and no error.
func TestHarness_Sync_Idempotent(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	run := func() {
		_ = captureStdout(t, func() {
			if err := cmd.RunSyncInDir(dir, nil); err != nil {
				t.Fatalf("RunSyncInDir: %v", err)
			}
		})
	}

	run()

	// Capture state after first run.
	firstContent, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md after first sync: %v", err)
	}

	run()

	// Content must be identical after second run.
	secondContent, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md after second sync: %v", err)
	}
	if string(firstContent) != string(secondContent) {
		t.Errorf("sync is not idempotent: content changed between runs")
	}
}

// Contract assertion S6: the sync subcommand is reachable via Run.
func TestHarness_Run_Sync(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := harness.New(newTestArts())
	_ = captureStdout(t, func() {
		if err := cmd.Run(context.Background(), []string{"sync"}); err != nil {
			t.Fatalf("Run sync: %v", err)
		}
	})

	// At least one adapter file should have been created.
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md should exist after 'carl harness sync': %v", err)
	}
}

// Contract assertion S7: usage output includes "sync".
func TestHarness_NoArgs_UsageIncludesSync(t *testing.T) {
	cmd := harness.New(newTestArts())
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Run(context.Background(), nil)
	})
	if runErr != nil {
		t.Fatalf("Run (no args): unexpected error: %v", runErr)
	}
	if !strings.Contains(output, "sync") {
		t.Errorf("usage output missing 'sync'; got:\n%s", output)
	}
}

// Contract assertion S8: supported adapters must have a non-empty SourceFile.
func TestHarness_Adapters_SupportedHaveSourceFile(t *testing.T) {
	for _, a := range harness.Adapters() {
		if a.Support == "supported" && a.SourceFile == "" {
			t.Errorf("supported adapter %q has no SourceFile; all supported adapters must define one for sync", a.ID)
		}
	}
}
