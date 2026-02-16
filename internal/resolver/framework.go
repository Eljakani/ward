package resolver

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/eljakani/ward/internal/models"
)

// FrameworkResolver detects Laravel version, PHP version, project name,
// and populates the dependency map from composer.json.
type FrameworkResolver struct{}

func NewFrameworkResolver() *FrameworkResolver {
	return &FrameworkResolver{}
}

func (r *FrameworkResolver) Name() string     { return "framework" }
func (r *FrameworkResolver) Priority() int    { return 10 }

func (r *FrameworkResolver) Resolve(_ context.Context, root string, pc *models.ProjectContext) error {
	pc.RootPath = root
	pc.FrameworkType = "laravel"

	r.resolveComposer(root, pc)
	r.resolveEnv(root, pc)

	return nil
}

func (r *FrameworkResolver) resolveComposer(root string, pc *models.ProjectContext) {
	data, err := os.ReadFile(filepath.Join(root, "composer.json"))
	if err != nil {
		return
	}

	var composer struct {
		Name    string            `json:"name"`
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return
	}

	if composer.Name != "" && pc.ProjectName == "" {
		pc.ProjectName = composer.Name
	}

	pc.ComposerDeps = make(map[string]string, len(composer.Require))
	for pkg, ver := range composer.Require {
		pc.ComposerDeps[pkg] = ver
	}

	if v, ok := composer.Require["laravel/framework"]; ok {
		pc.LaravelVersion = v
	}
	if v, ok := composer.Require["php"]; ok {
		pc.PHPVersion = v
	}
}

func (r *FrameworkResolver) resolveEnv(root string, pc *models.ProjectContext) {
	envVars := parseEnvFile(filepath.Join(root, ".env"))

	if name, ok := envVars["APP_NAME"]; ok && pc.ProjectName == "" {
		pc.ProjectName = name
	}

	// Store variable names (mask values for context — scanners will read raw)
	pc.EnvVariables = make(map[string]string, len(envVars))
	for k := range envVars {
		pc.EnvVariables[k] = "***"
	}

	// Discover config files
	configDir := filepath.Join(root, "config")
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".php") {
			pc.ConfigFiles = append(pc.ConfigFiles, filepath.Join("config", entry.Name()))
		}
	}
}

// parseEnvFile reads a .env file into a key-value map.
func parseEnvFile(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip surrounding quotes
		val = strings.Trim(val, `"'`)
		vars[key] = val
	}
	return vars
}
