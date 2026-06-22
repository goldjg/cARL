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

// newTestArts returns a testArts pre-loaded with the shared loader and
// per-harness shim content used by sync tests.
func newTestArts() *testArts {
	return &testArts{
		files: map[string][]byte{
			".github/copilot-instructions.md": []byte("# cARL Governance Instructions\n\nTest canonical content.\n"),
			"CLAUDE.md":                       []byte("# Claude Code cARL Adapter\n\nTest shim content.\n"),
			"AGENTS.md":                       []byte("# Codex cARL Adapter\n\nTest shim content.\n"),
			".cursor/rules/carl.mdc":           []byte("# Cursor cARL Adapter\n\nTest shim content.\n"),
			".agents/rules/carl.md":            []byte("# Antigravity cARL Adapter\n\nTest shim content.\n"),
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

func syncHarnesses(t *testing.T, dir string, ids ...string) {
	t.Helper()
	cmd := harness.New(newTestArts())
	_ = captureStdout(t, func() {
		if err := cmd.RunSyncInDir(dir, ids); err != nil {
			t.Fatalf("RunSyncInDir: %v", err)
		}
	})
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

// Contract assertion 2: copilot is "production"; adapters have varied support tiers.
// Also verifies count summary line.
func TestHarness_List_SupportStatus(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New(newTestArts())

	output := captureStdout(t, func() {
		if err := cmd.RunListInDir(dir); err != nil {
			t.Fatalf("RunListInDir: %v", err)
		}
	})

	if !strings.Contains(output, "production") {
		t.Errorf("expected 'production' in list output; got:\n%s", output)
	}

	// Derive expected counts from the registry itself for robustness.
	adapters := harness.Adapters()
	total := len(adapters)
	production, experimental, theoretical := 0, 0, 0
	for _, a := range adapters {
		switch a.Support {
		case "production":
			production++
		case "experimental":
			experimental++
		case "theoretical":
			theoretical++
		}
	}
	wantLine := fmt.Sprintf("%d production, %d experimental, %d theoretical (%d total).",
		production, experimental, theoretical, total)
	if !strings.Contains(output, wantLine) {
		t.Errorf("expected %q in list output; got:\n%s", wantLine, output)
	}
}

// Contract assertion H1: a synced adapter is reported as Present + Synced.
func TestHarness_Status_CopilotSynced(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "copilot")

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "copilot") || !strings.Contains(output, "Present") || !strings.Contains(output, "Synced") {
		t.Errorf("expected synced copilot adapter in status output; got:\n%s", output)
	}

	wantLine := "1 active, 4 missing, 0 drifted, 1 healthy."
	if !strings.Contains(output, wantLine) {
		t.Errorf("expected %q in status output; got:\n%s", wantLine, output)
	}
}

// Contract assertion H3: a deleted adapter is reported as Missing.
func TestHarness_Status_CopilotMissing(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "copilot")
	if err := os.Remove(filepath.Join(dir, ".github", "copilot-instructions.md")); err != nil {
		t.Fatalf("remove adapter file: %v", err)
	}

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "copilot") || strings.Count(output, "Missing") < 2 {
		t.Errorf("expected missing copilot adapter in status output; got:\n%s", output)
	}

	wantLine := "0 active, 5 missing, 0 drifted, 0 healthy."
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

// TestHarness_Adapters_CopilotIsProduction verifies the exported registry.
func TestHarness_Adapters_CopilotIsProduction(t *testing.T) {
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
	if copilot.Support != "production" {
		t.Errorf("copilot.Support = %q; want 'production'", copilot.Support)
	}
	if copilot.DetectionFile == "" {
		t.Error("copilot.DetectionFile must not be empty")
	}
}

// TestHarness_Adapters_ClaudeIsExperimental verifies Claude Code support tier.
func TestHarness_Adapters_ClaudeIsExperimental(t *testing.T) {
	adapters := harness.Adapters()
	for _, a := range adapters {
		if a.ID == "claude" {
			if a.Support != "experimental" {
				t.Errorf("claude.Support = %q; want 'experimental'", a.Support)
			}
			return
		}
	}
	t.Fatal("claude adapter not found in registry")
}

// TestHarness_Adapters_TheoreticalAdapters verifies Codex, Cursor, Antigravity support tier.
func TestHarness_Adapters_TheoreticalAdapters(t *testing.T) {
	theoretical := []string{"codex", "cursor", "antigravity"}
	adapters := harness.Adapters()
	index := make(map[string]harness.Adapter, len(adapters))
	for _, a := range adapters {
		index[a.ID] = a
	}
	for _, id := range theoretical {
		a, ok := index[id]
		if !ok {
			t.Errorf("adapter %q not found in registry", id)
			continue
		}
		if a.Support != "theoretical" {
			t.Errorf("%s.Support = %q; want 'theoretical'", id, a.Support)
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

// Contract assertion H1 variant: Claude is reported healthy after sync.
func TestHarness_Status_ClaudeSynced(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "claude")

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "claude") {
		t.Errorf("expected 'claude' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "Present") || !strings.Contains(output, "Synced") {
		t.Errorf("expected claude adapter to be Present + Synced; got:\n%s", output)
	}
}

// Contract assertion H2: a modified adapter is reported as Drifted.
func TestHarness_Status_ClaudeDrifted(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "claude")
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("modified"), 0644); err != nil {
		t.Fatalf("modify adapter file: %v", err)
	}

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "claude") {
		t.Errorf("expected 'claude' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "Present") || !strings.Contains(output, "Drifted") {
		t.Errorf("expected claude adapter to be Present + Drifted; got:\n%s", output)
	}
}

// TestHarness_Status_CodexSynced verifies Codex is reported healthy after sync.
func TestHarness_Status_CodexSynced(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "codex")

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "codex") {
		t.Errorf("expected 'codex' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "Present") || !strings.Contains(output, "Synced") {
		t.Errorf("expected codex adapter to be Present + Synced; got:\n%s", output)
	}
}

