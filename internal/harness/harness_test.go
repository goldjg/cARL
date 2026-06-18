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
	cmd := harness.New()

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

// Contract assertion 2: copilot is "supported"; others are "planned".
// Also verifies count summary line.
func TestHarness_List_SupportStatus(t *testing.T) {
	dir := t.TempDir()
	cmd := harness.New()

	output := captureStdout(t, func() {
		if err := cmd.RunListInDir(dir); err != nil {
			t.Fatalf("RunListInDir: %v", err)
		}
	})

	if !strings.Contains(output, "supported") {
		t.Errorf("expected 'supported' in list output; got:\n%s", output)
	}
	if !strings.Contains(output, "planned") {
		t.Errorf("expected 'planned' in list output; got:\n%s", output)
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

	cmd := harness.New()
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

	cmd := harness.New()
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
	cmd := harness.New()
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
	cmd := harness.New()
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
	cmd := harness.New()
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
	cmd := harness.New()

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

	cmd := harness.New()
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

	cmd := harness.New()
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
	cmd := harness.New()
	if cmd.Name() != "harness" {
		t.Errorf("Name() = %q; want 'harness'", cmd.Name())
	}
	if cmd.Synopsis() == "" {
		t.Error("Synopsis() must not be empty")
	}
}
