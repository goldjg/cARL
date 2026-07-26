package pack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/cmdutil"
)

type rejectingPolicyFetcher struct {
	called bool
}

func (f *rejectingPolicyFetcher) Fetch(
	_ context.Context,
	location string,
	_ int64,
) ([]byte, error) {
	f.called = true
	return nil, fmt.Errorf("unexpected registry fetch: %s", location)
}

func policyArtifacts(files map[string]string) *fakeArtifacts {
	artifacts := &fakeArtifacts{files: map[string][]byte{}}
	for path, content := range files {
		artifacts.files[path] = []byte(content)
	}
	return artifacts
}

func writePolicyProfiles(
	t *testing.T,
	rootDir string,
	selected []string,
	profiles *Profiles,
) {
	t.Helper()
	if err := WriteSelection(rootDir, selected); err != nil {
		t.Fatalf("WriteSelection: %v", err)
	}
	if profiles != nil {
		if err := WriteProfiles(rootDir, profiles); err != nil {
			t.Fatalf("WriteProfiles: %v", err)
		}
	}
}

// Contract assertion 1: explain reports structured profile and dependency
// provenance plus the canonical definition for an effective pack.
func TestExplainProfileDependencyProvenanceJSON(t *testing.T) {
	dir := t.TempDir()
	artifacts := policyArtifacts(map[string]string{
		".github/instructions/core/baseline.instructions.md": "<!-- version: 1.2.0 -->\n# Baseline\nBaseline guidance.\n",
		".github/instructions/languages/go.instructions.md":  "<!-- version: 1.0.0 -->\n<!-- requires: core/baseline -->\n# Go\nGo guidance.\n",
		".github/instructions/core/carl.instructions.md":     "<!-- version: 1.3.0 -->\n# cARL\nCognition guidance.\n",
	})
	writePolicyProfiles(t, dir, []string{"languages/go"}, &Profiles{
		SchemaVersion: metadataSchemaVersion,
		Defaults: ProfileDefaults{
			Organization: []string{},
			Repository:   []string{},
		},
		Profiles: []PolicyProfile{{
			ID:    "developer",
			Packs: []string{"languages/go"},
			Roles: map[string][]string{},
			Tasks: map[string][]string{},
		}},
		Active: ActiveProfileContext{Profile: "developer"},
	})

	cmd := NewExplain(artifacts)
	out := captureStdout(t, func() {
		if err := cmd.RunExplainInDir(dir, "core/baseline", true); err != nil {
			t.Fatalf("RunExplainInDir: %v", err)
		}
	})

	var payload explainPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("explain output is not valid JSON: %v\n%s", err, out)
	}
	policy := payload.Policy
	if payload.SchemaVersion != metadataSchemaVersion {
		t.Fatalf("schema version = %d", payload.SchemaVersion)
	}
	if payload.Context.Mode != "profiles" ||
		payload.Context.Profile != "developer" ||
		payload.Context.Source != ProfileFileName {
		t.Fatalf("context = %#v", payload.Context)
	}
	if !policy.Applied || policy.Status != "effective" {
		t.Fatalf("policy application = %#v", policy)
	}
	if policy.CanonicalDefinition != ".github/instructions/core/baseline.instructions.md" ||
		policy.Source != "bundled" {
		t.Fatalf(
			"definition/source = %q/%q",
			policy.CanonicalDefinition,
			policy.Source,
		)
	}
	if len(policy.Activation) != 1 ||
		policy.Activation[0].Kind != "dependency" ||
		policy.Activation[0].RelatedPack != "languages/go" ||
		policy.Activation[0].Source != ".github/instructions/languages/go.instructions.md" {
		t.Fatalf("activation = %#v", policy.Activation)
	}
	if !reflect.DeepEqual(policy.RequiredBy, []string{"languages/go"}) {
		t.Fatalf("requiredBy = %#v", policy.RequiredBy)
	}
	if policy.Order == nil || *policy.Order != 1 {
		t.Fatalf("order = %#v", policy.Order)
	}

	goOut := captureStdout(t, func() {
		if err := cmd.RunExplainInDir(dir, "languages/go", true); err != nil {
			t.Fatalf("RunExplainInDir languages/go: %v", err)
		}
	})
	var goPayload explainPayload
	if err := json.Unmarshal([]byte(goOut), &goPayload); err != nil {
		t.Fatalf("Go explanation is not valid JSON: %v", err)
	}
	if len(goPayload.Policy.Activation) != 1 ||
		goPayload.Policy.Activation[0].Kind != "profile" ||
		goPayload.Policy.Activation[0].Profile != "developer" ||
		goPayload.Policy.Activation[0].Source != ProfileFileName {
		t.Fatalf("profile activation = %#v", goPayload.Policy.Activation)
	}
}