// TestHarness_Status_CursorSynced verifies Cursor is reported healthy after sync.
func TestHarness_Status_CursorSynced(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "cursor")

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "cursor") {
		t.Errorf("expected 'cursor' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "Present") || !strings.Contains(output, "Synced") {
		t.Errorf("expected cursor adapter to be Present + Synced; got:\n%s", output)
	}
}

// TestHarness_Status_AntigravitySynced verifies Antigravity is reported healthy after sync.
func TestHarness_Status_AntigravitySynced(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "antigravity")

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	if !strings.Contains(output, "antigravity") {
		t.Errorf("expected 'antigravity' in status output; got:\n%s", output)
	}
	if !strings.Contains(output, "Present") || !strings.Contains(output, "Synced") {
		t.Errorf("expected antigravity adapter to be Present + Synced; got:\n%s", output)
	}
}

// TestHarness_Status_AllSynced verifies all 5 adapters report synced when
// all adapter files are present and canonical.
func TestHarness_Status_AllDetected(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir)

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	wantLine := "5 active, 0 missing, 0 drifted, 5 healthy."
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
		"cursor":      ".cursor/rules/carl.mdc",
		"antigravity": ".agents/rules/carl.md",
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

// Contract assertion S1: sync with no harness IDs writes all unique adapter
// files (shared loader once + each harness-specific shim).
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
		"cursor":      filepath.Join(dir, ".cursor", "rules", "carl.mdc"),
		"antigravity": filepath.Join(dir, ".agents", "rules", "carl.md"),
	}

	for id, path := range wantFiles {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("harness %q: expected adapter file %q to exist after sync; stat: %v", id, path, err)
		}
	}

	// Summary line must report unique file count (shared loader written once).
	adapters := harness.Adapters()
	seen := make(map[string]bool)
	for _, a := range adapters {
		for _, af := range a.Files {
			seen[af.Path] = true
		}
	}
	wantSummary := fmt.Sprintf("%d adapter file(s) synced.", len(seen))
	if !strings.Contains(output, wantSummary) {
		t.Errorf("expected %q in sync output; got:\n%s", wantSummary, output)
	}
}

