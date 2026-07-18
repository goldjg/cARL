// Package version implements the `carl version` command.
package version

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/goldjg/carl/internal/harness"
	"github.com/goldjg/carl/internal/manifest"
)

var versionHeaderRE = regexp.MustCompile(`(?i)^\s*(?:<!--\s*version:\s*([^\s]+)\s*-->|#\s*version:\s*([^\s]+))\s*$`)
var semverRE = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)

// Artifacts provides read access to embedded canonical runtime files.
// It is used to render bundled component versions.
type Artifacts interface {
	Open(targetPath string) ([]byte, error)
	List() ([]string, error)
}

// Command implements `carl version`.
type Command struct {
	cliVersion            string
	bundledRuntimeVersion string
	bundledRuntimeSource  string
	bundledRuntimeTag     string
	bundledRuntimeCommit  string
	arts                  Artifacts
}

// New returns a new version Command.
func New(
	cliVersion string,
	bundledRuntimeVersion string,
	bundledRuntimeSource string,
	bundledRuntimeTag string,
	bundledRuntimeCommit string,
	arts Artifacts,
) *Command {
	return &Command{
		cliVersion:            cliVersion,
		bundledRuntimeVersion: bundledRuntimeVersion,
		bundledRuntimeSource:  bundledRuntimeSource,
		bundledRuntimeTag:     bundledRuntimeTag,
		bundledRuntimeCommit:  bundledRuntimeCommit,
		arts:                  arts,
	}
}

// Name returns the command name.
func (c *Command) Name() string { return "version" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Show CLI, bundled runtime, and repository runtime version information"
}

// Run executes `carl version` in the current working directory.
func (c *Command) Run(_ context.Context, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	showComponents := false
	for _, arg := range args {
		switch arg {
		case "--components":
			showComponents = true
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument %q", arg)
		}
	}

	return c.runInDir(cwd, showComponents)
}

// RunInDir executes the default version output rooted at rootDir.
// Exported for testing without changing the process working directory.
func (c *Command) RunInDir(rootDir string) error {
	return c.runInDir(rootDir, false)
}

