// Package trustboundary inspects the optional repository-owned trust-boundary
// extension without reading or interpreting its Markdown content.
package trustboundary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtensionPath is the fixed repository-relative path for additional
// repository-owned trust-boundary guidance.
const ExtensionPath = ".github/carl/trust-boundaries.repository.md"

// ManagedPath is the canonical trust-boundary artifact shipped by cARL.
const ManagedPath = ".github/carl/trust-boundaries.md"

// ExtensionState describes the optional trust-boundary extension on disk.
type ExtensionState struct {
	Present       bool
	InvalidReason string
}

// Inspect performs a point-in-time check that the optional extension is a
// regular file reached without symlinked path components. It does not open the
// file or mediate later reads; the repository root itself is trusted.
func Inspect(rootDir string) (ExtensionState, error) {
	root, err := filepath.Abs(rootDir)
	if err != nil {
		return ExtensionState{}, fmt.Errorf("resolve repository root: %w", err)
	}

	parts := strings.Split(filepath.FromSlash(ExtensionPath), string(filepath.Separator))
	current := root
	for index, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return ExtensionState{}, nil
			}
			return ExtensionState{}, fmt.Errorf("inspect %s: %w", ExtensionPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ExtensionState{
				InvalidReason: fmt.Sprintf("%s is a symbolic link", filepath.ToSlash(filepath.Join(parts[:index+1]...))),
			}, nil
		}
		if index < len(parts)-1 && !info.IsDir() {
			return ExtensionState{
				InvalidReason: fmt.Sprintf("%s is not a directory", filepath.ToSlash(filepath.Join(parts[:index+1]...))),
			}, nil
		}
		if index == len(parts)-1 {
			if !info.Mode().IsRegular() {
				return ExtensionState{InvalidReason: "extension is not a regular file"}, nil
			}
			return ExtensionState{Present: true}, nil
		}
	}
	return ExtensionState{}, nil
}
