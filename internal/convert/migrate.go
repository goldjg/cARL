package convert

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Canonical cARL destination artefacts (repository-relative).
const (
	invariantsFile = ".github/carl/invariants.yml"
	memoryFile     = ".github/carl/memory.md"
)

// idPrefix namespaces invariant ids generated from migrated AADLC content so
// they are easy to recognise and unlikely to collide with native cARL ids.
const idPrefix = "aadlc-"

// conflict records a migration item that cannot be safely applied because it
// collides with — but is not identical to — existing cARL content. Conflicts
// are reported and never written; the human must reconcile them.
type conflict struct {
	item   Item
	reason string
}

// migrationPlan is the converter-agnostic result of analysing a set of items
// against the current cARL artefacts. It is the single source for both the
// dry-run report and the apply step.
type migrationPlan struct {
	source     string
	artefacts  []Artefact
	invariants []invariant // new invariant entries to write
	memory     []string    // new durable-memory entries to write
	governance []string    // new governance entries to write
	duplicates []Item      // items already present (skipped)
	conflicts  []conflict  // items requiring human review

	updated map[string]bool // destination files that would change

	memoryContent      string // existing memory.md content
	invariantsContent  string // existing invariants.yml content
	existingMigrations migratedEntries
}

// RunInDir runs the converter against rootDir and prints a migration report.
// When apply is false (dry-run) no files are modified. When apply is true the
// planned changes are written to the canonical cARL artefacts.
//
// Exported for testing without changing the process working directory.
func RunInDir(rootDir string, conv Converter, apply bool) error {
	artefacts, err := conv.Discover(rootDir)
	if err != nil {
		return fmt.Errorf("discover %s artefacts: %w", conv.Name(), err)
	}

	items, err := conv.Classify(artefacts)
	if err != nil {
		return fmt.Errorf("classify %s artefacts: %w", conv.Name(), err)
	}

	plan, err := buildPlan(rootDir, conv.ID(), artefacts, items)
	if err != nil {
		return err
	}

	if apply {
		if err := plan.applyTo(rootDir); err != nil {
			return err
		}
	}

	fmt.Print(plan.report(conv, apply))
	return nil
}

// buildPlan analyses items against the existing cARL artefacts under rootDir,
// classifying each as a planned write, a duplicate, or a conflict.
func buildPlan(rootDir, source string, artefacts []Artefact, items []Item) (*migrationPlan, error) {
	plan := &migrationPlan{
		source:    source,
		artefacts: artefacts,
		updated:   map[string]bool{},
	}

	// Load existing cARL content. Missing files are treated as empty, which
	// lets convert run before the destinations are fully populated.
	invContent, err := readOptional(filepath.Join(rootDir, filepath.FromSlash(invariantsFile)))
	if err != nil {
		return nil, err
	}
	memContent, err := readOptional(filepath.Join(rootDir, filepath.FromSlash(memoryFile)))
	if err != nil {
		return nil, err
	}
	plan.invariantsContent = invContent
	plan.memoryContent = memContent
	plan.existingMigrations = extractMigratedEntries(memContent)

	existingInvs := parseInvariants(invContent)
	existingRules := make(map[string]bool, len(existingInvs))
	existingIDs := make(map[string]string, len(existingInvs))
	for _, inv := range existingInvs {
		existingRules[normalize(inv.rule)] = true
		existingIDs[inv.id] = inv.rule
	}

	// Track ids assigned during this run to avoid in-run id collisions.
	assignedIDs := make(map[string]string)

	for _, it := range items {
		switch it.Category {
		case CategoryInvariant:
			plan.routeInvariant(it, existingRules, existingIDs, assignedIDs)
		default:
			plan.routeMemory(it)
		}
	}

	return plan, nil
}

// routeInvariant classifies a single invariant item.
func (p *migrationPlan) routeInvariant(it Item, existingRules map[string]bool, existingIDs map[string]string, assignedIDs map[string]string) {
	norm := normalize(it.Text)

	// Duplicate: an existing invariant already states this rule.
	if existingRules[norm] {
		p.duplicates = append(p.duplicates, it)
		return
	}

	id := idPrefix + slugify(it.Text)

	// Conflict: the generated id collides with an existing, different invariant.
	if rule, ok := existingIDs[id]; ok && normalize(rule) != norm {
		p.conflicts = append(p.conflicts, conflict{
			item:   it,
			reason: fmt.Sprintf("invariant id %q already exists with different content", id),
		})
		return
	}
	// Conflict: two migrated invariants reduce to the same id but differ.
	if text, ok := assignedIDs[id]; ok && normalize(text) != norm {
		p.conflicts = append(p.conflicts, conflict{
			item:   it,
			reason: fmt.Sprintf("two AADLC invariants map to the same id %q", id),
		})
		return
	}

	assignedIDs[id] = it.Text
	p.invariants = append(p.invariants, invariant{
		id:       id,
		name:     deriveName(it.Text),
		rule:     it.Text,
		severity: "high",
	})
	p.updated[invariantsFile] = true
}

// routeMemory classifies a single durable-memory or governance item.
func (p *migrationPlan) routeMemory(it Item) {
	// Duplicate: the statement already appears in memory.md (human-authored
	// prose or a previously migrated entry).
	if memoryContains(p.memoryContent, it.Text) {
		p.duplicates = append(p.duplicates, it)
		return
	}

	if it.Category == CategoryGovernance {
		p.governance = append(p.governance, it.Text)
	} else {
		p.memory = append(p.memory, it.Text)
	}
	p.updated[memoryFile] = true
}

