package pack

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type profileListPayload struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Defaults      ProfileDefaults      `json:"defaults"`
	Profiles      []PolicyProfile      `json:"profiles"`
	Active        ActiveProfileContext `json:"active"`
}

type profileShowPayload struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Profile       PolicyProfile        `json:"profile"`
	Active        ActiveProfileContext `json:"active"`
}

type profileActivationPayload struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Active        ActiveProfileContext `json:"active"`
}

func (c *Command) runProfile(rootDir string, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printProfileUsage()
		return nil
	}

	switch args[0] {
	case "list":
		jsonOut, help, err := parseJSONFlag(args[1:])
		if err != nil {
			return err
		}
		if help {
			fmt.Println("Usage: carl pack profile list [--json]")
			return nil
		}
		return c.RunProfileListInDir(rootDir, jsonOut)
	case "show":
		if len(args) >= 2 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Println("Usage: carl pack profile show <profile-id> [--json]")
			return nil
		}
		if len(args) < 2 {
			return fmt.Errorf("missing profile ID\n\nUsage: carl pack profile show <profile-id> [--json]")
		}
		jsonOut, help, err := parseJSONFlag(args[2:])
		if err != nil {
			return err
		}
		if help {
			fmt.Println("Usage: carl pack profile show <profile-id> [--json]")
			return nil
		}
		return c.RunProfileShowInDir(rootDir, args[1], jsonOut)
	case "activate":
		return c.runProfileActivate(rootDir, args[1:])
	case "clear":
		jsonOut, help, err := parseJSONFlag(args[1:])
		if err != nil {
			return err
		}
		if help {
			fmt.Println("Usage: carl pack profile clear [--json]")
			return nil
		}
		return c.RunProfileClearInDir(rootDir, jsonOut)
	default:
		return fmt.Errorf("unknown profile subcommand %q\n\nRun 'carl pack profile --help' for usage", args[0])
	}
}

func (c *Command) runProfileActivate(rootDir string, args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Println("Usage: carl pack profile activate <profile-id> [--role <role-id>] [--task <task-id>] [--json]")
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("missing profile ID\n\nUsage: carl pack profile activate <profile-id> [--role <role-id>] [--task <task-id>] [--json]")
	}
	profileID := args[0]
	var roleID, taskID string
	jsonOut := false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--help", "-h":
			fmt.Println("Usage: carl pack profile activate <profile-id> [--role <role-id>] [--task <task-id>] [--json]")
			return nil
		case "--role":
			if i+1 >= len(args) {
				return fmt.Errorf("--role requires a role ID")
			}
			i++
			roleID = args[i]
		case "--task":
			if i+1 >= len(args) {
				return fmt.Errorf("--task requires a task ID")
			}
			i++
			taskID = args[i]
		default:
			return fmt.Errorf("unknown argument %q", args[i])
		}
	}
	return c.RunProfileActivateInDir(rootDir, profileID, roleID, taskID, jsonOut)
}

// RunProfileListInDir executes `carl pack profile list`.
func (c *Command) RunProfileListInDir(rootDir string, jsonOut bool) error {
	if _, err := c.discover(rootDir); err != nil {
		return profileInputError(err, jsonOut)
	}
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return profileInputError(err, jsonOut)
	}
	if profiles == nil {
		profiles = emptyProfiles()
	}

	if jsonOut {
		return printProfileJSON(profileListPayload{
			SchemaVersion: metadataSchemaVersion,
			Defaults:      profiles.Defaults,
			Profiles:      profiles.Profiles,
			Active:        profiles.Active,
		}, "marshal profile list JSON")
	}

	fmt.Printf("Policy Profiles (%s):\n\n", ProfileFileName)
	if len(profiles.Profiles) == 0 {
		fmt.Println("  none configured")
	} else {
		fmt.Println("  ID                        Active  Packs  Roles  Tasks  Description")
		for _, profile := range profiles.Profiles {
			active := ""
			if profile.ID == profiles.Active.Profile {
				active = "yes"
			}
			fmt.Printf(
				"  %-25s %-7s %-6d %-6d %-6d %s\n",
				profile.ID,
				active,
				len(profile.Packs),
				len(profile.Roles),
				len(profile.Tasks),
				shorten(profile.Description, 56),
			)
		}
	}
	printActiveContext(profiles.Active)
	return nil
}

// RunProfileShowInDir executes `carl pack profile show <profile-id>`.
func (c *Command) RunProfileShowInDir(rootDir, profileID string, jsonOut bool) error {
	if _, err := c.discover(rootDir); err != nil {
		return profileInputError(err, jsonOut)
	}
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return profileInputError(err, jsonOut)
	}
	if profiles == nil {
		return profileNotFound(profileID, jsonOut)
	}
	profile, ok := findProfile(profiles.Profiles, profileID)
	if !ok {
		return profileNotFound(profileID, jsonOut)
	}

	if jsonOut {
		return printProfileJSON(profileShowPayload{
			SchemaVersion: metadataSchemaVersion,
			Profile:       profile,
			Active:        profiles.Active,
		}, "marshal profile show JSON")
	}

	fmt.Printf("Profile: %s\n\n", profile.ID)
	fmt.Printf("Description: %s\n", valueOrNone(profile.Description))
	fmt.Printf("Active:      %t\n", profile.ID == profiles.Active.Profile)
	printPackRefs("Packs", profile.Packs)
	printOverlayRefs("Roles", profile.Roles)
	printOverlayRefs("Tasks", profile.Tasks)
	if profile.ID == profiles.Active.Profile {
		printActiveContext(profiles.Active)
	}
	return nil
}

