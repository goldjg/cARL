// Package status implements the `carl status` command.
package status

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/goldjg/carl/internal/manifest"
	"github.com/goldjg/carl/internal/repair"
)

// Artifacts provides read access to the embedded canonical runtime files.
type Artifacts interface {
	Open(targetPath string) ([]byte, error)
}

// Command implements `carl status`.
type Command struct {
	cliVersion string
	arts       Artifacts
}

// New returns a new status Command.
// cliVersion is the CLI binary version set at build time via -ldflags.
// arts is the embedded asset store used for content-based drift detection.
func New(cliVersion string, arts Artifacts) *Command {
	return &Command{cliVersion: cliVersion, arts: arts}
}

// Name returns the command name.
func (c *Command) Name() string { return "status" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Report whether the installed cARL runtime is healthy, missing, or drifted"
}

// Run executes `carl status` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the status logic rooted at rootDir.
// Exported for testing without changing the process working directory.
func (c *Command) RunInDir(rootDir string) error {
	if !manifest.Exists(rootDir) {
		fmt.Println("No cARL runtime installed.")
		return nil
	}

	rt, err := manifest.Read(rootDir)
	if err != nil {
		return fmt.Errorf("read runtime manifest: %w", err)
	}

	fmt.Printf("CLI Version:      %s\n", c.cliVersion)
	fmt.Printf("Runtime Version:  %s\n", rt.RuntimeVersion)
	fmt.Printf("Source:           %s\n", rt.Source)
	fmt.Printf("Tag:              %s\n", rt.SourceTag)
	fmt.Printf("Commit:           %s\n", rt.SourceCommit)
	fmt.Println()

	packs := extractPacks(rt.ManagedArtifacts)
	fmt.Println("Installed Packs:")
	if len(packs) == 0 {
		fmt.Println("  none")
	} else {
		for _, p := range packs {
			fmt.Printf("  %s\n", p)
		}
	}
	fmt.Println()

	missing, drifted, err := repair.Inspect(rootDir, rt.ManagedArtifacts, c.arts)
	if err != nil {
		return fmt.Errorf("inspect runtime: %w", err)
	}

	fmt.Println("Missing Artefacts:")
	if len(missing) == 0 {
		fmt.Println("  none")
	} else {
		for _, f := range missing {
			fmt.Printf("  %s\n", f)
		}
	}
	fmt.Println()

	fmt.Println("Drifted Artefacts:")
	if len(drifted) == 0 {
		fmt.Println("  none")
	} else {
		for _, f := range drifted {
			fmt.Printf("  %s\n", f)
		}
	}
	fmt.Println()

	fmt.Printf("Status:           %s\n", overallStatus(missing, drifted))
	return nil
}

// overallStatus derives the status label from the classification results.
// Missing artefacts (Incomplete) take precedence over content drift (Drifted).
func overallStatus(missing, drifted []string) string {
	if len(missing) > 0 {
		return "Incomplete"
	}
	if len(drifted) > 0 {
		return "Drifted"
	}
	return "Healthy"
}

// extractPacks derives pack names from managed artifact paths of the form
// ".github/instructions/<category>/<name>.instructions.md".
// Results are returned in lexicographic order.
func extractPacks(artifacts []string) []string {
	seen := map[string]bool{}
	for _, a := range artifacts {
		// Normalise to forward slashes for consistent matching.
		a = path.Clean(strings.ReplaceAll(a, "\\", "/"))
		// Match .github/instructions/<category>/<name>.instructions.md
		parts := strings.SplitN(a, "/", 5)
		if len(parts) != 4 {
			continue
		}
		if parts[0] != ".github" || parts[1] != "instructions" {
			continue
		}
		name := strings.TrimSuffix(parts[3], ".instructions.md")
		if name == parts[3] {
			continue
		}
		pack := parts[2] + "/" + name
		if !seen[pack] {
			seen[pack] = true
		}
	}

	result := make([]string, 0, len(seen))
	for p := range seen {
		result = append(result, p)
	}
	sortStrings(result)
	return result
}

// sortStrings sorts a string slice in place using insertion sort.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
