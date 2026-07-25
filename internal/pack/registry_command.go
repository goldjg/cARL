package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type registryListPayload struct {
	SchemaVersion int        `json:"schemaVersion"`
	Registries    []Registry `json:"registries"`
}

type registrySearchResult struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Registry    string `json:"registry"`
	Artifact    string `json:"artifact"`
	SHA256      string `json:"sha256"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type registrySearchPayload struct {
	SchemaVersion int                    `json:"schemaVersion"`
	Packs         []registrySearchResult `json:"packs"`
}

type registryChangeResult struct {
	ID            string `json:"id"`
	Action        string `json:"action"`
	FromVersion   string `json:"fromVersion,omitempty"`
	Version       string `json:"version"`
	Registry      string `json:"registry"`
	SHA256        string `json:"sha256"`
	InstalledPath string `json:"installedPath"`
}

type registryChangePayload struct {
	SchemaVersion int                    `json:"schemaVersion"`
	Packs         []registryChangeResult `json:"packs"`
}

func (c *Command) runRegistry(ctx context.Context, rootDir string, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printRegistryUsage()
		return nil
	}
	switch args[0] {
	case "list":
		jsonOut, help, err := parseJSONFlag(args[1:])
		if err != nil {
			return err
		}
		if help {
			fmt.Println("Usage: carl pack registry list [--json]")
			return nil
		}
		return c.RunRegistryListInDir(rootDir, jsonOut)
	case "search":
		query, registryID, jsonOut, help, err := parseRegistrySearchArgs(args[1:])
		if err != nil {
			return registryCommandError("invalid_registry_arguments", err, jsonOut)
		}
		if help {
			fmt.Println("Usage: carl pack registry search [<query>] [--registry <registry-id>] [--json]")
			return nil
		}
		return c.RunRegistrySearchInDir(ctx, rootDir, query, registryID, jsonOut)
	default:
		return fmt.Errorf("unknown registry subcommand %q\n\nRun 'carl pack registry --help' for usage", args[0])
	}
}

func (c *Command) runInstall(ctx context.Context, rootDir string, args []string) error {
	packID, version, registryID, jsonOut, help, err := parseInstallArgs(args)
	if err != nil {
		return registryCommandError("invalid_install_arguments", err, jsonOut)
	}
	if help {
		fmt.Println("Usage: carl pack install <pack-id> [--version <version>] [--registry <registry-id>] [--json]")
		return nil
	}
	return c.RunInstallInDir(ctx, rootDir, packID, version, registryID, jsonOut)
}

func (c *Command) runUpdate(ctx context.Context, rootDir string, args []string) error {
	ids, jsonOut, help, err := parseUpdateArgs(args)
	if err != nil {
		return registryCommandError("invalid_update_arguments", err, jsonOut)
	}
	if help {
		fmt.Println("Usage: carl pack update [<pack-id>...] [--json]")
		return nil
	}
	return c.RunUpdateInDir(ctx, rootDir, ids, jsonOut)
}

// RunRegistryListInDir lists configured registries without fetching them.
func (c *Command) RunRegistryListInDir(rootDir string, jsonOut bool) error {
	config, err := ReadRegistries(rootDir)
	if err != nil {
		return registryCommandError("invalid_registries", err, jsonOut)
	}
	registries := []Registry{}
	if config != nil {
		registries = config.Registries
	}
	if jsonOut {
		return printRegistryJSON(registryListPayload{
			SchemaVersion: metadataSchemaVersion,
			Registries:    registries,
		}, "marshal registry list JSON")
	}
	fmt.Printf("Configured Pack Registries (%s):\n\n", RegistryFileName)
	if len(registries) == 0 {
		fmt.Println("  none configured")
		return nil
	}
	fmt.Println("  ID                        Location")
	for _, registry := range registries {
		fmt.Printf("  %-25s %s\n", registry.ID, registry.Location)
	}
	return nil
}

// RunRegistrySearchInDir fetches explicitly configured indexes and returns
// deterministic matching releases.
func (c *Command) RunRegistrySearchInDir(
	ctx context.Context,
	rootDir, query, registryID string,
	jsonOut bool,
) error {
	loaded, err := c.loadRegistries(ctx, rootDir, registryID)
	if err != nil {
		return registryCommandError("registry_unavailable", err, jsonOut)
	}
	query = strings.ToLower(strings.TrimSpace(query))
	var results []registrySearchResult
	for _, candidate := range allCandidates(loaded) {
		release := candidate.Release
		haystack := strings.ToLower(release.ID + "\n" + release.Title + "\n" + release.Description)
		if query != "" && !strings.Contains(haystack, query) {
			continue
		}
		results = append(results, registrySearchResult{
			ID:          release.ID,
			Version:     release.Version,
			Registry:    candidate.Registry.ID,
			Artifact:    release.Artifact,
			SHA256:      release.SHA256,
			Title:       release.Title,
			Description: release.Description,
		})
	}
	if results == nil {
		results = []registrySearchResult{}
	}
	if jsonOut {
		return printRegistryJSON(registrySearchPayload{
			SchemaVersion: metadataSchemaVersion,
			Packs:         results,
		}, "marshal registry search JSON")
	}
	fmt.Println("Registry Pack Releases:")
	fmt.Println()
	if len(results) == 0 {
		fmt.Println("  none found")
		return nil
	}
	fmt.Println("  ID                                Version       Registry                  SHA-256")
	for _, result := range results {
		fmt.Printf(
			"  %-33s %-13s %-25s %s\n",
			result.ID, result.Version, result.Registry, result.SHA256,
		)
	}
	return nil
}

// RunInstallInDir resolves, verifies, and installs a registry pack and any
// unavailable required dependencies without changing selection or activation.
func (c *Command) RunInstallInDir(
	ctx context.Context,
	rootDir, packID, version, registryID string,
	jsonOut bool,
) error {
	if !packIDRE.MatchString(packID) {
		return registryCommandError("invalid_pack_id", fmt.Errorf("malformed pack id %q", packID), jsonOut)
	}
	if version != "" {
		if _, err := parseSemanticVersion(version); err != nil {
			return registryCommandError("invalid_pack_version", err, jsonOut)
		}
	}
	if registryID != "" && !contextIDRE.MatchString(registryID) {
		return registryCommandError("invalid_registry_id", fmt.Errorf("malformed registry id %q", registryID), jsonOut)
	}
	loaded, err := c.loadRegistries(ctx, rootDir, registryID)
	if err != nil {
		return registryCommandError("registry_unavailable", err, jsonOut)
	}
	candidate, err := resolveCandidate(allCandidates(loaded), packID, version)
	if err != nil {
		return registryCommandError("pack_release_not_found", err, jsonOut)
	}
	existing, err := c.discover(rootDir)
	if err != nil {
		return registryCommandError("invalid_pack_state", err, jsonOut)
	}
	if issues := validatePackSet(existing); len(issues) > 0 {
		return registryCommandError(
			"invalid_pack_state",
			fmt.Errorf("pack metadata validation failed:\n- %s", joinIssues(issues)),
			jsonOut,
		)
	}
	provenance, err := ReadInstalledPacks(rootDir)
	if err != nil {
		return registryCommandError("invalid_installed_provenance", err, jsonOut)
	}
	plan, err := c.planInstall(
		ctx, rootDir, loaded, candidate, existing, provenance, map[string]bool{},
	)
	if err != nil {
		return registryCommandError("pack_install_failed", err, jsonOut)
	}
	if err := validatePlannedSet(existing, plan); err != nil {
		return registryCommandError("pack_install_failed", err, jsonOut)
	}
	if err := applyInstallPlan(rootDir, provenance, plan); err != nil {
		return registryCommandError("pack_install_failed", err, jsonOut)
	}

	results := make([]registryChangeResult, 0, len(plan))
	for _, item := range plan {
		release := item.Candidate.Release
		results = append(results, registryChangeResult{
			ID:            release.ID,
			Action:        "installed",
			Version:       release.Version,
			Registry:      item.Candidate.Registry.ID,
			SHA256:        release.SHA256,
			InstalledPath: expectedPackPath(release.ID),
		})
	}
	return reportRegistryChanges("Installed Packs", results, jsonOut)
}

// RunUpdateInDir updates named or all registry-managed packs from their
// recorded source registries. Equal or older releases are reported unchanged.
func (c *Command) RunUpdateInDir(
	ctx context.Context,
	rootDir string,
	ids []string,
	jsonOut bool,
) error {
	provenance, err := ReadInstalledPacks(rootDir)
	if err != nil {
		return registryCommandError("invalid_installed_provenance", err, jsonOut)
	}
	requested := map[string]bool{}
	if len(ids) == 0 {
		for _, entry := range provenance.Packs {
			requested[entry.ID] = true
		}
	} else {
		for _, id := range ids {
			if !packIDRE.MatchString(id) {
				return registryCommandError("invalid_pack_id", fmt.Errorf("malformed pack id %q", id), jsonOut)
			}
			if requested[id] {
				continue
			}
			if _, ok := installedPackByID(provenance, id); !ok {
				return registryCommandError(
					"pack_not_registry_managed",
					fmt.Errorf("pack %s is not recorded in %s", id, InstalledPacksFileName),
					jsonOut,
				)
			}
			requested[id] = true
		}
	}
	if len(requested) == 0 {
		return reportRegistryChanges("Updated Packs", []registryChangeResult{}, jsonOut)
	}

	requestedIDs := make([]string, 0, len(requested))
	for id := range requested {
		requestedIDs = append(requestedIDs, id)
	}
	sort.Strings(requestedIDs)
	for _, id := range requestedIDs {
		entry, _ := installedPackByID(provenance, id)
		if err := validateInstalledDigest(rootDir, entry); err != nil {
			return registryCommandError("installed_pack_drifted", err, jsonOut)
		}
	}

	loadedByRegistry := map[string][]loadedRegistry{}
	for _, id := range requestedIDs {
		entry, _ := installedPackByID(provenance, id)
		if _, ok := loadedByRegistry[entry.Registry]; ok {
			continue
		}
		loaded, err := c.loadRegistries(ctx, rootDir, entry.Registry)
		if err != nil {
			return registryCommandError("registry_unavailable", err, jsonOut)
		}
		if len(loaded) != 1 || loaded[0].Registry.Location != entry.RegistryLocation {
			return registryCommandError(
				"registry_provenance_mismatch",
				fmt.Errorf("registry %s location does not match recorded provenance", entry.Registry),
				jsonOut,
			)
		}
		loadedByRegistry[entry.Registry] = loaded
	}

	existing, err := c.discover(rootDir)
	if err != nil {
		return registryCommandError("invalid_pack_state", err, jsonOut)
	}
	if issues := validatePackSet(existing); len(issues) > 0 {
		return registryCommandError(
			"invalid_pack_state",
			fmt.Errorf("pack metadata validation failed:\n- %s", joinIssues(issues)),
			jsonOut,
		)
	}

	plannedByID := map[string]plannedPack{}
	resultsByID := map[string]registryChangeResult{}
	for _, id := range requestedIDs {
		current, _ := installedPackByID(provenance, id)
		loaded := loadedByRegistry[current.Registry]
		candidate, err := resolveCandidate(allCandidates(loaded), id, "")
		if err != nil {
			return registryCommandError("pack_release_not_found", err, jsonOut)
		}
		cmp, err := compareSemanticVersions(candidate.Release.Version, current.Version)
		if err != nil {
			return registryCommandError("invalid_pack_version", err, jsonOut)
		}
		if cmp == 0 && candidate.Release.SHA256 != current.SHA256 {
			return registryCommandError(
				"registry_release_mutated",
				fmt.Errorf(
					"registry %s changed the SHA-256 for immutable release %s@%s",
					current.Registry, id, current.Version,
				),
				jsonOut,
			)
		}
		if cmp <= 0 {
			resultsByID[id] = registryChangeResult{
				ID:            id,
				Action:        "unchanged",
				FromVersion:   current.Version,
				Version:       current.Version,
				Registry:      current.Registry,
				SHA256:        current.SHA256,
				InstalledPath: current.InstalledPath,
			}
			continue
		}
		plan, err := c.planInstall(ctx, rootDir, loaded, candidate, existing, provenance, requested)
		if err != nil {
			return registryCommandError("pack_update_failed", err, jsonOut)
		}
		for _, item := range plan {
			if previous, ok := plannedByID[item.Candidate.Release.ID]; ok &&
				(previous.Candidate.Registry.ID != item.Candidate.Registry.ID ||
					previous.Candidate.Release.Version != item.Candidate.Release.Version) {
				return registryCommandError(
					"pack_update_failed",
					fmt.Errorf(
						"conflicting planned releases for %s: %s@%s and %s@%s",
						item.Candidate.Release.ID,
						previous.Candidate.Registry.ID,
						previous.Candidate.Release.Version,
						item.Candidate.Registry.ID,
						item.Candidate.Release.Version,
					),
					jsonOut,
				)
			}
			plannedByID[item.Candidate.Release.ID] = item
		}
		resultsByID[id] = registryChangeResult{
			ID:            id,
			Action:        "updated",
			FromVersion:   current.Version,
			Version:       candidate.Release.Version,
			Registry:      current.Registry,
			SHA256:        candidate.Release.SHA256,
			InstalledPath: current.InstalledPath,
		}
	}

	plan := make([]plannedPack, 0, len(plannedByID))
	for _, item := range plannedByID {
		plan = append(plan, item)
	}
	sort.Slice(plan, func(i, j int) bool {
		return plan[i].Candidate.Release.ID < plan[j].Candidate.Release.ID
	})
	if len(plan) > 0 {
		if err := validatePlannedSet(existing, plan); err != nil {
			return registryCommandError("pack_update_failed", err, jsonOut)
		}
		if err := applyInstallPlan(rootDir, provenance, plan); err != nil {
			return registryCommandError("pack_update_failed", err, jsonOut)
		}
		for _, item := range plan {
			id := item.Candidate.Release.ID
			if _, exists := resultsByID[id]; exists {
				continue
			}
			release := item.Candidate.Release
			resultsByID[id] = registryChangeResult{
				ID:            id,
				Action:        "installed",
				Version:       release.Version,
				Registry:      item.Candidate.Registry.ID,
				SHA256:        release.SHA256,
				InstalledPath: expectedPackPath(id),
			}
		}
	}
	results := make([]registryChangeResult, 0, len(resultsByID))
	for _, result := range resultsByID {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return reportRegistryChanges("Updated Packs", results, jsonOut)
}

func parseRegistrySearchArgs(args []string) (query, registryID string, jsonOut, help bool, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			help = true
		case "--registry":
			if i+1 >= len(args) {
				err = fmt.Errorf("--registry requires a registry ID")
				return
			}
			i++
			registryID = args[i]
		default:
			if strings.HasPrefix(args[i], "-") {
				err = fmt.Errorf("unknown argument %q", args[i])
				return
			}
			if query != "" {
				err = fmt.Errorf("registry search accepts at most one query")
				return
			}
			query = args[i]
		}
	}
	return
}

func parseInstallArgs(args []string) (packID, version, registryID string, jsonOut, help bool, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			help = true
		case "--version":
			if i+1 >= len(args) {
				err = fmt.Errorf("--version requires a semantic version")
				return
			}
			i++
			version = args[i]
		case "--registry":
			if i+1 >= len(args) {
				err = fmt.Errorf("--registry requires a registry ID")
				return
			}
			i++
			registryID = args[i]
		default:
			if strings.HasPrefix(args[i], "-") {
				err = fmt.Errorf("unknown argument %q", args[i])
				return
			}
			if packID != "" {
				err = fmt.Errorf("install accepts exactly one pack ID")
				return
			}
			packID = args[i]
		}
	}
	if !help && packID == "" {
		err = fmt.Errorf("missing pack ID")
	}
	return
}

func parseUpdateArgs(args []string) (ids []string, jsonOut, help bool, err error) {
	for _, arg := range args {
		switch arg {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			help = true
		default:
			if strings.HasPrefix(arg, "-") {
				err = fmt.Errorf("unknown argument %q", arg)
				return
			}
			ids = append(ids, arg)
		}
	}
	return
}

func registryCommandError(code string, err error, jsonOut bool) error {
	if jsonOut {
		return newJSONExitError(code, err.Error())
	}
	return err
}

func printRegistryJSON(payload any, context string) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %w", context, err)
	}
	fmt.Println(string(data))
	return nil
}

func reportRegistryChanges(title string, results []registryChangeResult, jsonOut bool) error {
	if results == nil {
		results = []registryChangeResult{}
	}
	if jsonOut {
		return printRegistryJSON(registryChangePayload{
			SchemaVersion: metadataSchemaVersion,
			Packs:         results,
		}, "marshal registry change JSON")
	}
	fmt.Printf("%s:\n\n", title)
	if len(results) == 0 {
		fmt.Println("  none")
		return nil
	}
	fmt.Println("  ID                                Action     Version       Registry")
	for _, result := range results {
		version := result.Version
		if result.FromVersion != "" && result.FromVersion != result.Version {
			version = result.FromVersion + " -> " + result.Version
		}
		fmt.Printf("  %-33s %-10s %-13s %s\n", result.ID, result.Action, version, result.Registry)
	}
	fmt.Println()
	fmt.Println("Every written artifact passed SHA-256 integrity verification.")
	fmt.Println("SHA-256 does not authenticate publisher identity.")
	return nil
}

func printRegistryUsage() {
	fmt.Println("Usage: carl pack registry <subcommand> [arguments]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list                    List explicitly configured registries")
	fmt.Println("  search [<query>]        Search registry pack releases")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --registry <id>  Restrict search to one registry")
	fmt.Println("  --json           Print machine-readable JSON output")
}
