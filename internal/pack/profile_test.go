package pack

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goldjg/carl/embedded"
	"github.com/goldjg/carl/internal/cmdutil"
	"github.com/goldjg/carl/internal/manifest"
)

var defaultProfilePackIDs = []string{
	"cloud/azure",
	"cloud/entra",
	"cloud/gcp",
	"cloud/microsoft-graph",
	"cloud/netlify",
	"core/baseline",
	"core/carl",
	"core/cognition-governance",
	"core/dependency",
	"core/identity",
	"core/memory-cache",
	"core/pr-contract",
	"core/security",
	"core/tool-permission-tiers",
	"languages/go",
	"languages/html",
	"languages/javascript",
	"languages/powershell",
	"languages/python",
	"languages/terraform",
	"languages/typescript",
	"platform/cicd",
	"platform/docker",
	"platform/kubernetes",
}

func testProfiles() *Profiles {
	return &Profiles{
		SchemaVersion: metadataSchemaVersion,
		Defaults: ProfileDefaults{
			Organization: []string{"core/security"},
			Repository:   []string{"core/baseline"},
		},
		Profiles: []PolicyProfile{
			{
				ID:          "developer",
				Description: "Default implementation context.",
				Packs:       []string{"languages/go"},
				Roles: map[string][]string{
					"reviewer": {"core/pr-contract"},
				},
				Tasks: map[string][]string{
					"security-review": {"core/identity"},
				},
			},
		},
		Active: ActiveProfileContext{
			Profile: "developer",
			Role:    "reviewer",
			Task:    "security-review",
		},
	}
}

