// Package plan implements the `carl plan` command.
// It discovers, validates, and summarises plan files in .github/carl/plans/.
package plan

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PlansDir is the path of the plans directory relative to the repository root.
const PlansDir = ".github/carl/plans"

// Plan represents a parsed plan file.
type Plan struct {
	// Filename is the base name of the plan file (e.g. "my-plan.md").
	Filename string
	// Title is extracted from the first # heading in the file.
	Title string
	// Status is the lifecycle state extracted from the ## Plan metadata section.
	// Typical values: Draft, Active, Completed, Archived.
	Status string
	// Purpose is extracted from the ## Task summary, ## Task, or ## Goal section.
	Purpose string
	// Warnings contains validation issues found during parsing.
	Warnings []string
}

// Command implements `carl plan`.
type Command struct{}

// New returns a new plan Command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "plan" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Discover, validate, and summarise plans in .github/carl/plans/"
}

// Run executes `carl plan` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the plan logic rooted at rootDir.
// Exported for testing without changing the process working directory.
// Always returns nil — the command is read-only and never modifies files.
func (c *Command) RunInDir(rootDir string) error {
	plans, err := Scan(rootDir)
	if err != nil {
		return err
	}
	printPlans(plans)
	return nil
}

// Scan discovers and parses all .md files in PlansDir under rootDir.
// Returns a slice of Plan sorted lexicographically by filename.
// Returns nil (not an error) if the directory does not exist or contains no .md files.
func Scan(rootDir string) ([]Plan, error) {
	dir := filepath.Join(rootDir, filepath.FromSlash(PlansDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read plans directory: %w", err)
	}

	var plans []Plan
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(e.Name())) != ".md" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read plan file %s: %w", e.Name(), err)
		}
		plans = append(plans, parse(e.Name(), data))
	}

	sort.Slice(plans, func(i, j int) bool {
		return plans[i].Filename < plans[j].Filename
	})
	return plans, nil
}

// parse parses a single plan file and returns a Plan struct.
func parse(filename string, data []byte) Plan {
	p := Plan{Filename: filename}

	sections := parseSections(data)

	p.Title = extractTitle(data)

	// Validate and extract the ## Plan metadata section.
	if metadata, ok := sections["Plan metadata"]; ok {
		p.Status = extractListField(metadata, "Status")
		if p.Status == "" {
			p.Warnings = append(p.Warnings, "Status not set in ## Plan metadata")
		}
	} else {
		p.Warnings = append(p.Warnings, "missing ## Plan metadata section")
	}

	// Extract purpose: prefer ## Task summary, then ## Task, then ## Goal.
	for _, name := range []string{"Task summary", "Task", "Goal"} {
		if content, ok := sections[name]; ok {
			if s := firstParagraph(content); s != "" {
				p.Purpose = s
				break
			}
		}
	}

	return p
}

// parseSections splits the document body into a map of level-2 section name
// (the text after "## ") to the lines that follow until the next ## heading.
// Only level-2 headings ("## ") define section boundaries.
func parseSections(data []byte) map[string][]string {
	sections := map[string][]string{}
	var current string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if _, exists := sections[current]; !exists {
				sections[current] = nil
			}
			continue
		}
		if current != "" {
			sections[current] = append(sections[current], line)
		}
	}
	return sections
}

// extractTitle returns the text of the first level-1 heading ("# ") in data.
// Returns "(not set)" if no level-1 heading is found.
func extractTitle(data []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return "(not set)"
}

// extractListField scans lines for a markdown list item of the form
// "- FieldName: <value>" and returns the trimmed value.
// Returns "" if the field is absent or its value is empty.
func extractListField(lines []string, field string) string {
	prefix := "- " + field + ":"
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

// firstParagraph returns the first non-empty paragraph from lines, joining
// continuation lines with a space. A paragraph ends at a blank line or at
// a line that starts a new section heading. Code fences (``` ... ```) are skipped.
func firstParagraph(lines []string) string {
	var parts []string
	inCode := false
	started := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		trimmed := strings.TrimSpace(line)
		// A heading line (any level) ends the first paragraph if we have started.
		if strings.HasPrefix(trimmed, "#") {
			if started {
				break
			}
			continue
		}
		if trimmed == "" {
			if started {
				break // End of first paragraph.
			}
			continue
		}
		parts = append(parts, trimmed)
		started = true
	}
	return strings.Join(parts, " ")
}

// printPlans writes the plan listing to stdout.
func printPlans(plans []Plan) {
	if len(plans) == 0 {
		fmt.Println("No plans found.")
		return
	}

	fmt.Printf("Plans in %s\n\n", PlansDir)

	totalWarnings := 0
	for _, p := range plans {
		status := p.Status
		if status == "" {
			status = "(not set)"
		}
		purpose := p.Purpose
		if purpose == "" {
			purpose = "(not set)"
		}

		fmt.Printf("  %s\n", p.Filename)
		fmt.Printf("    Title:    %s\n", p.Title)
		fmt.Printf("    Status:   %s\n", status)
		fmt.Printf("    Purpose:  %s\n", purpose)
		for _, w := range p.Warnings {
			fmt.Printf("    Warning:  %s\n", w)
			totalWarnings++
		}
		fmt.Println()
	}

	if totalWarnings > 0 {
		fmt.Printf("%d plan(s) found. %d warning(s).\n", len(plans), totalWarnings)
	} else {
		fmt.Printf("%d plan(s) found.\n", len(plans))
	}
}
