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

// AdapterFile describes one output file managed by a harness adapter.
// Each file has its own embedded source, enabling shim adapters to write
// a shared loader alongside their own harness-specific shim content.
type AdapterFile struct {
	// Path is the repo-relative output path for this file.
	Path string
	// SourceFile is the embedded asset path whose content is written to Path.
	SourceFile string
}

// Adapter describes a harness adapter — the bridge between cARL canonical
// artefacts and a specific AI coding agent's context injection mechanism.
type Adapter struct {
	// ID is the machine-readable identifier (e.g. "copilot").
	ID string
	// Name is the human-readable display name (e.g. "GitHub Copilot").
	Name string
	// Support indicates implementation maturity:
	//   "production"   -- tested, production-validated, primary development target
	//   "experimental" -- partial validation, governance loading under investigation
	//   "theoretical"  -- adapter exists, not yet validated end-to-end
	Support string
	// DetectionFile is the repo-relative path whose presence indicates this
	// harness is active in the repository. For shim adapters this is the
	// harness-specific shim file, not the shared loader. Empty for planned adapters.
	DetectionFile string
	// Files lists all files managed by this adapter, each mapped to its own
	// embedded source. For shim adapters (all except copilot) this includes
	// the shared loader (.github/copilot-instructions.md) as well as the
	// harness-specific shim. Empty for planned adapters.
	Files []AdapterFile
}

// loaderPath is the shared cARL loader that every harness shim points to.
const loaderPath = ".github/copilot-instructions.md"

// knownAdapters is the canonical ordered registry of all harnesses cARL
// is aware of. The order determines display order in list and status output.
var knownAdapters = []Adapter{
	{
		ID:            "copilot",
		Name:          "GitHub Copilot",
		Support:       "production",
		DetectionFile: loaderPath,
		Files: []AdapterFile{
			{Path: loaderPath, SourceFile: loaderPath},
		},
	},
	{
		ID:            "claude",
		Name:          "Claude Code",
		Support:       "experimental",
		DetectionFile: "CLAUDE.md",
		Files: []AdapterFile{
			{Path: loaderPath, SourceFile: loaderPath},
			{Path: "CLAUDE.md", SourceFile: "CLAUDE.md"},
		},
	},
	{
		ID:            "codex",
		Name:          "Codex",
		Support:       "theoretical",
		DetectionFile: "AGENTS.md",
		Files: []AdapterFile{
			{Path: loaderPath, SourceFile: loaderPath},
			{Path: "AGENTS.md", SourceFile: "AGENTS.md"},
		},
	},
	{
		ID:            "cursor",
		Name:          "Cursor",
		Support:       "theoretical",
		DetectionFile: ".cursor/rules/carl.mdc",
		Files: []AdapterFile{
			{Path: loaderPath, SourceFile: loaderPath},
			{Path: ".cursor/rules/carl.mdc", SourceFile: ".cursor/rules/carl.mdc"},
		},
	},
	{
		ID:            "antigravity",
		Name:          "Antigravity",
		Support:       "theoretical",
		DetectionFile: ".agents/rules/carl.md",
		Files: []AdapterFile{
			{Path: loaderPath, SourceFile: loaderPath},
			{Path: ".agents/rules/carl.md", SourceFile: ".agents/rules/carl.md"},
		},
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

	production, experimental, theoretical := 0, 0, 0
	for _, a := range knownAdapters {
		switch a.Support {
		case "production":
			production++
		case "experimental":
			experimental++
		case "theoretical":
			theoretical++
		}
	}
	fmt.Printf("%d production, %d experimental, %d theoretical (%d total).\n",
		production, experimental, theoretical, len(knownAdapters))
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
// When multiple adapters share a common file (e.g. the shared loader at
// .github/copilot-instructions.md), that file is written only once to avoid
// redundant writes and to keep the summary count accurate.
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

	// Collect the ordered set of unique file writes. When the same output path
	// appears in multiple adapters (e.g. the shared loader), write it only once
	// under the adapter that claims it first in registry order.
	type pendingWrite struct {
		adapterID  string
		path       string
		sourceFile string
	}
	seen := make(map[string]bool)
	var queue []pendingWrite
	for _, a := range targets {
		for _, af := range a.Files {
			if seen[af.Path] {
				continue
			}
			seen[af.Path] = true
			queue = append(queue, pendingWrite{adapterID: a.ID, path: af.Path, sourceFile: af.SourceFile})
		}
	}

	written := 0
	for _, w := range queue {
		content, err := c.arts.Open(w.sourceFile)
		if err != nil {
			return fmt.Errorf("read canonical source %q for harness %q: %w", w.sourceFile, w.adapterID, err)
		}
		target := filepath.Join(rootDir, filepath.FromSlash(w.path))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("create directory for %s: %w", w.path, err)
		}
		if err := os.WriteFile(target, content, 0644); err != nil {
			return fmt.Errorf("write adapter file %s: %w", w.path, err)
		}
		fmt.Printf("  %-13s  %s\n", w.adapterID, w.path)
		written++
	}

	fmt.Println()
	fmt.Printf("%d adapter file(s) synced.\n", written)
	return nil
}

// resolveAdapters returns the list of adapters to sync. If ids is empty, all
// adapters with defined Files are returned. If ids is non-empty, only the
// named adapters are returned; an error is returned if any id is unrecognised.
func resolveAdapters(ids []string) ([]Adapter, error) {
	if len(ids) == 0 {
		result := make([]Adapter, 0, len(knownAdapters))
		for _, a := range knownAdapters {
			if len(a.Files) > 0 {
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
