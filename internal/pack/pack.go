// Package pack implements the `carl pack` command and its subcommands.
package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/goldjg/carl/internal/cmdutil"
	"github.com/goldjg/carl/internal/manifest"
)

const metadataSchemaVersion = 1

var (
	packIDRE        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*/[a-z0-9]+(?:-[a-z0-9]+)*$`)
	semverRE        = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)
	versionHeaderRE = regexp.MustCompile(`(?i)^\s*(?:<!--\s*version:\s*([^\s]+)\s*-->|#\s*version:\s*([^\s]+))\s*$`)
)

// Artifacts provides read access to embedded runtime files.
type Artifacts interface {
	List() ([]string, error)
	Open(targetPath string) ([]byte, error)
}

// Command implements `carl pack`.
type Command struct {
	arts Artifacts
}

// New returns a new pack Command.
func New(arts Artifacts) *Command { return &Command{arts: arts} }

// Name returns the command name.
func (c *Command) Name() string { return "pack" }

// Synopsis returns a short description.
func (c *Command) Synopsis() string {
	return "Discover and inspect available instruction packs"
}

// PackMetadata is the versioned runtime metadata model for a pack.
type PackMetadata struct {
	SchemaVersion  int              `json:"schemaVersion"`
	ID             string           `json:"id"`
	Version        string           `json:"version"`
	Title          string           `json:"title,omitempty"`
	Description    string           `json:"description,omitempty"`
	Category       string           `json:"category"`
	Source         string           `json:"source"`
	State          PackState        `json:"state"`
	OwnedArtifacts []string         `json:"ownedArtifacts"`
	Dependencies   []PackDependency `json:"dependencies,omitempty"`
	Compatibility  *Compatibility   `json:"compatibility,omitempty"`
	Precedence     *Precedence      `json:"precedence,omitempty"`
}

// PackState captures current observed pack state.
type PackState struct {
	Bundled   bool `json:"bundled"`
	Installed bool `json:"installed"`
	Selected  bool `json:"selected"`
	Active    bool `json:"active"`
}

// PackDependency declares a dependency on another pack.
type PackDependency struct {
	ID       string `json:"id"`
	Required bool   `json:"required"`
}

// Compatibility captures compatibility constraints when declared.
type Compatibility struct {
	MinimumRuntimeVersion string `json:"minimumRuntimeVersion,omitempty"`
}

// Precedence captures explicit precedence metadata where declared.
type Precedence struct {
	Mode      string   `json:"mode,omitempty"`
	Priority  *int     `json:"priority,omitempty"`
	Overrides []string `json:"overrides,omitempty"`
}

type listPayload struct {
	SchemaVersion int            `json:"schemaVersion"`
	Packs         []PackMetadata `json:"packs"`
}

type showPayload struct {
	SchemaVersion int          `json:"schemaVersion"`
	Pack          PackMetadata `json:"pack"`
}

type errorPayload struct {
	SchemaVersion int `json:"schemaVersion"`
	Error         struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type packFileMetadata struct {
	Path        string
	Version     string
	Title       string
	Description string
}

// Run dispatches to pack subcommands.
func (c *Command) Run(_ context.Context, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	switch args[0] {
	case "list":
		jsonOut, help, err := parseJSONFlag(args[1:])
		if err != nil {
			return err
		}
		if help {
			fmt.Println("Usage: carl pack list [--json]")
			return nil
		}
		return c.RunListInDir(cwd, jsonOut)
	case "show":
		return c.runShow(cwd, args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q\n\nRun 'carl pack --help' for usage", args[0])
	}
}

// RunListInDir executes `carl pack list`.
func (c *Command) RunListInDir(rootDir string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return err
	}
	if issues := validatePackSet(packs); len(issues) > 0 {
		return fmt.Errorf("pack metadata validation failed:\n- %s", strings.Join(issues, "\n- "))
	}
	if jsonOut {
		data, err := json.MarshalIndent(listPayload{
			SchemaVersion: metadataSchemaVersion,
			Packs:         packs,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal pack list JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	printPackList(packs)
	return nil
}

// RunShowInDir executes `carl pack show <pack>`.
func (c *Command) RunShowInDir(rootDir, packID string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return err
	}
	if issues := validatePackSet(packs); len(issues) > 0 {
		return fmt.Errorf("pack metadata validation failed:\n- %s", strings.Join(issues, "\n- "))
	}

	var found *PackMetadata
	for i := range packs {
		if packs[i].ID == packID {
			found = &packs[i]
			break
		}
	}
	if found == nil {
		msg := fmt.Sprintf("unknown pack %q", packID)
		if jsonOut {
			return newJSONExitError("pack_not_found", msg)
		}
		return fmt.Errorf("%s", msg)
	}

	if jsonOut {
		data, err := json.MarshalIndent(showPayload{
			SchemaVersion: metadataSchemaVersion,
			Pack:          *found,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal pack show JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	printPackDetails(*found)
	return nil
}

func (c *Command) runShow(rootDir string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing pack ID\n\nUsage: carl pack show <pack-id> [--json]")
	}
	packID := args[0]
	jsonOut, help, err := parseJSONFlag(args[1:])
	if err != nil {
		return err
	}
	if help {
		fmt.Println("Usage: carl pack show <pack-id> [--json]")
		return nil
	}
	return c.RunShowInDir(rootDir, packID, jsonOut)
}

func parseJSONFlag(args []string) (bool, bool, error) {
	jsonOut := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			return false, true, nil
		default:
			return false, false, fmt.Errorf("unknown argument %q", arg)
		}
	}
	return jsonOut, false, nil
}

func printUsage() {
	fmt.Println("Usage: carl pack <subcommand> [arguments]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list           List available packs")
	fmt.Println("  show <pack-id> Show metadata for a pack")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --json         Print machine-readable JSON output")
}

func printPackList(packs []PackMetadata) {
	fmt.Println("Available Packs:")
	fmt.Println()
	fmt.Println("  ID                                Version   Category    Source                    State       Description")
	for _, p := range packs {
		fmt.Printf(
			"  %-33s %-9s %-11s %-25s %-11s %s\n",
			p.ID,
			p.Version,
			p.Category,
			p.Source,
			summarizeState(p.State),
			shorten(p.Description, 64),
		)
	}
}

func printPackDetails(p PackMetadata) {
	fmt.Printf("Pack: %s\n\n", p.ID)
	fmt.Printf("Schema Version:    %d\n", p.SchemaVersion)
	fmt.Printf("Version:           %s\n", p.Version)
	fmt.Printf("Title:             %s\n", valueOrNone(p.Title))
	fmt.Printf("Description:       %s\n", valueOrNone(p.Description))
	fmt.Printf("Category:          %s\n", p.Category)
	fmt.Printf("Source:            %s\n", p.Source)
	fmt.Printf("State:             %s\n", summarizeState(p.State))
	fmt.Println()

	fmt.Println("Dependencies:")
	if len(p.Dependencies) == 0 {
		fmt.Println("  none")
	} else {
		for _, d := range p.Dependencies {
			req := "optional"
			if d.Required {
				req = "required"
			}
			fmt.Printf("  %s (%s)\n", d.ID, req)
		}
	}
	fmt.Println()

	fmt.Println("Owned Artefacts:")
	if len(p.OwnedArtifacts) == 0 {
		fmt.Println("  none")
	} else {
		for _, f := range p.OwnedArtifacts {
			fmt.Printf("  %s\n", f)
		}
	}
	fmt.Println()

	fmt.Println("Compatibility:")
	if p.Compatibility == nil || p.Compatibility.MinimumRuntimeVersion == "" {
		fmt.Println("  none declared")
	} else {
		fmt.Printf("  minimumRuntimeVersion: %s\n", p.Compatibility.MinimumRuntimeVersion)
	}
	fmt.Println()

	fmt.Println("Precedence:")
	if p.Precedence == nil {
		fmt.Println("  no explicit precedence metadata")
		return
	}
	fmt.Printf("  mode: %s\n", valueOrNone(p.Precedence.Mode))
	if p.Precedence.Priority != nil {
		fmt.Printf("  priority: %d\n", *p.Precedence.Priority)
	} else {
		fmt.Println("  priority: none")
	}
	if len(p.Precedence.Overrides) == 0 {
		fmt.Println("  overrides: none")
	} else {
		for _, o := range p.Precedence.Overrides {
			fmt.Printf("  override: %s\n", o)
		}
	}
}

func summarizeState(s PackState) string {
	switch {
	case s.Active:
		return "active"
	case s.Selected:
		return "selected"
	case s.Installed && s.Bundled:
		return "installed"
	case s.Installed:
		return "repository"
	case s.Bundled:
		return "bundled"
	default:
		return "available"
	}
}

func valueOrNone(v string) string {
	if strings.TrimSpace(v) == "" {
		return "none"
	}
	return v
}

func shorten(v string, max int) string {
	v = strings.TrimSpace(v)
	if len(v) <= max {
		return v
	}
	if max <= 1 {
		return v[:max]
	}
	return v[:max-1] + "…"
}

func (c *Command) discover(rootDir string) ([]PackMetadata, error) {
	bundled, err := c.discoverBundled()
	if err != nil {
		return nil, err
	}
	local, err := discoverLocal(rootDir)
	if err != nil {
		return nil, err
	}
	selected := discoverSelected(rootDir)

	idsMap := map[string]bool{}
	for id := range bundled {
		idsMap[id] = true
	}
	for id := range local {
		idsMap[id] = true
	}
	for id := range selected {
		idsMap[id] = true
	}

	ids := make([]string, 0, len(idsMap))
	for id := range idsMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]PackMetadata, 0, len(ids))
	for _, id := range ids {
		b, hasBundled := bundled[id]
		l, hasLocal := local[id]

		version := firstNonEmpty(l.Version, b.Version)
		title := firstNonEmpty(l.Title, b.Title)
		description := firstNonEmpty(l.Description, b.Description)
		category := categoryFromID(id)

		owned := make([]string, 0, 2)
		if hasBundled {
			owned = append(owned, b.Path)
		}
		if hasLocal && l.Path != b.Path {
			owned = append(owned, l.Path)
		}
		if len(owned) == 0 {
			owned = append(owned, expectedPackPath(id))
		}
		sort.Strings(owned)

		state := PackState{
			Bundled:   hasBundled,
			Installed: hasLocal,
			Selected:  selected[id],
		}
		state.Active = state.Selected

		out = append(out, PackMetadata{
			SchemaVersion:  metadataSchemaVersion,
			ID:             id,
			Version:        version,
			Title:          title,
			Description:    description,
			Category:       category,
			Source:         deriveSource(state),
			State:          state,
			OwnedArtifacts: owned,
		})
	}

	return out, nil
}

func (c *Command) discoverBundled() (map[string]packFileMetadata, error) {
	result := map[string]packFileMetadata{}
	if c.arts == nil {
		return result, nil
	}
	paths, err := c.arts.List()
	if err != nil {
		return nil, fmt.Errorf("list embedded artefacts: %w", err)
	}

	for _, p := range paths {
		id, _, ok := parsePackPath(p)
		if !ok {
			continue
		}
		data, err := c.arts.Open(p)
		if err != nil {
			return nil, fmt.Errorf("read embedded pack %s: %w", p, err)
		}
		result[id] = parsePackFileMetadata(p, data)
	}
	return result, nil
}

func discoverLocal(rootDir string) (map[string]packFileMetadata, error) {
	result := map[string]packFileMetadata{}
	base := filepath.Join(rootDir, ".github", "instructions")
	if _, err := os.Stat(base); err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("inspect instructions directory: %w", err)
	}
	err := filepath.WalkDir(base, func(filePath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(filePath, ".instructions.md") {
			return nil
		}
		rel, err := filepath.Rel(rootDir, filePath)
		if err != nil {
			return err
		}
		rel = path.Clean(strings.ReplaceAll(rel, "\\", "/"))
		id, _, ok := parsePackPath(rel)
		if !ok {
			return nil
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		result[id] = parsePackFileMetadata(rel, data)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan local packs: %w", err)
	}
	return result, nil
}

func discoverSelected(rootDir string) map[string]bool {
	selected := map[string]bool{}
	if !manifest.Exists(rootDir) {
		return selected
	}
	rt, err := manifest.Read(rootDir)
	if err != nil {
		return selected
	}
	for _, m := range rt.ManagedArtifacts {
		id, _, ok := parsePackPath(m)
		if ok {
			selected[id] = true
		}
	}
	return selected
}

func parsePackFileMetadata(packPath string, data []byte) packFileMetadata {
	return packFileMetadata{
		Path:        path.Clean(strings.ReplaceAll(packPath, "\\", "/")),
		Version:     extractVersion(data),
		Title:       extractTitle(data),
		Description: extractDescription(data),
	}
}

func extractVersion(data []byte) string {
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
		v := m[1]
		if v == "" {
			v = m[2]
		}
		return strings.TrimSpace(v)
	}
	return ""
}

func extractTitle(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func extractDescription(data []byte) string {
	lines := strings.Split(string(data), "\n")
	start := 0
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			start = i + 1
			break
		}
	}
	var parts []string
	for i := start; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if len(parts) > 0 {
				break
			}
			continue
		}
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "<!--") {
			if len(parts) > 0 {
				break
			}
			continue
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, " ")
}

