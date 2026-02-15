package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WardConfig is the top-level configuration loaded from ~/.ward/config.yaml.
type WardConfig struct {
	Severity  string          `yaml:"severity"`  // minimum severity to report: info, low, medium, high, critical
	Output    OutputConfig    `yaml:"output"`
	Scanners  ScannersConfig  `yaml:"scanners"`
	Rules     RulesConfig     `yaml:"rules"`
	AI        AIConfig        `yaml:"ai"`
	Providers ProvidersConfig `yaml:"providers"`
}

// OutputConfig controls report formats and destinations.
type OutputConfig struct {
	Formats []string `yaml:"formats"` // terminal, json, sarif, html, markdown
	Dir     string   `yaml:"dir"`     // output directory for file reports
}

// ScannersConfig controls which scanners are enabled.
type ScannersConfig struct {
	Enable  []string `yaml:"enable"`  // explicit list; if empty, all are enabled
	Disable []string `yaml:"disable"` // scanners to skip
}

// RulesConfig controls rule overrides and custom rules.
type RulesConfig struct {
	Disable  []string                `yaml:"disable"`  // rule IDs to disable
	Override map[string]RuleOverride `yaml:"override"`  // rule ID → overrides
	CustomDirs []string              `yaml:"custom_dirs"` // extra dirs to load rules from
}

// RuleOverride lets users change severity or disable a built-in rule.
type RuleOverride struct {
	Severity string `yaml:"severity,omitempty"`
	Enabled  *bool  `yaml:"enabled,omitempty"`
}

// AIConfig holds settings for AI-assisted scanning.
type AIConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"` // openai, anthropic, ollama
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"` // can also come from WARD_AI_API_KEY env
	Endpoint string `yaml:"endpoint,omitempty"`
}

// ProvidersConfig controls source provider behaviour.
type ProvidersConfig struct {
	GitDepth int `yaml:"git_depth"` // shallow clone depth, 0 = full
}

// Default returns the default configuration.
func Default() *WardConfig {
	return &WardConfig{
		Severity: "info",
		Output: OutputConfig{
			Formats: []string{"terminal"},
			Dir:     ".",
		},
		Scanners: ScannersConfig{},
		Rules:    RulesConfig{},
		AI: AIConfig{
			Enabled:  false,
			Provider: "openai",
			Model:    "gpt-4o",
		},
		Providers: ProvidersConfig{
			GitDepth: 1,
		},
	}
}

// Load reads the config from ~/.ward/config.yaml.
// If the file doesn't exist it returns the defaults.
func Load() (*WardConfig, error) {
	cfg := Default()

	path, err := FilePath("config.yaml")
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}

// Save writes the config to ~/.ward/config.yaml.
func Save(cfg *WardConfig) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	path, err := FilePath("config.yaml")
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	header := []byte("# Ward configuration\n# https://github.com/eljakani/ward\n\n")
	return os.WriteFile(path, append(header, data...), 0644)
}
