// Package convert implements the `carl convert` command and its converter
// framework.
//
// `carl convert <source>` migrates durable governance knowledge from a legacy
// or foreign source into canonical cARL artefacts. The first supported source
// is AADLC (`carl convert aadlc`).
//
// The design intentionally separates two concerns:
//
//   - A Converter (e.g. the AADLC converter) knows how to *discover* and
//     *classify* source-specific artefacts into category-tagged migration
//     Items. Adding a new source (claude, copilot, repo, ...) only requires
//     implementing the Converter interface and registering it below.
//   - A converter-agnostic migration engine (see migrate.go) routes Items to
//     canonical cARL destinations, performs duplicate and conflict detection,
//     and produces a deterministic migration report. This logic is shared by
//     all converters.
//
// The guiding principle is preservation: existing durable cARL knowledge is
// more valuable than freshly migrated knowledge. The command never deletes or
// modifies source artefacts and never overwrites existing cARL content without
// explicit conflict handling.
package convert

import (
	"context"
	"fmt"
	"os"
	"sort"
)

// Category identifies the kind of durable knowledge a migration Item carries.
// Each category maps to a canonical cARL destination.
type Category string

const (
	// CategoryInvariant is a repository constraint or assumption.
	// Destination: .github/carl/invariants.yml
	CategoryInvariant Category = "invariant"
	// CategoryMemory is repository-specific context or a lesson learned.
	// Destination: .github/carl/memory.md
	CategoryMemory Category = "memory"
	// CategoryGovernance is a governance rule (PR contract, planning, or
	// approval requirement). Destination: .github/carl/memory.md
	CategoryGovernance Category = "governance"
)

// Artefact is a single discovered source file and its raw content.
type Artefact struct {
	// Path is the repository-relative path of the source artefact.
	Path string
	// Content is the raw file content.
	Content []byte
}

// Item is one unit of durable knowledge classified out of a source artefact.
type Item struct {
	// Category determines the canonical destination.
	Category Category
	// Text is the normalised, single-line statement of the knowledge.
	Text string
	// Source is the repository-relative path of the artefact it came from.
	Source string
}

// Converter knows how to discover and classify a particular family of source
// artefacts. Implementations must be deterministic: given the same repository
// contents they must return the same artefacts and items in the same order.
type Converter interface {
	// ID is the machine-readable source identifier used on the command line
	// (e.g. "aadlc").
	ID() string
	// Name is the human-readable source name (e.g. "AADLC").
	Name() string
	// Discover finds source artefacts under rootDir. It returns an empty
	// slice (not an error) when no artefacts are present.
	Discover(rootDir string) ([]Artefact, error)
	// Classify turns discovered artefacts into category-tagged items.
	Classify(artefacts []Artefact) ([]Item, error)
}

// converters is the ordered registry of all known converters. Adding a new
// source converter only requires appending it here.
var converters = []Converter{
	aadlcConverter{},
}

// converterByID returns the registered converter with the given id.
func converterByID(id string) (Converter, bool) {
	for _, c := range converters {
		if c.ID() == id {
			return c, true
		}
	}
	return nil, false
}

// Command implements `carl convert`.
type Command struct{}

// New returns a new convert Command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "convert" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Migrate durable governance knowledge from legacy sources into cARL artefacts"
}

// Run dispatches to a source converter subcommand, or prints usage.
func (c *Command) Run(_ context.Context, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return nil
	}

	source := args[0]
	conv, ok := converterByID(source)
	if !ok {
		return fmt.Errorf("unknown convert source %q\n\nRun 'carl convert --help' for usage", source)
	}

	apply, err := parseMode(args[1:])
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return RunInDir(cwd, conv, apply)
}

// parseMode interprets the mode flags for a convert subcommand. The default
// (no flag) is dry-run. --apply enables writing; --dry-run is explicit dry-run.
// The two flags are mutually exclusive.
func parseMode(args []string) (apply bool, err error) {
	seen := ""
	for _, a := range args {
		switch a {
		case "--apply":
			if seen == "--dry-run" {
				return false, fmt.Errorf("--apply and --dry-run are mutually exclusive")
			}
			seen, apply = "--apply", true
		case "--dry-run":
			if seen == "--apply" {
				return false, fmt.Errorf("--apply and --dry-run are mutually exclusive")
			}
			seen, apply = "--dry-run", false
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		default:
			return false, fmt.Errorf("unknown flag %q\n\nRun 'carl convert --help' for usage", a)
		}
	}
	return apply, nil
}

// printUsage prints the convert subcommand usage to stdout.
func printUsage() {
	fmt.Println("Usage: carl convert <source> [--dry-run | --apply]")
	fmt.Println()
	fmt.Println("Migrate durable governance knowledge from a legacy source into")
	fmt.Println("canonical cARL artefacts. Defaults to --dry-run (no changes).")
	fmt.Println()
	fmt.Println("Sources:")
	sorted := make([]Converter, len(converters))
	copy(sorted, converters)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID() < sorted[j].ID() })
	for _, c := range sorted {
		fmt.Printf("  %-10s Migrate from %s artefacts\n", c.ID(), c.Name())
	}
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --dry-run  Analyse and report migration opportunities without writing (default)")
	fmt.Println("  --apply    Perform the migration and update cARL artefacts")
}
