// Package repomap implements the `carl map` command.
// It derives a cognitive repository map from the filesystem and writes it
// to .github/carl/repo-map.json.
package repomap

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// OutputFile is the path of the generated repo map relative to the repository root.
const OutputFile = ".github/carl/repo-map.json"

// Map is the top-level structure written to repo-map.json.
type Map struct {
	Note          string            `json:"_note"`
	GeneratedBy   string            `json:"generated_by"`
	LastUpdated   string            `json:"last_updated"`
	Languages     []string          `json:"languages"`
	EntryPoints   []File            `json:"entry_points"`
	Directories   map[string]string `json:"directories"`
	Workflows     []File            `json:"workflows"`
	Governance    []File            `json:"governance"`
	Documentation []File            `json:"documentation"`
}

// File is a path with an optional human-readable purpose description.
type File struct {
	Path    string `json:"path"`
	Purpose string `json:"purpose,omitempty"`
}

// Command implements `carl map`.
type Command struct{}

// New returns a new map Command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "map" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Generate and update .github/carl/repo-map.json from repository structure"
}

// Run executes `carl map` in the current working directory.
func (c *Command) Run(_ context.Context, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	return c.RunInDir(cwd)
}

// RunInDir executes the map logic rooted at rootDir.
// Exported for testing without changing the process working directory.
func (c *Command) RunInDir(rootDir string) error {
	m, err := Build(rootDir)
	if err != nil {
		return fmt.Errorf("build repo map: %w", err)
	}
	if err := writeMap(rootDir, m); err != nil {
		return fmt.Errorf("write repo map: %w", err)
	}
	fmt.Printf("Repo map updated: %s\n", OutputFile)
	fmt.Printf("  Languages:     %s\n", strings.Join(m.Languages, ", "))
	fmt.Printf("  Entry points:  %d\n", len(m.EntryPoints))
	fmt.Printf("  Directories:   %d\n", len(m.Directories))
	fmt.Printf("  Workflows:     %d\n", len(m.Workflows))
	fmt.Printf("  Governance:    %d\n", len(m.Governance))
	fmt.Printf("  Documentation: %d\n", len(m.Documentation))
	return nil
}

// Build scans rootDir and derives a Map from its filesystem structure.
// The scan is bounded to rootDir; it never follows symlinks outside it.
func Build(rootDir string) (*Map, error) {
	m := &Map{
		Note:        "Repository map derived by `carl map`. Re-run to update after structural changes.",
		GeneratedBy: "carl map",
		LastUpdated: time.Now().UTC().Format("2006-01-02"),
	}

	langs, err := detectLanguages(rootDir)
	if err != nil {
		return nil, fmt.Errorf("detect languages: %w", err)
	}
	m.Languages = langs

	eps, err := detectEntryPoints(rootDir)
	if err != nil {
		return nil, fmt.Errorf("detect entry points: %w", err)
	}
	m.EntryPoints = eps

	dirs, err := detectDirectories(rootDir)
	if err != nil {
		return nil, fmt.Errorf("detect directories: %w", err)
	}
	m.Directories = dirs

	wfs, err := detectWorkflows(rootDir)
	if err != nil {
		return nil, fmt.Errorf("detect workflows: %w", err)
	}
	m.Workflows = wfs

	gov, err := detectGovernance(rootDir)
	if err != nil {
		return nil, fmt.Errorf("detect governance: %w", err)
	}
	m.Governance = gov

	docs, err := detectDocumentation(rootDir)
	if err != nil {
		return nil, fmt.Errorf("detect documentation: %w", err)
	}
	m.Documentation = docs

	return m, nil
}

// writeMap serialises m as indented JSON and writes it to rootDir/OutputFile.
// It creates the parent directory if it does not exist.
func writeMap(rootDir string, m *Map) error {
	path := filepath.Join(rootDir, filepath.FromSlash(OutputFile))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// skipDir reports whether a directory should be excluded from all scans.
func skipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor":
		return true
	}
	return false
}

// extToLanguage maps common source-file extensions to programming language names.
var extToLanguage = map[string]string{
	".go":   "Go",
	".py":   "Python",
	".ts":   "TypeScript",
	".tsx":  "TypeScript",
	".js":   "JavaScript",
	".jsx":  "JavaScript",
	".rs":   "Rust",
	".java": "Java",
	".rb":   "Ruby",
	".sh":   "Shell",
	".bash": "Shell",
	".ps1":  "PowerShell",
	".tf":   "Terraform",
	".cs":   "C#",
	".cpp":  "C++",
	".cc":   "C++",
	".c":    "C",
	".swift": "Swift",
	".kt":   "Kotlin",
}