// Contract assertion 1: a discoverable inactive pack remains explainable and
// is never represented as applied.
func TestExplainInactivePack(t *testing.T) {
	dir := t.TempDir()
	artifacts := policyArtifacts(map[string]string{
		".github/instructions/core/baseline.instructions.md": "<!-- version: 1.2.0 -->\n# Baseline\nBaseline guidance.\n",
		".github/instructions/core/carl.instructions.md":     "<!-- version: 1.3.0 -->\n# cARL\nCognition guidance.\n",
	})
	writePolicyProfiles(t, dir, []string{"core/baseline"}, nil)

	cmd := NewExplain(artifacts)
	out := captureStdout(t, func() {
		if err := cmd.RunExplainInDir(dir, "core/carl", true); err != nil {
			t.Fatalf("RunExplainInDir: %v", err)
		}
	})

	var payload explainPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal explain output: %v", err)
	}
	if payload.Policy.Applied ||
		payload.Policy.Status != "inactive" ||
		payload.Policy.Order != nil ||
		len(payload.Policy.Activation) != 0 ||
		payload.Policy.Effect.AddsConstraints {
		t.Fatalf("inactive policy = %#v", payload.Policy)
	}
}

func TestTraceWithoutSelectionHasNoSyntheticSource(t *testing.T) {
	dir := t.TempDir()
	cmd := NewTrace(policyArtifacts(map[string]string{
		".github/instructions/core/baseline.instructions.md": "<!-- version: 1.2.0 -->\n# Baseline\nBaseline guidance.\n",
	}))

	out := captureStdout(t, func() {
		if err := cmd.RunTraceInDir(dir, true); err != nil {
			t.Fatalf("RunTraceInDir: %v", err)
		}
	})
	var trace PolicyTrace
	if err := json.Unmarshal([]byte(out), &trace); err != nil {
		t.Fatalf("trace output is not valid JSON: %v", err)
	}
	if trace.Context.Mode != "none" ||
		trace.Context.Source != "none" ||
		len(trace.Policies) != 0 {
		t.Fatalf("empty context trace = %#v", trace)
	}
}

