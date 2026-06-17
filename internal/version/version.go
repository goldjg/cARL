// Package version implements the `carl version` command.
package version

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/goldjg/carl/internal/manifest"
)

// Command implements `carl version`.
type Command struct {
	cliVersion string
}

// New returns a new version Command.
// cliVersion is the CLI binary version set at build time via -ldflags.
func New(cliVersion string) *Command {
	return &Command{cliVersion: cliVersion}
}

// Name returns the command name.
func (c *Command) Name() string { return "version" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Show CLI and installed runtime version information"
}

// Run executes `carl version` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the version logic rooted at rootDir.
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
	if len(packs) > 0 {
		fmt.Println("Installed Packs:")
		for _, p := range packs {
			fmt.Printf("  %s\n", p)
		}
		fmt.Println()
	}

	fmt.Println("Runtime Status:")
	fmt.Printf("  %s\n", runtimeStatus(rootDir, rt))
	return nil
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
			// Did not have the expected suffix — skip.
			continue
		}
		pack := parts[2] + "/" + name
		if !seen[pack] {
			seen[pack] = true
		}
	}

	// Sort for deterministic output.
	result := make([]string, 0, len(seen))
	for p := range seen {
		result = append(result, p)
	}
	sortStrings(result)
	return result
}

// runtimeStatus returns a one-line status string.
// "Healthy" means all non-protected managed artefacts are present.
func runtimeStatus(rootDir string, rt *manifest.Runtime) string {
	missing := 0
	for _, f := range rt.ManagedArtifacts {
		if f == manifest.FileName || f == ".github/carl/memory.md" {
			continue
		}
		target := rootDir + "/" + strings.ReplaceAll(f, "\\", "/")
		if _, err := os.Stat(target); os.IsNotExist(err) {
			missing++
		}
	}
	if missing == 0 {
		return "Healthy"
	}
	return fmt.Sprintf("Drift detected (%d artefact(s) missing)", missing)
}

// sortStrings sorts a string slice in place using insertion sort.
// Used to avoid importing "sort" for a small slice.
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
