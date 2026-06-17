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

const (
	runtimeVersion = "1.0.0"
	source         = "goldjg/cARL"
	sourceTag      = "v1.0.0"
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
	arts         Artifacts
	sourceCommit string
}

// New returns a new init Command backed by the given Artifacts.
// sourceCommit is the VCS commit hash recorded in runtime.json; set at build
// time via -ldflags "-X main.sourceCommit=<hash>" and threaded in from main.
func New(arts Artifacts, sourceCommit string) *Command {
	return &Command{arts: arts, sourceCommit: sourceCommit}
}

// Name returns the command name.
func (c *Command) Name() string { return "init" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Install the cARL runtime into the current repository"
}

// Run executes `carl init` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the init logic rooted at rootDir.
// Exported for testing without changing the process working directory.
func (c *Command) RunInDir(rootDir string) error {
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

	var conflicts []string
	for _, f := range files {
		target := filepath.Join(rootDir, filepath.FromSlash(f))
		if _, err := os.Stat(target); err == nil {
			conflicts = append(conflicts, f)
		}
	}
	if len(conflicts) > 0 {
		msg := "cARL artefacts already exist — remove them first or run `carl repair`:\n"
		for _, c := range conflicts {
			msg += fmt.Sprintf("  %s\n", c)
		}
		return fmt.Errorf("%s", msg)
	}

	// Install all embedded artefacts.
	for _, f := range files {
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
		RuntimeVersion:   runtimeVersion,
		Source:           source,
		SourceTag:        sourceTag,
		SourceCommit:     c.sourceCommit,
		InstalledAt:      time.Now().UTC(),
		ManagedArtifacts: files,
	}
	if err := manifest.Write(rootDir, rt); err != nil {
		return fmt.Errorf("write runtime manifest: %w", err)
	}

	fmt.Printf("cARL runtime installed successfully.\n")
	fmt.Printf("  Runtime version:  %s\n", rt.RuntimeVersion)
	fmt.Printf("  Source:           %s @ %s\n", rt.Source, rt.SourceTag)
	fmt.Printf("  Artefacts:        %d files installed\n", len(files))
	return nil
}
