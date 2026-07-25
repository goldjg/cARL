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
	composeHeaderRE = regexp.MustCompile(`(?i)^\s*<!--\s*(requires|precedence-mode|priority|overrides):\s*(.+?)\s*-->\s*$`)
	intRE           = regexp.MustCompile(`^-?\d+$`)
)

// headerScanLimit bounds how many leading lines of a pack file are scanned
// for metadata headers.
const headerScanLimit = 10

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
	return "Discover, inspect, select, and compose instruction packs"
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
	Bundled       bool     `json:"bundled"`
	Installed     bool     `json:"installed"`
	Selected      bool     `json:"selected"`
	Active        bool     `json:"active"`
	ActiveReasons []string `json:"activeReasons,omitempty"`
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
	Requires    []string
	Mode        string
	Priority    *int
	Overrides   []string
	HasCompose  bool
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
	case "select":
		return c.runSelect(cwd, args[1:], true)
	case "unselect":
		return c.runSelect(cwd, args[1:], false)
	case "effective":
		jsonOut, help, err := parseJSONFlag(args[1:])
		if err != nil {
			return err
		}
		if help {
			fmt.Println("Usage: carl pack effective [--json]")
			return nil
		}
		return c.RunEffectiveInDir(cwd, jsonOut)
	case "profile":
		return c.runProfile(cwd, args[1:])
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

func (c *Command) runSelect(rootDir string, args []string, selecting bool) error {
	verb := "select"
	if !selecting {
		verb = "unselect"
	}
	var ids []string
	jsonOut := false
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			fmt.Printf("Usage: carl pack %s <pack-id>... [--json]\n", verb)
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown argument %q", arg)
			}
			ids = append(ids, arg)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("missing pack ID\n\nUsage: carl pack %s <pack-id>... [--json]", verb)
	}
	if selecting {
		return c.RunSelectInDir(rootDir, ids, jsonOut)
	}
	return c.RunUnselectInDir(rootDir, ids, jsonOut)
}

type selectionPayload struct {
	SchemaVersion int      `json:"schemaVersion"`
	Selected      []string `json:"selected"`
}

// RunSelectInDir executes `carl pack select <pack-id>...`.
func (c *Command) RunSelectInDir(rootDir string, ids []string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return err
	}
	if issues := validatePackSet(packs); len(issues) > 0 {
		return fmt.Errorf("pack metadata validation failed:\n- %s", strings.Join(issues, "\n- "))
	}
	known := map[string]bool{}
	current := make([]string, 0, len(packs))
	for _, p := range packs {
		known[p.ID] = true
		if p.State.Selected {
			current = append(current, p.ID)
		}
	}
	for _, id := range ids {
		if !known[id] {
			msg := fmt.Sprintf("unknown pack %q", id)
			if jsonOut {
				return newJSONExitError("pack_not_found", msg)
			}
			return fmt.Errorf("%s", msg)
		}
	}
	if err := WriteSelection(rootDir, append(current, ids...)); err != nil {
		return err
	}
	return c.reportSelection(rootDir, jsonOut)
}

// RunUnselectInDir executes `carl pack unselect <pack-id>...`.
func (c *Command) RunUnselectInDir(rootDir string, ids []string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return err
	}
	if issues := validatePackSet(packs); len(issues) > 0 {
		return fmt.Errorf("pack metadata validation failed:\n- %s", strings.Join(issues, "\n- "))
	}
	known := map[string]bool{}
	current := make([]string, 0, len(packs))
	for _, p := range packs {
		known[p.ID] = true
		if p.State.Selected {
			current = append(current, p.ID)
		}
	}
	for _, id := range ids {
		if !known[id] {
			msg := fmt.Sprintf("unknown pack %q", id)
			if jsonOut {
				return newJSONExitError("pack_not_found", msg)
			}
			return fmt.Errorf("%s", msg)
		}
	}
	remove := map[string]bool{}
	for _, id := range ids {
		remove[id] = true
	}
	remaining := make([]string, 0, len(current))
	for _, id := range current {
		if !remove[id] {
			remaining = append(remaining, id)
		}
	}
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return err
	}
	if profiles != nil {
		selected := make(map[string]bool, len(remaining))
		for _, id := range remaining {
			selected[id] = true
		}
		if err := ValidateProfileReferences(profiles, selected); err != nil {
			return fmt.Errorf("cannot unselect packs: %w", err)
		}
	}
	if err := WriteSelection(rootDir, remaining); err != nil {
		return err
	}
	return c.reportSelection(rootDir, jsonOut)
}

