package semantic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigMissingDisablesSemanticEvaluation(t *testing.T) {
	cfg, exists, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if exists {
		t.Fatal("config should not exist")
	}
	if cfg.Enabled {
		t.Fatal("missing config must not enable semantic evaluation")
	}
	if cfg.FailMode != DefaultFailMode || cfg.JEV.APIKeyEnv != DefaultAPIKeyEnv {
		t.Fatalf("defaults not applied: %+v", cfg)
	}
}

func TestLoadConfigParsesJEVConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".github", "carl", "config.yml")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`semantic_evaluation:
  enabled: true
  provider: jev
  fail_mode: open
  jev:
    api_key_env: TYPESAFE_TEST_KEY
    endpoint: "https://example.invalid/jev"
    model: "jev-review"
    timeout_seconds: 7
`), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, exists, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !exists || !cfg.Enabled || cfg.Provider != "jev" || cfg.JEV.APIKeyEnv != "TYPESAFE_TEST_KEY" || cfg.JEV.TimeoutSeconds != 7 {
		t.Fatalf("unexpected config: exists=%t cfg=%+v", exists, cfg)
	}
}