// Contract assertion S2: sync with a specific harness ID writes the shared
// loader and that harness's shim; unrelated shim files remain absent.
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
	// Shared loader must also exist — syncing any shim includes the shared loader.
	if _, err := os.Stat(filepath.Join(dir, ".github", "copilot-instructions.md")); err != nil {
		t.Errorf(".github/copilot-instructions.md should exist after sync claude (shared loader); stat: %v", err)
	}
	// Unrelated shims should NOT exist.
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Errorf("AGENTS.md should not exist when only claude was synced")
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
// artefacts to the adapter file. This test uses its own testArts fixture
// (distinct from newTestArts) to verify that whatever content the Artifacts
// implementation returns is written verbatim — the content itself is the
// contract, not the placeholder string used by other tests.
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
// same file state and no error across all synced adapter files.
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

	// Capture content of all adapter files after first run.
	adapterPaths := map[string]string{
		"copilot":     filepath.Join(dir, ".github", "copilot-instructions.md"),
		"claude":      filepath.Join(dir, "CLAUDE.md"),
		"codex":       filepath.Join(dir, "AGENTS.md"),
		"cursor":      filepath.Join(dir, ".cursor", "rules", "carl.mdc"),
		"antigravity": filepath.Join(dir, ".agents", "rules", "carl.md"),
	}
	firstContents := make(map[string][]byte, len(adapterPaths))
	for id, path := range adapterPaths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s after first sync: %v", id, err)
		}
		firstContents[id] = content
	}

	run()

	// Content must be identical after the second run.
	for id, path := range adapterPaths {
		second, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s after second sync: %v", id, err)
		}
		if string(firstContents[id]) != string(second) {
			t.Errorf("sync is not idempotent for %s: content changed between runs", id)
		}
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

// Contract assertion S8: all adapters with defined Files must have at least
// one entry with a non-empty SourceFile.
func TestHarness_Adapters_SupportedHaveSourceFile(t *testing.T) {
	for _, a := range harness.Adapters() {
		if len(a.Files) == 0 {
			continue
		}
		for i, af := range a.Files {
			if af.SourceFile == "" {
				t.Errorf("adapter %q Files[%d].SourceFile is empty; all adapter files must define a source", a.ID, i)
			}
			if af.Path == "" {
				t.Errorf("adapter %q Files[%d].Path is empty; all adapter files must define a target path", a.ID, i)
			}
		}
	}
}

// --- shim model health contract assertions ---

// TestHarness_ShimAdapter_IncludesLoaderInFiles verifies that every non-copilot
// adapter includes the shared loader (.github/copilot-instructions.md) in its
// Files list. This enforces the shim model: every shim must also manage the
// shared loader so health checks fail if the loader is absent.
func TestHarness_ShimAdapter_IncludesLoaderInFiles(t *testing.T) {
	const loader = ".github/copilot-instructions.md"
	for _, a := range harness.Adapters() {
		if a.ID == "copilot" {
			continue // copilot IS the shared loader
		}
		hasLoader := false
		for _, af := range a.Files {
			if af.Path == loader {
				hasLoader = true
				break
			}
		}
		if !hasLoader {
			t.Errorf("shim adapter %q does not include shared loader %q in Files; shim adapters must manage the shared loader", a.ID, loader)
		}
	}
}

