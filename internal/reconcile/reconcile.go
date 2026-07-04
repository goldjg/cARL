// Package reconcile implements the `carl reconcile` command.
// It updates repository-specific memory sections in .github/carl/memory.md
// using data from .github/carl/repo-map.json, preserving human-authored content.
package reconcile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goldjg/carl/internal/repomap"
)

// MemoryFile is the path of memory.md relative to the repository root.
const MemoryFile = ".github/carl/memory.md"

// genBeginMarker and genEndMarker delimit the generated section in memory.md.
// Content between these markers is owned by `carl reconcile` and will be
// overwritten on each run. Do not place human-authored content inside them.
const (
	genBeginMarker = "<!-- BEGIN GENERATED: reconcile -->"
	genEndMarker   = "<!-- END GENERATED: reconcile -->"
)

// Command implements `carl reconcile`.
type Command struct{}

// New returns a new reconcile Command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "reconcile" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Update repository-specific memory sections from current repo-map data"
}

// Run executes `carl reconcile` in the current working directory.
func (c *Command) Run(_ context.Context, args []string) error {
	repairMarkers := false
	for _, a := range args {
		switch a {
		case "--repair-markers":
			repairMarkers = true
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown flag %q for `carl reconcile`\n\nRun 'carl reconcile --help' for usage", a)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDirWithOptions(cwd, repairMarkers)
}

// printUsage prints `carl reconcile` usage to stdout.
func printUsage() {
	fmt.Println("Usage: carl reconcile [--repair-markers]")
	fmt.Println()
	fmt.Println("Update repository-specific memory sections from current repo-map data.")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --repair-markers  Remove a single orphaned generated-section marker line")
	fmt.Println("                    (a BEGIN with no END, or an END with no BEGIN) before")
	fmt.Println("                    reconciling. Never removes any other content. Fails if")
	fmt.Println("                    the marker state is ambiguous.")
}

// RunInDir executes the reconcile logic rooted at rootDir, without repairing
// any orphaned markers. Exported for testing without changing the process
// working directory.
func (c *Command) RunInDir(rootDir string) error {
	return c.RunInDirWithOptions(rootDir, false)
}

// RunInDirWithOptions executes the reconcile logic rooted at rootDir.
// When repairMarkers is true and memory.md contains exactly one orphaned
// generated-section marker (a BEGIN with no END, or an END with no BEGIN),
// that single marker line is removed before reconciling proceeds. No other
// content is ever touched by this repair. Exported for testing without
// changing the process working directory or going through Run's flag parsing.
func (c *Command) RunInDirWithOptions(rootDir string, repairMarkers bool) error {
	// Verify repo-map.json exists.
	mapPath := filepath.Join(rootDir, filepath.FromSlash(repomap.OutputFile))
	mapData, err := os.ReadFile(mapPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repo map not found — run `carl map` first")
		}
		return fmt.Errorf("read repo map: %w", err)
	}

	// Verify memory.md exists.
	memPath := filepath.Join(rootDir, filepath.FromSlash(MemoryFile))
	memData, err := os.ReadFile(memPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("memory.md not found — run `carl init` first")
		}
		return fmt.Errorf("read memory.md: %w", err)
	}

	// Parse repo map.
	var m repomap.Map
	if err := json.Unmarshal(mapData, &m); err != nil {
		return fmt.Errorf("parse repo map: %w", err)
	}

	memContent := string(memData)

	// Validate that any existing generated-section markers are well-formed.
	if err := checkMarkers(memContent); err != nil {
		if !repairMarkers {
			return err
		}
		repaired, repairErr := repairOrphanMarker(memContent)
		if repairErr != nil {
			return fmt.Errorf("%w; --repair-markers could not resolve this automatically: %v", err, repairErr)
		}
		memContent = repaired
	}

	// Build and inject the generated section.
	generated := buildGeneratedSection(&m)
	newContent, changed := injectSection(memContent, generated)

	if !changed {
		fmt.Println("No reconciliation needed.")
		return nil
	}

	if err := os.WriteFile(memPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("write memory.md: %w", err)
	}

	fmt.Println("Reconciled durable artefacts.")
	fmt.Printf("  %s\n", MemoryFile)
	return nil
}

