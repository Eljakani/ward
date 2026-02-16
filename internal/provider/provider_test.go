package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalProvider_ValidLaravelProject(t *testing.T) {
	dir := t.TempDir()

	// Create artisan file
	os.WriteFile(filepath.Join(dir, "artisan"), []byte("#!/usr/bin/env php"), 0755)
	// Create composer.json
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"require":{"laravel/framework":"^11.0"}}`), 0644)

	p := NewLocalProvider()
	result, err := p.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsLaravel {
		t.Error("expected IsLaravel = true")
	}
	if result.RootPath != dir {
		t.Errorf("RootPath = %q, want %q", result.RootPath, dir)
	}
}

func TestLocalProvider_ComposerOnlyDetection(t *testing.T) {
	dir := t.TempDir()

	// No artisan, but composer.json has laravel/framework
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"require":{"laravel/framework":"^11.0"}}`), 0644)

	p := NewLocalProvider()
	result, err := p.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsLaravel {
		t.Error("expected IsLaravel = true from composer.json detection")
	}
}

func TestLocalProvider_NonLaravelProject(t *testing.T) {
	dir := t.TempDir()

	// Just a random PHP project
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"require":{"some/package":"^1.0"}}`), 0644)

	p := NewLocalProvider()
	result, err := p.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsLaravel {
		t.Error("expected IsLaravel = false for non-Laravel project")
	}
}

func TestLocalProvider_NonExistentPath(t *testing.T) {
	p := NewLocalProvider()
	_, err := p.Acquire(context.Background(), "/nonexistent/path/xyz")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestLocalProvider_FilePath(t *testing.T) {
	f, err := os.CreateTemp("", "ward-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	p := NewLocalProvider()
	_, err = p.Acquire(context.Background(), f.Name())
	if err == nil {
		t.Error("expected error for file path (not directory)")
	}
}

func TestLocalProvider_GitDetection(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, ".git"), 0755)
	os.WriteFile(filepath.Join(dir, "artisan"), []byte(""), 0644)

	p := NewLocalProvider()
	result, err := p.Acquire(context.Background(), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasGit {
		t.Error("expected HasGit = true")
	}
}

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://github.com/user/repo.git", true},
		{"http://github.com/user/repo", true},
		{"git@github.com:user/repo.git", true},
		{"ssh://git@github.com/user/repo", true},
		{"/local/path", false},
		{"./relative/path", false},
		{"some-repo.git", true},
	}

	for _, tt := range tests {
		if got := IsGitURL(tt.input); got != tt.want {
			t.Errorf("IsGitURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