// TestHarness_Status_ShimUnhealthy_WhenLoaderMissing verifies that a shim
// harness is not healthy when only its shim file exists but the shared loader
// (.github/copilot-instructions.md) is absent.
func TestHarness_Status_ShimUnhealthy_WhenLoaderMissing(t *testing.T) {
	dir := t.TempDir()
	// Write only the shim file for claude, not the shared loader.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude Code cARL Adapter\n\nTest shim content.\n"), 0644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	// claude should be Present (CLAUDE.md exists) but not Synced (loader missing).
	if !strings.Contains(output, "claude") {
		t.Errorf("expected 'claude' in status output; got:\n%s", output)
	}
	// Must not appear in the healthy count.
	if strings.Contains(output, "1 healthy.") {
		t.Errorf("expected 0 healthy when shared loader is absent; got:\n%s", output)
	}
}

// TestHarness_Status_ShimUnhealthy_WhenShimMissing verifies that a shim
// harness is not healthy when only the shared loader exists but the
// harness-specific shim is absent.
func TestHarness_Status_ShimUnhealthy_WhenShimMissing(t *testing.T) {
	dir := t.TempDir()
	// Write only the shared loader, not the claude shim.
	if err := os.MkdirAll(filepath.Join(dir, ".github"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github", "copilot-instructions.md"), []byte("# cARL Governance Instructions\n\nTest canonical content.\n"), 0644); err != nil {
		t.Fatalf("write copilot-instructions.md: %v", err)
	}

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	// claude should be Missing (CLAUDE.md absent) and Sync should show Missing.
	if !strings.Contains(output, "claude") {
		t.Errorf("expected 'claude' in status output; got:\n%s", output)
	}
	// claude must not be healthy — its shim is missing.
	if strings.Contains(output, "Missing") == false {
		t.Errorf("expected 'Missing' in status output when shim is absent; got:\n%s", output)
	}
}

// TestHarness_Status_ShimDrifted_WhenLoaderModified verifies that a shim
// harness is reported as Drifted when the shared loader is modified.
func TestHarness_Status_ShimDrifted_WhenLoaderModified(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "claude")

	// Modify the shared loader.
	if err := os.WriteFile(filepath.Join(dir, ".github", "copilot-instructions.md"), []byte("modified loader"), 0644); err != nil {
		t.Fatalf("modify shared loader: %v", err)
	}

	cmd := harness.New(newTestArts())
	output := captureStdout(t, func() {
		if err := cmd.RunStatusInDir(dir); err != nil {
			t.Fatalf("RunStatusInDir: %v", err)
		}
	})

	// Both copilot and claude should be drifted (both share the loader file).
	if !strings.Contains(output, "Drifted") {
		t.Errorf("expected 'Drifted' in status output when shared loader is modified; got:\n%s", output)
	}
}

// TestHarness_Status_CursorDetectionFile verifies cursor uses the new
// .cursor/rules/carl.mdc detection file path.
func TestHarness_Status_CursorDetectionFile(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "cursor")

	// .cursor/rules/carl.mdc must exist (new detection file).
	if _, err := os.Stat(filepath.Join(dir, ".cursor", "rules", "carl.mdc")); err != nil {
		t.Errorf(".cursor/rules/carl.mdc should exist after sync cursor; stat: %v", err)
	}
	// Old .cursorrules must NOT have been created.
	if _, err := os.Stat(filepath.Join(dir, ".cursorrules")); err == nil {
		t.Errorf(".cursorrules should not be created by the new cursor adapter")
	}
}

// TestHarness_Status_AntigravityDetectionFile verifies antigravity uses the
// new .agents/rules/carl.md detection file path.
func TestHarness_Status_AntigravityDetectionFile(t *testing.T) {
	dir := t.TempDir()
	syncHarnesses(t, dir, "antigravity")

	// .agents/rules/carl.md must exist (new detection file).
	if _, err := os.Stat(filepath.Join(dir, ".agents", "rules", "carl.md")); err != nil {
		t.Errorf(".agents/rules/carl.md should exist after sync antigravity; stat: %v", err)
	}
	// Old ANTIGRAVITY.md must NOT have been created.
	if _, err := os.Stat(filepath.Join(dir, "ANTIGRAVITY.md")); err == nil {
		t.Errorf("ANTIGRAVITY.md should not be created by the new antigravity adapter")
	}
}
