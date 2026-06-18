// Package repair implements the `carl repair` command.
package repair

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goldjg/carl/internal/manifest"
)

// protectedArtifacts lists paths that must never be overwritten by repair.
// memory.md is per-repository state; runtime.json is managed by init.
var protectedArtifacts = map[string]bool{
	".github/carl/memory.md":      true,
	manifest.FileName:             true,
}

// Artifacts provides read access to the embedded canonical runtime files.
type Artifacts interface {
	Open(targetPath string) ([]byte, error)
}

// Command implements `carl repair`.
type Command struct {
	arts Artifacts
}

// New returns a new repair Command backed by the given Artifacts.
func New(arts Artifacts) *Command {
	return &Command{arts: arts}
}

// Name returns the command name.
func (c *Command) Name() string { return "repair" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Restore modified managed cARL artefacts to their canonical state"
}

// Run executes `carl repair` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the repair logic rooted at rootDir.
// Exported for testing without changing the process working directory.
func (c *Command) RunInDir(rootDir string) error {
	if !manifest.Exists(rootDir) {
		return fmt.Errorf("no cARL runtime installed — run `carl init` first")
	}

	rt, err := manifest.Read(rootDir)
	if err != nil {
		return fmt.Errorf("read runtime manifest: %w", err)
	}

	// Identify drifted artefacts.
	drifted, err := c.detectDrift(rootDir, rt.ManagedArtifacts)
	if err != nil {
		return err
	}

	if len(drifted) == 0 {
		fmt.Println("No drift detected.")
		return nil
	}

	fmt.Println("Drift detected:")
	for _, f := range drifted {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println()

	fmt.Println("Repairing...")
	for _, f := range drifted {
		if err := c.restoreFile(rootDir, f); err != nil {
			return fmt.Errorf("repair %s: %w", f, err)
		}
	}
	fmt.Println("Done.")
	return nil
}

// Inspect classifies managed artefacts for the given rootDir.
// It skips protected paths (memory.md, runtime.json) and paths that have no
// embedded canonical (future artefact types the current CLI does not bundle).
// missing contains paths that are absent from disk.
// drifted contains paths that exist but whose content differs from the embedded
// canonical.
// Callers that only need a combined list (e.g. repair) should use detectDrift.
func Inspect(rootDir string, managed []string, arts Artifacts) (missing, drifted []string, err error) {
	for _, f := range managed {
		if protectedArtifacts[f] {
			continue
		}
		canonical, canonErr := arts.Open(f)
		if canonErr != nil {
			// Not in embedded FS — skip (future artefact type).
			continue
		}
		target := filepath.Join(rootDir, filepath.FromSlash(f))
		installed, readErr := os.ReadFile(target)
		if readErr != nil {
			missing = append(missing, f)
			continue
		}
		if !bytes.Equal(canonical, installed) {
			drifted = append(drifted, f)
		}
	}
	return missing, drifted, nil
}

// detectDrift returns the combined list of managed artefacts that are missing
// from disk or differ from the embedded canonical versions, skipping protected
// paths. The returned slice preserves the order in which managed is iterated:
// missing entries come before drifted entries within each scan.
func (c *Command) detectDrift(rootDir string, managed []string) ([]string, error) {
	missing, drifted, err := Inspect(rootDir, managed, c.arts)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(missing)+len(drifted))
	result = append(result, missing...)
	result = append(result, drifted...)
	return result, nil
}

// restoreFile writes the embedded canonical version of f to rootDir.
func (c *Command) restoreFile(rootDir, f string) error {
	content, err := c.arts.Open(f)
	if err != nil {
		return err
	}
	target := filepath.Join(rootDir, filepath.FromSlash(f))
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, content, 0644)
}