func (c *Command) reportSelection(rootDir string, jsonOut bool) error {
	sel, err := ReadSelection(rootDir)
	if err != nil {
		return err
	}
	selected := []string{}
	if sel != nil {
		selected = sel.Selected
	}
	if jsonOut {
		data, err := json.MarshalIndent(selectionPayload{
			SchemaVersion: metadataSchemaVersion,
			Selected:      selected,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal selection JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Selection written to %s\n\n", SelectionFileName)
	fmt.Println("Selected Packs:")
	if len(selected) == 0 {
		fmt.Println("  none")
		return nil
	}
	for _, id := range selected {
		fmt.Printf("  %s\n", id)
	}
	return nil
}

type effectivePayload struct {
	SchemaVersion int             `json:"schemaVersion"`
	Packs         []EffectivePack `json:"packs"`
	Conflicts     []Conflict      `json:"conflicts,omitempty"`
}

// RunEffectiveInDir executes `carl pack effective`.
func (c *Command) RunEffectiveInDir(rootDir string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return err
	}
	if issues := validatePackSet(packs); len(issues) > 0 {
		return fmt.Errorf("pack metadata validation failed:\n- %s", strings.Join(issues, "\n- "))
	}
	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		return err
	}

	if jsonOut {
		data, err := json.MarshalIndent(effectivePayload{
			SchemaVersion: metadataSchemaVersion,
			Packs:         set.Packs,
			Conflicts:     set.Conflicts,
		}, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal effective set JSON: %w", err)
		}
		fmt.Println(string(data))
		if len(set.Conflicts) > 0 {
			return &cmdutil.ExitError{Code: 1}
		}
		return nil
	}

	fmt.Println("Effective Pack Set (precedence order):")
	fmt.Println()
	if len(set.Packs) == 0 {
		fmt.Println("  none (no packs selected)")
	} else {
		fmt.Println("  ID                                Version   Priority  Mode               Reasons")
		for _, p := range set.Packs {
			line := fmt.Sprintf(
				"  %-33s %-9s %-9d %-18s %s",
				p.ID, p.Version, p.Priority, p.Mode, strings.Join(p.Reasons, "; "),
			)
			if len(p.OverriddenBy) > 0 {
				line += fmt.Sprintf(" [overridden by: %s]", strings.Join(p.OverriddenBy, ", "))
			}
			fmt.Println(line)
		}
	}
	if len(set.Conflicts) > 0 {
		fmt.Println()
		fmt.Println("Conflicts:")
		fmt.Printf("- %s\n", ConflictSummary(set.Conflicts))
		return &cmdutil.ExitError{Code: 1, Message: "pack composition conflicts detected"}
	}
	return nil
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
	fmt.Println("  list                   List available packs")
	fmt.Println("  show <pack-id>         Show metadata for a pack")
	fmt.Println("  select <pack-id>...    Select packs for this repository")
	fmt.Println("  unselect <pack-id>...  Remove packs from the repository selection")
	fmt.Println("  effective              Show the computed effective pack set")
	fmt.Println("  profile                Inspect and activate policy profiles")
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
	selected, err := discoverSelected(rootDir)
	if err != nil {
		return nil, err
	}

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
		// Legacy compatibility fallback. When profiles.json exists this is
		// replaced below with explicit profile-driven activation.
		state.Active = state.Selected

		// The repository-local copy is authoritative for composition
		// metadata when it declares any; otherwise the bundled copy applies.
		compose := b
		if hasLocal && l.HasCompose {
			compose = l
		}
		var deps []PackDependency
		for _, req := range compose.Requires {
			deps = append(deps, PackDependency{ID: req, Required: true})
		}
		var precedence *Precedence
		if compose.Mode != "" || compose.Priority != nil || len(compose.Overrides) > 0 {
			precedence = &Precedence{
				Mode:      compose.Mode,
				Priority:  compose.Priority,
				Overrides: compose.Overrides,
			}
		}

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
			Dependencies:   deps,
			Precedence:     precedence,
		})
	}

	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return nil, err
	}
	if profiles != nil {
		selectedSet := make(map[string]bool, len(out))
		for _, p := range out {
			if p.State.Selected {
				selectedSet[p.ID] = true
			}
		}
		if err := ValidateProfileReferences(profiles, selectedSet); err != nil {
			return nil, err
		}
		reasons, err := ResolveActivation(profiles, selectedSet)
		if err != nil {
			return nil, err
		}
		for i := range out {
			out[i].State.Active = len(reasons[out[i].ID]) > 0
			out[i].State.ActiveReasons = reasons[out[i].ID]
		}
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
		m, err := parsePackFileMetadata(p, data)
		if err != nil {
			return nil, fmt.Errorf("embedded pack metadata: %w", err)
		}
		result[id] = m
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
		m, err := parsePackFileMetadata(rel, data)
		if err != nil {
			return fmt.Errorf("local pack metadata: %w", err)
		}
		result[id] = m
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan local packs: %w", err)
	}
	return result, nil
}