// Contract assertion 1: profile persistence is schema-versioned,
// deterministic, and strict about malformed or unknown content.
func TestProfilesWriteReadDeterministic(t *testing.T) {
	dir := t.TempDir()
	profiles := testProfiles()
	profiles.Defaults.Repository = append(profiles.Defaults.Repository, "core/baseline")
	profiles.Profiles = append(profiles.Profiles, PolicyProfile{
		ID:    "auditor",
		Packs: []string{"core/security", "core/security"},
	})

	if err := WriteProfiles(dir, profiles); err != nil {
		t.Fatalf("WriteProfiles: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(ProfileFileName)))
	if err != nil {
		t.Fatal(err)
	}

	read, err := ReadProfiles(dir)
	if err != nil {
		t.Fatalf("ReadProfiles: %v", err)
	}
	if len(read.Profiles) != 2 || read.Profiles[0].ID != "auditor" || read.Profiles[1].ID != "developer" {
		t.Fatalf("profiles are not sorted: %+v", read.Profiles)
	}
	if len(read.Profiles[0].Packs) != 1 || len(read.Defaults.Repository) != 1 {
		t.Fatalf("pack references are not deduplicated: %+v", read)
	}

	if err := WriteProfiles(dir, read); err != nil {
		t.Fatalf("second WriteProfiles: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(ProfileFileName)))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("profile output is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestReadProfilesAbsentAndMalformed(t *testing.T) {
	profiles, err := ReadProfiles(t.TempDir())
	if err != nil {
		t.Fatalf("ReadProfiles absent: %v", err)
	}
	if profiles != nil {
		t.Fatalf("expected nil profiles, got %+v", profiles)
	}

	cases := map[string]string{
		"invalid json":     `{not json`,
		"unknown field":    `{"schemaVersion":1,"defaults":{"organization":[],"repository":[]},"profiles":[],"active":{},"extra":true}`,
		"unknown profile":  `{"schemaVersion":1,"defaults":{"organization":[],"repository":[]},"profiles":[],"active":{"profile":"missing"}}`,
		"role without one": `{"schemaVersion":1,"defaults":{"organization":[],"repository":[]},"profiles":[],"active":{"role":"reviewer"}}`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeProfileFile(t, dir, content)
			if _, err := ReadProfiles(dir); err == nil {
				t.Fatal("expected explicit profile validation error")
			}
		})
	}
}

func TestProfileCommandJSONErrorIsStructured(t *testing.T) {
	dir := t.TempDir()
	writeProfileFile(t, dir, `{"schemaVersion":99}`)
	cmd := New(bundledArtifacts())

	err := cmd.RunProfileListInDir(dir, true)
	var exitErr *cmdutil.ExitError
	if !errors.As(err, &exitErr) || !exitErr.SuppressPrefix {
		t.Fatalf("expected structured ExitError, got %v", err)
	}
	var payload errorPayload
	if err := json.Unmarshal([]byte(exitErr.Message), &payload); err != nil {
		t.Fatalf("error payload is not JSON: %v\n%s", err, exitErr.Message)
	}
	if payload.Error.Code != "invalid_profiles" {
		t.Fatalf("error code = %q", payload.Error.Code)
	}
}

// Contract assertion 2: defaults, profile packs, role overlays, and task
// overlays compose additively with explicit provenance.
func TestResolveProfileActivation(t *testing.T) {
	profiles := testProfiles()
	selected := map[string]bool{
		"core/security":    true,
		"core/baseline":    true,
		"languages/go":     true,
		"core/pr-contract": true,
		"core/identity":    true,
	}
	reasons, err := ResolveActivation(profiles, selected)
	if err != nil {
		t.Fatalf("ResolveActivation: %v", err)
	}

	want := map[string]string{
		"core/security":    "organization default",
		"core/baseline":    "repository default",
		"languages/go":     "profile developer",
		"core/pr-contract": "role reviewer in profile developer",
		"core/identity":    "task security-review in profile developer",
	}
	if len(reasons) != len(want) {
		t.Fatalf("activation = %+v; want %d packs", reasons, len(want))
	}
	for id, reason := range want {
		if len(reasons[id]) != 1 || reasons[id][0] != reason {
			t.Fatalf("%s reasons = %v; want %q", id, reasons[id], reason)
		}
	}
}

// Contract assertion 3: profile references must resolve to selected packs.
func TestProfileReferencesRequireSelectedPacks(t *testing.T) {
	profiles := testProfiles()
	selected := map[string]bool{
		"core/security":    true,
		"core/baseline":    true,
		"languages/go":     true,
		"core/pr-contract": true,
	}
	if err := ValidateProfileReferences(profiles, selected); err == nil ||
		!strings.Contains(err.Error(), "core/identity") {
		t.Fatalf("expected unselected core/identity error, got %v", err)
	}
}

func TestDefaultProfileExampleSchemaAndReferences(t *testing.T) {
	dir := t.TempDir()
	profiles := installDefaultProfileExample(t, dir)

	if profiles.SchemaVersion != metadataSchemaVersion {
		t.Fatalf("schemaVersion = %d; want %d", profiles.SchemaVersion, metadataSchemaVersion)
	}
	if len(profiles.Defaults.Organization) != 0 || len(profiles.Defaults.Repository) != 0 {
		t.Fatalf("default context = %+v; want empty organization and repository defaults", profiles.Defaults)
	}
	if profiles.Active != (ActiveProfileContext{Profile: "default"}) {
		t.Fatalf("active context = %+v; want role-neutral default profile", profiles.Active)
	}
	if len(profiles.Profiles) != 1 {
		t.Fatalf("profiles = %+v; want one default profile", profiles.Profiles)
	}
	profile := profiles.Profiles[0]
	if profile.ID != "default" || profile.Description == "" {
		t.Fatalf("default profile identity = %+v", profile)
	}
	if len(profile.Roles) != 0 || len(profile.Tasks) != 0 {
		t.Fatalf("default profile invents role/task overlays: %+v", profile)
	}
	if strings.Join(profile.Packs, ",") != strings.Join(defaultProfilePackIDs, ",") {
		t.Fatalf("default profile packs = %v; want %v", profile.Packs, defaultProfilePackIDs)
	}

	selected := make(map[string]bool, len(defaultProfilePackIDs))
	for _, id := range defaultProfilePackIDs {
		selected[id] = true
	}
	if err := ValidateProfileReferences(profiles, selected); err != nil {
		t.Fatalf("ValidateProfileReferences: %v", err)
	}
	delete(selected, defaultProfilePackIDs[len(defaultProfilePackIDs)-1])
	if err := ValidateProfileReferences(profiles, selected); err == nil {
		t.Fatal("expected an unselected example reference to fail validation")
	}
}

func TestDefaultProfileExampleMatchesLegacyEffectiveSet(t *testing.T) {
	managed, err := embedded.Assets.List()
	if err != nil {
		t.Fatal(err)
	}

	legacyDir := t.TempDir()
	writeLegacyRuntime(t, legacyDir, managed)
	profileDir := t.TempDir()
	writeLegacyRuntime(t, profileDir, managed)
	installDefaultProfileExample(t, profileDir)

	cmd := New(embedded.Assets)
	legacy := discoverEffectiveSet(t, cmd, legacyDir)
	profile := discoverEffectiveSet(t, cmd, profileDir)
	legacyIDs := effectivePackIDs(legacy)
	profileIDs := effectivePackIDs(profile)

	if strings.Join(legacyIDs, ",") != strings.Join(defaultProfilePackIDs, ",") {
		t.Fatalf("legacy effective packs = %v; want complete baseline %v", legacyIDs, defaultProfilePackIDs)
	}
	if strings.Join(profileIDs, ",") != strings.Join(legacyIDs, ",") {
		t.Fatalf("default profile effective packs = %v; want legacy baseline %v", profileIDs, legacyIDs)
	}
	if len(profile.Conflicts) != 0 {
		t.Fatalf("default profile conflicts = %+v", profile.Conflicts)
	}
	for _, p := range profile.Packs {
		if p.Priority != defaultPriority || p.Mode != defaultMode || len(p.OverriddenBy) != 0 {
			t.Fatalf("default profile changed ordinary composition for %s: %+v", p.ID, p)
		}
	}
}

// Contract assertions 2 and 5: discovery marks profile seeds active, and
// effective composition uses those seeds with their profile provenance.
func TestProfileDrivenActiveAndEffectiveSet(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())
	selected := []string{
		"core/baseline",
		"core/identity",
		"core/pr-contract",
		"core/security",
		"languages/go",
		"cloud/azure",
	}
	if err := WriteSelection(dir, selected); err != nil {
		t.Fatal(err)
	}
	if err := WriteProfiles(dir, testProfiles()); err != nil {
		t.Fatal(err)
	}

	packs, err := cmd.discover(dir)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	active := map[string]bool{}
	for _, p := range packs {
		if p.State.Active {
			active[p.ID] = true
		}
	}
	if len(active) != 5 || active["cloud/azure"] {
		t.Fatalf("profile-driven active set = %v", active)
	}

	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	foundProfileReason := false
	for _, p := range set.Packs {
		if p.ID == "languages/go" && len(p.Reasons) == 1 && p.Reasons[0] == "profile developer" {
			foundProfileReason = true
		}
		if p.ID == "cloud/azure" {
			t.Fatal("selected but inactive pack must not be effective")
		}
	}
	if !foundProfileReason {
		t.Fatalf("effective set lacks profile provenance: %+v", set.Packs)
	}
}

// Contract assertion 4: activate and clear update only the active profile
// context and provide schema-versioned JSON output.
func TestProfileActivateAndClearCommands(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())
	if err := WriteSelection(dir, []string{
		"core/baseline",
		"core/identity",
		"core/pr-contract",
		"core/security",
		"languages/go",
	}); err != nil {
		t.Fatal(err)
	}
	profiles := testProfiles()
	profiles.Active = ActiveProfileContext{}
	if err := WriteProfiles(dir, profiles); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		if err := cmd.RunProfileActivateInDir(
			dir, "developer", "reviewer", "security-review", true,
		); err != nil {
			t.Fatalf("RunProfileActivateInDir: %v", err)
		}
	})
	var payload profileActivationPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("activation output is not JSON: %v\n%s", err, out)
	}
	if payload.SchemaVersion != metadataSchemaVersion ||
		payload.Active.Profile != "developer" ||
		payload.Active.Role != "reviewer" ||
		payload.Active.Task != "security-review" {
		t.Fatalf("activation payload = %+v", payload)
	}

	_ = captureStdout(t, func() {
		if err := cmd.RunProfileClearInDir(dir, false); err != nil {
			t.Fatalf("RunProfileClearInDir: %v", err)
		}
	})
	read, err := ReadProfiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if read.Active != (ActiveProfileContext{}) {
		t.Fatalf("active context was not cleared: %+v", read.Active)
	}
}

