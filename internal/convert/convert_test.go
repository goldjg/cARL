package convert

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn and returns
// everything written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

const baseInvariants = `# version: 1.0.0
invariants:
  - id: no-hardcoded-secrets
    name: No hardcoded secrets
    rule: Secrets, tokens, and credentials must not be committed to repository files.
    severity: critical
`

const baseMemory = `<!-- version: 1.2.0 -->
# Durable Architectural Truth Cache

## Project purpose
Example project.

## Last updated
2026-01-01
`

// newRepo creates a temporary repo with the canonical cARL destinations
// populated and optional AADLC artefacts written from files.
func newRepo(t *testing.T, aadlc map[string]string) string {
	t.Helper()
	root := t.TempDir()
	carlDir := filepath.Join(root, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(carlDir, "invariants.yml"), baseInvariants)
	writeFile(t, filepath.Join(carlDir, "memory.md"), baseMemory)
	for rel, content := range aadlc {
		writeFile(t, filepath.Join(root, filepath.FromSlash(rel)), content)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

const sampleAADLC = `# AADLC

## Invariants
- Do not patch production Terraform directly
- Cloud Functions must remain Gen2
- All PRs require security review

## Lessons learned
- The legacy auth module cannot handle concurrent refresh

## Governance
- Every PR requires a populated PR contract
`

func runConvert(t *testing.T, root string, apply bool) string {
	t.Helper()
	var err error
	out := captureStdout(t, func() {
		err = RunInDir(root, aadlcConverter{}, apply)
	})
	if err != nil {
		t.Fatalf("RunInDir: %v", err)
	}
	return out
}

func TestNoArtefacts(t *testing.T) {
	root := newRepo(t, nil)
	out := runConvert(t, root, false)

	if !strings.Contains(out, "0 artefact(s)") {
		t.Errorf("expected zero artefacts reported, got:\n%s", out)
	}
	if !strings.Contains(out, "No AADLC artefacts found") {
		t.Errorf("expected nothing-to-migrate note, got:\n%s", out)
	}
	// Destinations must be untouched.
	if got := readFile(t, filepath.Join(root, ".github", "carl", "invariants.yml")); got != baseInvariants {
		t.Errorf("invariants.yml modified unexpectedly")
	}
	if got := readFile(t, filepath.Join(root, ".github", "carl", "memory.md")); got != baseMemory {
		t.Errorf("memory.md modified unexpectedly")
	}
}

func TestDryRunMakesNoChanges(t *testing.T) {
	root := newRepo(t, map[string]string{"AADLC.md": sampleAADLC})
	out := runConvert(t, root, false)

	if !strings.Contains(out, "Would update:") {
		t.Errorf("expected 'Would update:' heading in dry-run, got:\n%s", out)
	}
	if !strings.Contains(out, "Dry run — no changes written") {
		t.Errorf("expected dry-run note, got:\n%s", out)
	}
	if got := readFile(t, filepath.Join(root, ".github", "carl", "invariants.yml")); got != baseInvariants {
		t.Errorf("dry-run modified invariants.yml")
	}
	if got := readFile(t, filepath.Join(root, ".github", "carl", "memory.md")); got != baseMemory {
		t.Errorf("dry-run modified memory.md")
	}
}

func TestApplyWritesDestinations(t *testing.T) {
	root := newRepo(t, map[string]string{"AADLC.md": sampleAADLC})
	out := runConvert(t, root, true)

	if !strings.Contains(out, "Updated:") || !strings.Contains(out, "Migration applied.") {
		t.Errorf("expected applied report, got:\n%s", out)
	}

	inv := readFile(t, filepath.Join(root, ".github", "carl", "invariants.yml"))
	for _, want := range []string{
		"aadlc-do-not-patch-production-terraform-directly",
		"aadlc-cloud-functions-must-remain-gen2",
		"aadlc-all-prs-require-security-review",
	} {
		if !strings.Contains(inv, want) {
			t.Errorf("invariants.yml missing %q:\n%s", want, inv)
		}
	}

	mem := readFile(t, filepath.Join(root, ".github", "carl", "memory.md"))
	if !strings.Contains(mem, memBeginMarker) || !strings.Contains(mem, memEndMarker) {
		t.Errorf("memory.md missing managed convert block:\n%s", mem)
	}
	if !strings.Contains(mem, "The legacy auth module cannot handle concurrent refresh") {
		t.Errorf("memory.md missing migrated durable entry:\n%s", mem)
	}
	if !strings.Contains(mem, "Every PR requires a populated PR contract") {
		t.Errorf("memory.md missing migrated governance entry:\n%s", mem)
	}
	// Pre-existing human content must be preserved.
	if !strings.Contains(mem, "## Project purpose") {
		t.Errorf("memory.md lost human-authored content:\n%s", mem)
	}
}

func TestDuplicateDetection(t *testing.T) {
	// An invariant that already exists verbatim in invariants.yml must be
	// reported as a duplicate and skipped (not re-added).
	dupAADLC := `# AADLC

## Invariants
- Secrets, tokens, and credentials must not be committed to repository files
`
	root := newRepo(t, map[string]string{"AADLC.md": dupAADLC})
	out := runConvert(t, root, true)

	if !strings.Contains(out, "1 duplicate(s)") {
		t.Errorf("expected 1 duplicate reported, got:\n%s", out)
	}
	if !strings.Contains(out, "0 invariant(s)") {
		t.Errorf("expected no convertible invariants, got:\n%s", out)
	}
	if got := readFile(t, filepath.Join(root, ".github", "carl", "invariants.yml")); got != baseInvariants {
		t.Errorf("duplicate invariant was written:\n%s", got)
	}
}

func TestConflictDetection(t *testing.T) {
	// An invariant whose generated id collides with an existing invariant but
	// whose content differs must be reported as a conflict, not written.
	conflictInv := baseInvariants +
		"\n  - id: aadlc-cloud-functions-must-remain-gen2\n" +
		"    name: \"Existing\"\n" +
		"    rule: \"Cloud Functions must use Gen1 for legacy compatibility\"\n" +
		"    severity: high\n"
	root := t.TempDir()
	carlDir := filepath.Join(root, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(carlDir, "invariants.yml"), conflictInv)
	writeFile(t, filepath.Join(carlDir, "memory.md"), baseMemory)
	writeFile(t, filepath.Join(root, "AADLC.md"), `# AADLC

## Invariants
- Cloud Functions must remain Gen2
`)

	out := runConvert(t, root, true)
	if !strings.Contains(out, "1 item(s) requiring review") {
		t.Errorf("expected 1 conflict, got:\n%s", out)
	}
	if !strings.Contains(out, "already exists with different content") {
		t.Errorf("expected conflict reason, got:\n%s", out)
	}
	if got := readFile(t, filepath.Join(carlDir, "invariants.yml")); got != conflictInv {
		t.Errorf("conflict was written to invariants.yml:\n%s", got)
	}
}

func TestIdempotentReRun(t *testing.T) {
	root := newRepo(t, map[string]string{"AADLC.md": sampleAADLC})

	runConvert(t, root, true)
	inv1 := readFile(t, filepath.Join(root, ".github", "carl", "invariants.yml"))
	mem1 := readFile(t, filepath.Join(root, ".github", "carl", "memory.md"))

	out := runConvert(t, root, true)
	inv2 := readFile(t, filepath.Join(root, ".github", "carl", "invariants.yml"))
	mem2 := readFile(t, filepath.Join(root, ".github", "carl", "memory.md"))

	if inv1 != inv2 {
		t.Errorf("invariants.yml changed on idempotent re-run:\nfirst:\n%s\nsecond:\n%s", inv1, inv2)
	}
	if mem1 != mem2 {
		t.Errorf("memory.md changed on idempotent re-run:\nfirst:\n%s\nsecond:\n%s", mem1, mem2)
	}
	if !strings.Contains(out, "0 invariant(s)") || !strings.Contains(out, "(none)") {
		t.Errorf("re-run should report nothing convertible, got:\n%s", out)
	}
}

func TestNeverModifiesSourceArtefacts(t *testing.T) {
	root := newRepo(t, map[string]string{"AADLC.md": sampleAADLC})
	runConvert(t, root, true)
	if got := readFile(t, filepath.Join(root, "AADLC.md")); got != sampleAADLC {
		t.Errorf("source AADLC.md was modified:\n%s", got)
	}
}

func TestDiscoveryAcrossRoots(t *testing.T) {
	root := newRepo(t, map[string]string{
		".aadlc/notes.md":        "## Invariants\n- A must hold\n",
		".github/aadlc/rules.md": "## Governance\n- Approval required\n",
		"aadlc/lessons.md":       "## Lessons learned\n- We learned X\n",
		"AADLC.md":               "## Memory\n- Historical context Y\n",
		".aadlc/ignore.txt":      "- not convertible",
		".aadlc/invariants.yaml": "invariants:\n  - id: x\n    name: X\n    rule: Yaml rule holds\n    severity: high\n",
	})
	conv := aadlcConverter{}
	artefacts, err := conv.Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	// .txt must be excluded; the other five included.
	if len(artefacts) != 5 {
		var paths []string
		for _, a := range artefacts {
			paths = append(paths, a.Path)
		}
		t.Fatalf("expected 5 artefacts, got %d: %v", len(artefacts), paths)
	}
	// Deterministic sorted order.
	for i := 1; i < len(artefacts); i++ {
		if artefacts[i-1].Path > artefacts[i].Path {
			t.Errorf("artefacts not sorted: %q before %q", artefacts[i-1].Path, artefacts[i].Path)
		}
	}
}

func TestDeterministicReport(t *testing.T) {
	root := newRepo(t, map[string]string{"AADLC.md": sampleAADLC})
	first := runConvert(t, root, false)
	second := runConvert(t, root, false)
	if first != second {
		t.Errorf("dry-run report not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestUnknownSourceAndFlags(t *testing.T) {
	cmd := New()
	if err := cmd.Run(nil, []string{"bogus"}); err == nil {
		t.Error("expected error for unknown source")
	}
	if _, err := parseMode([]string{"--apply", "--dry-run"}); err == nil {
		t.Error("expected mutual-exclusion error")
	}
	if _, err := parseMode([]string{"--bogus"}); err == nil {
		t.Error("expected unknown-flag error")
	}
	if apply, err := parseMode(nil); err != nil || apply {
		t.Errorf("default mode should be dry-run, got apply=%v err=%v", apply, err)
	}
	if apply, err := parseMode([]string{"--apply"}); err != nil || !apply {
		t.Errorf("--apply should enable apply, got apply=%v err=%v", apply, err)
	}
}

// newRepoWithMemory creates a temporary repo with canonical destinations, but
// uses the supplied memory.md content (instead of baseMemory) and an AADLC
// source that yields convertible items. It returns the repo root and the
// memory.md path.
func newRepoWithMemory(t *testing.T, memory string) (root, memPath string) {
	t.Helper()
	root = t.TempDir()
	carlDir := filepath.Join(root, ".github", "carl")
	if err := os.MkdirAll(carlDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(carlDir, "invariants.yml"), baseInvariants)
	memPath = filepath.Join(carlDir, "memory.md")
	writeFile(t, memPath, memory)
	writeFile(t, filepath.Join(root, "AADLC.md"), sampleAADLC)
	return root, memPath
}

// malformedMarkerCases enumerates the three invalid managed-convert-block
// marker states. Each must: return a non-nil error mentioning "malformed",
// leave memory.md unchanged, and write nothing (whether applying or not).
var malformedMarkerCases = []struct {
	name   string
	memory string
}{
	{
		name:   "begin-only",
		memory: "<!-- version: 1.2.0 -->\n# Memory\n\n<!-- BEGIN GENERATED: convert aadlc -->\norphan content\n\n## Last updated\n2026-01-01\n",
	},
	{
		name:   "end-only",
		memory: "<!-- version: 1.2.0 -->\n# Memory\n\norphan content\n<!-- END GENERATED: convert aadlc -->\n\n## Last updated\n2026-01-01\n",
	},
	{
		name:   "end-before-begin",
		memory: "<!-- version: 1.2.0 -->\n# Memory\n\n<!-- END GENERATED: convert aadlc -->\n\nsome text\n\n<!-- BEGIN GENERATED: convert aadlc -->\n\n## Last updated\n2026-01-01\n",
	},
}

func TestMalformedMarkersFailSafely(t *testing.T) {
	for _, tc := range malformedMarkerCases {
		tc := tc
		// Malformed markers must fail safely in both dry-run and apply modes.
		for _, apply := range []bool{false, true} {
			mode := "dry-run"
			if apply {
				mode = "apply"
			}
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				root, memPath := newRepoWithMemory(t, tc.memory)

				err := RunInDir(root, aadlcConverter{}, apply)
				if err == nil {
					t.Fatalf("expected error for malformed markers (%s)", tc.name)
				}
				if !strings.Contains(err.Error(), "malformed") {
					t.Errorf("error should describe markers as malformed; got: %q", err.Error())
				}
				if got := readFile(t, memPath); got != tc.memory {
					t.Errorf("memory.md must not be modified on malformed marker error;\ngot:\n%s", got)
				}
			})
		}
	}
}

func TestInvariantsRoundTrip(t *testing.T) {
	parsed := parseInvariants(baseInvariants)
	if len(parsed) != 1 {
		t.Fatalf("expected 1 invariant, got %d", len(parsed))
	}
	if parsed[0].id != "no-hardcoded-secrets" {
		t.Errorf("unexpected id %q", parsed[0].id)
	}
	// Quoted values written by this command must round-trip.
	rendered := appendInvariants(baseInvariants, []invariant{{
		id: "aadlc-x", name: `A "tricky": name`, rule: `Rule with: colon`, severity: "high",
	}})
	reparsed := parseInvariants(rendered)
	if len(reparsed) != 2 {
		t.Fatalf("expected 2 invariants after append, got %d", len(reparsed))
	}
	if reparsed[1].rule != "Rule with: colon" {
		t.Errorf("quoted rule did not round-trip: %q", reparsed[1].rule)
	}
	if reparsed[1].name != `A "tricky": name` {
		t.Errorf("quoted name did not round-trip: %q", reparsed[1].name)
	}
}