// discoverSelected returns the set of explicitly selected pack IDs.
// The committed selection artefact (.github/carl/packs.json) is
// authoritative when present; otherwise selection falls back to the legacy
// derivation from runtime.json managed artefacts.
func discoverSelected(rootDir string) (map[string]bool, error) {
	selected := map[string]bool{}

	sel, err := ReadSelection(rootDir)
	if err != nil {
		return nil, err
	}
	if sel != nil {
		for _, id := range sel.Selected {
			selected[id] = true
		}
		return selected, nil
	}

	if !manifest.Exists(rootDir) {
		return selected, nil
	}
	rt, err := manifest.Read(rootDir)
	if err != nil {
		return selected, nil
	}
	for _, m := range rt.ManagedArtifacts {
		id, _, ok := parsePackPath(m)
		if ok {
			selected[id] = true
		}
	}
	return selected, nil
}

func parsePackFileMetadata(packPath string, data []byte) (packFileMetadata, error) {
	m := packFileMetadata{
		Path:        path.Clean(strings.ReplaceAll(packPath, "\\", "/")),
		Version:     extractVersion(data),
		Title:       extractTitle(data),
		Description: extractDescription(data),
	}
	if err := extractComposeHeaders(data, &m); err != nil {
		return m, fmt.Errorf("%s: %w", m.Path, err)
	}
	return m, nil
}

// extractComposeHeaders parses explicit composition metadata headers from the
// leading lines of a pack file. Absent headers are valid: composition
// metadata defaults to no dependencies, additive mode, no priority, and no
// overrides. Malformed values are explicit errors — never silently ignored.
func extractComposeHeaders(data []byte, m *packFileMetadata) error {
	lines := strings.Split(string(data), "\n")
	limit := len(lines)
	if limit > headerScanLimit {
		limit = headerScanLimit
	}
	for i := 0; i < limit; i++ {
		match := composeHeaderRE.FindStringSubmatch(lines[i])
		if len(match) == 0 {
			continue
		}
		key := strings.ToLower(match[1])
		value := strings.TrimSpace(match[2])
		switch key {
		case "requires":
			ids, err := parsePackIDList(value)
			if err != nil {
				return fmt.Errorf("invalid requires header: %w", err)
			}
			m.Requires = ids
			m.HasCompose = true
		case "overrides":
			ids, err := parsePackIDList(value)
			if err != nil {
				return fmt.Errorf("invalid overrides header: %w", err)
			}
			m.Overrides = ids
			m.HasCompose = true
		case "precedence-mode":
			mode := strings.ToLower(value)
			switch mode {
			case "additive", "overridable", "restrictable-only", "immutable":
				m.Mode = mode
				m.HasCompose = true
			default:
				return fmt.Errorf("invalid precedence-mode %q", value)
			}
		case "priority":
			if !intRE.MatchString(value) {
				return fmt.Errorf("invalid priority %q", value)
			}
			var n int
			if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
				return fmt.Errorf("invalid priority %q", value)
			}
			if n < 0 {
				return fmt.Errorf("invalid priority %d: must be non-negative", n)
			}
			m.Priority = &n
			m.HasCompose = true
		}
	}
	return nil
}

func parsePackIDList(value string) ([]string, error) {
	var out []string
	for _, part := range strings.Split(value, ",") {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		if !packIDRE.MatchString(id) {
			return nil, fmt.Errorf("malformed pack id %q", id)
		}
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
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