// RunProfileActivateInDir persists an active profile context.
func (c *Command) RunProfileActivateInDir(rootDir, profileID, roleID, taskID string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return profileInputError(err, jsonOut)
	}
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return profileInputError(err, jsonOut)
	}
	if profiles == nil {
		msg := fmt.Sprintf("%s does not exist; define profiles before activation", ProfileFileName)
		if jsonOut {
			return newJSONExitError("profiles_not_configured", msg)
		}
		return fmt.Errorf("%s", msg)
	}
	if _, ok := findProfile(profiles.Profiles, profileID); !ok {
		return profileNotFound(profileID, jsonOut)
	}

	profiles.Active = ActiveProfileContext{
		Profile: profileID,
		Role:    roleID,
		Task:    taskID,
	}
	if issues := ValidateProfiles(profiles); len(issues) > 0 {
		msg := strings.Join(issues, "; ")
		if jsonOut {
			return newJSONExitError("invalid_profile_context", msg)
		}
		return fmt.Errorf("invalid profile context: %s", msg)
	}
	if err := ValidateProfileReferences(profiles, selectedPackSet(packs)); err != nil {
		if jsonOut {
			return newJSONExitError("invalid_profile_reference", err.Error())
		}
		return err
	}
	if err := WriteProfiles(rootDir, profiles); err != nil {
		return err
	}
	return reportProfileActivation(profiles.Active, jsonOut)
}

// RunProfileClearInDir clears the active profile and overlays. Organisation
// and repository defaults remain active.
func (c *Command) RunProfileClearInDir(rootDir string, jsonOut bool) error {
	packs, err := c.discover(rootDir)
	if err != nil {
		return profileInputError(err, jsonOut)
	}
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		return profileInputError(err, jsonOut)
	}
	if profiles == nil {
		msg := fmt.Sprintf("%s does not exist", ProfileFileName)
		if jsonOut {
			return newJSONExitError("profiles_not_configured", msg)
		}
		return fmt.Errorf("%s", msg)
	}
	profiles.Active = ActiveProfileContext{}
	if err := ValidateProfileReferences(profiles, selectedPackSet(packs)); err != nil {
		return err
	}
	if err := WriteProfiles(rootDir, profiles); err != nil {
		return err
	}
	return reportProfileActivation(profiles.Active, jsonOut)
}

func reportProfileActivation(active ActiveProfileContext, jsonOut bool) error {
	if jsonOut {
		return printProfileJSON(profileActivationPayload{
			SchemaVersion: metadataSchemaVersion,
			Active:        active,
		}, "marshal profile activation JSON")
	}
	fmt.Printf("Active profile context written to %s\n", ProfileFileName)
	printActiveContext(active)
	return nil
}

func printProfileUsage() {
	fmt.Println("Usage: carl pack profile <subcommand> [arguments]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list                                      List policy profiles")
	fmt.Println("  show <profile-id>                         Show a policy profile")
	fmt.Println("  activate <profile-id> [--role] [--task]  Activate a profile context")
	fmt.Println("  clear                                     Clear the active profile context")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --json         Print machine-readable JSON output")
}

func printActiveContext(active ActiveProfileContext) {
	fmt.Println()
	fmt.Println("Active Context:")
	if active.Profile == "" {
		fmt.Println("  profile: none (defaults only)")
		return
	}
	fmt.Printf("  profile: %s\n", active.Profile)
	fmt.Printf("  role:    %s\n", valueOrNone(active.Role))
	fmt.Printf("  task:    %s\n", valueOrNone(active.Task))
}

func printPackRefs(label string, refs []string) {
	fmt.Printf("\n%s:\n", label)
	if len(refs) == 0 {
		fmt.Println("  none")
		return
	}
	for _, ref := range refs {
		fmt.Printf("  %s\n", ref)
	}
}

func printOverlayRefs(label string, overlays map[string][]string) {
	fmt.Printf("\n%s:\n", label)
	if len(overlays) == 0 {
		fmt.Println("  none")
		return
	}
	keys := make([]string, 0, len(overlays))
	for key := range overlays {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  %s: %s\n", key, strings.Join(overlays[key], ", "))
	}
}

func selectedPackSet(packs []PackMetadata) map[string]bool {
	selected := make(map[string]bool, len(packs))
	for _, pack := range packs {
		if pack.State.Selected {
			selected[pack.ID] = true
		}
	}
	return selected
}

func findProfile(profiles []PolicyProfile, id string) (PolicyProfile, bool) {
	for _, profile := range profiles {
		if profile.ID == id {
			return profile, true
		}
	}
	return PolicyProfile{}, false
}

func profileNotFound(profileID string, jsonOut bool) error {
	msg := fmt.Sprintf("unknown profile %q", profileID)
	if jsonOut {
		return newJSONExitError("profile_not_found", msg)
	}
	return fmt.Errorf("%s", msg)
}

func profileInputError(err error, jsonOut bool) error {
	if jsonOut {
		return newJSONExitError("invalid_profiles", err.Error())
	}
	return err
}

func emptyProfiles() *Profiles {
	return &Profiles{
		SchemaVersion: metadataSchemaVersion,
		Defaults: ProfileDefaults{
			Organization: []string{},
			Repository:   []string{},
		},
		Profiles: []PolicyProfile{},
	}
}

func printProfileJSON(payload any, context string) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %w", context, err)
	}
	fmt.Println(string(data))
	return nil
}