// buildGeneratedSection produces the full generated block including its markers.
// The block opens with genBeginMarker and closes with genEndMarker, each on
// their own line. A trailing newline is always appended after genEndMarker.
func buildGeneratedSection(m *repomap.Map) string {
	var b strings.Builder

	b.WriteString(genBeginMarker + "\n")
	b.WriteString("## Repository snapshot\n\n")
	b.WriteString("This section is regenerated by `carl reconcile`. Do not edit manually.\n\n")

	if len(m.Languages) > 0 {
		b.WriteString("**Languages:** " + strings.Join(m.Languages, ", ") + "  \n")
	}
	b.WriteString("**Last reconciled:** " + time.Now().UTC().Format("2006-01-02") + "\n")

	if len(m.EntryPoints) > 0 {
		b.WriteString("\n### Entry points\n\n")
		for _, ep := range m.EntryPoints {
			if ep.Purpose != "" {
				b.WriteString(fmt.Sprintf("- `%s` — %s\n", ep.Path, ep.Purpose))
			} else {
				b.WriteString(fmt.Sprintf("- `%s`\n", ep.Path))
			}
		}
	}

	if len(m.Directories) > 0 {
		b.WriteString("\n### Key directories\n\n")
		keys := make([]string, 0, len(m.Directories))
		for k := range m.Directories {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := m.Directories[k]
			if v != "" {
				b.WriteString(fmt.Sprintf("- `%s` — %s\n", k, v))
			} else {
				b.WriteString(fmt.Sprintf("- `%s`\n", k))
			}
		}
	}

	if len(m.Workflows) > 0 {
		b.WriteString("\n### Workflows\n\n")
		for _, wf := range m.Workflows {
			if wf.Purpose != "" {
				b.WriteString(fmt.Sprintf("- `%s` — %s\n", wf.Path, wf.Purpose))
			} else {
				b.WriteString(fmt.Sprintf("- `%s`\n", wf.Path))
			}
		}
	}

	if len(m.Governance) > 0 {
		b.WriteString("\n### Governance artefacts\n\n")
		for _, g := range m.Governance {
			if g.Purpose != "" {
				b.WriteString(fmt.Sprintf("- `%s` — %s\n", g.Path, g.Purpose))
			} else {
				b.WriteString(fmt.Sprintf("- `%s`\n", g.Path))
			}
		}
	}

	if len(m.Documentation) > 0 {
		b.WriteString("\n### Documentation\n\n")
		for _, d := range m.Documentation {
			if d.Purpose != "" {
				b.WriteString(fmt.Sprintf("- `%s` — %s\n", d.Path, d.Purpose))
			} else {
				b.WriteString(fmt.Sprintf("- `%s`\n", d.Path))
			}
		}
	}

	b.WriteString(genEndMarker + "\n")
	return b.String()
}

// markerAnalysis records every occurrence of the generated-section markers
// found in a memory.md content string, in document order.
type markerAnalysis struct {
	begins []int
	ends   []int
}

// analyzeMarkers scans content for all occurrences of genBeginMarker and
// genEndMarker.
func analyzeMarkers(content string) markerAnalysis {
	return markerAnalysis{
		begins: markerOccurrences(content, genBeginMarker),
		ends:   markerOccurrences(content, genEndMarker),
	}
}

// markerOccurrences returns the byte offsets of every non-overlapping
// occurrence of marker in content, in ascending order.
func markerOccurrences(content, marker string) []int {
	var idxs []int
	offset := 0
	for {
		i := strings.Index(content[offset:], marker)
		if i < 0 {
			return idxs
		}
		idxs = append(idxs, offset+i)
		offset += i + len(marker)
	}
}

// empty reports whether no generated-section markers of either kind are
// present.
func (a markerAnalysis) empty() bool {
	return len(a.begins) == 0 && len(a.ends) == 0
}

// wellFormedPairs reports whether every begin/end marker forms a strictly
// alternating, non-overlapping sequence of matched pairs:
// begins[0] < ends[0] < begins[1] < ends[1] < ...
// A single matched pair also qualifies. Content strictly between any matched
// pair is cARL-owned generated content (see genBeginMarker/genEndMarker
// doc comment), so any number of well-formed pairs — including duplicates —
// can be safely collapsed without risk to human-authored content.
func (a markerAnalysis) wellFormedPairs() bool {
	if len(a.begins) == 0 || len(a.begins) != len(a.ends) {
		return false
	}
	for i := range a.begins {
		if a.begins[i] >= a.ends[i] {
			return false
		}
		if i > 0 && a.ends[i-1] >= a.begins[i] {
			return false
		}
	}
	return true
}

// singleOrphanMarker reports whether content has exactly one marker of one
// kind and none of the other — the only marker state that
// `--repair-markers` will resolve automatically, since there is exactly one
// unambiguous, cARL-owned marker line to remove.
func (a markerAnalysis) singleOrphanMarker() (marker string, idx int, ok bool) {
	switch {
	case len(a.begins) == 1 && len(a.ends) == 0:
		return genBeginMarker, a.begins[0], true
	case len(a.begins) == 0 && len(a.ends) == 1:
		return genEndMarker, a.ends[0], true
	default:
		return "", 0, false
	}
}

