package models

// PipelineStage represents a stage in the scan pipeline.
type PipelineStage int

const (
	StageProvider    PipelineStage = iota
	StageResolvers
	StageScanners
	StagePostProcess
	StageReport
)

func (s PipelineStage) String() string {
	switch s {
	case StageProvider:
		return "Provider"
	case StageResolvers:
		return "Resolvers"
	case StageScanners:
		return "Scanners"
	case StagePostProcess:
		return "Post-Process"
	case StageReport:
		return "Report"
	default:
		return "Unknown"
	}
}

// AllStages returns all pipeline stages in execution order.
func AllStages() []PipelineStage {
	return []PipelineStage{
		StageProvider,
		StageResolvers,
		StageScanners,
		StagePostProcess,
		StageReport,
	}
}