// Contract assertion 2: trace preserves effective precedence order and
// explains an explicitly permitted override as a resolved decision.
func TestTraceResolvedOverrideJSON(t *testing.T) {
	dir := t.TempDir()
	artifacts := policyArtifacts(map[string]string{
		".github/instructions/core/base.instructions.md":   "<!-- version: 1.0.0 -->\n<!-- precedence-mode: overridable -->\n# Base\nBase guidance.\n",
		".github/instructions/core/custom.instructions.md": "<!-- version: 1.0.0 -->\n<!-- overrides: core/base -->\n# Custom\nCustom guidance.\n",
	})
	writePolicyProfiles(t, dir, []string{"core/base", "core/custom"}, nil)

	cmd := NewTrace(artifacts)
	out := captureStdout(t, func() {
		if err := cmd.RunTraceInDir(dir, true); err != nil {
			t.Fatalf("RunTraceInDir: %v", err)
		}
	})

	var trace PolicyTrace
	if err := json.Unmarshal([]byte(out), &trace); err != nil {
		t.Fatalf("trace output is not valid JSON: %v\n%s", err, out)
	}
	if len(trace.Policies) != 2 ||
		trace.Policies[0].ID != "core/base" ||
		trace.Policies[1].ID != "core/custom" {
		t.Fatalf("policy order = %#v", trace.Policies)
	}
	if !strings.Contains(trace.Notice, "does not prove model adherence") {
		t.Fatalf("trace evidence boundary missing from notice: %q", trace.Notice)
	}
	if trace.Policies[0].Status != "overridden" ||
		trace.Policies[0].Applied ||
		trace.Policies[0].Effect.AddsConstraints ||
		!reflect.DeepEqual(trace.Policies[0].Effect.OverriddenBy, []string{"core/custom"}) {
		t.Fatalf("base policy = %#v", trace.Policies[0])
	}
	if !reflect.DeepEqual(
		trace.Policies[1].Effect.ResolvedOverrides,
		[]string{"core/base"},
	) {
		t.Fatalf("custom policy = %#v", trace.Policies[1])
	}
	foundOverride := false
	foundNotApplied := false
	for _, decision := range trace.Decisions {
		if decision.Kind == "override" &&
			decision.Outcome == "resolved" &&
			decision.Subject == "core/custom" &&
			decision.Target == "core/base" &&
			strings.Contains(decision.Reason, "overridable") {
			foundOverride = true
		}
		if decision.Kind == "constraint" &&
			decision.Outcome == "not-applied" &&
			decision.Subject == "core/base" {
			foundNotApplied = true
		}
	}
	if !foundOverride {
		t.Fatalf("resolved override decision missing: %#v", trace.Decisions)
	}
	if !foundNotApplied {
		t.Fatalf("overridden definition must be explicitly not applied: %#v", trace.Decisions)
	}
}

// Contract assertion 2: unresolved conflicts remain explicit and cause a
// non-zero exit after deterministic JSON is emitted.
func TestTraceUnresolvedConflictExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	artifacts := policyArtifacts(map[string]string{
		".github/instructions/core/base.instructions.md":   "<!-- version: 1.0.0 -->\n# Base\nBase guidance.\n",
		".github/instructions/core/custom.instructions.md": "<!-- version: 1.0.0 -->\n<!-- overrides: core/base -->\n# Custom\nCustom guidance.\n",
	})
	writePolicyProfiles(t, dir, []string{"core/base", "core/custom"}, nil)

	cmd := NewTrace(artifacts)
	var runErr error
	out := captureStdout(t, func() {
		runErr = cmd.RunTraceInDir(dir, true)
	})
	if runErr == nil {
		t.Fatal("expected non-zero exit for unresolved conflict")
	}
	var exitErr *cmdutil.ExitError
	if !errors.As(runErr, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError code 1, got %v", runErr)
	}

	var trace PolicyTrace
	if err := json.Unmarshal([]byte(out), &trace); err != nil {
		t.Fatalf("trace output is not valid JSON: %v\n%s", err, out)
	}
	if len(trace.Conflicts) != 1 ||
		trace.Conflicts[0].Code != "override_not_permitted" {
		t.Fatalf("conflicts = %#v", trace.Conflicts)
	}
	found := false
	for _, decision := range trace.Decisions {
		if decision.Kind == "conflict" &&
			decision.Outcome == "unresolved" &&
			decision.Code == "override_not_permitted" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unresolved conflict decision missing: %#v", trace.Decisions)
	}

	humanOut := captureStdout(t, func() {
		runErr = cmd.RunTraceInDir(dir, false)
	})
	if runErr == nil {
		t.Fatal("expected non-zero human exit for unresolved conflict")
	}
	if !errors.As(runErr, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected human ExitError code 1, got %v", runErr)
	}
	if !strings.Contains(humanOut, "Conflicts:") ||
		!strings.Contains(humanOut, "override_not_permitted") {
		t.Fatalf("human conflict output missing details:\n%s", humanOut)
	}
}

