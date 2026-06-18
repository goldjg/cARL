// Package harness implements the `carl harness` command and its subcommands.
//
// Harness adapters bridge cARL canonical artefacts to the context injection
// mechanisms of specific AI coding agents. cARL artefacts (.github/carl/) are
// the canonical source of truth. Harness files (e.g. .github/copilot-instructions.md)
// are adapters that present cARL artefacts to a particular agent runtime.
package harness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Artifacts provides read access to embedded canonical runtime files.
// Implementations are expected to be backed by the binary-embedded assets.
type Artifacts interface {
	// Open returns the content of an embedded file at the given target path.
	// targetPath is relative to the repository root.
	Open(targetPath string) ([]byte, error)
}

// Adapter describes a harness adapter — the bridge between cARL canonical
// artefacts and a specific AI coding agent's context injection mechanism.
type Adapter struct {
	// ID is the machine-readable identifier (e.g. "copilot").
	ID string
	// Name is the human-readable display name (e.g. "GitHub Copilot").
	Name string
	// Support indicates implementation maturity: "supported" or "planned".
	// A "supported" adapter has its DetectionFile and AdapterFiles defined,
	// enabling detection and status reporting.
	Support string
	// DetectionFile is the repo-relative path whose presence indicates this
	// harness is active in the repository. Empty for planned adapters.
	DetectionFile string
	// AdapterFiles lists repo-relative paths that serve as the adapter layer
	// for this harness. These files are managed by cARL; the harness is the
	// consumer. Empty for planned adapters.
	AdapterFiles []string
	// SourceFile is the repo-relative path of the embedded canonical artefact
	// used as the content source when generating this adapter's files via
	// `carl harness sync`. Empty for planned adapters.
	SourceFile string
}

// knownAdapters is the canonical ordered registry of all harnesses cARL
// is aware of. The order determines display order in list and status output.
var knownAdapters = []Adapter{
	{
		ID:            "copilot",
		Name:          "GitHub Copilot",
		Support:       "supported",
		DetectionFile: ".github/copilot-instructions.md",
		AdapterFiles:  []string{".github/copilot-instructions.md"},
		SourceFile:    ".github/copilot-instructions.md",
	},
	{
		ID:            "claude",
		Name:          "Claude Code",
		Support:       "supported",
		DetectionFile: "CLAUDE.md",
		AdapterFiles:  []string{"CLAUDE.md"},
		SourceFile:    ".github/copilot-instructions.md",
	},
	{
		ID:            "codex",
		Name:          "Codex",
		Support:       "supported",
		DetectionFile: "AGENTS.md",
		AdapterFiles:  []string{"AGENTS.md"},
		SourceFile:    ".github/copilot-instructions.md",
	},
	{
		ID:            "cursor",
		Name:          "Cursor",
		Support:       "supported",
		DetectionFile: ".cursorrules",
		AdapterFiles:  []string{".cursorrules"},
		SourceFile:    ".github/copilot-instructions.md",
	},
	{
		ID:            "antigravity",
		Name:          "Antigravity",
		Support:       "supported",
		DetectionFile: "ANTIGRAVITY.md",
		AdapterFiles:  []string{"ANTIGRAVITY.md"},
		SourceFile:    ".github/copilot-instructions.md",
	},
}

// Adapters returns a copy of the known adapter registry.
// Callers may inspect the registry without modifying the canonical list.
func Adapters() []Adapter {
	result := make([]Adapter, len(knownAdapters))
	copy(result, knownAdapters)
	return result
}

// isDetected reports whether a is active in the repository at rootDir.
// Detection is based solely on the presence of a.DetectionFile.
// An adapter with an empty DetectionFile is never detected.
func isDetected(a Adapter, rootDir string) bool {
	if a.DetectionFile == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(rootDir, filepath.FromSlash(a.DetectionFile)))
	return err == nil
}

// Command implements `carl harness`.
type Command struct {
	arts Artifacts
}

// New returns a new harness Command backed by the given Artifacts.
// arts is used by the sync subcommand to read canonical adapter content.
func New(arts Artifacts) *Command { return &Command{arts: arts} }

// Name returns the command name.
func (c *Command) Name() string { return "harness" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Manage and inspect harness adapters for AI coding agents"
}

