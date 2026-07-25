package pack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ProfileFileName is the path of the user-owned profile artefact relative to
// the repository root.
const ProfileFileName = ".github/carl/profiles.json"

var contextIDRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Profiles is the schema-versioned policy profile artefact.
type Profiles struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Defaults      ProfileDefaults      `json:"defaults"`
	Profiles      []PolicyProfile      `json:"profiles"`
	Active        ActiveProfileContext `json:"active"`
}

// ProfileDefaults are additive pack sets that apply to every active context.
type ProfileDefaults struct {
	Organization []string `json:"organization"`
	Repository   []string `json:"repository"`
}

// PolicyProfile defines base packs and optional role/task overlays.
type PolicyProfile struct {
	ID          string              `json:"id"`
	Description string              `json:"description,omitempty"`
	Packs       []string            `json:"packs"`
	Roles       map[string][]string `json:"roles,omitempty"`
	Tasks       map[string][]string `json:"tasks,omitempty"`
}

// ActiveProfileContext identifies the profile and optional overlays currently
// active for the repository.
type ActiveProfileContext struct {
	Profile string `json:"profile,omitempty"`
	Role    string `json:"role,omitempty"`
	Task    string `json:"task,omitempty"`
}

// ReadProfiles reads and strictly validates profiles.json. A missing file is
// not an error: callers retain the legacy selected-as-active behaviour.
func ReadProfiles(rootDir string) (*Profiles, error) {
	p := filepath.Join(rootDir, filepath.FromSlash(ProfileFileName))
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", ProfileFileName, err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var profiles Profiles
	if err := dec.Decode(&profiles); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ProfileFileName, err)
	}
	if err := ensureJSONEOF(dec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ProfileFileName, err)
	}
	if issues := ValidateProfiles(&profiles); len(issues) > 0 {
		return nil, fmt.Errorf("%s validation failed:\n- %s", ProfileFileName, joinIssues(issues))
	}
	normalizeProfiles(&profiles)
	return &profiles, nil
}

// WriteProfiles validates and persists profiles.json deterministically.
func WriteProfiles(rootDir string, profiles *Profiles) error {
	if profiles == nil {
		return fmt.Errorf("profiles must not be nil")
	}
	normalizeProfiles(profiles)
	if issues := ValidateProfiles(profiles); len(issues) > 0 {
		return fmt.Errorf("%s validation failed:\n- %s", ProfileFileName, joinIssues(issues))
	}

	p := filepath.Join(rootDir, filepath.FromSlash(ProfileFileName))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("create %s directory: %w", ProfileFileName, err)
	}
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", ProfileFileName, err)
	}
	return os.WriteFile(p, append(data, '\n'), 0644)
}

// ValidateProfiles validates the profile schema without relying on repository
// state. Pack availability and selection are checked by ResolveActivation.
func ValidateProfiles(profiles *Profiles) []string {
	if profiles == nil {
		return []string{"profile document is nil"}
	}
	var issues []string
	if profiles.SchemaVersion != metadataSchemaVersion {
		issues = append(issues, fmt.Sprintf("unsupported schema version %d", profiles.SchemaVersion))
	}
	if profiles.Defaults.Organization == nil {
		issues = append(issues, "missing defaults.organization array")
	}
	if profiles.Defaults.Repository == nil {
		issues = append(issues, "missing defaults.repository array")
	}
	if profiles.Profiles == nil {
		issues = append(issues, "missing profiles array")
	}

	validatePackRefs := func(owner string, refs []string) {
		seen := map[string]bool{}
		for _, id := range refs {
			if !packIDRE.MatchString(id) {
				issues = append(issues, fmt.Sprintf("%s: malformed pack id %q", owner, id))
			}
			if seen[id] {
				issues = append(issues, fmt.Sprintf("%s: duplicate pack id %q", owner, id))
			}
			seen[id] = true
		}
	}
	validatePackRefs("defaults.organization", profiles.Defaults.Organization)
	validatePackRefs("defaults.repository", profiles.Defaults.Repository)

	seenProfiles := map[string]bool{}
	index := map[string]PolicyProfile{}
	for _, profile := range profiles.Profiles {
		if !contextIDRE.MatchString(profile.ID) {
			issues = append(issues, fmt.Sprintf("profile: malformed id %q", profile.ID))
		}
		if seenProfiles[profile.ID] {
			issues = append(issues, fmt.Sprintf("duplicate profile id %q", profile.ID))
		}
		seenProfiles[profile.ID] = true
		index[profile.ID] = profile
		if profile.Packs == nil {
			issues = append(issues, fmt.Sprintf("profile %s: missing packs array", profile.ID))
		}
		validatePackRefs(fmt.Sprintf("profile %s", profile.ID), profile.Packs)
		for role, refs := range profile.Roles {
			if !contextIDRE.MatchString(role) {
				issues = append(issues, fmt.Sprintf("profile %s: malformed role id %q", profile.ID, role))
			}
			validatePackRefs(fmt.Sprintf("profile %s role %s", profile.ID, role), refs)
		}
		for task, refs := range profile.Tasks {
			if !contextIDRE.MatchString(task) {
				issues = append(issues, fmt.Sprintf("profile %s: malformed task id %q", profile.ID, task))
			}
			validatePackRefs(fmt.Sprintf("profile %s task %s", profile.ID, task), refs)
		}
	}

	active := profiles.Active
	if active.Profile == "" {
		if active.Role != "" || active.Task != "" {
			issues = append(issues, "active role or task requires an active profile")
		}
	} else if !contextIDRE.MatchString(active.Profile) {
		issues = append(issues, fmt.Sprintf("active: malformed profile id %q", active.Profile))
	} else if profile, ok := index[active.Profile]; !ok {
		issues = append(issues, fmt.Sprintf("active: unknown profile %q", active.Profile))
	} else {
		if active.Role != "" {
			if !contextIDRE.MatchString(active.Role) {
				issues = append(issues, fmt.Sprintf("active: malformed role id %q", active.Role))
			} else if _, ok := profile.Roles[active.Role]; !ok {
				issues = append(issues, fmt.Sprintf("active: unknown role %q for profile %q", active.Role, active.Profile))
			}
		}
		if active.Task != "" {
			if !contextIDRE.MatchString(active.Task) {
				issues = append(issues, fmt.Sprintf("active: malformed task id %q", active.Task))
			} else if _, ok := profile.Tasks[active.Task]; !ok {
				issues = append(issues, fmt.Sprintf("active: unknown task %q for profile %q", active.Task, active.Profile))
			}
		}
	}
	sort.Strings(issues)
	return issues
}

