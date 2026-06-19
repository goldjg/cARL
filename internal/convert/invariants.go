package convert

import (
	"fmt"
	"strings"
)

// invariant is a minimal representation of a single entry in the cARL
// invariants.yml file. Only the fields cARL itself uses are modelled.
type invariant struct {
	id       string
	name     string
	rule     string
	severity string
}

// parseInvariants reads the cARL invariants.yml format into a slice of
// invariants. The format is a fixed two-space-indented list under an
// `invariants:` key, each entry exposing id/name/rule/severity scalars.
//
// The parser is intentionally small and tolerant: it understands both plain
// and double-quoted scalar values (this command always writes quoted values),
// ignores comments and blank lines, and silently skips fields it does not
// recognise. It is sufficient for round-tripping cARL invariants and for
// reading AADLC files that adopt the same schema.
func parseInvariants(content string) []invariant {
	var (
		result  []invariant
		current *invariant
	)

	flush := func() {
		if current != nil {
			result = append(result, *current)
			current = nil
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if key, value, ok := splitListEntry(line); ok {
			// A new list entry ("  - key: value") starts a fresh invariant.
			flush()
			current = &invariant{}
			assignField(current, key, value)
			continue
		}

		if current == nil {
			continue
		}
		if key, value, ok := splitField(trimmed); ok {
			assignField(current, key, value)
		}
	}
	flush()
	return result
}

// splitListEntry parses a line of the form "  - key: value" into key and
// value. The third return value is false when the line is not a list entry.
func splitListEntry(line string) (key, value string, ok bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- ") {
		return "", "", false
	}
	return splitField(strings.TrimSpace(trimmed[len("- "):]))
}

// splitField parses a "key: value" pair, unquoting the value when quoted.
func splitField(s string) (key, value string, ok bool) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(s[:idx])
	value = unquoteScalar(strings.TrimSpace(s[idx+1:]))
	return key, value, key != ""
}

// assignField stores a recognised scalar field on inv.
func assignField(inv *invariant, key, value string) {
	switch key {
	case "id":
		inv.id = value
	case "name":
		inv.name = value
	case "rule":
		inv.rule = value
	case "severity":
		inv.severity = value
	}
}

// unquoteScalar removes surrounding double quotes from a YAML scalar and
// unescapes \" and \\ sequences. Plain (unquoted) scalars are returned as-is.
func unquoteScalar(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		inner := s[1 : len(s)-1]
		inner = strings.ReplaceAll(inner, `\"`, `"`)
		inner = strings.ReplaceAll(inner, `\\`, `\`)
		return inner
	}
	return s
}

// quoteScalar renders a value as a double-quoted YAML scalar. Quoting
// unconditionally keeps output unambiguous regardless of colons, hashes, or
// other characters that would otherwise need escaping in plain scalars.
func quoteScalar(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// renderInvariant formats a single invariant as a YAML list entry block using
// the canonical two-space indentation. The block is preceded by a blank line
// so successive entries are separated like the existing file.
func renderInvariant(inv invariant) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  - id: %s\n", inv.id))
	b.WriteString(fmt.Sprintf("    name: %s\n", quoteScalar(inv.name)))
	b.WriteString(fmt.Sprintf("    rule: %s\n", quoteScalar(inv.rule)))
	b.WriteString(fmt.Sprintf("    severity: %s\n", inv.severity))
	return b.String()
}

// appendInvariants returns content with the given invariant blocks appended
// after the existing entries, preserving the original content verbatim. The
// existing file is expected to end the `invariants:` list at EOF, which is the
// canonical cARL layout.
func appendInvariants(content string, invs []invariant) string {
	out := content
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	for _, inv := range invs {
		out += renderInvariant(inv)
	}
	return out
}
