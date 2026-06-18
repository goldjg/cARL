package plan_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/plan"
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

// writePlan creates a plan file in dir with the given filename and content.
func writePlan(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// planDir returns the .github/carl/plans path under rootDir.
func planDir(rootDir string) string {
	return filepath.Join(rootDir, filepath.FromSlash(plan.PlansDir))
}

// Contract assertion 1: no plans directory → output "No plans found." and return nil.
func TestPlan_NoPlansDirReturnsNilAndNoPlansFound(t *testing.T) {
	dir := t.TempDir()
	cmd := plan.New()

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir: unexpected error: %v", runErr)
	}
	if !strings.Contains(output, "No plans found.") {
		t.Errorf("expected 'No plans found.'; got: %q", output)
	}
}

// Contract assertion 1 (variant): empty plans directory → output "No plans found."
func TestPlan_EmptyPlansDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(planDir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := plan.New()

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir: unexpected error: %v", runErr)
	}
	if !strings.Contains(output, "No plans found.") {
		t.Errorf("expected 'No plans found.'; got: %q", output)
	}
}

// Contract assertion 2: plan with all metadata fields is parsed and shown with
// correct title, status, and purpose.
func TestPlan_FullyPopulatedPlan(t *testing.T) {
	dir := t.TempDir()
	content := `# My Feature Plan

## Plan metadata
- PR / branch: feature/my-feature
- Status: Active
- Author: alice
- Created: 2026-01-01
- Last updated: 2026-01-02

## Task summary
Add the widget subsystem.

## Goal
Provide a fast, testable widget implementation.
`
	writePlan(t, planDir(dir), "my-feature.md", content)

	cmd := plan.New()
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir: unexpected error: %v", runErr)
	}

	for _, want := range []string{
		"my-feature.md",
		"Title:    My Feature Plan",
		"Status:   Active",
		"Purpose:  Add the widget subsystem.",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Warning:") {
		t.Errorf("expected no warnings for fully populated plan; got: %q", output)
	}
}

// Contract assertion 3: plan missing ## Plan metadata section produces a warning.
func TestPlan_MissingMetadataSection(t *testing.T) {
	dir := t.TempDir()
	content := `# Old Style Plan

## Goal
Do something interesting.
`
	writePlan(t, planDir(dir), "old-style.md", content)

	cmd := plan.New()
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir: unexpected error: %v", runErr)
	}
	if !strings.Contains(output, "Warning:") {
		t.Errorf("expected Warning for missing metadata section; got: %q", output)
	}
	if !strings.Contains(output, "missing ## Plan metadata section") {
		t.Errorf("expected 'missing ## Plan metadata section' warning; got: %q", output)
	}
}

// Contract assertion 4: plan with empty Status field produces a warning.
func TestPlan_EmptyStatus(t *testing.T) {
	dir := t.TempDir()
	content := `# Draft Plan

## Plan metadata
- PR / branch:
- Status:
- Author:
- Created:
- Last updated:

## Task summary
This is a draft.
`
	writePlan(t, planDir(dir), "draft.md", content)

	cmd := plan.New()
	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.RunInDir(dir)
	})

	if runErr != nil {
		t.Fatalf("RunInDir: unexpected error: %v", runErr)
	}
	if !strings.Contains(output, "Warning:") {
		t.Errorf("expected Warning for empty Status; got: %q", output)
	}
	if !strings.Contains(output, "Status not set in ## Plan metadata") {
		t.Errorf("expected 'Status not set' warning; got: %q", output)
	}
}