func parsePackPath(p string) (id string, category string, ok bool) {
	clean := path.Clean(strings.ReplaceAll(p, "\\", "/"))
	parts := strings.Split(clean, "/")
	if len(parts) != 4 {
		return "", "", false
	}
	if parts[0] != ".github" || parts[1] != "instructions" {
		return "", "", false
	}
	name := strings.TrimSuffix(parts[3], ".instructions.md")
	if name == parts[3] || name == "" {
		return "", "", false
	}
	category = parts[2]
	return category + "/" + name, category, true
}

func deriveSource(state PackState) string {
	switch {
	case state.Bundled && state.Installed:
		return "bundled+repository-local"
	case state.Bundled:
		return "bundled"
	case state.Installed:
		return "repository-local"
	default:
		return "unknown"
	}
}

func validatePackSet(packs []PackMetadata) []string {
	var issues []string
	seen := map[string]bool{}
	index := map[string]PackMetadata{}

	for _, p := range packs {
		if p.SchemaVersion != metadataSchemaVersion {
			issues = append(issues, fmt.Sprintf("%s: unsupported schema version %d", p.ID, p.SchemaVersion))
		}
		if p.ID == "" {
			issues = append(issues, "pack with missing id")
		} else if !packIDRE.MatchString(p.ID) {
			issues = append(issues, fmt.Sprintf("%s: malformed pack id", p.ID))
		}
		if seen[p.ID] {
			issues = append(issues, fmt.Sprintf("%s: duplicate pack ID", p.ID))
		}
		seen[p.ID] = true
		index[p.ID] = p
		if p.Version == "" || !isSemver(p.Version) {
			issues = append(issues, fmt.Sprintf("%s: invalid version %q", p.ID, p.Version))
		}
		if p.Category == "" || p.Category != categoryFromID(p.ID) {
			issues = append(issues, fmt.Sprintf("%s: invalid or contradictory category", p.ID))
		}
		for _, owned := range p.OwnedArtifacts {
			if !isValidOwnedArtifact(owned) {
				issues = append(issues, fmt.Sprintf("%s: invalid owned artefact %q", p.ID, owned))
			}
		}
		if p.State.Active && !p.State.Selected {
			issues = append(issues, fmt.Sprintf("%s: contradictory state (active requires selected)", p.ID))
		}
		if p.Precedence != nil {
			if p.Precedence.Priority != nil && *p.Precedence.Priority < 0 {
				issues = append(issues, fmt.Sprintf("%s: invalid priority value %d", p.ID, *p.Precedence.Priority))
			}
			if p.Precedence.Mode != "" {
				switch p.Precedence.Mode {
				case "additive", "overridable", "restrictable-only", "immutable":
				default:
					issues = append(issues, fmt.Sprintf("%s: invalid precedence mode %q", p.ID, p.Precedence.Mode))
				}
			}
		}
	}

	adj := map[string][]string{}
	for _, p := range packs {
		for _, d := range p.Dependencies {
			if d.ID == "" || !packIDRE.MatchString(d.ID) {
				issues = append(issues, fmt.Sprintf("%s: malformed dependency id %q", p.ID, d.ID))
				continue
			}
			if _, ok := index[d.ID]; !ok {
				issues = append(issues, fmt.Sprintf("%s: missing dependency %s", p.ID, d.ID))
				continue
			}
			adj[p.ID] = append(adj[p.ID], d.ID)
		}
	}
	issues = append(issues, findDependencyCycles(adj)...)

	return issues
}

