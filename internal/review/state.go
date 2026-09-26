package review

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goldjg/carl/internal/semantic"
)

const (
	maxDiffBytes    = 200000
	maxContextBytes = 12000
)

type StateOptions struct {
	Base     string
	Head     string
	TaskFile string
	TaskText string
}

// BuildState collects bounded local repository state for semantic review.
func BuildState(ctx context.Context, rootDir string, opts StateOptions) (semantic.EvaluationState, error) {
	repoRoot, err := gitOutput(ctx, rootDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return semantic.EvaluationState{}, fmt.Errorf("locate git repository: %w", err)
	}
	rootDir = strings.TrimSpace(repoRoot)
	base := opts.Base
	if base == "" {
		base = defaultBase(ctx, rootDir)
	}
	head := opts.Head

	changed, err := changedFiles(ctx, rootDir, base, head)
	if err != nil {
		return semantic.EvaluationState{}, err
	}
	diff, err := collectDiff(ctx, rootDir, base, head, changed)
	if err != nil {
		return semantic.EvaluationState{}, err
	}
	contextItems, categories, err := collectCarlContext(rootDir)
	if err != nil {
		return semantic.EvaluationState{}, err
	}
	task, taskCategory, err := collectTask(rootDir, opts)
	if err != nil {
		return semantic.EvaluationState{}, err
	}
	if taskCategory.Name != "" {
		categories = append(categories, taskCategory)
	}
	categories = append(categories, semantic.DataCategory{
		Name:        "git_diff",
		Description: "Redacted git diff for changed non-secret-bearing files",
		Included:    diff.Text != "",
		Redacted:    diff.Redacted,
		Truncated:   diff.Truncated,
	})
	categories = append(categories, semantic.DataCategory{
		Name:        "changed_files",
		Description: "Changed file paths and addition/deletion counts from git diff metadata",
		Included:    len(changed) > 0,
	})
	categories = append(categories, semantic.DataCategory{
		Name:        "repository_metadata",
		Description: "Repository name, branch, sanitized remote, and base/head refs",
		Included:    true,
	})

	baseSHA, _ := gitOutput(ctx, rootDir, "rev-parse", "--verify", base)
	headSHA := ""
	if head != "" {
		headSHA, _ = gitOutput(ctx, rootDir, "rev-parse", "--verify", head)
	} else {
		headSHA, _ = gitOutput(ctx, rootDir, "rev-parse", "--verify", "HEAD")
	}
	branch, _ := gitOutput(ctx, rootDir, "rev-parse", "--abbrev-ref", "HEAD")
	remote, _ := gitOutput(ctx, rootDir, "config", "--get", "remote.origin.url")
	remote, _ = semantic.RedactText(strings.TrimSpace(remote))

	return semantic.EvaluationState{
		Repository: semantic.RepositoryMetadata{
			Name:          filepath.Base(rootDir),
			CurrentBranch: strings.TrimSpace(branch),
			RemoteOrigin:  strings.TrimSpace(remote),
		},
		Review: semantic.ReviewState{
			Base:         base,
			Head:         head,
			BaseSHA:      strings.TrimSpace(baseSHA),
			HeadSHA:      strings.TrimSpace(headSHA),
			ChangedFiles: changed,
		},
		Diff:           diff,
		Context:        contextItems,
		Task:           task,
		DataCategories: categories,
	}, nil
}

func defaultBase(ctx context.Context, rootDir string) string {
	for _, candidate := range []string{"origin/main", "main", "HEAD~1"} {
		if _, err := gitOutput(ctx, rootDir, "rev-parse", "--verify", candidate); err == nil {
			return candidate
		}
	}
	return "HEAD~1"
}

func changedFiles(ctx context.Context, rootDir, base, head string) ([]semantic.ChangedFile, error) {
	args := diffRangeArgs(base, head)
	args = append([]string{"diff", "--numstat"}, args...)
	out, err := gitOutput(ctx, rootDir, args...)
	if err != nil {
		return nil, fmt.Errorf("collect changed files: %w", err)
	}
	var files []semantic.ChangedFile
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		cf := semantic.ChangedFile{Path: parts[2]}
		if parts[0] == "-" || parts[1] == "-" {
			cf.Binary = true
		} else {
			cf.Additions, _ = strconv.Atoi(parts[0])
			cf.Deletions, _ = strconv.Atoi(parts[1])
		}
		if semantic.IsSecretPath(cf.Path) {
			cf.Redacted = true
		}
		files = append(files, cf)
	}
	return files, nil
}

