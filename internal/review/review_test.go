package review

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/cmdutil"
	"github.com/goldjg/carl/internal/semantic"
)

type fakeEvaluator struct {
	called bool
	result *semantic.EvaluationResult
	err    error
}

func (f *fakeEvaluator) Evaluate(_ context.Context, _ semantic.EvaluationState, _ []semantic.Question) (*semantic.EvaluationResult, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestReviewDryRunDoesNotCallEvaluatorAndReportsPayload(t *testing.T) {
	dir := fixtureRepo(t)
	var out bytes.Buffer
	fake := &fakeEvaluator{}
	cmd := New()
	cmd.out = &out
	cmd.newEvaluator = func(semantic.Config, string) semantic.Evaluator { return fake }

	if err := cmd.RunInDir(context.Background(), dir, []string{"--base", "HEAD~1", "--dry-run", "--format", "json"}); err != nil {
		t.Fatalf("RunInDir: %v", err)
	}
	if fake.called {
		t.Fatal("dry-run must not call evaluator")
	}
	var doc outputDocument
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out.String())
	}
	if doc.DryRun == nil || len(doc.DryRun.Questions) == 0 || len(doc.DryRun.DataCategories) == 0 {
		t.Fatalf("dry-run payload missing: %+v", doc)
	}
}

func TestReviewMissingAPIKeyFailOpen(t *testing.T) {
	dir := fixtureRepo(t)
	writeConfig(t, dir, `semantic_evaluation:
  enabled: true
  provider: jev
  fail_mode: open
  jev:
    api_key_env: TYPESAFE_MISSING_TEST_KEY
    endpoint: https://example.invalid/jev
`)
	var out bytes.Buffer
	cmd := New()
	cmd.out = &out
	err := cmd.RunInDir(context.Background(), dir, []string{"--base", "HEAD~1", "--format", "json"})
	if err != nil {
		t.Fatalf("fail-open missing API key should not fail command: %v", err)
	}
	if !strings.Contains(out.String(), "missing_api_key") {
		t.Fatalf("expected missing_api_key diagnostic: %s", out.String())
	}
}

func TestReviewProviderErrorFailOpen(t *testing.T) {
	dir := fixtureRepo(t)
	writeConfig(t, dir, `semantic_evaluation:
  enabled: true
  provider: jev
  fail_mode: open
  jev:
    api_key_env: TYPESAFE_TEST_KEY
    endpoint: https://example.invalid/jev
`)
	t.Setenv("TYPESAFE_TEST_KEY", "secret-test-key")
	var out bytes.Buffer
	fake := &fakeEvaluator{err: &semantic.ProviderError{Code: "rate_limited", Message: "JEV rate limit exceeded"}}
	cmd := New()
	cmd.out = &out
	cmd.newEvaluator = func(semantic.Config, string) semantic.Evaluator { return fake }

	err := cmd.RunInDir(context.Background(), dir, []string{"--base", "HEAD~1"})
	if err != nil {
		t.Fatalf("fail-open provider error should not fail command: %v", err)
	}
	if !fake.called || !strings.Contains(out.String(), "rate_limited") {
		t.Fatalf("expected fail-open diagnostic, called=%t output=%s", fake.called, out.String())
	}
	if strings.Contains(out.String(), "secret-test-key") {
		t.Fatalf("API key leaked in output: %s", out.String())
	}
}

func TestReviewEvaluatorResultOutputAndExitCode(t *testing.T) {
	dir := fixtureRepo(t)
	writeConfig(t, dir, `semantic_evaluation:
  enabled: true
  provider: jev
  fail_mode: open
  jev:
    api_key_env: TYPESAFE_TEST_KEY
    endpoint: https://example.invalid/jev
`)
	t.Setenv("TYPESAFE_TEST_KEY", "secret-test-key")
	var out bytes.Buffer
	fake := &fakeEvaluator{result: &semantic.EvaluationResult{Signals: []semantic.Signal{
		{ID: "trust_boundary_change", Value: true, Source: "test"},
	}}}
	cmd := New()
	cmd.out = &out
	cmd.newEvaluator = func(semantic.Config, string) semantic.Evaluator { return fake }

	err := cmd.RunInDir(context.Background(), dir, []string{"--base", "HEAD~1", "--format", "markdown"})
	var exitErr *cmdutil.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Fatalf("expected semantic non-pass exit 2, got %T %[1]v", err)
	}
	if !strings.Contains(out.String(), "review_required") || !strings.Contains(out.String(), "Trust-boundary change") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestReviewDisabledConfigPassesWithDiagnostic(t *testing.T) {
	dir := fixtureRepo(t)
	var out bytes.Buffer
	cmd := New()
	cmd.out = &out
	err := cmd.RunInDir(context.Background(), dir, []string{"--base", "HEAD~1"})
	if err != nil {
		t.Fatalf("disabled semantic evaluation should pass: %v", err)
	}
	if !strings.Contains(out.String(), "semantic_evaluation_disabled") {
		t.Fatalf("expected disabled diagnostic: %s", out.String())
	}
}

func TestBuildStateRedactsSecretFilesAndDiffLines(t *testing.T) {
	dir := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("API_KEY=secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safe.txt"), []byte("hello\npassword = secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", ".env", "safe.txt")
	state, err := BuildState(context.Background(), dir, StateOptions{Base: "HEAD~1"})
	if err != nil {
		t.Fatalf("BuildState: %v", err)
	}
	if strings.Contains(state.Diff.Text, "secret") {
		t.Fatalf("secret leaked in diff: %s", state.Diff.Text)
	}
	foundEnv := false
	for _, f := range state.Review.ChangedFiles {
		if f.Path == ".env" && f.Redacted {
			foundEnv = true
		}
	}
	if !foundEnv {
		t.Fatalf("expected .env changed file to be marked redacted: %+v", state.Review.ChangedFiles)
	}
}

func fixtureRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init")
	git(t, dir, "config", "user.email", "test@example.invalid")
	git(t, dir, "config", "user.name", "Test User")
	writeFile(t, dir, "safe.txt", "hello\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "initial")
	writeFile(t, dir, "safe.txt", "hello\nworld\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "change")
	return dir
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	writeFile(t, dir, ".github/carl/config.yml", content)
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
