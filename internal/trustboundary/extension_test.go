package trustboundary_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goldjg/carl/internal/trustboundary"
)

func TestInspectOptionalExtension(t *testing.T) {
	root := t.TempDir()
	state, err := trustboundary.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if state.Present || state.InvalidReason != "" {
		t.Fatalf("absent extension state = %+v, want absent and valid", state)
	}

	path := filepath.Join(root, filepath.FromSlash(trustboundary.ExtensionPath))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# Repository trust boundaries\n"), 0644); err != nil {
		t.Fatal(err)
	}
	state, err = trustboundary.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Present || state.InvalidReason != "" {
		t.Fatalf("regular extension state = %+v, want present and valid", state)
	}
}

func TestInspectRejectsNonRegularExtension(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, filepath.FromSlash(trustboundary.ExtensionPath))
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	state, err := trustboundary.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if state.Present || state.InvalidReason == "" {
		t.Fatalf("directory extension state = %+v, want invalid", state)
	}
}

func TestInspectRejectsSymlinkedPathComponents(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".github")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	state, err := trustboundary.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if state.Present || state.InvalidReason == "" {
		t.Fatalf("symlinked parent state = %+v, want invalid", state)
	}
}

func TestInspectRejectsSymlinkedExtensionFile(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("outside content"), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, filepath.FromSlash(trustboundary.ExtensionPath))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, path); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}

	state, err := trustboundary.Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if state.Present || state.InvalidReason == "" {
		t.Fatalf("symlinked extension state = %+v, want invalid", state)
	}
}