// Contract assertions 3 and 4: explain and trace neither fetch registries nor
// write repository files, and their output states the reasoning boundary.
func TestPolicyCommandsAreReadOnlyOfflineAndBounded(t *testing.T) {
	dir := t.TempDir()
	artifacts := policyArtifacts(map[string]string{
		".github/instructions/core/baseline.instructions.md": "<!-- version: 1.2.0 -->\n# Baseline\nBaseline guidance.\n",
	})
	writePolicyProfiles(t, dir, []string{"core/baseline"}, nil)
	writeFile(t, dir, RegistryFileName, []byte(
		"{\n  \"schemaVersion\": 1,\n  \"registries\": [{\"id\": \"remote\", \"location\": \"https://example.invalid/index.json\"}]\n}\n",
	))
	before := snapshotPolicyTree(t, dir)

	explain := NewExplain(artifacts)
	explainFetcher := &rejectingPolicyFetcher{}
	explain.pack.fetcher = explainFetcher
	explainOut := captureStdout(t, func() {
		if err := explain.RunExplainInDir(dir, "core/baseline", true); err != nil {
			t.Fatalf("RunExplainInDir: %v", err)
		}
	})

	trace := NewTrace(artifacts)
	traceFetcher := &rejectingPolicyFetcher{}
	trace.explain.pack.fetcher = traceFetcher
	traceOut := captureStdout(t, func() {
		if err := trace.RunTraceInDir(dir, true); err != nil {
			t.Fatalf("RunTraceInDir: %v", err)
		}
	})
	traceOutAgain := captureStdout(t, func() {
		if err := trace.RunTraceInDir(dir, true); err != nil {
			t.Fatalf("second RunTraceInDir: %v", err)
		}
	})

	after := snapshotPolicyTree(t, dir)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("policy commands changed repository files\nbefore=%#v\nafter=%#v", before, after)
	}
	if explainFetcher.called || traceFetcher.called {
		t.Fatal("policy commands must not fetch configured registries")
	}
	if traceOut != traceOutAgain {
		t.Fatalf("trace output is not deterministic\nfirst:\n%s\nsecond:\n%s", traceOut, traceOutAgain)
	}
	for name, out := range map[string]string{
		"explain": explainOut,
		"trace":   traceOut,
	} {
		if !strings.Contains(out, "does not interpret individual natural-language rules") ||
			!strings.Contains(out, "hidden model reasoning") ||
			!strings.Contains(out, "chain-of-thought") {
			t.Fatalf("%s output lacks epistemic boundary:\n%s", name, out)
		}
		if strings.Contains(out, `"chainOfThought":`) ||
			strings.Contains(out, `"rules":`) {
			t.Fatalf("%s output exposes a forbidden field:\n%s", name, out)
		}
	}
}

func TestExplainUnknownPackJSONError(t *testing.T) {
	cmd := NewExplain(policyArtifacts(map[string]string{
		".github/instructions/core/baseline.instructions.md": "<!-- version: 1.2.0 -->\n# Baseline\nBaseline guidance.\n",
	}))
	err := cmd.RunExplainInDir(t.TempDir(), "core/missing", true)
	if err == nil {
		t.Fatal("expected unknown-pack error")
	}
	var exitErr *cmdutil.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError code 1, got %v", err)
	}
	var payload errorPayload
	if unmarshalErr := json.Unmarshal([]byte(exitErr.Message), &payload); unmarshalErr != nil {
		t.Fatalf("error payload is not valid JSON: %v", unmarshalErr)
	}
	if payload.Error.Code != "pack_not_found" {
		t.Fatalf("error code = %q", payload.Error.Code)
	}
}

func snapshotPolicyTree(t *testing.T, rootDir string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	err := filepath.WalkDir(rootDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		snapshot[filepath.ToSlash(relative)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot repository: %v", err)
	}
	return snapshot
}
