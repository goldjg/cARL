// Package embedded provides access to the cARL runtime assets bundled into
// the CLI binary. All managed runtime files are embedded here so the CLI
// requires no network access.
package embedded

import (
	"embed"
	"io/fs"
	"strings"
)

// The assets directory mirrors the target installation path structure under
// a leading "assets/" prefix so that the embed directive can include dotfiles.
//
//go:embed all:assets
var rawFS embed.FS

const assetsPrefix = "assets"

// FS wraps the embedded file system and provides helper methods.
type FS struct {
	raw embed.FS
}

// Assets is the package-level instance backed by the embedded binary data.
var Assets = FS{raw: rawFS}

// Open returns the content of an embedded file at the given target path.
// targetPath is relative to the repository root (e.g. ".github/copilot-instructions.md").
func (e FS) Open(targetPath string) ([]byte, error) {
	return e.raw.ReadFile(assetsPrefix + "/" + targetPath)
}

// List returns all embedded file paths relative to the repository root,
// in lexicographic order.
func (e FS) List() ([]string, error) {
	var paths []string
	err := fs.WalkDir(e.raw, assetsPrefix, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// Strip the "assets/" prefix to get the repo-relative path.
			paths = append(paths, strings.TrimPrefix(p, assetsPrefix+"/"))
		}
		return nil
	})
	return paths, err
}