// ResolveActivation returns the explicitly active pack seeds and their
// provenance. Every referenced pack must exist in the repository selection.
func ResolveActivation(profiles *Profiles, selected map[string]bool) (map[string][]string, error) {
	if issues := ValidateProfiles(profiles); len(issues) > 0 {
		return nil, fmt.Errorf("%s validation failed:\n- %s", ProfileFileName, joinIssues(issues))
	}

	reasons := map[string][]string{}
	add := func(owner string, refs []string) error {
		for _, id := range refs {
			if !selected[id] {
				return fmt.Errorf("%s: %s references pack %q, which is not selected", ProfileFileName, owner, id)
			}
			reasons[id] = appendUnique(reasons[id], owner)
		}
		return nil
	}
	if err := add("organization default", profiles.Defaults.Organization); err != nil {
		return nil, err
	}
	if err := add("repository default", profiles.Defaults.Repository); err != nil {
		return nil, err
	}

	if profiles.Active.Profile != "" {
		profile := profileByID(profiles.Profiles, profiles.Active.Profile)
		if err := add("profile "+profile.ID, profile.Packs); err != nil {
			return nil, err
		}
		if profiles.Active.Role != "" {
			reason := fmt.Sprintf("role %s in profile %s", profiles.Active.Role, profile.ID)
			if err := add(reason, profile.Roles[profiles.Active.Role]); err != nil {
				return nil, err
			}
		}
		if profiles.Active.Task != "" {
			reason := fmt.Sprintf("task %s in profile %s", profiles.Active.Task, profile.ID)
			if err := add(reason, profile.Tasks[profiles.Active.Task]); err != nil {
				return nil, err
			}
		}
	}
	for id := range reasons {
		sort.Strings(reasons[id])
	}
	return reasons, nil
}

// ValidateProfileReferences checks every profile/default pack reference
// against the selected pack set, including currently inactive profiles.
func ValidateProfileReferences(profiles *Profiles, selected map[string]bool) error {
	if issues := ValidateProfiles(profiles); len(issues) > 0 {
		return fmt.Errorf("%s validation failed:\n- %s", ProfileFileName, joinIssues(issues))
	}
	check := func(owner string, refs []string) error {
		for _, id := range refs {
			if !selected[id] {
				return fmt.Errorf("%s: %s references pack %q, which is not selected", ProfileFileName, owner, id)
			}
		}
		return nil
	}
	if err := check("organization defaults", profiles.Defaults.Organization); err != nil {
		return err
	}
	if err := check("repository defaults", profiles.Defaults.Repository); err != nil {
		return err
	}
	for _, profile := range profiles.Profiles {
		if err := check("profile "+profile.ID, profile.Packs); err != nil {
			return err
		}
		for role, refs := range profile.Roles {
			if err := check(fmt.Sprintf("profile %s role %s", profile.ID, role), refs); err != nil {
				return err
			}
		}
		for task, refs := range profile.Tasks {
			if err := check(fmt.Sprintf("profile %s task %s", profile.ID, task), refs); err != nil {
				return err
			}
		}
	}
	return nil
}

func profileByID(profiles []PolicyProfile, id string) PolicyProfile {
	for _, profile := range profiles {
		if profile.ID == id {
			return profile
		}
	}
	return PolicyProfile{}
}

func normalizeProfiles(profiles *Profiles) {
	profiles.Defaults.Organization = sortedUnique(profiles.Defaults.Organization)
	profiles.Defaults.Repository = sortedUnique(profiles.Defaults.Repository)
	for i := range profiles.Profiles {
		profiles.Profiles[i].Packs = sortedUnique(profiles.Profiles[i].Packs)
		for role, refs := range profiles.Profiles[i].Roles {
			profiles.Profiles[i].Roles[role] = sortedUnique(refs)
		}
		for task, refs := range profiles.Profiles[i].Tasks {
			profiles.Profiles[i].Tasks[task] = sortedUnique(refs)
		}
	}
	sort.Slice(profiles.Profiles, func(i, j int) bool {
		return profiles.Profiles[i].ID < profiles.Profiles[j].ID
	})
}

func sortedUnique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func ensureJSONEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected additional JSON value")
		}
		return err
	}
	return nil
}

func joinIssues(issues []string) string {
	return strings.Join(issues, "\n- ")
}
