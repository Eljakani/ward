package resolver

import (
	"context"

	"github.com/eljakani/ward/internal/models"
)

// ContextResolver builds a section of the ProjectContext.
type ContextResolver interface {
	Name() string
	Resolve(ctx context.Context, root string, pc *models.ProjectContext) error
	Priority() int // lower runs first
}
