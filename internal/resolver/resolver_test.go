package resolver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eljakani/ward/internal/models"
)

func TestFrameworkResolver_ComposerJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{
		"name": "acme/myapp",
		"require": {
			"php": "^8.2",
			"laravel/framework": "^11.0",
			"guzzlehttp/guzzle": "^7.8"
		}
	}`), 0644)

	r := NewFrameworkResolver()
	pc := &models.ProjectContext{}
	err := r.Resolve(context.Background(), dir, pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pc.ProjectName != "acme/myapp" {
		t.Errorf("ProjectName = %q, want %q", pc.ProjectName, "acme/myapp")
	}
	if pc.LaravelVersion != "^11.0" {
		t.Errorf("LaravelVersion = %q, want %q", pc.LaravelVersion, "^11.0")
	}
	if pc.PHPVersion != "^8.2" {
		t.Errorf("PHPVersion = %q, want %q", pc.PHPVersion, "^8.2")
	}
	if len(pc.ComposerDeps) != 3 {
		t.Errorf("ComposerDeps count = %d, want 3", len(pc.ComposerDeps))
	}
}

func TestFrameworkResolver_EnvFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"name":"test/app"}`), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("APP_NAME=MyLaravelApp\nAPP_ENV=production\n"), 0644)

	r := NewFrameworkResolver()
	pc := &models.ProjectContext{}
	r.Resolve(context.Background(), dir, pc)

	// composer.json name takes priority, but env vars should be discovered
	if len(pc.EnvVariables) != 2 {
		t.Errorf("EnvVariables count = %d, want 2", len(pc.EnvVariables))
	}
}

func TestFrameworkResolver_ConfigDiscovery(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{}`), 0644)

	configDir := filepath.Join(dir, "config")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "app.php"), []byte("<?php return [];"), 0644)
	os.WriteFile(filepath.Join(configDir, "auth.php"), []byte("<?php return [];"), 0644)

	r := NewFrameworkResolver()
	pc := &models.ProjectContext{}
	r.Resolve(context.Background(), dir, pc)

	if len(pc.ConfigFiles) != 2 {
		t.Errorf("ConfigFiles count = %d, want 2", len(pc.ConfigFiles))
	}
}

func TestPackageResolver_ComposerLock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.lock"), []byte(`{
		"packages": [
			{"name": "laravel/framework", "version": "v11.5.0"},
			{"name": "guzzlehttp/guzzle", "version": "7.8.1"}
		],
		"packages-dev": [
			{"name": "phpunit/phpunit", "version": "11.0.1"}
		]
	}`), 0644)

	r := NewPackageResolver()
	pc := &models.ProjectContext{}
	err := r.Resolve(context.Background(), dir, pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pc.InstalledPackages) != 3 {
		t.Errorf("InstalledPackages count = %d, want 3", len(pc.InstalledPackages))
	}
	if pc.InstalledPackages["laravel/framework"] != "v11.5.0" {
		t.Errorf("laravel/framework = %q, want %q", pc.InstalledPackages["laravel/framework"], "v11.5.0")
	}
}

func TestPackageResolver_NoLockFile(t *testing.T) {
	dir := t.TempDir()

	r := NewPackageResolver()
	pc := &models.ProjectContext{}
	err := r.Resolve(context.Background(), dir, pc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be nil or empty — no error
	if len(pc.InstalledPackages) != 0 {
		t.Errorf("InstalledPackages should be empty, got %d", len(pc.InstalledPackages))
	}
}
