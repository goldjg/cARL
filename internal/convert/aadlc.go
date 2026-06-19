package convert

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// aadlcConverter migrates artefacts from legacy AADLC (Autonomous Agent
// Development Lifecycle) repositories. cARL is the productised form of AADLC,
// so many repositories already carry durable invariants, lessons, and
// governance rules under AADLC paths that should be preserved on adoption.
type aadlcConverter struct{}

func (aadlcConverter) ID() string   { return "aadlc" }
func (aadlcConverter) Name() string { return "AADLC" }

// aadlcSearchRoots are the conventional locations AADLC artefacts live in.
// Directories are scanned recursively for Markdown and YAML files; single
// files are read directly. Order is fixed for deterministic discovery.
var aadlcSearchRoots = []string{
	".aadlc",
	".github/aadlc",
	"aadlc",
	"AADLC.md",
}

// Discover finds AADLC artefacts under rootDir. Directories listed in
// aadlcSearchRoots are walked recursively; files are read directly. Only
// Markdown (.md) and YAML (.yml/.yaml) files are considered. Results are
// returned sorted by path. Missing roots are skipped silently.
func (a aadlcConverter) Discover(rootDir string) ([]Artefact, error) {
	seen := make(map[string]bool)
	var artefacts []Artefact

	add := func(rel string) error {
		if seen[rel] {
			return nil
		}
		content, err := os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		seen[rel] = true
		artefacts = append(artefacts, Artefact{Path: rel, Content: content})
		return nil
	}

	for _, root := range aadlcSearchRoots {
		abs := filepath.Join(rootDir, filepath.FromSlash(root))
		info, err := os.Stat(abs)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		if !info.IsDir() {
			if isConvertibleFile(root) {
				if err := add(root); err != nil {
					return nil, err
				}
			}
			continue
		}

		err = filepath.WalkDir(abs, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !isConvertibleFile(p) {
				return nil
			}
			rel, err := filepath.Rel(rootDir, p)
			if err != nil {
				return err
			}
			return add(filepath.ToSlash(rel))
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(artefacts, func(i, j int) bool { return artefacts[i].Path < artefacts[j].Path })
	return artefacts, nil
}

// isConvertibleFile reports whether path is a Markdown or YAML file.
func isConvertibleFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".yml", ".yaml":
		return true
	default:
		return false
	}
}

// Classify turns discovered AADLC artefacts into category-tagged items.
//
// YAML files containing an `invariants:` list (the AADLC/cARL invariant
// format) contribute their rules as invariant items. Markdown files are
// scanned section by section; each section heading is matched against
// category keywords, and the bullet-list items beneath a matched heading
// become migration items of that category.
//
// Content that cannot be confidently classified is ignored — the command
// prefers safety over speculative migration.
func (a aadlcConverter) Classify(artefacts []Artefact) ([]Item, error) {
	var items []Item
	for _, art := range artefacts {
		switch strings.ToLower(filepath.Ext(art.Path)) {
		case ".yml", ".yaml":
			items = append(items, classifyYAML(art)...)
		case ".md":
			items = append(items, classifyMarkdown(art)...)
		}
	}

	// Deterministic ordering and de-duplication within the classified set.
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Category != items[j].Category {
			return items[i].Category < items[j].Category
		}
		if items[i].Text != items[j].Text {
			return items[i].Text < items[j].Text
		}
		return items[i].Source < items[j].Source
	})

	deduped := items[:0]
	seen := make(map[string]bool)
	for _, it := range items {
		key := string(it.Category) + "\x00" + normalize(it.Text)
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, it)
	}
	return deduped, nil
}

// classifyYAML extracts invariant rules from an AADLC invariants YAML file.
// It reuses the cARL invariants parser, so any file using the canonical
// `invariants:` schema is understood. Files without that schema yield nothing.
func classifyYAML(art Artefact) []Item {
	invs := parseInvariants(string(art.Content))
	items := make([]Item, 0, len(invs))
	for _, inv := range invs {
		if strings.TrimSpace(inv.rule) == "" {
			continue
		}
		items = append(items, Item{
			Category: CategoryInvariant,
			Text:     strings.TrimSpace(inv.rule),
			Source:   art.Path,
		})
	}
	return items
}

// classifyMarkdown scans a Markdown artefact section by section. The category
// of each section is derived from its heading text; bullet-list items beneath
// a categorised heading become migration items. Headings with no recognised
// category reset the active category so unrelated prose is not migrated.
func classifyMarkdown(art Artefact) []Item {
	var items []Item
	active := Category("")

	for _, raw := range strings.Split(string(art.Content), "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") {
			active = categoryForHeading(trimmed)
			continue
		}

		if active == "" {
			continue
		}

		text, ok := bulletText(trimmed)
		if !ok {
			continue
		}
		items = append(items, Item{Category: active, Text: text, Source: art.Path})
	}
	return items
}

// bulletText returns the content of a Markdown list item (`-`, `*`, or `+`
// bullet) with the marker and surrounding whitespace stripped. The second
// return value is false when the line is not a non-empty list item.
func bulletText(line string) (string, bool) {
	for _, marker := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, marker) {
			text := strings.TrimSpace(line[len(marker):])
			if text == "" {
				return "", false
			}
			return text, true
		}
	}
	return "", false
}

// categoryForHeading maps a Markdown heading to a migration category based on
// keywords in the heading text. Governance keywords take precedence over the
// more general memory keywords. An unrecognised heading returns the empty
// category, meaning its content is not migrated.
func categoryForHeading(heading string) Category {
	h := strings.ToLower(strings.TrimSpace(strings.TrimLeft(heading, "#")))

	switch {
	case containsAny(h, "invariant", "constraint", "must remain", "must not"):
		return CategoryInvariant
	case containsAny(h, "governance", "pr contract", "pull request", "approval", "planning", "review requirement"):
		return CategoryGovernance
	case containsAny(h, "memory", "lesson", "decision", "mistake", "limitation", "history", "historical", "context", "known"):
		return CategoryMemory
	default:
		return ""
	}
}

// containsAny reports whether s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
