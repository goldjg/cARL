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

// Adapter describes a harness adapter — the bridge between cARL canonical
// artefacts and a specific AI coding agent's context injection mechanism.
type Adapter struct {
	// ID is the machine-readable identifier (e.g. "copilot").
	ID string
	// Name is the human-readable display name (e.g. "GitHub Copilot").
	Name string
	// Support indicates implementation maturity: "supported" or "planned".
	Support string
	// DetectionFile is the repo-relative path whose presence indicates this
	// harness is active in the repository. Empty for planned adapters.
	DetectionFile string
	// AdapterFiles lists repo-relative paths that serve as the adapter layer
	// for this harness. These files are managed by cARL; the harness is the
	// consumer. Empty for planned adapters.
	AdapterFiles []string
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
	},
	{
		ID:      "claude",
		Name:    "Claude Code",
		Support: "planned",
	},
	{
		ID:      "codex",
		Name:    "Codex",
		Support: "planned",
	},
	{
		ID:      "cursor",
		Name:    "Cursor",
		Support: "planned",
	},
	{
		ID:      "antigravity",
		Name:    "Antigravity",
		Support: "planned",
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
type Command struct{}

// New returns a new harness Command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "harness" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Manage and inspect harness adapters for AI coding agents"
}

// Run dispatches to the list or status subcommand, or prints usage.
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
	fmt.Println("  status  Report harness adapter detection status in the current repository")
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

// RunStatusInDir reports the detection status of all known harness adapters
// in rootDir. Detection is based on the presence of each adapter's DetectionFile.
// Exported for testing without changing the process working directory.
func (c *Command) RunStatusInDir(rootDir string) error {
	fmt.Println("Harness Adapter Status:")
	fmt.Println()

	active := 0
	for _, a := range knownAdapters {
		detected := isDetected(a, rootDir)
		if detected {
			active++
		}
		// "supported" adapters report "active" or "not active".
		// "planned" adapters report "-" — no detection is attempted.
		detectionStatus := "-"
		if a.Support == "supported" {
			if detected {
				detectionStatus = "active"
			} else {
				detectionStatus = "not active"
			}
		}
		fmt.Printf("  %-13s %-20s %-11s %s\n", a.ID, a.Name, a.Support, detectionStatus)
	}
	fmt.Println()
	fmt.Printf("%d of %d harness(es) active.\n", active, len(knownAdapters))
	return nil
}
