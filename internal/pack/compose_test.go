package pack

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/cmdutil"
)

func intPtr(n int) *int { return &n }

func composePack(id, version string, selected bool, deps []string, prec *Precedence) PackMetadata {
	var pd []PackDependency
	for _, d := range deps {
		pd = append(pd, PackDependency{ID: d, Required: true})
	}
	return PackMetadata{
		SchemaVersion:  metadataSchemaVersion,
		ID:             id,
		Version:        version,
		Category:       categoryFromID(id),
		Source:         "bundled",
		State:          PackState{Bundled: true, Selected: selected, Active: selected},
		OwnedArtifacts: []string{expectedPackPath(id)},
		Dependencies:   pd,
		Precedence:     prec,
	}
}

// Contract assertion 2: effective set = explicit selection + transitive
// required dependencies, each entry carrying an explicit reason.
func TestEffectiveSetDependencyExpansionWithReasons(t *testing.T) {
	packs := []PackMetadata{
		composePack("core/baseline", "1.0.0", false, nil, nil),
		composePack("core/security", "1.0.0", false, []string{"core/baseline"}, nil),
		composePack("languages/go", "1.0.0", true, []string{"core/security"}, nil),
		composePack("cloud/azure", "1.0.0", false, nil, nil),
	}

	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	if len(set.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts: %+v", set.Conflicts)
	}
	got := map[string][]string{}
	for _, p := range set.Packs {
		got[p.ID] = p.Reasons
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 effective packs, got %v", got)
	}
	if _, ok := got["cloud/azure"]; ok {
		t.Fatal("unselected pack without dependents must not be in effective set")
	}
	if r := got["languages/go"]; len(r) != 1 || r[0] != "selected" {
		t.Fatalf("languages/go reasons = %v", r)
	}
	if r := got["core/security"]; len(r) != 1 || r[0] != "dependency of languages/go" {
		t.Fatalf("core/security reasons = %v", r)
	}
	if r := got["core/baseline"]; len(r) != 1 || r[0] != "dependency of core/security" {
		t.Fatalf("core/baseline reasons = %v", r)
	}
}

// Contract assertion 3: precedence order is priority descending, ties broken
// by pack ID — never load order.
func TestEffectiveSetPrecedenceOrder(t *testing.T) {
	packs := []PackMetadata{
		composePack("core/zeta", "1.0.0", true, nil, nil),
		composePack("core/alpha", "1.0.0", true, nil, nil),
		composePack("core/low", "1.0.0", true, nil, &Precedence{Priority: intPtr(1)}),
		composePack("core/high", "1.0.0", true, nil, &Precedence{Priority: intPtr(10)}),
	}

	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	var order []string
	for _, p := range set.Packs {
		order = append(order, p.ID)
	}
	want := []string{"core/high", "core/low", "core/alpha", "core/zeta"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Fatalf("precedence order = %v; want %v", order, want)
	}
}

// Contract assertion 4: overrides only apply to explicitly overridable
// targets; overridden packs stay in the set flagged with overriddenBy.
func TestEffectiveSetExplicitOverride(t *testing.T) {
	packs := []PackMetadata{
		composePack("core/base", "1.0.0", true, nil, &Precedence{Mode: "overridable"}),
		composePack("core/custom", "1.0.0", true, nil, &Precedence{Overrides: []string{"core/base"}}),
	}

	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	if len(set.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts: %+v", set.Conflicts)
	}
	var base *EffectivePack
	for i := range set.Packs {
		if set.Packs[i].ID == "core/base" {
			base = &set.Packs[i]
		}
	}
	if base == nil {
		t.Fatal("overridden pack must remain in the effective set")
	}
	if len(base.OverriddenBy) != 1 || base.OverriddenBy[0] != "core/custom" {
		t.Fatalf("overriddenBy = %v", base.OverriddenBy)
	}
}

// Contract assertion 4: overriding a non-overridable pack is a conflict.
func TestEffectiveSetOverrideNotPermittedConflict(t *testing.T) {
	for _, mode := range []string{"", "additive", "restrictable-only", "immutable"} {
		var prec *Precedence
		if mode != "" {
			prec = &Precedence{Mode: mode}
		}
		packs := []PackMetadata{
			composePack("core/base", "1.0.0", true, nil, prec),
			composePack("core/custom", "1.0.0", true, nil, &Precedence{Overrides: []string{"core/base"}}),
		}
		set, err := ComputeEffectiveSet(packs)
		if err != nil {
			t.Fatalf("ComputeEffectiveSet: %v", err)
		}
		if len(set.Conflicts) != 1 || set.Conflicts[0].Code != "override_not_permitted" {
			t.Fatalf("mode %q: conflicts = %+v", mode, set.Conflicts)
		}
	}
}

// Contract assertion 4: mutual overrides are a conflict.
func TestEffectiveSetMutualOverrideConflict(t *testing.T) {
	packs := []PackMetadata{
		composePack("core/one", "1.0.0", true, nil, &Precedence{Mode: "overridable", Overrides: []string{"core/two"}}),
		composePack("core/two", "1.0.0", true, nil, &Precedence{Mode: "overridable", Overrides: []string{"core/one"}}),
	}
	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	found := false
	for _, c := range set.Conflicts {
		if c.Code == "mutual_override" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mutual_override conflict, got %+v", set.Conflicts)
	}
}