// Contract assertion 3: selection cannot be made contradictory with profile
// references.
func TestUnselectRejectsProfileReference(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())
	if err := WriteSelection(dir, []string{"core/baseline", "core/security"}); err != nil {
		t.Fatal(err)
	}
	profiles := &Profiles{
		SchemaVersion: metadataSchemaVersion,
		Defaults: ProfileDefaults{
			Organization: []string{},
			Repository:   []string{"core/security"},
		},
		Profiles: []PolicyProfile{},
	}
	if err := WriteProfiles(dir, profiles); err != nil {
		t.Fatal(err)
	}

	if err := cmd.RunUnselectInDir(dir, []string{"core/security"}, false); err == nil {
		t.Fatal("expected referenced-pack rejection")
	}
	selection, err := ReadSelection(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(selection.Selected) != 2 {
		t.Fatalf("selection changed after rejected unselect: %+v", selection)
	}
}

func writeProfileFile(t *testing.T, rootDir, content string) {
	t.Helper()
	p := filepath.Join(rootDir, filepath.FromSlash(ProfileFileName))
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func installDefaultProfileExample(t *testing.T, rootDir string) *Profiles {
	t.Helper()
	data, err := embedded.Assets.Open(".github/carl/profiles.example.json")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, rootDir, ProfileFileName, data)
	profiles, err := ReadProfiles(rootDir)
	if err != nil {
		t.Fatalf("ReadProfiles(default example): %v", err)
	}
	return profiles
}

func writeLegacyRuntime(t *testing.T, rootDir string, managed []string) {
	t.Helper()
	err := manifest.Write(rootDir, &manifest.Runtime{
		RuntimeVersion:   "test",
		Source:           "test",
		SourceTag:        "test",
		SourceCommit:     "test",
		InstalledAt:      time.Unix(0, 0).UTC(),
		ManagedArtifacts: managed,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func discoverEffectiveSet(t *testing.T, cmd *Command, rootDir string) *EffectiveSet {
	t.Helper()
	packs, err := cmd.discover(rootDir)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	return set
}

func effectivePackIDs(set *EffectiveSet) []string {
	ids := make([]string, 0, len(set.Packs))
	for _, p := range set.Packs {
		ids = append(ids, p.ID)
	}
	return ids
}