func (c *Command) runInDir(rootDir string, showComponents bool) error {
	printCLIMetadata(c.cliVersion)
	fmt.Println()
	printBundledRuntimeMetadata(
		c.bundledRuntimeVersion,
		c.bundledRuntimeSource,
		c.bundledRuntimeTag,
		c.bundledRuntimeCommit,
	)
	fmt.Println()

	runtimeInstalled := manifest.Exists(rootDir)
	var rt *manifest.Runtime
	if runtimeInstalled {
		var err error
		rt, err = manifest.Read(rootDir)
		if err != nil {
			return fmt.Errorf("read runtime manifest: %w", err)
		}
	}

	printRepositoryRuntime(runtimeInstalled, rt, c.bundledRuntimeVersion)
	fmt.Println()

	if runtimeInstalled {
		printInstalledPacks(rootDir, rt.ManagedArtifacts)
		fmt.Println()
	}

	printHarnessShims(rootDir, harness.Adapters())

	if showComponents {
		fmt.Println()
		if err := c.printComponents(rootDir); err != nil {
			return err
		}
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: carl version [--components]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --components  Show bundled vs installed versions for packs and harness shims")
}

func printCLIMetadata(cliVersion string) {
	fmt.Println("cARL CLI:")
	fmt.Printf("  Version:          %s\n", cliVersion)
}

func printBundledRuntimeMetadata(version, source, tag, commit string) {
	fmt.Println("Bundled Runtime:")
	fmt.Printf("  Version:          %s\n", version)
	fmt.Printf("  Source:           %s\n", source)
	fmt.Printf("  Tag:              %s\n", tag)
	fmt.Printf("  Commit:           %s\n", commit)
}

func printRepositoryRuntime(installed bool, rt *manifest.Runtime, bundledVersion string) {
	fmt.Println("Repository Runtime:")
	if !installed {
		fmt.Println("  Not installed in the current repository.")
		return
	}
	fmt.Printf("  Version:          %s\n", rt.RuntimeVersion)
	fmt.Printf("  Source:           %s\n", rt.Source)
	fmt.Printf("  Tag:              %s\n", rt.SourceTag)
	fmt.Printf("  Commit:           %s\n", rt.SourceCommit)
	fmt.Printf("  Status:           %s\n", compareRuntimeStatus(rt.RuntimeVersion, bundledVersion))
}

func compareRuntimeStatus(repositoryRuntimeVersion, bundledRuntimeVersion string) string {
	cmp, ok := compareSemanticVersions(repositoryRuntimeVersion, bundledRuntimeVersion)
	if !ok {
		return "Unknown"
	}
	switch {
	case cmp == 0:
		return "Current"
	case cmp < 0:
		return "Upgrade available"
	default:
		return "Repository runtime is newer"
	}
}

type packVersion struct {
	name    string
	version string
}

func printInstalledPacks(rootDir string, managedArtifacts []string) {
	packFiles := extractPackFiles(managedArtifacts)
	fmt.Println("Installed Packs:")
	if len(packFiles) == 0 {
		fmt.Println("  none")
		return
	}
	versions := make([]packVersion, 0, len(packFiles))
	for _, p := range packFiles {
		installedVersion, ok := readInstalledVersion(rootDir, p.path)
		if !ok {
			installedVersion = "unknown"
		}
		versions = append(versions, packVersion{name: p.name, version: installedVersion})
	}
	sortPackVersions(versions)
	for _, p := range versions {
		fmt.Printf("  %-33s %s\n", p.name, p.version)
	}
}

type harnessShimRow struct {
	id      string
	path    string
	version string
}

func printHarnessShims(rootDir string, adapters []harness.Adapter) {
	fmt.Println("Harness Shims:")
	rows := make([]harnessShimRow, 0, len(adapters))
	for _, a := range adapters {
		installedVersion, installed := readInstalledVersion(rootDir, a.DetectionFile)
		switch {
		case !installed:
			installedVersion = "not installed"
		case installedVersion == "":
			installedVersion = "unknown"
		}
		rows = append(rows, harnessShimRow{
			id:      a.ID,
			path:    a.DetectionFile,
			version: installedVersion,
		})
	}
	for _, r := range rows {
		fmt.Printf("  %-12s %-35s %s\n", r.id, r.path, r.version)
	}
}

func (c *Command) printComponents(rootDir string) error {
	packRows, err := c.collectPackComponentRows(rootDir)
	if err != nil {
		return err
	}
	fmt.Println("Instruction Packs:")
	fmt.Println("  Pack                              Bundled   Installed  State")
	for _, row := range packRows {
		fmt.Printf("  %-33s %-9s %-10s %s\n", row.name, row.bundled, row.installed, row.state)
	}
	fmt.Println()

	shimRows, err := c.collectShimComponentRows(rootDir)
	if err != nil {
		return err
	}
	fmt.Println("Harness Shims:")
	fmt.Println("  Harness       File                              Bundled   Installed  State")
	for _, row := range shimRows {
		fmt.Printf(
			"  %-13s %-33s %-9s %-10s %s\n",
			row.name, row.path, row.bundled, row.installed, row.state,
		)
	}
	return nil
}

type componentRow struct {
	name      string
	path      string
	bundled   string
	installed string
	state     string
}

func (c *Command) collectPackComponentRows(rootDir string) ([]componentRow, error) {
	if c.arts == nil {
		return nil, nil
	}
	paths, err := c.arts.List()
	if err != nil {
		return nil, fmt.Errorf("list embedded artefacts: %w", err)
	}
	packFiles := extractPackFiles(paths)
	rows := make([]componentRow, 0, len(packFiles))
	for _, p := range packFiles {
		bundledVersion, bundledOK := readBundledVersion(c.arts, p.path)
		if !bundledOK {
			bundledVersion = "unknown"
		}

		installedVersion, installed := readInstalledVersion(rootDir, p.path)
		if !installed {
			installedVersion = "missing"
		}

		rows = append(rows, componentRow{
			name:      p.name,
			path:      p.path,
			bundled:   bundledVersion,
			installed: installedVersion,
			state:     compareComponentState(bundledVersion, installedVersion),
		})
	}
	sortComponentRows(rows)
	return rows, nil
}

func (c *Command) collectShimComponentRows(rootDir string) ([]componentRow, error) {
	rows := make([]componentRow, 0, len(harness.Adapters()))
	for _, a := range harness.Adapters() {
		bundledVersion := "unknown"
		if c.arts != nil {
			if v, ok := readBundledVersion(c.arts, a.DetectionFile); ok {
				bundledVersion = v
			}
		}
		installedVersion, installed := readInstalledVersion(rootDir, a.DetectionFile)
		if !installed {
			installedVersion = "missing"
		}

		rows = append(rows, componentRow{
			name:      a.ID,
			path:      a.DetectionFile,
			bundled:   bundledVersion,
			installed: installedVersion,
			state:     compareComponentState(bundledVersion, installedVersion),
		})
	}
	return rows, nil
}

func compareComponentState(bundledVersion, installedVersion string) string {
	if installedVersion == "missing" {
		return "missing"
	}
	cmp, ok := compareSemanticVersions(installedVersion, bundledVersion)
	if !ok {
		return "unknown"
	}
	switch {
	case cmp == 0:
		return "current"
	case cmp < 0:
		return "older"
	default:
		return "newer"
	}
}

type packFile struct {
	name string
	path string
}

func extractPackFiles(artifacts []string) []packFile {
	seen := map[string]string{}
	for _, a := range artifacts {
		a = path.Clean(strings.ReplaceAll(a, "\\", "/"))
		parts := strings.Split(a, "/")
		if len(parts) != 4 {
			continue
		}
		if parts[0] != ".github" || parts[1] != "instructions" {
			continue
		}
		name := strings.TrimSuffix(parts[3], ".instructions.md")
		if name == parts[3] {
			continue
		}
		packName := parts[2] + "/" + name
		if _, exists := seen[packName]; !exists {
			seen[packName] = a
		}
	}
	result := make([]packFile, 0, len(seen))
	for name, p := range seen {
		result = append(result, packFile{name: name, path: p})
	}
	sortPackFiles(result)
	return result
}

func readBundledVersion(arts Artifacts, targetPath string) (string, bool) {
	data, err := arts.Open(targetPath)
	if err != nil {
		return "", false
	}
	return extractVersionHeader(data)
}

func readInstalledVersion(rootDir, targetPath string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(targetPath)))
	if err != nil {
		return "", false
	}
	v, ok := extractVersionHeader(data)
	if !ok {
		return "unknown", true
	}
	return v, true
}

