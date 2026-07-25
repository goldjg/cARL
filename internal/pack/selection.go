package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// SelectionFileName is the path of the pack selection artefact relative to
// the repository root. It is a user-owned committed artefact, distinct from
// runtime.json (which is init-only and never written by pack commands).
const SelectionFileName = ".github/carl/packs.json"

// Selection is the schema-versioned pack selection artefact.
type Selection struct {
	SchemaVersion int      `json:"schemaVersion"`
	Selected      []string `json:"selected"`
}

// ReadSelection reads the pack selection artefact from rootDir.
// It returns (nil, nil) when the file does not exist so callers can fall
// back to legacy runtime.json-derived selection. Malformed content is an
// explicit error — never a silent fallback.
func ReadSelection(rootDir string) (*Selection, error) {
	p := filepath.Join(rootDir, filepath.FromSlash(SelectionFileName))
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", SelectionFileName, err)
	}
	var s Selection
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", SelectionFileName, err)
	}
	if s.SchemaVersion != metadataSchemaVersion {
		return nil, fmt.Errorf("%s: unsupported schema version %d", SelectionFileName, s.SchemaVersion)
	}
	for _, id := range s.Selected {
		if !packIDRE.MatchString(id) {
			return nil, fmt.Errorf("%s: malformed pack id %q", SelectionFileName, id)
		}
	}
	return &s, nil
}

// WriteSelection persists the selection artefact deterministically:
// IDs are deduplicated and sorted before writing.
func WriteSelection(rootDir string, ids []string) error {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)

	p := filepath.Join(rootDir, filepath.FromSlash(SelectionFileName))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create %s directory: %w", SelectionFileName, err)
	}
	data, err := json.MarshalIndent(Selection{
		SchemaVersion: metadataSchemaVersion,
		Selected:      out,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", SelectionFileName, err)
	}
	return os.WriteFile(p, append(data, '\n'), 0644)
}