// Contract assertion 5: absent composition metadata yields safe defaults.
func TestEffectiveSetDefaultsWithoutMetadata(t *testing.T) {
	packs := []PackMetadata{composePack("core/plain", "1.0.0", true, nil, nil)}
	set, err := ComputeEffectiveSet(packs)
	if err != nil {
		t.Fatalf("ComputeEffectiveSet: %v", err)
	}
	if len(set.Packs) != 1 {
		t.Fatalf("packs = %+v", set.Packs)
	}
	p := set.Packs[0]
	if p.Priority != 0 || p.Mode != "additive" || len(p.OverriddenBy) != 0 {
		t.Fatalf("unexpected defaults: %+v", p)
	}
}

// Contract assertion 5: composition headers are parsed from pack files and
// malformed headers are explicit errors.
func TestComposeHeaderParsing(t *testing.T) {
	data := []byte(`<!-- version: 1.0.0 -->
<!-- requires: core/baseline, core/security -->
<!-- precedence-mode: overridable -->
<!-- priority: 20 -->
<!-- overrides: cloud/azure -->
# Sample Pack

Sample description.
`)
	m, err := parsePackFileMetadata(".github/instructions/core/sample.instructions.md", data)
	if err != nil {
		t.Fatalf("parsePackFileMetadata: %v", err)
	}
	if len(m.Requires) != 2 || m.Requires[0] != "core/baseline" || m.Requires[1] != "core/security" {
		t.Fatalf("requires = %v", m.Requires)
	}
	if m.Mode != "overridable" {
		t.Fatalf("mode = %q", m.Mode)
	}
	if m.Priority == nil || *m.Priority != 20 {
		t.Fatalf("priority = %v", m.Priority)
	}
	if len(m.Overrides) != 1 || m.Overrides[0] != "cloud/azure" {
		t.Fatalf("overrides = %v", m.Overrides)
	}

	for name, bad := range map[string]string{
		"bad requires id": "<!-- requires: Not A Pack -->\n# X\n",
		"bad mode":        "<!-- precedence-mode: destructive -->\n# X\n",
		"bad priority":    "<!-- priority: high -->\n# X\n",
		"negative":        "<!-- priority: -3 -->\n# X\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parsePackFileMetadata(".github/instructions/core/sample.instructions.md", []byte(bad)); err == nil {
				t.Fatal("expected explicit error for malformed header")
			}
		})
	}
}

// Contract assertions 2-4 end to end: `carl pack effective --json` produces
// valid schema-versioned JSON; conflicts exit non-zero.
func TestPackEffectiveCommandJSON(t *testing.T) {
	dir := t.TempDir()
	arts := bundledArtifacts()
	arts.files[".github/instructions/languages/go.instructions.md"] = []byte(
		"<!-- version: 1.0.0 -->\n<!-- requires: core/baseline -->\n# Go\nGo guidance.\n")
	cmd := New(arts)

	_ = captureStdout(t, func() {
		if err := cmd.RunSelectInDir(dir, []string{"languages/go"}, false); err != nil {
			t.Fatalf("RunSelectInDir: %v", err)
		}
	})

	out := captureStdout(t, func() {
		if err := cmd.RunEffectiveInDir(dir, true); err != nil {
			t.Fatalf("RunEffectiveInDir: %v", err)
		}
	})
	var payload effectivePayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("effective output is not valid JSON: %v\n%s", err, out)
	}
	if payload.SchemaVersion != metadataSchemaVersion {
		t.Fatalf("schema version = %d", payload.SchemaVersion)
	}
	if len(payload.Packs) != 2 {
		t.Fatalf("expected 2 effective packs, got %+v", payload.Packs)
	}
	ids := map[string]bool{}
	for _, p := range payload.Packs {
		ids[p.ID] = true
	}
	if !ids["languages/go"] || !ids["core/baseline"] {
		t.Fatalf("unexpected effective set: %+v", payload.Packs)
	}
}

func TestPackEffectiveConflictExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	arts := bundledArtifacts()
	arts.files[".github/instructions/core/custom.instructions.md"] = []byte(
		"<!-- version: 1.0.0 -->\n<!-- overrides: core/baseline -->\n# Custom\nCustom guidance.\n")
	cmd := New(arts)

	_ = captureStdout(t, func() {
		if err := cmd.RunSelectInDir(dir, []string{"core/custom", "core/baseline"}, false); err != nil {
			t.Fatalf("RunSelectInDir: %v", err)
		}
	})

	var runErr error
	out := captureStdout(t, func() {
		runErr = cmd.RunEffectiveInDir(dir, true)
	})
	if runErr == nil {
		t.Fatal("expected non-zero exit for composition conflict")
	}
	exitErr := &cmdutil.ExitError{}
	if !errors.As(runErr, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError code 1, got %v", runErr)
	}
	var payload effectivePayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("conflict output is not valid JSON: %v\n%s", err, out)
	}
	if len(payload.Conflicts) == 0 || payload.Conflicts[0].Code != "override_not_permitted" {
		t.Fatalf("conflicts = %+v", payload.Conflicts)
	}
	// Conservative composition: the overridden pack remains listed.
	found := false
	for _, p := range payload.Packs {
		if p.ID == "core/baseline" {
			found = true
		}
	}
	if !found {
		t.Fatal("overridden pack must remain in the effective set")
	}
}
