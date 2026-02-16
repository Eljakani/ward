package provider

import "context"

// SourceProvider abstracts where the project code comes from.
type SourceProvider interface {
	Acquire(ctx context.Context, path string) (*SourceResult, error)
	Cleanup() error
}

// SourceResult holds information about the acquired source.
type SourceResult struct {
	RootPath  string
	IsLaravel bool
	HasGit    bool
}
