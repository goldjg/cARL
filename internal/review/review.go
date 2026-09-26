// Package review implements the `carl review` command.
package review

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/goldjg/carl/internal/cmdutil"
	"github.com/goldjg/carl/internal/semantic"
)

type Command struct {
	out          io.Writer
	err          io.Writer
	newEvaluator func(semantic.Config, string) semantic.Evaluator
}

func New() *Command {
	return &Command{
		out: os.Stdout,
		err: os.Stderr,
		newEvaluator: func(cfg semantic.Config, apiKey string) semantic.Evaluator {
			return semantic.NewTypeSafeJevEvaluator(cfg.JEV, apiKey, nil)
		},
	}
}

func (c *Command) Name() string { return "review" }

func (c *Command) Synopsis() string {
	return "Run bounded semantic review over a git diff"
}

func (c *Command) Run(ctx context.Context, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(ctx, cwd, args)
}

func (c *Command) RunInDir(ctx context.Context, rootDir string, args []string) error {
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}
	if opts.help {
		printUsage(c.out)
		return nil
	}
	state, err := BuildState(ctx, rootDir, StateOptions{
		Base:     opts.base,
		Head:     opts.head,
		TaskFile: opts.taskFile,
		TaskText: opts.task,
	})
	if err != nil {
		return err
	}
	questions := DefaultQuestions()
	cfg, configured, err := semantic.LoadConfig(rootDir)
	if err != nil {
		return err
	}
	if !configured {
		cfg = semantic.Config{Enabled: false, Provider: "jev", FailMode: semantic.DefaultFailMode, JEV: semantic.JEVConfig{APIKeyEnv: semantic.DefaultAPIKeyEnv}}
	}
	provider := firstNonEmpty(cfg.Provider, "jev")
	if opts.dryRun {
		result := Result{Decision: DecisionPass}
		return writeResult(c.out, opts.format, result, &dryRunDocument{
			Provider:       provider,
			Questions:      questions,
			State:          state,
			DataCategories: state.DataCategories,
		})
	}
	if !cfg.Enabled {
		result := Result{
			Decision: DecisionPass,
			Diagnostics: []semantic.Diagnostic{{
				Severity: "warning",
				Code:     "semantic_evaluation_disabled",
				Message:  "semantic evaluation is disabled or unconfigured; no external evaluator was called",
			}},
		}
		return writeResult(c.out, opts.format, result, nil)
	}
	if provider != "jev" {
		return c.handleUnavailable(opts.format, cfg, "unsupported_provider", fmt.Sprintf("semantic evaluation provider %q is not supported", provider))
	}
	apiKeyEnv := firstNonEmpty(cfg.JEV.APIKeyEnv, semantic.DefaultAPIKeyEnv)
	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		return c.handleUnavailable(opts.format, cfg, "missing_api_key", fmt.Sprintf("semantic evaluator API key environment variable %s is not set", apiKeyEnv))
	}
	evaluator := c.newEvaluator(cfg, apiKey)
	eval, err := evaluator.Evaluate(ctx, state, questions)
	if err != nil {
		code := "provider_unavailable"
		var providerErr *semantic.ProviderError
		if errors.As(err, &providerErr) {
			code = providerErr.Code
		}
		return c.handleUnavailable(opts.format, cfg, code, sanitizedProviderMessage(err))
	}
	result := ApplyPolicy(eval, state.Review.ChangedFiles)
	if err := writeResult(c.out, opts.format, result, nil); err != nil {
		return err
	}
	return exitForDecision(result.Decision)
}

func (c *Command) handleUnavailable(format OutputFormat, cfg semantic.Config, code, message string) error {
	result := Result{
		Decision: DecisionPass,
		Diagnostics: []semantic.Diagnostic{{
			Severity: "warning",
			Code:     code,
			Message:  message,
		}},
	}
	if cfg.FailMode == "closed" {
		result.Decision = DecisionFail
		if err := writeResult(c.out, format, result, nil); err != nil {
			return err
		}
		return &cmdutil.ExitError{Code: 1}
	}
	return writeResult(c.out, format, result, nil)
}

type options struct {
	base     string
	head     string
	format   OutputFormat
	dryRun   bool
	taskFile string
	task     string
	help     bool
}

func parseArgs(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.base, "base", "", "base ref for git diff (default: origin/main, main, then HEAD~1)")
	fs.StringVar(&opts.head, "head", "", "head ref for git diff; omitted reviews working tree against base")
	format := fs.String("format", string(FormatText), "output format: text, json, markdown")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "show semantic payload categories/questions without calling the evaluator")
	fs.StringVar(&opts.taskFile, "task-file", "", "repository-relative task or PR description file to include")
	fs.StringVar(&opts.task, "task", "", "task or PR description text to include")
	fs.BoolVar(&opts.help, "help", false, "show help")
	if err := fs.Parse(args); err != nil {
		return opts, fmt.Errorf("parse review flags: %w", err)
	}
	if fs.NArg() > 0 {
		return opts, fmt.Errorf("unexpected review argument %q", fs.Arg(0))
	}
	switch OutputFormat(*format) {
	case FormatText, FormatJSON, FormatMarkdown:
		opts.format = OutputFormat(*format)
	default:
		return opts, fmt.Errorf("unsupported review format %q", *format)
	}
	return opts, nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: carl review [--base <ref>] [--head <ref>] [--format text|json|markdown] [--dry-run] [--task-file <path>] [--task <text>]")
}

func exitForDecision(decision Decision) error {
	switch decision {
	case DecisionPass:
		return nil
	case DecisionWarn, DecisionReviewRequired, DecisionFail:
		return &cmdutil.ExitError{Code: 2}
	default:
		return &cmdutil.ExitError{Code: 1, Message: "unknown review decision"}
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func sanitizedProviderMessage(err error) string {
	var providerErr *semantic.ProviderError
	if errors.As(err, &providerErr) && providerErr.Message != "" {
		return providerErr.Message
	}
	return "semantic evaluator is unavailable"
}
