package reporter

import (
	"context"

	"github.com/eljakani/ward/internal/models"
)

// Reporter generates output from a scan report.
type Reporter interface {
	Name() string
	Format() string // file extension
	Generate(ctx context.Context, report *models.ScanReport) error
}