// detectLanguages walks rootDir and returns a sorted, deduplicated list of
// programming languages inferred from file extensions. Hidden directories
// (except the root), .git, node_modules, and vendor are excluded.
func detectLanguages(rootDir string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if lang, ok := extToLanguage[ext]; ok {
			seen[lang] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	langs := make([]string, 0, len(seen))
	for l := range seen {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	return langs, nil
}

// candidateEntryPoints lists root-level filenames recognised as project entry
// points together with a base purpose description.
var candidateEntryPoints = []struct {
	name    string
	purpose string
}{
	{"go.mod", "Go module definition"},
	{"Makefile", "Build automation"},
	{"makefile", "Build automation"},
	{"package.json", "Node.js project"},
	{"Cargo.toml", "Rust project"},
	{"setup.py", "Python project"},
	{"pyproject.toml", "Python project"},
	{"Dockerfile", "Container definition"},
	{"docker-compose.yml", "Compose services"},
	{"docker-compose.yaml", "Compose services"},
}

// detectEntryPoints returns project entry point files. It checks well-known
// root-level files and discovers cmd/**/main.go patterns.
func detectEntryPoints(rootDir string) ([]File, error) {
	var eps []File

	for _, ep := range candidateEntryPoints {
		p := filepath.Join(rootDir, ep.name)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		purpose := ep.purpose
		if ep.name == "go.mod" {
			if mod := readGoModuleName(p); mod != "" {
				purpose = fmt.Sprintf("Go module definition: %s", mod)
			}
		}
		eps = append(eps, File{Path: ep.name, Purpose: purpose})
	}

	// Discover cmd/**/main.go entry points.
	cmdDir := filepath.Join(rootDir, "cmd")
	if _, err := os.Stat(cmdDir); err == nil {
		if err := filepath.WalkDir(cmdDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || d.Name() != "main.go" {
				return nil
			}
			rel, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			// Derive the command name from the parent directory.
			// e.g. cmd/carl/main.go → "carl"
			parts := strings.Split(rel, "/")
			cmdName := ""
			if len(parts) >= 2 {
				cmdName = parts[len(parts)-2]
			}
			purpose := "CLI entry point"
			if cmdName != "" && cmdName != "cmd" {
				purpose = fmt.Sprintf("%s CLI entry point", cmdName)
			}
			eps = append(eps, File{Path: rel, Purpose: purpose})
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return eps, nil
}

// readGoModuleName reads and returns the module name from a go.mod file.
// Returns an empty string on any read or parse error.
func readGoModuleName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

// knownDirPurposes maps repo-relative directory paths (forward-slash) to
// known human-readable purpose descriptions.
var knownDirPurposes = map[string]string{
	".github":                        "GitHub configuration and Copilot instruction root",
	".github/carl":                   "cARLv2 governance artefacts and templates",
	".github/carl/plans":             "Prompt-as-code planning artefacts",
	".github/instructions":           "Copilot instruction packs",
	".github/instructions/core":      "Core governance packs",
	".github/instructions/languages": "Language-specific guidance packs",
	".github/instructions/platform":  "Platform guidance packs",
	".github/instructions/cloud":     "Cloud guidance packs",
	".github/workflows":              "GitHub Actions workflows",
	"embedded/assets":                "Embedded runtime asset mirror",
	// Common Go project directories.
	"cmd":      "CLI command entry points",
	"internal": "Internal implementation packages",
	"pkg":      "Public library packages",
	"api":      "API definitions",
	"docs":     "Documentation",
	"scripts":  "Build and automation scripts",
	"tools":    "Development tooling",
	"test":     "Integration and end-to-end tests",
}

// noDescendDir lists directories that are included in the map but whose
// subdirectories are not recursed into, to avoid mirrored tree duplication.
var noDescendDir = map[string]bool{
	"embedded/assets": true,
}

// detectDirectories returns a map of key directory paths (relative,
// forward-slash) to human-readable purpose descriptions. The scan covers
// up to three levels of depth. .git, node_modules, and vendor are excluded.
func detectDirectories(rootDir string) (map[string]string, error) {
	result := map[string]string{}

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || path == rootDir {
			return nil
		}
		if skipDir(d.Name()) {
			return filepath.SkipDir
		}

		rel, relErr := filepath.Rel(rootDir, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		depth := strings.Count(rel, "/") + 1
		result[rel] = dirPurpose(rootDir, path, rel)

		if noDescendDir[rel] || depth >= 3 {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// dirPurpose returns a human-readable description for the directory at absPath
// whose repo-relative path is relPath.
func dirPurpose(rootDir, absPath, relPath string) string {
	if p, ok := knownDirPurposes[relPath]; ok {
		return p
	}
	return goPackageDoc(absPath)
}

// goPackageDoc reads Go source files in dir and returns the description
// extracted from the first `// Package <name> ...` or `// Command <name> ...`
// doc comment found. Returns an empty string if none is found.
func goPackageDoc(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if doc := readGoDoc(filepath.Join(dir, e.Name())); doc != "" {
			return doc
		}
	}
	return ""
}

// readGoDoc reads a Go source file and extracts the description from the first
// `// Package <name> ...` or `// Command <name> ...` doc comment block.
// It reads continuation `//` lines to assemble the full first sentence, then
// trims to the first sentence end (a period at end of string or before a space).
// The first character of the result is capitalised.
func readGoDoc(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var parts []string
	inDoc := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !inDoc {
			var rest string
			switch {
			case strings.HasPrefix(line, "// Package "):
				text := strings.TrimPrefix(line, "// Package ")
				idx := strings.Index(text, " ")
				if idx < 0 {
					return ""
				}
				rest = strings.TrimSpace(text[idx+1:])
			case strings.HasPrefix(line, "// Command "):
				text := strings.TrimPrefix(line, "// Command ")
				idx := strings.Index(text, " ")
				if idx < 0 {
					return ""
				}
				rest = strings.TrimSpace(text[idx+1:])
			default:
				continue
			}
			if rest == "" {
				return ""
			}
			parts = append(parts, rest)
			inDoc = true
		} else {
			// Continue reading the doc block until a blank comment or non-comment line.
			if !strings.HasPrefix(line, "//") {
				break
			}
			cont := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if cont == "" {
				break // blank comment line ends the doc block
			}
			parts = append(parts, cont)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	full := strings.Join(parts, " ")

	// Trim to the first sentence: a period that ends the string or is
	// followed by a space (not a period inside a path like .github/foo).
	if i := sentenceEnd(full); i >= 0 {
		full = full[:i+1]
	}

	if len(full) == 0 {
		return ""
	}
	return strings.ToUpper(full[:1]) + full[1:]
}

// sentenceEnd returns the index of the first sentence-terminating period in s.
// A period terminates a sentence when it is the last character of s or when
// the next character is a space. Periods inside path-like tokens (e.g.
// ".github/foo" or "v1.2.3") are skipped.
func sentenceEnd(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != '.' {
			continue
		}
		next := i + 1
		if next >= len(s) || s[next] == ' ' {
			return i
		}
	}
	return -1
}

// detectWorkflows returns a sorted list of workflow files found under
// .github/workflows/. Non-YAML files are ignored.
func detectWorkflows(rootDir string) ([]File, error) {
	wfDir := filepath.Join(rootDir, ".github", "workflows")
	entries, err := os.ReadDir(wfDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var wfs []File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		wfs = append(wfs, File{
			Path:    ".github/workflows/" + e.Name(),
			Purpose: stem + " workflow",
		})
	}
	sort.Slice(wfs, func(i, j int) bool { return wfs[i].Path < wfs[j].Path })
	return wfs, nil
}

// knownGovernancePurposes maps known .github/carl/ filenames to purpose descriptions.
var knownGovernancePurposes = map[string]string{
	"invariants.yml":                   "Runtime invariants enforced by all implementation PRs",
	"memory.md":                        "Durable architectural truth cache",
	"current-pr-contract.md":           "Active PR scope and constraints",
	"current-pr-contract.template.md":  "PR contract template",
	"tool-policy.yml":                  "Tool permission tier definitions",
	"trust-boundaries.md":              "Trust boundary documentation",
	"repo-map.example.json":            "Repo map template example",
	"repo-map.json":                    "Generated cognitive repository map",
}

// detectGovernance returns a sorted list of files found directly under
// .github/carl/ with their purpose descriptions.
func detectGovernance(rootDir string) ([]File, error) {
	carlDir := filepath.Join(rootDir, ".github", "carl")
	entries, err := os.ReadDir(carlDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var gov []File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		relPath := ".github/carl/" + e.Name()
		gov = append(gov, File{
			Path:    relPath,
			Purpose: knownGovernancePurposes[e.Name()],
		})
	}
	sort.Slice(gov, func(i, j int) bool { return gov[i].Path < gov[j].Path })
	return gov, nil
}

// knownDocPurposes maps known root-level documentation filenames to purpose descriptions.
var knownDocPurposes = map[string]string{
	"README.md":       "Repository overview and pack catalogue",
	"ARCHITECTURE.md": "Architecture documentation",
	"ROADMAP.md":      "Feature roadmap and backlog",
	"CLI.md":          "CLI command reference",
	"VISION.md":       "Project vision",
	"GLOSSARY.md":     "Terminology glossary",
	"CONTRIBUTING.md": "Contribution guidelines",
	"CHANGELOG.md":    "Version changelog",
	"LICENSE":         "Licence",
}

// detectDocumentation returns a sorted list of root-level documentation files
// (.md files and LICENSE).
func detectDocumentation(rootDir string) ([]File, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	var docs []File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".md" && name != "LICENSE" {
			continue
		}
		docs = append(docs, File{
			Path:    name,
			Purpose: knownDocPurposes[name],
		})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Path < docs[j].Path })
	return docs, nil
}
