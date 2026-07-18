// Command carl is the cARL CLI — a governance runtime manager for coding agents.
//
// Usage:
//
//	carl <command> [arguments]
//
// Available commands:
//
//	convert    Migrate durable governance knowledge from legacy sources into cARL artefacts
//	doctor     Diagnose runtime issues and provide actionable remediation guidance
//	harness    Manage and inspect harness adapters for AI coding agents
//	init       Install the cARL runtime into the current repository
//	map        Generate and update .github/carl/repo-map.json from repository structure
//	plan       Discover, validate, and summarise plans in .github/carl/plans/
//	reconcile  Update repository-specific memory sections from current repo-map data
//	repair     Restore modified managed cARL artefacts to their canonical state
//	status     Report whether the installed cARL runtime is healthy, missing, or drifted
//	version    Show CLI, bundled runtime, and repository runtime version information
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goldjg/carl/embedded"
	"github.com/goldjg/carl/internal/cmdutil"
	"github.com/goldjg/carl/internal/convert"
	"github.com/goldjg/carl/internal/doctor"
	"github.com/goldjg/carl/internal/harness"
	"github.com/goldjg/carl/internal/install"
	"github.com/goldjg/carl/internal/plan"
	"github.com/goldjg/carl/internal/reconcile"
	"github.com/goldjg/carl/internal/repair"
	"github.com/goldjg/carl/internal/repomap"
	"github.com/goldjg/carl/internal/status"
	"github.com/goldjg/carl/internal/version"
)

// cliVersion is set at build time via:
//
//	go build -ldflags "-X main.cliVersion=1.0.0"
var cliVersion = "dev"

// bundledRuntimeVersion is the canonical cARL runtime version embedded in
// this CLI binary. It is set at build time and may differ from cliVersion.
var bundledRuntimeVersion = "1.0.0"

// bundledRuntimeSource is the canonical source repository for the embedded
// runtime payload.
var bundledRuntimeSource = "goldjg/cARL"

// bundledRuntimeTag is the source release tag for the embedded runtime
// payload.
var bundledRuntimeTag = "v1.0.0"

// bundledRuntimeCommit is the source commit for the embedded runtime payload.
// It is set at build time via:
//
//	go build -ldflags "-X main.bundledRuntimeCommit=<hash>"
var bundledRuntimeCommit = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cmds := []cmdutil.Command{
		convert.New(),
		doctor.New(embedded.Assets),
		harness.New(embedded.Assets),
		install.New(
			embedded.Assets,
			bundledRuntimeVersion,
			bundledRuntimeSource,
			bundledRuntimeTag,
			bundledRuntimeCommit,
		),
		repomap.New(),
		plan.New(),
		reconcile.New(),
		repair.New(embedded.Assets),
		status.New(cliVersion, embedded.Assets),
		version.New(
			cliVersion,
			bundledRuntimeVersion,
			bundledRuntimeSource,
			bundledRuntimeTag,
			bundledRuntimeCommit,
			embedded.Assets,
		),
	}

	if err := run(ctx, os.Args[1:], cmds); err != nil {
		fmt.Fprintf(os.Stderr, "carl: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, cmds []cmdutil.Command) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printUsage(cmds)
		return nil
	}

	name := args[0]
	if name == "--version" || name == "-v" {
		name = "version"
	}

	for _, cmd := range cmds {
		if cmd.Name() == name {
			return cmd.Run(ctx, args[1:])
		}
	}

	fmt.Fprintf(os.Stderr, "carl: unknown command %q\n\n", name)
	printUsage(cmds)
	os.Exit(1)
	return nil
}

func printUsage(cmds []cmdutil.Command) {
	fmt.Fprintln(os.Stderr, "Usage: carl <command> [arguments]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	for _, cmd := range cmds {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", cmd.Name(), cmd.Synopsis())
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Run 'carl <command> --help' for more information on a command.")
}
