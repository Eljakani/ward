package resolver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/eljakani/ward/internal/models"
)

// PackageResolver reads composer.lock to populate resolved package versions.
type PackageResolver struct{}

func NewPackageResolver() *PackageResolver {
	return &PackageResolver{}
}

func (r *PackageResolver) Name() string  { return "packages" }
func (r *PackageResolver) Priority() int { return 20 }

func (r *PackageResolver) Resolve(_ context.Context, root string, pc *models.ProjectContext) error {
	data, err := os.ReadFile(filepath.Join(root, "composer.lock"))
	if err != nil {
		return nil // no lock file is fine
	}

	var lock composerLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil
	}

	if pc.InstalledPackages == nil {
		pc.InstalledPackages = make(map[string]string)
	}

	for _, pkg := range lock.Packages {
		pc.InstalledPackages[pkg.Name] = pkg.Version
	}
	for _, pkg := range lock.PackagesDev {
		pc.InstalledPackages[pkg.Name] = pkg.Version
	}

	return nil
}

type composerLock struct {
	Packages    []composerPackage `json:"packages"`
	PackagesDev []composerPackage `json:"packages-dev"`
}

type composerPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