func extractVersionHeader(data []byte) (string, bool) {
	lines := strings.Split(string(data), "\n")
	limit := len(lines)
	if limit > 10 {
		limit = 10
	}
	for i := 0; i < limit; i++ {
		m := versionHeaderRE.FindStringSubmatch(lines[i])
		if len(m) == 0 {
			continue
		}
		candidate := m[1]
		if candidate == "" {
			candidate = m[2]
		}
		if _, ok := parseSemanticVersion(candidate); !ok {
			return "", false
		}
		return candidate, true
	}
	return "", false
}

type semanticVersion struct {
	major int
	minor int
	patch int
	pre   []semverToken
}

type semverToken struct {
	numeric bool
	num     int
	text    string
}

func compareSemanticVersions(a, b string) (int, bool) {
	av, ok := parseSemanticVersion(a)
	if !ok {
		return 0, false
	}
	bv, ok := parseSemanticVersion(b)
	if !ok {
		return 0, false
	}

	if av.major != bv.major {
		if av.major < bv.major {
			return -1, true
		}
		return 1, true
	}
	if av.minor != bv.minor {
		if av.minor < bv.minor {
			return -1, true
		}
		return 1, true
	}
	if av.patch != bv.patch {
		if av.patch < bv.patch {
			return -1, true
		}
		return 1, true
	}
	return comparePreRelease(av.pre, bv.pre), true
}

func parseSemanticVersion(input string) (semanticVersion, bool) {
	s := strings.TrimSpace(input)
	s = strings.TrimPrefix(s, "v")
	m := semverRE.FindStringSubmatch(s)
	if len(m) == 0 {
		return semanticVersion{}, false
	}

	major, err := strconv.Atoi(m[1])
	if err != nil {
		return semanticVersion{}, false
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return semanticVersion{}, false
	}
	patch, err := strconv.Atoi(m[3])
	if err != nil {
		return semanticVersion{}, false
	}

	v := semanticVersion{major: major, minor: minor, patch: patch}
	if m[4] != "" {
		tokens := strings.Split(m[4], ".")
		pre := make([]semverToken, 0, len(tokens))
		for _, token := range tokens {
			if token == "" {
				return semanticVersion{}, false
			}
			if isNumeric(token) {
				num, err := strconv.Atoi(token)
				if err != nil {
					return semanticVersion{}, false
				}
				pre = append(pre, semverToken{numeric: true, num: num, text: token})
				continue
			}
			pre = append(pre, semverToken{text: token})
		}
		v.pre = pre
	}
	return v, true
}

func comparePreRelease(a, b []semverToken) int {
	// A release version has higher precedence than a pre-release.
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	if len(a) == 0 {
		return 1
	}
	if len(b) == 0 {
		return -1
	}

	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		if a[i].numeric && b[i].numeric {
			if a[i].num < b[i].num {
				return -1
			}
			if a[i].num > b[i].num {
				return 1
			}
			continue
		}
		if a[i].numeric && !b[i].numeric {
			return -1
		}
		if !a[i].numeric && b[i].numeric {
			return 1
		}
		if a[i].text < b[i].text {
			return -1
		}
		if a[i].text > b[i].text {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func isNumeric(token string) bool {
	for _, r := range token {
		if r < '0' || r > '9' {
			return false
		}
	}
	return token != ""
}

func sortPackFiles(items []packFile) {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 && items[j].name > key.name {
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
}

func sortPackVersions(items []packVersion) {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 && items[j].name > key.name {
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
}

func sortComponentRows(items []componentRow) {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 && items[j].name > key.name {
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
}