// applyTo writes the planned changes to the canonical artefacts under rootDir.
// invariants.yml gains appended entries; memory.md's managed convert block is
// regenerated to include both existing and newly migrated entries. Files with
// no planned changes are left untouched.
func (p *migrationPlan) applyTo(rootDir string) error {
	if len(p.invariants) > 0 {
		path := filepath.Join(rootDir, filepath.FromSlash(invariantsFile))
		if p.invariantsContent == "" {
			return fmt.Errorf("%s not found — run `carl init` first", invariantsFile)
		}
		updated := appendInvariants(p.invariantsContent, p.invariants)
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return fmt.Errorf("write %s: %w", invariantsFile, err)
		}
	}

	if len(p.memory) > 0 || len(p.governance) > 0 {
		path := filepath.Join(rootDir, filepath.FromSlash(memoryFile))
		if p.memoryContent == "" {
			return fmt.Errorf("%s not found — run `carl init` first", memoryFile)
		}
		merged := migratedEntries{
			memory:     append(append([]string{}, p.existingMigrations.memory...), p.memory...),
			governance: append(append([]string{}, p.existingMigrations.governance...), p.governance...),
		}
		block := renderMigratedBlock(merged)
		updated := injectMigratedBlock(p.memoryContent, block)
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return fmt.Errorf("write %s: %w", memoryFile, err)
		}
	}
	return nil
}

// report renders the deterministic migration report. The same report is
// produced for dry-run and apply; only the trailing action note differs.
func (p *migrationPlan) report(conv Converter, applied bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s Migration Report\n\n", conv.Name())

	fmt.Fprintf(&b, "Discovered:\n")
	fmt.Fprintf(&b, "  %d artefact(s)\n", len(p.artefacts))
	for _, a := range p.artefacts {
		fmt.Fprintf(&b, "    %s\n", a.Path)
	}

	fmt.Fprintf(&b, "\nConvertible:\n")
	fmt.Fprintf(&b, "  %d invariant(s)\n", len(p.invariants))
	fmt.Fprintf(&b, "  %d memory entry(ies)\n", len(p.memory))
	fmt.Fprintf(&b, "  %d governance rule(s)\n", len(p.governance))

	fmt.Fprintf(&b, "\nSkipped:\n")
	fmt.Fprintf(&b, "  %d duplicate(s)\n", len(p.duplicates))

	fmt.Fprintf(&b, "\nConflicts:\n")
	fmt.Fprintf(&b, "  %d item(s) requiring review\n", len(p.conflicts))
	for _, c := range p.conflicts {
		fmt.Fprintf(&b, "    %s — %s\n", c.item.Text, c.reason)
	}

	updated := sortedKeys(p.updated)
	fmt.Fprintf(&b, "\n%s:\n", updatedHeading(applied))
	if len(updated) == 0 {
		fmt.Fprintf(&b, "  (none)\n")
	}
	for _, f := range updated {
		fmt.Fprintf(&b, "  %s\n", f)
	}

	if len(p.artefacts) == 0 {
		fmt.Fprintf(&b, "\nNo %s artefacts found — nothing to migrate.\n", conv.Name())
		return b.String()
	}

	if applied {
		fmt.Fprintf(&b, "\nMigration applied.\n")
	} else {
		fmt.Fprintf(&b, "\nDry run — no changes written. Re-run with --apply to migrate.\n")
	}
	return b.String()
}

// updatedHeading returns the heading for the destinations section, reflecting
// whether the changes were applied or merely planned.
func updatedHeading(applied bool) string {
	if applied {
		return "Updated"
	}
	return "Would update"
}

// readOptional returns a file's content, or an empty string if it is absent.
func readOptional(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

// sortedKeys returns the keys of m in lexicographic order.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// normalize reduces a statement to a comparison key: lowercased, with internal
// whitespace collapsed and surrounding whitespace and trailing punctuation
// removed. Used for duplicate and conflict detection so cosmetic differences
// do not defeat de-duplication.
func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimRight(s, ".!,;: ")
}

// memoryContains reports whether memory already states text, comparing on the
// normalised form so cosmetic differences are ignored.
func memoryContains(memory, text string) bool {
	target := normalize(text)
	if target == "" {
		return false
	}
	return strings.Contains(normalize(memory), target)
}

var slugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts text into a stable, url-safe id fragment: lowercased,
// non-alphanumeric runs collapsed to single hyphens, trimmed, and truncated to
// a bounded length so ids stay readable.
func slugify(text string) string {
	s := slugInvalid.ReplaceAllString(strings.ToLower(text), "-")
	s = strings.Trim(s, "-")
	const max = 50
	if len(s) > max {
		s = strings.Trim(s[:max], "-")
	}
	if s == "" {
		return "item"
	}
	return s
}

// deriveName produces a short human-readable name from a rule statement: the
// first sentence (or the whole text), truncated to a bounded length.
func deriveName(text string) string {
	name := strings.TrimSpace(text)
	if idx := strings.IndexAny(name, ".!?"); idx > 0 {
		name = strings.TrimSpace(name[:idx])
	}
	const max = 72
	if len(name) > max {
		name = strings.TrimSpace(name[:max])
	}
	return name
}
