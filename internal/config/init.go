package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed defaults/rules/*.yaml
var defaultRulesFS embed.FS

const defaultConfigYAML = `# Ward configuration
# https://github.com/eljakani/ward

# Minimum severity to report: info, low, medium, high, critical
severity: info

output:
  formats:
    - json
    - sarif
    - html
    - markdown
  dir: .

scanners:
  # enable: []   # if empty, all scanners run
  disable: []    # scanner names to skip

rules:
  disable: []    # rule IDs to disable globally
  override: {}   # rule ID -> {severity, enabled}
  # custom_dirs: # extra directories to load rules from
  #   - /path/to/my-rules

ai:
  enabled: false
  provider: openai
  model: gpt-4o
  # api_key: sk-...         # or set WARD_AI_API_KEY env var
  # endpoint: http://...    # for ollama / custom endpoints

providers:
  git_depth: 1
`

// Init creates the ~/.ward directory structure with default files.
// If force is true, existing files are overwritten.
func Init(force bool) (string, error) {
	if err := EnsureDir(); err != nil {
		return "", err
	}

	dir, err := Dir()
	if err != nil {
		return "", err
	}

	configPath, err := FilePath("config.yaml")
	if err != nil {
		return "", err
	}

	if err := writeIfMissing(configPath, defaultConfigYAML, force); err != nil {
		return "", fmt.Errorf("writing config.yaml: %w", err)
	}

	// Copy all embedded default rules to ~/.ward/rules/
	rulesDir, err := RulesDir()
	if err != nil {
		return "", err
	}

	entries, err := defaultRulesFS.ReadDir("defaults/rules")
	if err != nil {
		return "", fmt.Errorf("reading embedded rules: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := defaultRulesFS.ReadFile("defaults/rules/" + entry.Name())
		if err != nil {
			return "", fmt.Errorf("reading embedded rule %s: %w", entry.Name(), err)
		}

		targetPath := filepath.Join(rulesDir, entry.Name())
		if err := writeIfMissing(targetPath, string(content), force); err != nil {
			return "", fmt.Errorf("writing rule %s: %w", entry.Name(), err)
		}
	}

	return dir, nil
}

func writeIfMissing(path, content string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return os.WriteFile(path, []byte(content), 0644)
}
