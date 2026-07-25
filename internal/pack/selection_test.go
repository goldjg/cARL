package pack

import (
	"os"
	"path/filepath"
	"testing"
)

// Contract assertion 1: selection persists deterministically (sorted,
// deduplicated) and malformed selection files are explicit errors.
func TestSelectionWriteReadDeterministic(t *testing.T) {
	dir := t.TempDir()

	if err := WriteSelection(dir, []string{"core/carl", "cloud/azure", "core/carl", "core/baseline"}); err != nil {
		t.Fatalf("WriteSelection: %v", err)
	}

	sel, err := ReadSelection(dir)
	if err != nil {
		t.Fatalf("ReadSelection: %v", err)
	}
	if sel == nil {
		t.Fatal("expected selection")
	}
	want := []string{"cloud/azure", "core/baseline", "core/carl"}
	if len(sel.Selected) != len(want) {
		t.Fatalf("selected = %v; want %v", sel.Selected, want)
	}
	for i := range want {
		if sel.Selected[i] != want[i] {
			t.Fatalf("selected = %v; want %v", sel.Selected, want)
		}
	}
}

func TestReadSelectionAbsentReturnsNil(t *testing.T) {
	sel, err := ReadSelection(t.TempDir())
	if err != nil {
		t.Fatalf("ReadSelection: %v", err)
	}
	if sel != nil {
		t.Fatalf("expected nil selection, got %+v", sel)
	}
}

func TestReadSelectionMalformed(t *testing.T) {
	cases := map[string]string{
		"invalid json":       `{not json`,
		"bad schema version": `{"schemaVersion": 99, "selected": []}`,
		"malformed pack id":  `{"schemaVersion": 1, "selected": ["Not A Pack"]}`,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, filepath.FromSlash(SelectionFileName))
			if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadSelection(dir); err == nil {
				t.Fatal("expected explicit error for malformed selection file")
			}
		})
	}
}

// Contract assertion 1: select/unselect commands validate pack existence and
// persist selection; the selection artefact is authoritative for discovery.
func TestPackSelectUnselectCommands(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())

	_ = captureStdout(t, func() {
		if err := cmd.RunSelectInDir(dir, []string{"core/carl", "core/baseline"}, false); err != nil {
			t.Fatalf("RunSelectInDir: %v", err)
		}
	})

	sel, err := ReadSelection(dir)
	if err != nil {
		t.Fatalf("ReadSelection: %v", err)
	}
	if sel == nil || len(sel.Selected) != 2 || sel.Selected[0] != "core/baseline" || sel.Selected[1] != "core/carl" {
		t.Fatalf("unexpected selection: %+v", sel)
	}

	packs, err := cmd.discover(dir)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	for _, p := range packs {
		wantSelected := p.ID == "core/baseline" || p.ID == "core/carl"
		if p.State.Selected != wantSelected {
			t.Fatalf("%s: selected = %v; want %v", p.ID, p.State.Selected, wantSelected)
		}
	}

	_ = captureStdout(t, func() {
		if err := cmd.RunUnselectInDir(dir, []string{"core/carl"}, false); err != nil {
			t.Fatalf("RunUnselectInDir: %v", err)
		}
	})
	sel, err = ReadSelection(dir)
	if err != nil {
		t.Fatalf("ReadSelection: %v", err)
	}
	if sel == nil || len(sel.Selected) != 1 || sel.Selected[0] != "core/baseline" {
		t.Fatalf("unexpected selection after unselect: %+v", sel)
	}
}

func TestPackSelectUnknownPack(t *testing.T) {
	dir := t.TempDir()
	cmd := New(bundledArtifacts())

	if err := cmd.RunSelectInDir(dir, []string{"core/nonexistent"}, false); err == nil {
		t.Fatal("expected error for unknown pack")
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(SelectionFileName))); !os.IsNotExist(err) {
		t.Fatal("selection file must not be written when selection fails")
	}
}