// Run dispatches to the list, status, or sync subcommand, or prints usage.
// An unknown subcommand returns a non-nil error.
func (c *Command) Run(_ context.Context, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	switch args[0] {
	case "list":
		return c.RunListInDir(cwd)
	case "status":
		return c.RunStatusInDir(cwd)
	case "sync":
		return c.RunSyncInDir(cwd, args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q\n\nRun 'carl harness --help' for usage", args[0])
	}
}

// printUsage prints the harness subcommand usage to stdout.
func printUsage() {
	fmt.Println("Usage: carl harness <subcommand> [arguments]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list    List known harness adapters and their support status")
	fmt.Println("  status  Report harness adapter presence and sync health in the current repository")
	fmt.Println("  sync    Generate adapter files for supported harnesses from canonical cARL artefacts")
	fmt.Println()
	fmt.Println("Run 'carl harness <subcommand> --help' for more information.")
}

// RunListInDir prints the known harness adapters and their support status.
// rootDir is accepted for API consistency but unused by this subcommand.
// Exported for testing without changing the process working directory.
func (c *Command) RunListInDir(_ string) error {
	fmt.Println("Harness Adapters:")
	fmt.Println()
	for _, a := range knownAdapters {
		fmt.Printf("  %-13s %-20s %s\n", a.ID, a.Name, a.Support)
	}
	fmt.Println()

	supported := 0
	for _, a := range knownAdapters {
		if a.Support == "supported" {
			supported++
		}
	}
	fmt.Printf("%d of %d adapter(s) supported.\n", supported, len(knownAdapters))
	return nil
}

// RunStatusInDir reports the presence and sync health of all known harness
// adapters in rootDir. Detection is based on DetectionFile presence; sync
// health compares adapter file content against the canonical embedded source.
// Exported for testing without changing the process working directory.
func (c *Command) RunStatusInDir(rootDir string) error {
	health, err := Inspect(rootDir, c.arts)
	if err != nil {
		return err
	}

	fmt.Println("Harness Adapter Status:")
	fmt.Println()

	for _, h := range health {
		fmt.Printf("  %-13s %-20s %-11s %-8s %s\n",
			h.Adapter.ID, h.Adapter.Name, h.Adapter.Support, h.Presence, h.Sync)
	}
	fmt.Println()

	summary := Summarize(health)
	fmt.Printf("%d active, %d missing, %d drifted, %d healthy.\n",
		summary.Active, summary.Missing, summary.Drifted, summary.Healthy)
	return nil
}

// RunSyncInDir generates adapter files for supported harnesses in rootDir.
// It reads the canonical content from embedded artefacts and writes each
// adapter's file(s) to disk. Existing files are overwritten — adapter files
// are disposable and always regenerated from the canonical source.
//
// harnessIDs restricts the operation to the named harnesses. If empty, all
// supported harnesses are synced. An unknown harness ID returns an error.
//
// Exported for testing without changing the process working directory.
func (c *Command) RunSyncInDir(rootDir string, harnessIDs []string) error {
	// Validate and resolve target adapters.
	targets, err := resolveAdapters(harnessIDs)
	if err != nil {
		return err
	}

	fmt.Println("Syncing harness adapters...")
	fmt.Println()

	written := 0
	for _, a := range targets {
		// resolveAdapters returns only supported adapters; this guard is a
		// belt-and-suspenders check against malformed registry entries.
		if len(a.AdapterFiles) == 0 || a.SourceFile == "" {
			continue
		}

		content, err := c.arts.Open(a.SourceFile)
		if err != nil {
			return fmt.Errorf("read canonical source for harness %q: %w", a.ID, err)
		}

		for _, af := range a.AdapterFiles {
			target := filepath.Join(rootDir, filepath.FromSlash(af))
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("create directory for %s: %w", af, err)
			}
			if err := os.WriteFile(target, content, 0644); err != nil {
				return fmt.Errorf("write adapter file %s: %w", af, err)
			}
			fmt.Printf("  %-13s  %s\n", a.ID, af)
			written++
		}
	}

	fmt.Println()
	fmt.Printf("%d adapter file(s) synced.\n", written)
	return nil
}

// resolveAdapters returns the list of adapters to sync. If ids is empty, all
// supported adapters are returned. If ids is non-empty, only the named adapters
// are returned; an error is returned if any id is unrecognised.
func resolveAdapters(ids []string) ([]Adapter, error) {
	if len(ids) == 0 {
		result := make([]Adapter, 0, len(knownAdapters))
		for _, a := range knownAdapters {
			if a.Support == "supported" {
				result = append(result, a)
			}
		}
		return result, nil
	}

	index := make(map[string]Adapter, len(knownAdapters))
	for _, a := range knownAdapters {
		index[a.ID] = a
	}

	result := make([]Adapter, 0, len(ids))
	for _, id := range ids {
		a, ok := index[id]
		if !ok {
			return nil, fmt.Errorf("unknown harness %q — run 'carl harness list' to see available adapters", id)
		}
		result = append(result, a)
	}
	return result, nil
}