// checkMarkers returns an error when the generated-section markers in content
// are in a malformed state — any shape other than "no markers at all" or
// "one or more well-formed, non-overlapping BEGIN/END pairs" (see
// wellFormedPairs). Malformed states include an orphaned marker (one kind
// present without the other), an end marker appearing before its begin
// marker, or any other unpaired/overlapping combination.
//
// In any malformed case the caller should not write anything; the user must
// repair memory.md manually, or — when the state is a single orphaned marker
// — re-run with `carl reconcile --repair-markers`.
func checkMarkers(content string) error {
	a := analyzeMarkers(content)
	if a.empty() || a.wellFormedPairs() {
		return nil
	}

	const manualAdvice = "repair the markers manually"
	const orphanAdvice = manualAdvice + ", or run `carl reconcile --repair-markers` " +
		"to remove the single orphaned marker line automatically"

	switch {
	case len(a.begins) == 1 && len(a.ends) == 0:
		return fmt.Errorf("memory.md contains %q but is missing %q — "+
			"the generated section markers are malformed; %s",
			genBeginMarker, genEndMarker, orphanAdvice)
	case len(a.ends) == 1 && len(a.begins) == 0:
		return fmt.Errorf("memory.md contains %q but is missing %q — "+
			"the generated section markers are malformed; %s",
			genEndMarker, genBeginMarker, orphanAdvice)
	case len(a.begins) == 1 && len(a.ends) == 1 && a.ends[0] < a.begins[0]:
		return fmt.Errorf("memory.md has the end marker appearing before the begin marker — "+
			"the generated section markers are malformed; %s", manualAdvice)
	default:
		return fmt.Errorf("memory.md has %d %q marker(s) and %d %q marker(s) that do not form "+
			"well-formed pairs — the generated section markers are malformed; %s",
			len(a.begins), genBeginMarker, len(a.ends), genEndMarker, manualAdvice)
	}
}

// repairOrphanMarker removes exactly one unpaired generated-section marker
// line from content — a BEGIN with no END, or an END with no BEGIN — and
// nothing else. It returns an error, touching no content, if content is not
// in that exact single-orphan-marker state; ambiguous marker states (e.g.
// an end-before-begin pair, or multiple stray markers) are never guessed at.
func repairOrphanMarker(content string) (string, error) {
	a := analyzeMarkers(content)
	marker, idx, ok := a.singleOrphanMarker()
	if !ok {
		return "", fmt.Errorf("marker state is ambiguous (%d begin marker(s), %d end marker(s)); "+
			"--repair-markers only handles a single orphaned marker line",
			len(a.begins), len(a.ends))
	}

	end := idx + len(marker)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	return content[:idx] + content[end:], nil
}

// spanEnd returns the byte offset immediately after the generated-section end
// marker at endIdx, including its trailing newline if present.
func spanEnd(content string, endIdx int) int {
	end := endIdx + len(genEndMarker)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	return end
}

// injectSection inserts or replaces the generated section in memoryContent.
// It returns the updated content and a boolean indicating whether the content
// changed. The caller must have already validated markers via checkMarkers
// (after any --repair-markers pass), so memoryContent here always has either
// no markers or one-or-more well-formed BEGIN/END pairs.
//
// Replacement: when one or more well-formed pairs are present, every matched
// pair's span (including its markers and trailing newline) is removed and a
// single generatedSection is inserted at the position of the first pair —
// collapsing duplicates onto one canonical block. Any content between
// consecutive pairs (e.g. stray text separating duplicated blocks) is
// preserved in place. If exactly one pair was present and its span is
// byte-identical to generatedSection, the content is unchanged and
// changed=false is returned.
//
// Insertion: if no markers exist, generatedSection is inserted before the last
// "## Last updated" heading (preceded by a blank line) so the human-authored
// last-updated note stays at the end. If no such heading exists, generatedSection
// is appended after a blank line at the end of the file.
func injectSection(memoryContent, generatedSection string) (string, bool) {
	a := analyzeMarkers(memoryContent)

	if len(a.begins) > 0 && a.wellFormedPairs() {
		if len(a.begins) == 1 {
			existing := memoryContent[a.begins[0]:spanEnd(memoryContent, a.ends[0])]
			if existing == generatedSection {
				return memoryContent, false
			}
		}

		var b strings.Builder
		b.WriteString(memoryContent[:a.begins[0]])
		b.WriteString(generatedSection)
		cursor := spanEnd(memoryContent, a.ends[0])
		for i := 1; i < len(a.begins); i++ {
			b.WriteString(memoryContent[cursor:a.begins[i]])
			cursor = spanEnd(memoryContent, a.ends[i])
		}
		b.WriteString(memoryContent[cursor:])
		return b.String(), true
	}

	// No existing generated section — insert before "## Last updated" or at end.
	const lastUpdatedHeader = "\n## Last updated"
	if insertIdx := strings.LastIndex(memoryContent, lastUpdatedHeader); insertIdx >= 0 {
		return memoryContent[:insertIdx] + "\n" + generatedSection + memoryContent[insertIdx:], true
	}

	// Fallback: append at end, ensuring a blank-line separator.
	if !strings.HasSuffix(memoryContent, "\n\n") {
		if strings.HasSuffix(memoryContent, "\n") {
			return memoryContent + "\n" + generatedSection, true
		}
		return memoryContent + "\n\n" + generatedSection, true
	}
	return memoryContent + generatedSection, true
}
