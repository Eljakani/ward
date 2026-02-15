package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const wardDir = ".ward"

// Dir returns the absolute path to ~/.ward.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, wardDir), nil
}

// FilePath returns the absolute path to a file inside ~/.ward.
func FilePath(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// SubDir returns the absolute path to a subdirectory inside ~/.ward
// and ensures it exists.
func SubDir(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("creating %s: %w", path, err)
	}
	return path, nil
}

// EnsureDir creates ~/.ward and its standard subdirectories if they don't exist.
func EnsureDir() error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	dirs := []string{
		dir,
		filepath.Join(dir, "rules"),
		filepath.Join(dir, "reports"),
		filepath.Join(dir, "store"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", d, err)
		}
	}

	return nil
}

// RulesDir returns the path to ~/.ward/rules.
func RulesDir() (string, error) {
	return SubDir("rules")
}

// ReportsDir returns the path to ~/.ward/reports.
func ReportsDir() (string, error) {
	return SubDir("reports")
}

// StoreDir returns the path to ~/.ward/store.
func StoreDir() (string, error) {
	return SubDir("store")
}
