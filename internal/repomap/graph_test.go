package repomap_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/goldjg/carl/internal/repomap"
)

func TestGraph_DeterministicStableAndRepositoryRelative(t *testing.T) {
	dir := graphFixture(t)

	first, err := repomap.Build(dir)
	if err != nil {
		t.Fatalf("first Build: %v", err)
	}
	second, err := repomap.Build(dir)
	if err != nil {
		t.Fatalf("second Build: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatal("Build must produce deterministic graph data for unchanged repository state")
	}
	if first.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", first.SchemaVersion)
	}

	seen := map[string]bool{}
	previous := ""
	for _, node := range first.Graph.Nodes {
		if seen[node.ID] {
			t.Errorf("duplicate node ID %q", node.ID)
		}
		seen[node.ID] = true
		if previous != "" && node.ID < previous {
			t.Errorf("nodes not sorted: %q before %q", previous, node.ID)
		}
		previous = node.ID
		if filepath.IsAbs(node.Path) {
			t.Errorf("node path must be repository-relative: %q", node.Path)
		}
		if node.AgentContext == "" {
			t.Errorf("node %q has empty agent context", node.ID)
		}
	}

	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), filepath.ToSlash(dir)) {
		t.Errorf("graph output exposes repository root %q", dir)
	}
}

func TestGraph_GoDependenciesAndDirectChangeImpact(t *testing.T) {
	dir := graphFixture(t)
	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	edge := findGraphEdge(t, m, "depends_on", "component:cmd/app", "component:internal/lib")
	if !slices.Equal(edge.Evidence, []string{"cmd/app/main.go"}) {
		t.Errorf("dependency evidence = %v, want cmd/app/main.go", edge.Evidence)
	}

	lib := findGraphNode(t, m, "component:internal/lib")
	if !slices.Contains(lib.ChangeImpact, "component:cmd/app") {
		t.Errorf("lib change impact = %v, want direct dependant component:cmd/app", lib.ChangeImpact)
	}
	app := findGraphNode(t, m, "component:cmd/app")
	if len(app.ChangeImpact) != 0 {
		t.Errorf("app change impact = %v, want no reverse dependants", app.ChangeImpact)
	}
}

func TestGraph_ClassifiesContextTrustCriticalityAndPolicyPoints(t *testing.T) {
	dir := graphFixture(t)
	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	root := findGraphNode(t, m, "repository:.")
	if root.Criticality != "high" || root.TrustBoundary != "repository" || !root.PolicyAttachmentPoint {
		t.Errorf("unexpected root classification: %+v", root)
	}

	pkg := findGraphNode(t, m, "component:internal/lib")
	if pkg.Kind != "package" || pkg.Criticality != "medium" || !pkg.PolicyAttachmentPoint {
		t.Errorf("unexpected package classification: %+v", pkg)
	}

	policy := findGraphNode(t, m, "policy:core/security")
	if policy.TrustBoundary != "policy" || policy.Criticality != "high" {
		t.Errorf("unexpected policy classification: %+v", policy)
	}
	if !strings.Contains(policy.AgentContext, "carl trace") {
		t.Errorf("policy context must preserve activation boundary: %q", policy.AgentContext)
	}

	governance := findGraphNode(t, m, "governance:.github/carl/invariants.yml")
	if governance.TrustBoundary != "governance" || governance.Criticality != "high" {
		t.Errorf("unexpected governance classification: %+v", governance)
	}

	findGraphEdge(t, m, "contains", "component:.github/instructions/core", "policy:core/security")
}

func TestGraph_CoverageMakesEvidenceLimitsExplicit(t *testing.T) {
	dir := graphFixture(t)
	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}

	coverage := m.Graph.Coverage
	if coverage.Ownership.Status != "unavailable" {
		t.Errorf("ownership status = %q, want unavailable", coverage.Ownership.Status)
	}
	if coverage.Dependencies.Status != "partial" {
		t.Errorf("dependencies status = %q, want partial", coverage.Dependencies.Status)
	}
	if coverage.DataFlows.Status != "unavailable" {
		t.Errorf("data flow status = %q, want unavailable", coverage.DataFlows.Status)
	}
	if coverage.TrustBoundaries.Status != "derived" || coverage.Criticality.Status != "derived" {
		t.Errorf("derived classifications missing: %+v", coverage)
	}
	if coverage.PolicyAttachments.Status != "partial" || !strings.Contains(coverage.PolicyAttachments.Detail, "carl trace") {
		t.Errorf("policy coverage must preserve activation boundary: %+v", coverage.PolicyAttachments)
	}
	if coverage.ChangeImpact.Status != "partial" || !strings.Contains(coverage.ChangeImpact.Detail, "direct") {
		t.Errorf("change-impact limitation missing: %+v", coverage.ChangeImpact)
	}
}

func TestGraph_MalformedGoImportFailsWithoutExecution(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "go.mod", "module example.com/broken\n\ngo 1.24\n")
	writeFixtureFile(t, dir, "broken.go", "package broken\nimport \"unterminated\n")

	_, err := repomap.Build(dir)
	if err == nil {
		t.Fatal("Build should reject malformed Go import syntax")
	}
	if !strings.Contains(err.Error(), "parse Go imports") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGraph_DoesNotFollowSymlinkedModuleDefinition(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "go.mod")
	if err := os.WriteFile(outside, []byte("module example.com/outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "go.mod")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	writeFixtureFile(t, dir, "internal/lib/lib.go", "package lib\n")

	m, err := repomap.Build(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range m.EntryPoints {
		if entry.Path == "go.mod" {
			t.Fatal("symlinked go.mod must not be followed or reported as an entry point")
		}
	}
	for _, edge := range m.Graph.Edges {
		if edge.Type == "depends_on" {
			t.Fatalf("no local module dependencies should be derived through symlinked go.mod: %+v", edge)
		}
	}
}

func graphFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFixtureFile(t, dir, "go.mod", "module example.com/graph\n\ngo 1.24\n")
	writeFixtureFile(t, dir, "internal/lib/lib.go", "// Package lib provides graph evidence.\npackage lib\n")
	writeFixtureFile(t, dir, "cmd/app/main.go", "package main\n\nimport \"example.com/graph/internal/lib\"\n\nfunc main() { _ = lib.Value }\n")
	writeFixtureFile(t, dir, ".github/instructions/core/security.instructions.md", "# Security\n")
	writeFixtureFile(t, dir, ".github/carl/invariants.yml", "invariants: []\n")
	return dir
}

func writeFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findGraphNode(t *testing.T, m *repomap.Map, id string) repomap.GraphNode {
	t.Helper()
	for _, node := range m.Graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("graph node %q not found", id)
	return repomap.GraphNode{}
}

func findGraphEdge(t *testing.T, m *repomap.Map, edgeType, from, to string) repomap.GraphEdge {
	t.Helper()
	for _, edge := range m.Graph.Edges {
		if edge.Type == edgeType && edge.From == from && edge.To == to {
			return edge
		}
	}
	t.Fatalf("graph edge %s %s -> %s not found", edgeType, from, to)
	return repomap.GraphEdge{}
}