// Contract assertion 5: plans are listed sorted lexicographically by filename.
func TestPlan_SortedByFilename(t *testing.T) {
	dir := t.TempDir()
	pd := planDir(dir)

	for _, name := range []string{"zebra.md", "apple.md", "mango.md"} {
		writePlan(t, pd, name, "# "+name+"\n\n## Plan metadata\n- Status: Active\n\n## Task summary\nSome task.\n")
	}

	plans, err := plan.Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(plans) != 3 {
		t.Fatalf("expected 3 plans; got %d", len(plans))
	}
	want := []string{"apple.md", "mango.md", "zebra.md"}
	for i, p := range plans {
		if p.Filename != want[i] {
			t.Errorf("plans[%d].Filename = %q; want %q", i, p.Filename, want[i])
		}
	}
}

// Contract assertion 6: RunInDir always returns nil (read-only command).
func TestPlan_AlwaysReturnsNil(t *testing.T) {
	dir := t.TempDir()
	// No plans directory at all.
	cmd := plan.New()
	_ = captureStdout(t, func() {
		if err := cmd.RunInDir(dir); err != nil {
			t.Errorf("RunInDir: expected nil; got: %v", err)
		}
	})
}

// TestPlan_Run exercises the Run method (changes working directory).
func TestPlan_Run(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	cmd := plan.New()
	_ = captureStdout(t, func() {
		if err := cmd.Run(context.Background(), nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})
}

// TestPlan_PurposeFallback verifies that purpose falls back from ## Task summary
// to ## Task and then to ## Goal.
func TestPlan_PurposeFallback(t *testing.T) {
	dir := t.TempDir()
	pd := planDir(dir)

	// Plan with only ## Goal.
	writePlan(t, pd, "goal-only.md", `# Goal Only Plan

## Plan metadata
- Status: Active

## Goal
Deliver the feature on time.
`)
	// Plan with ## Task (not ## Task summary).
	writePlan(t, pd, "task-only.md", `# Task Only Plan

## Plan metadata
- Status: Active

## Task
Migrate the database schema.
`)

	plans, err := plan.Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byName := map[string]plan.Plan{}
	for _, p := range plans {
		byName[p.Filename] = p
	}

	if got := byName["goal-only.md"].Purpose; got != "Deliver the feature on time." {
		t.Errorf("goal-only purpose = %q; want 'Deliver the feature on time.'", got)
	}
	if got := byName["task-only.md"].Purpose; got != "Migrate the database schema." {
		t.Errorf("task-only purpose = %q; want 'Migrate the database schema.'", got)
	}
}

// TestPlan_NonMDFilesIgnored verifies that non-.md files in the plans directory
// are silently ignored.
func TestPlan_NonMDFilesIgnored(t *testing.T) {
	dir := t.TempDir()
	pd := planDir(dir)
	writePlan(t, pd, "notes.txt", "some text")
	writePlan(t, pd, "real-plan.md", "# Real Plan\n\n## Plan metadata\n- Status: Active\n\n## Task summary\nDo it.\n")

	plans, err := plan.Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 plan; got %d", len(plans))
	}
	if plans[0].Filename != "real-plan.md" {
		t.Errorf("expected 'real-plan.md'; got %q", plans[0].Filename)
	}
}

// TestPlan_SummaryLine verifies the trailing count line format:
// "N plan(s) found." when no warnings, "N plan(s) found. M warning(s)." when warnings present.
func TestPlan_SummaryLine(t *testing.T) {
	dir := t.TempDir()
	pd := planDir(dir)

	// One clean plan.
	writePlan(t, pd, "clean.md", "# Clean Plan\n\n## Plan metadata\n- Status: Completed\n\n## Task summary\nDone.\n")
	// One plan with a warning.
	writePlan(t, pd, "draft.md", "# Draft\n\n## Task summary\nStarting soon.\n")

	cmd := plan.New()
	output := captureStdout(t, func() {
		_ = cmd.RunInDir(dir)
	})

	if !strings.Contains(output, "2 plan(s) found.") {
		t.Errorf("expected '2 plan(s) found.' in output; got: %q", output)
	}
	if !strings.Contains(output, "1 warning(s).") {
		t.Errorf("expected '1 warning(s).' in output; got: %q", output)
	}
}
