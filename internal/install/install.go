// Package install implements the `carl init` command.
package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goldjg/carl/internal/manifest"
)

// Artifacts provides read access to embedded runtime files.
type Artifacts interface {
	// List returns all embedded file paths relative to the repo root.
	List() ([]string, error)
	// Open returns the content of an embedded file.
	Open(targetPath string) ([]byte, error)
}

// Command implements `carl init`.
type Command struct {
	arts                  Artifacts
	bundledRuntimeVersion string
	bundledRuntimeSource  string
	bundledRuntimeTag     string
	bundledRuntimeCommit  string
}

// New returns a new init Command backed by the given Artifacts.
// Bundled runtime metadata is set at CLI build time and identifies the
// canonical runtime payload embedded in the binary.
func New(
	arts Artifacts,
	bundledRuntimeVersion string,
	bundledRuntimeSource string,
	bundledRuntimeTag string,
	bundledRuntimeCommit string,
) *Command {
	return &Command{
		arts:                  arts,
		bundledRuntimeVersion: bundledRuntimeVersion,
		bundledRuntimeSource:  bundledRuntimeSource,
		bundledRuntimeTag:     bundledRuntimeTag,
		bundledRuntimeCommit:  bundledRuntimeCommit,
	}
}

// Name returns the command name.
func (c *Command) Name() string { return "init" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Install the cARL runtime into the current repository"
}

// Run executes `carl init` in the current working directory.
func (c *Command) Run(_ context.Context, args []string) error {
	return c.run(args)
}

func (c *Command) run(args []string) error {
	adopt := false
	for _, arg := range args {
		switch arg {
		case "--adopt":
			adopt = true
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown init argument %q", arg)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDirWithOptions(cwd, adopt)
}

// RunInDir executes the init logic rooted at rootDir.
// Exported for testing without changing the process working directory.
func (c *Command) RunInDir(rootDir string) error {
	return c.RunInDirWithOptions(rootDir, false)
}

// RunInDirWithOptions executes init rooted at rootDir. When adopt is true,
// existing managed artefacts are preserved and only missing artefacts are
// installed before the new runtime manifest is created.
func (c *Command) RunInDirWithOptions(rootDir string, adopt bool) error {
	// Fail early if runtime.json already exists.
	if manifest.Exists(rootDir) {
		return fmt.Errorf("cARL runtime already installed — %s already exists.\n"+
			"Run `carl repair` to restore any modified artefacts.", manifest.FileName)
	}

	// List all embedded files and check for pre-existing artefacts.
	files, err := c.arts.List()
	if err != nil {
		return fmt.Errorf("list embedded artefacts: %w", err)
	}

	var conflicts, missing []string
	for _, f := range files {
		target := filepath.Join(rootDir, filepath.FromSlash(f))
		_, statErr := os.Stat(target)
		switch {
		case statErr == nil:
			conflicts = append(conflicts, f)
		case os.IsNotExist(statErr):
			missing = append(missing, f)
		default:
			return fmt.Errorf("inspect %s: %w", f, statErr)
		}
	}
	if len(conflicts) > 0 && !adopt {
		msg := "cARL artefacts already exist — run `carl init --adopt` to preserve and adopt them, or remove them for a clean installation:\n"
		for _, c := range conflicts {
			msg += fmt.Sprintf("  %s\n", c)
		}
		return fmt.Errorf("%s", msg)
	}

	// Install all embedded artefacts for a fresh init, or only missing
	// artefacts during explicit non-destructive adoption.
	installFiles := files
	if adopt {
		installFiles = missing
	}
	for _, f := range installFiles {
		target := filepath.Join(rootDir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("create directory for %s: %w", f, err)
		}
		content, err := c.arts.Open(f)
		if err != nil {
			return fmt.Errorf("read embedded artefact %s: %w", f, err)
		}
		if err := os.WriteFile(target, content, 0644); err != nil {
			return fmt.Errorf("write %s: %w", f, err)
		}
	}

	// Create runtime.json.
	rt := &manifest.Runtime{
		RuntimeVersion:   c.bundledRuntimeVersion,
		Source:           c.bundledRuntimeSource,
		SourceTag:        c.bundledRuntimeTag,
		SourceCommit:     c.bundledRuntimeCommit,
		InstalledAt:      time.Now().UTC(),
		ManagedArtifacts: files,
	}
	if err := manifest.WriteNew(rootDir, rt); err != nil {
		return fmt.Errorf("write runtime manifest: %w", err)
	}

	if adopt {
		fmt.Println("Existing cARL artefacts adopted successfully.")
		fmt.Printf("  Existing files preserved: %d\n", len(conflicts))
		fmt.Printf("  Missing files installed:  %d\n", len(missing))
		fmt.Println("  Run `carl doctor` to inspect drift and `carl repair` to restore repairable artefacts.")
	} else {
		fmt.Println("cARL runtime installed successfully.")
	}
	fmt.Printf("  Runtime version:  %s\n", rt.RuntimeVersion)
	fmt.Printf("  Source:           %s @ %s\n", rt.Source, rt.SourceTag)
	fmt.Printf("  Artefacts:        %d files installed\n", len(files))
	return nil
}

func printUsage() {
	fmt.Println("Usage: carl init [--adopt]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --adopt  Preserve existing cARL artefacts, install missing files, and create runtime.json")
}