func findDependencyCycles(adj map[string][]string) []string {
	var issues []string
	visiting := map[string]bool{}
	visited := map[string]bool{}

	var dfs func(string, []string)
	dfs = func(node string, stack []string) {
		if visiting[node] {
			cycle := append(stack, node)
			issues = append(issues, fmt.Sprintf("dependency cycle detected: %s", strings.Join(cycle, " -> ")))
			return
		}
		if visited[node] {
			return
		}
		visiting[node] = true
		for _, next := range adj[node] {
			dfs(next, append(stack, node))
		}
		visiting[node] = false
		visited[node] = true
	}

	keys := make([]string, 0, len(adj))
	for k := range adj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		dfs(k, nil)
	}
	return issues
}

func isSemver(v string) bool {
	return semverRE.MatchString(strings.TrimPrefix(strings.TrimSpace(v), "v"))
}

func isValidOwnedArtifact(p string) bool {
	p = path.Clean(strings.ReplaceAll(p, "\\", "/"))
	_, _, ok := parsePackPath(p)
	return ok
}

func categoryFromID(id string) string {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func expectedPackPath(id string) string {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return path.Clean(".github/instructions/" + parts[0] + "/" + parts[1] + ".instructions.md")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func newJSONExitError(code, message string) *cmdutil.ExitError {
	payload := errorPayload{SchemaVersion: metadataSchemaVersion}
	payload.Error.Code = code
	payload.Error.Message = message
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf(`{"schemaVersion":%d,"error":{"code":"%s","message":%q}}`, metadataSchemaVersion, code, message))
	}
	return &cmdutil.ExitError{
		Code:           1,
		Message:        string(data),
		SuppressPrefix: true,
	}
}
