package semantic

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ConfigPathYAML   = ".github/carl/config.yml"
	ConfigPathYML    = ".github/carl/config.yaml"
	DefaultAPIKeyEnv = "TYPESAFE_API_KEY"
	DefaultFailMode  = "open"
)

// Config captures optional semantic-evaluation configuration.
type Config struct {
	Enabled  bool
	Provider string
	FailMode string
	JEV      JEVConfig
}

// JEVConfig captures TypeSafe JEV adapter configuration.
type JEVConfig struct {
	APIKeyEnv      string
	Endpoint       string
	Model          string
	TimeoutSeconds int
}

// LoadConfig reads the optional cARL config file. Missing config means
// semantic evaluation is disabled.
func LoadConfig(rootDir string) (Config, bool, error) {
	cfg := Config{FailMode: DefaultFailMode, JEV: JEVConfig{APIKeyEnv: DefaultAPIKeyEnv, TimeoutSeconds: 30}}
	path := filepath.Join(rootDir, filepath.FromSlash(ConfigPathYAML))
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return cfg, false, fmt.Errorf("inspect %s: %w", ConfigPathYAML, err)
		}
		path = filepath.Join(rootDir, filepath.FromSlash(ConfigPathYML))
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return cfg, false, nil
			}
			return cfg, false, fmt.Errorf("inspect %s: %w", ConfigPathYML, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, false, fmt.Errorf("read semantic evaluation config: %w", err)
	}
	if err := parseConfig(data, &cfg); err != nil {
		rel, _ := filepath.Rel(rootDir, path)
		return cfg, true, fmt.Errorf("%s: %w", filepath.ToSlash(rel), err)
	}
	if cfg.FailMode == "" {
		cfg.FailMode = DefaultFailMode
	}
	if cfg.JEV.APIKeyEnv == "" {
		cfg.JEV.APIKeyEnv = DefaultAPIKeyEnv
	}
	if cfg.JEV.TimeoutSeconds == 0 {
		cfg.JEV.TimeoutSeconds = 30
	}
	return cfg, true, nil
}

func parseConfig(data []byte, cfg *Config) error {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasSuffix(trimmed, ":") {
			key := strings.TrimSuffix(trimmed, ":")
			switch key {
			case "semantic_evaluation":
				section = "semantic"
			case "jev":
				if section != "semantic" {
					return fmt.Errorf("jev section must be nested under semantic_evaluation")
				}
				section = "jev"
			default:
				section = ""
			}
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return fmt.Errorf("unsupported config line %q", trimmed)
		}
		key = strings.TrimSpace(key)
		value = cleanScalar(value)
		switch section {
		case "semantic":
			switch key {
			case "enabled":
				enabled, err := strconv.ParseBool(value)
				if err != nil {
					return fmt.Errorf("semantic_evaluation.enabled must be true or false")
				}
				cfg.Enabled = enabled
			case "provider":
				cfg.Provider = value
			case "fail_mode":
				if value != "open" && value != "closed" {
					return fmt.Errorf("semantic_evaluation.fail_mode must be open or closed")
				}
				cfg.FailMode = value
			default:
				return fmt.Errorf("unsupported semantic_evaluation key %q", key)
			}
		case "jev":
			switch key {
			case "api_key_env":
				cfg.JEV.APIKeyEnv = value
			case "endpoint":
				cfg.JEV.Endpoint = value
			case "model":
				cfg.JEV.Model = value
			case "timeout_seconds":
				n, err := strconv.Atoi(value)
				if err != nil || n < 1 || n > 300 {
					return fmt.Errorf("semantic_evaluation.jev.timeout_seconds must be between 1 and 300")
				}
				cfg.JEV.TimeoutSeconds = n
			default:
				return fmt.Errorf("unsupported semantic_evaluation.jev key %q", key)
			}
		default:
			// Ignore unrelated top-level config keys so one file can grow later.
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan config: %w", err)
	}
	return nil
}

func cleanScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}
