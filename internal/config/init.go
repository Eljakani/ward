package config

import (
	"fmt"
	"os"
)

const defaultConfigYAML = `# Ward configuration
# https://github.com/eljakani/ward

# Minimum severity to report: info, low, medium, high, critical
severity: info

output:
  formats:
    - terminal
  # dir: ./reports

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

const exampleRulesYAML = `# Example custom rules for Ward
# Place .yaml files in this directory to add your own rules.
# Ward loads all .yaml/.yml files from ~/.ward/rules/ automatically.

rules:
  - id: CUSTOM-001
    title: "Hardcoded internal API key"
    description: "Detects hardcoded internal API keys in source files."
    severity: high
    category: secrets
    enabled: true
    tags:
      - secrets
      - cwe-798
    patterns:
      - type: regex
        target: php-files
        pattern: 'INTERNAL_API_KEY\s*=\s*[''"][a-zA-Z0-9]+'
    remediation: |
      Move API keys to environment variables.
      Use .env files or a secrets manager instead of hardcoding keys.
    references:
      - https://cwe.mitre.org/data/definitions/798.html

  # - id: CUSTOM-002
  #   title: "Your rule title"
  #   description: "What this rule checks for."
  #   severity: medium
  #   category: config
  #   enabled: true
  #   patterns:
  #     - type: regex
  #       target: config-files
  #       pattern: 'some_pattern'
  #   remediation: "How to fix it."
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

	rulesDir, err := RulesDir()
	if err != nil {
		return "", err
	}

	examplePath := rulesDir + "/example.yaml"
	if err := writeIfMissing(examplePath, exampleRulesYAML, force); err != nil {
		return "", fmt.Errorf("writing example rules: %w", err)
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
