package convert

import (
	"sort"
	"strings"
)

// Memory migration writes durable-memory and governance items into a single
// managed block in memory.md, delimited by the markers below. The block is
// owned by `carl convert` and is fully regenerated on each apply, so existing
// migrated entries are preserved and new ones merged in without duplication.
const (
	memBeginMarker = "<!-- BEGIN GENERATED: convert aadlc -->"
	memEndMarker   = "<!-- END GENERATED: convert aadlc -->"
)

// migratedEntries holds the durable-memory and governance bullet entries
// currently recorded in the managed convert block.
type migratedEntries struct {
	memory     []string
	governance []string
}

// extractMigratedEntries parses the managed convert block out of memory.md
// content and returns the durable-memory and governance entries it lists.
// When no block is present the result is empty.
func extractMigratedEntries(memory string) migratedEntries {
	begin := strings.Index(memory, memBeginMarker)
	end := strings.Index(memory, memEndMarker)
	if begin < 0 || end < 0 || end < begin {
		return migratedEntries{}
	}
	block := memory[begin:end]

	var entries migratedEntries
	active := ""
	for _, raw := range strings.Split(block, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### "):
			active = strings.ToLower(strings.TrimSpace(trimmed[len("### "):]))
		default:
			if text, ok := bulletText(trimmed); ok {
				if strings.Contains(active, "governance") {
					entries.governance = append(entries.governance, text)
				} else {
					entries.memory = append(entries.memory, text)
				}
			}
		}
	}
	return entries
}

// renderMigratedBlock builds the full managed convert block (including markers
// and a trailing newline) from the given entries. Entries are sorted for
// deterministic output. Empty subsections are omitted.
func renderMigratedBlock(entries migratedEntries) string {
	memory := dedupeSorted(entries.memory)
	governance := dedupeSorted(entries.governance)

	var b strings.Builder
	b.WriteString(memBeginMarker + "\n")
	b.WriteString("## Migrated from AADLC\n\n")
	b.WriteString("The following durable knowledge was migrated from legacy AADLC artefacts\n")
	b.WriteString("by `carl convert aadlc`. Do not edit inside the markers — re-running\n")
	b.WriteString("convert is idempotent and will not duplicate these entries.\n")

	if len(memory) > 0 {
		b.WriteString("\n### Durable memory\n\n")
		for _, e := range memory {
			b.WriteString("- " + e + "\n")
		}
	}
	if len(governance) > 0 {
		b.WriteString("\n### Governance rules\n\n")
		for _, e := range governance {
			b.WriteString("- " + e + "\n")
		}
	}

	b.WriteString(memEndMarker + "\n")
	return b.String()
}

// injectMigratedBlock inserts or replaces the managed convert block in memory.
// If the block already exists it is replaced in place; otherwise the block is
// appended at the end of the file after a blank-line separator.
func injectMigratedBlock(memory, block string) string {
	begin := strings.Index(memory, memBeginMarker)
	end := strings.Index(memory, memEndMarker)

	if begin >= 0 && end > begin {
		endOfEnd := end + len(memEndMarker)
		if endOfEnd < len(memory) && memory[endOfEnd] == '\n' {
			endOfEnd++
		}
		return memory[:begin] + block + memory[endOfEnd:]
	}

	switch {
	case strings.HasSuffix(memory, "\n\n"):
		return memory + block
	case strings.HasSuffix(memory, "\n"):
		return memory + "\n" + block
	default:
		return memory + "\n\n" + block
	}
}

// dedupeSorted returns a sorted copy of in with normalised duplicates removed,
// preserving the first textual form seen for each normalised key.
func dedupeSorted(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		key := normalize(s)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
