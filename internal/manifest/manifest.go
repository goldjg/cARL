// Package manifest manages the cARL runtime manifest (.github/carl/runtime.json).
// runtime.json is the authoritative source of truth for installed runtime state.
// It must never be modified by repair or any command other than init.
package manifest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// FileName is the path of the runtime manifest relative to the repository root.
const FileName = ".github/carl/runtime.json"

// Runtime is the cARL runtime manifest.
type Runtime struct {
	RuntimeVersion   string    `json:"runtimeVersion"`
	Source           string    `json:"source"`
	SourceTag        string    `json:"sourceTag"`
	SourceCommit     string    `json:"sourceCommit"`
	InstalledAt      time.Time `json:"installedAt"`
	ManagedArtifacts []string  `json:"managedArtifacts"`
}

// Read reads and parses the runtime manifest at rootDir/FileName.
func Read(rootDir string) (*Runtime, error) {
	path := filepath.Join(rootDir, FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Runtime
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Write serialises and writes the runtime manifest to rootDir/FileName.
// It creates the parent directory if it does not exist.
func Write(rootDir string, r *Runtime) error {
	path := filepath.Join(rootDir, FileName)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// Exists reports whether the runtime manifest exists at rootDir/FileName.
func Exists(rootDir string) bool {
	_, err := os.Stat(filepath.Join(rootDir, FileName))
	return !errors.Is(err, os.ErrNotExist)
}
