package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LocalProvider acquires source from a local filesystem path.
type LocalProvider struct{}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}

func (p *LocalProvider) Acquire(_ context.Context, path string) (*SourceResult, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	result := &SourceResult{
		RootPath:  absPath,
		IsLaravel: false,
		HasGit:    false,
	}

	// Check for .git directory
	if _, err := os.Stat(filepath.Join(absPath, ".git")); err == nil {
		result.HasGit = true
	}

	// Check for artisan file (strong Laravel signal)
	if _, err := os.Stat(filepath.Join(absPath, "artisan")); err == nil {
		result.IsLaravel = true
		return result, nil
	}

	// Fall back to checking composer.json for laravel/framework
	composerPath := filepath.Join(absPath, "composer.json")
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return result, nil // not an error, just not Laravel
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return result, nil
	}

	if _, ok := composer.Require["laravel/framework"]; ok {
		result.IsLaravel = true
	}

	return result, nil
}

func (p *LocalProvider) Cleanup() error {
	return nil // local provider doesn't create temp files
}