func collectDiff(ctx context.Context, rootDir, base, head string, changed []semantic.ChangedFile) (semantic.DiffState, error) {
	var safePaths []string
	for _, f := range changed {
		if !f.Redacted {
			safePaths = append(safePaths, f.Path)
		}
	}
	if len(safePaths) == 0 {
		return semantic.DiffState{LimitBytes: maxDiffBytes, Redacted: len(changed) > 0}, nil
	}
	args := append([]string{"diff", "--no-ext-diff"}, diffRangeArgs(base, head)...)
	args = append(args, "--")
	args = append(args, safePaths...)
	out, err := gitOutput(ctx, rootDir, args...)
	if err != nil {
		return semantic.DiffState{}, fmt.Errorf("collect git diff: %w", err)
	}
	truncated := false
	if len(out) > maxDiffBytes {
		out = out[:maxDiffBytes]
		truncated = true
	}
	redactedText, redacted := semantic.RedactText(out)
	return semantic.DiffState{Text: redactedText, Truncated: truncated, Redacted: redacted, LimitBytes: maxDiffBytes}, nil
}

func diffRangeArgs(base, head string) []string {
	if head == "" {
		return []string{base}
	}
	return []string{base, head}
}

func collectCarlContext(rootDir string) ([]semantic.ContextItem, []semantic.DataCategory, error) {
	paths := []string{
		".github/carl/current-pr-contract.md",
		".github/carl/invariants.yml",
		".github/carl/trust-boundaries.md",
		".github/carl/tool-policy.yml",
	}
	var items []semantic.ContextItem
	for _, rel := range paths {
		item, ok, err := readContextFile(rootDir, rel)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			items = append(items, item)
		}
	}
	return items, []semantic.DataCategory{{
		Name:        "carl_context",
		Description: "Bounded cARL governance context files relevant to review policy",
		Included:    len(items) > 0,
	}}, nil
}

func readContextFile(rootDir, rel string) (semantic.ContextItem, bool, error) {
	path := filepath.Join(rootDir, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return semantic.ContextItem{}, false, nil
		}
		return semantic.ContextItem{}, false, fmt.Errorf("read %s: %w", rel, err)
	}
	truncated := false
	if len(data) > maxContextBytes {
		data = data[:maxContextBytes]
		truncated = true
	}
	content, redacted := semantic.RedactText(string(data))
	return semantic.ContextItem{Path: rel, Content: content, Truncated: truncated, Redacted: redacted}, true, nil
}

func collectTask(rootDir string, opts StateOptions) (string, semantic.DataCategory, error) {
	if opts.TaskText != "" {
		text, redacted := semantic.RedactText(opts.TaskText)
		return text, semantic.DataCategory{Name: "task_description", Description: "Task or PR description supplied by the caller", Included: true, Redacted: redacted}, nil
	}
	if opts.TaskFile == "" {
		return "", semantic.DataCategory{Name: "task_description", Description: "Task or PR description supplied by the caller", Included: false}, nil
	}
	clean := filepath.Clean(opts.TaskFile)
	if filepath.IsAbs(clean) {
		return "", semantic.DataCategory{}, fmt.Errorf("task file must be repository-relative")
	}
	full := filepath.Join(rootDir, clean)
	rel, err := filepath.Rel(rootDir, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", semantic.DataCategory{}, fmt.Errorf("task file must remain inside the repository")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", semantic.DataCategory{}, fmt.Errorf("read task file: %w", err)
	}
	truncated := false
	if len(data) > maxContextBytes {
		data = data[:maxContextBytes]
		truncated = true
	}
	text, redacted := semantic.RedactText(string(data))
	return text, semantic.DataCategory{Name: "task_description", Description: "Task or PR description supplied by the caller", Included: true, Redacted: redacted, Truncated: truncated}, nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return string(out), nil
}
