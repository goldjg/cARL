// Package doctor implements the `carl doctor` command.
package doctor

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/goldjg/carl/internal/harness"
	"github.com/goldjg/carl/internal/manifest"
	"github.com/goldjg/carl/internal/repair"
)

// Level classifies the severity of a diagnostic finding.
type Level string

const (
	// LevelError indicates a condition that prevents normal operation.
	LevelError Level = "ERROR"
	// LevelWarning indicates a condition that should be addressed but does not
	// prevent operation.
	LevelWarning Level = "WARNING"
	// LevelInfo indicates a neutral or informational observation.
	LevelInfo Level = "INFO"
)

// Finding represents a single diagnostic observation.
type Finding struct {
	Level   Level
	Message string
	Action  string
}

// Artifacts provides read access to the embedded canonical runtime files.
type Artifacts interface {
	Open(targetPath string) ([]byte, error)
}

// Command implements `carl doctor`.
type Command struct {
	arts Artifacts
}

// New returns a new doctor Command backed by the given Artifacts.
func New(arts Artifacts) *Command {
	return &Command{arts: arts}
}

// Name returns the command name.
func (c *Command) Name() string { return "doctor" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Diagnose runtime issues and provide actionable remediation guidance"
}

// Run executes `carl doctor` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the doctor logic rooted at rootDir.
// Exported for testing without changing the process working directory.
// Returns nil (success) regardless of whether findings are reported; the
// command is diagnostic only and does not modify any files.
func (c *Command) RunInDir(rootDir string) error {
	findings, err := c.diagnose(rootDir)
	if err != nil {
		return err
	}
	printFindings(findings)
	return nil
}

// diagnose runs all checks and returns the resulting findings.
func (c *Command) diagnose(rootDir string) ([]Finding, error) {
	var findings []Finding

	if !manifest.Exists(rootDir) {
		findings = append(findings, Finding{
			Level:   LevelError,
			Message: "missing runtime manifest (.github/carl/runtime.json)",
			Action:  "run `carl init`",
		})
		return findings, nil
	}

	rt, err := manifest.Read(rootDir)
	if err != nil {
		findings = append(findings, Finding{
			Level:   LevelError,
			Message: fmt.Sprintf("runtime manifest is unreadable: %v", err),
			Action:  "run `carl init` after removing the corrupted manifest",
		})
		return findings, nil
	}

	if rt.RuntimeVersion == "" {
		findings = append(findings, Finding{
			Level:   LevelWarning,
			Message: "runtime manifest is missing runtimeVersion",
			Action:  "run `carl init` to reinstall with a valid manifest",
		})
	}

	missing, drifted, err := repair.Inspect(rootDir, rt.ManagedArtifacts, c.arts)
	if err != nil {
		return nil, fmt.Errorf("inspect runtime: %w", err)
	}

	for _, f := range missing {
		findings = append(findings, Finding{
			Level:   LevelError,
			Message: fmt.Sprintf("%s — artefact is missing from disk", f),
			Action:  "run `carl repair`",
		})
	}

	for _, f := range drifted {
		findings = append(findings, Finding{
			Level:   LevelWarning,
			Message: fmt.Sprintf("%s — artefact has drifted from its canonical version", f),
			Action:  "run `carl repair`",
		})
	}

	harnessHealth, err := harness.Inspect(rootDir, c.arts)
	if err != nil {
		return nil, fmt.Errorf("inspect harness adapters: %w", err)
	}
	for _, h := range harnessHealth {
		switch h.Sync {
		case harness.SyncMissing:
			findings = append(findings, Finding{
				Level:   LevelWarning,
				Message: fmt.Sprintf("%s (%s) — harness adapter file is missing", h.Adapter.ID, strings.Join(h.MissingFiles, ", ")),
				Action:  "run `carl harness sync`",
			})
		case harness.SyncDrifted:
			findings = append(findings, Finding{
				Level:   LevelWarning,
				Message: fmt.Sprintf("%s (%s) — harness adapter file has drifted from its canonical version", h.Adapter.ID, strings.Join(h.DriftedFiles, ", ")),
				Action:  "run `carl harness sync`",
			})
		}
	}

	if len(findings) == 0 {
		findings = append(findings, Finding{
			Level:   LevelInfo,
			Message: "runtime is healthy — all managed artefacts are present and canonical",
		})
	}

	return findings, nil
}

// printFindings writes the findings to stdout in a consistent, human-readable
// format. It summarises error and warning counts at the end unless only INFO
// findings are present.
func printFindings(findings []Finding) {
	errors, warnings, infos := 0, 0, 0
	for _, f := range findings {
		switch f.Level {
		case LevelError:
			errors++
		case LevelWarning:
			warnings++
		default:
			infos++
		}
	}

	for _, f := range findings {
		// %-7s pads to 7 characters — the width of "WARNING", the longest level.
		fmt.Printf("%-7s %s\n", string(f.Level), f.Message)
		if f.Action != "" {
			fmt.Printf("        Action: %s\n", f.Action)
		}
	}

	onlyInfo := errors == 0 && warnings == 0
	if !onlyInfo {
		fmt.Println()
		fmt.Printf("%d error(s), %d warning(s), %d info(s) found.\n", errors, warnings, infos)
	}
}
