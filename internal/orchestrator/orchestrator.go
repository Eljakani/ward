package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eljakani/ward/internal/baseline"
	"github.com/eljakani/ward/internal/config"
	"github.com/eljakani/ward/internal/eventbus"
	"github.com/eljakani/ward/internal/models"
	"github.com/eljakani/ward/internal/provider"
	"github.com/eljakani/ward/internal/reporter"
	"github.com/eljakani/ward/internal/resolver"
	configscanner "github.com/eljakani/ward/internal/scanner/configscan"
	depscanner "github.com/eljakani/ward/internal/scanner/dependency"
	envscanner "github.com/eljakani/ward/internal/scanner/env"
	rulesscanner "github.com/eljakani/ward/internal/scanner/rules"
	"github.com/eljakani/ward/internal/store"
)

// Orchestrator coordinates the full scan pipeline.
type Orchestrator struct {
	bus          *eventbus.EventBus
	cfg          *config.WardConfig
	target       string
	version      string
	baseline     *baseline.Baseline
	baselinePath string // if set, save baseline after scan
}

// New creates a new Orchestrator.
func New(bus *eventbus.EventBus, cfg *config.WardConfig, target string, version string) *Orchestrator {
	return &Orchestrator{bus: bus, cfg: cfg, target: target, version: version}
}

// SetBaseline configures an existing baseline for filtering.
func (o *Orchestrator) SetBaseline(b *baseline.Baseline) {
	o.baseline = b
}

// SetBaselinePath configures a path to save a new baseline after scanning.
func (o *Orchestrator) SetBaselinePath(path string) {
	o.baselinePath = path
}

// Run executes the full scan pipeline.
func (o *Orchestrator) Run(ctx context.Context) error {
	startTime := time.Now()

	scanners := []models.Scanner{
		envscanner.New(),
		configscanner.New(),
		depscanner.New(),
	}

	// Load custom YAML rules and add rules scanner if any rules found
	customRules, err := config.LoadAllRules(o.cfg)
	if err != nil {
		o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
			Level: "warn", Message: fmt.Sprintf("Failed to load custom rules: %v", err),
		}))
	} else if len(customRules) > 0 {
		scanners = append(scanners, rulesscanner.New(customRules))
		o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
			Level: "info", Message: fmt.Sprintf("Loaded %d custom rule(s)", len(customRules)),
		}))
	}

	// Filter scanners based on config enable/disable lists
	scanners = o.filterScanners(scanners)

	o.bus.Publish(eventbus.NewEvent(eventbus.EventScanStarted, eventbus.ScanStartedData{
		ProjectPath:  o.target,
		ProjectName:  o.target,
		ScannerCount: len(scanners),
	}))

	// --- Stage 1: Provider ---
	o.stageStart(models.StageProvider)

	var src provider.SourceProvider
	if provider.IsGitURL(o.target) {
		src = provider.NewGitProvider(o.cfg.Providers.GitDepth)
		o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
			Level: "info", Message: fmt.Sprintf("Cloning %s ...", o.target),
		}))
	} else {
		src = provider.NewLocalProvider()
	}

	result, err := src.Acquire(ctx, o.target)
	if err != nil {
		return o.fail(fmt.Errorf("provider: %w", err))
	}
	defer src.Cleanup()

	if !result.IsLaravel {
		o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
			Level: "warn", Message: "Path does not appear to be a Laravel project",
		}))
	}

	o.stageComplete(models.StageProvider)

	// --- Stage 2: Resolvers ---
	o.stageStart(models.StageResolvers)

	pc := &models.ProjectContext{}
	resolvers := []resolver.ContextResolver{
		resolver.NewFrameworkResolver(),
		resolver.NewPackageResolver(),
	}

	for _, r := range resolvers {
		if err := r.Resolve(ctx, result.RootPath, pc); err != nil {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "error", Message: fmt.Sprintf("Resolver %s failed: %v", r.Name(), err),
			}))
		}
	}

	// Emit resolved project context to TUI
	o.bus.Publish(eventbus.NewEvent(eventbus.EventContextResolved, eventbus.ContextResolvedData{
		ProjectName:    pc.ProjectName,
		LaravelVersion: pc.LaravelVersion,
		PHPVersion:     pc.PHPVersion,
		FrameworkType:  pc.FrameworkType,
		PackageCount:   len(pc.InstalledPackages),
	}))

	o.stageComplete(models.StageResolvers)

	// --- Stage 3: Scanners ---
	o.stageStart(models.StageScanners)

	// Register all scanners
	for _, sc := range scanners {
		o.bus.Publish(eventbus.NewEvent(eventbus.EventScannerRegistered, eventbus.ScannerRegisteredData{
			Name:        sc.Name(),
			Description: sc.Description(),
		}))
	}

	var allFindings []models.Finding
	scannersRun := make([]string, 0, len(scanners))
	scannerErrors := make(map[string]string)

	for _, sc := range scanners {
		if o.isScannerDisabled(sc.Name()) {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventScannerSkipped, eventbus.ScannerSkippedData{
				Name: sc.Name(), Reason: "disabled in config",
			}))
			continue
		}

		o.bus.Publish(eventbus.NewEvent(eventbus.EventScannerStarted, eventbus.ScannerStartedData{
			Name: sc.Name(),
		}))

		emit := func(f models.Finding) {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventFindingDiscovered, eventbus.FindingDiscoveredData{
				Finding: f,
			}))
		}

		findings, err := sc.Scan(ctx, *pc, emit)
		if err != nil {
			scannerErrors[sc.Name()] = err.Error()
			o.bus.Publish(eventbus.NewEvent(eventbus.EventScannerFailed, eventbus.ScannerFailedData{
				Name: sc.Name(), Error: err,
			}))
			continue
		}

		allFindings = append(allFindings, findings...)
		scannersRun = append(scannersRun, sc.Name())

		o.bus.Publish(eventbus.NewEvent(eventbus.EventScannerCompleted, eventbus.ScannerCompletedData{
			Name:         sc.Name(),
			FindingCount: len(findings),
		}))
	}

	o.stageComplete(models.StageScanners)

	// --- Stage 4: Post-Process ---
	o.stageStart(models.StagePostProcess)
	allFindings = deduplicate(allFindings)
	allFindings = filterBySeverity(allFindings, models.ParseSeverity(o.cfg.Severity))

	// Apply baseline filtering
	if o.baseline != nil {
		var suppressed int
		allFindings, suppressed = o.baseline.Filter(allFindings)
		if suppressed > 0 {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "info", Message: fmt.Sprintf("%d findings suppressed by baseline", suppressed),
			}))
		}
	}

	o.stageComplete(models.StagePostProcess)

	// --- Stage 5: Report ---
	o.stageStart(models.StageReport)

	endTime := time.Now()
	report := &models.ScanReport{
		ProjectContext: *pc,
		Findings:       allFindings,
		StartedAt:      startTime,
		CompletedAt:    endTime,
		Duration:       endTime.Sub(startTime),
		ScannersRun:    scannersRun,
		ScannerErrors:  scannerErrors,
	}

	reporters := o.buildReporters()
	for _, rep := range reporters {
		if err := rep.Generate(ctx, report); err != nil {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "error", Message: fmt.Sprintf("%s reporter failed: %v", rep.Name(), err),
			}))
		} else {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "info", Message: fmt.Sprintf("Report written to ward-report.%s", rep.Format()),
			}))
		}
	}

	// Compare with last scan and save to store
	diff, _ := store.CompareLast(report)
	if diff != nil {
		if len(diff.NewFindings) > 0 || len(diff.ResolvedFindings) > 0 {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "info", Message: fmt.Sprintf("vs last scan: %d new, %d resolved (%d→%d)",
					len(diff.NewFindings), len(diff.ResolvedFindings), diff.TotalBefore, diff.TotalAfter),
			}))
		}
	}

	if _, err := store.Save(report); err != nil {
		o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
			Level: "warn", Message: fmt.Sprintf("Failed to save scan history: %v", err),
		}))
	}

	o.stageComplete(models.StageReport)

	// Save baseline if requested
	if o.baselinePath != "" {
		if err := baseline.Save(o.baselinePath, allFindings); err != nil {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "warn", Message: fmt.Sprintf("Failed to save baseline: %v", err),
			}))
		} else {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "info", Message: fmt.Sprintf("Baseline saved to %s (%d findings)", o.baselinePath, len(allFindings)),
			}))
		}
	}

	// --- Done ---
	o.bus.Publish(eventbus.NewEvent(eventbus.EventScanCompleted, eventbus.ScanCompletedData{
		Report: report,
	}))

	return nil
}

func (o *Orchestrator) stageStart(stage models.PipelineStage) {
	o.bus.Publish(eventbus.NewEvent(eventbus.EventStageStarted, eventbus.StageStartedData{Stage: stage}))
}

func (o *Orchestrator) stageComplete(stage models.PipelineStage) {
	o.bus.Publish(eventbus.NewEvent(eventbus.EventStageCompleted, eventbus.StageCompletedData{Stage: stage}))
}

func (o *Orchestrator) fail(err error) error {
	o.bus.Publish(eventbus.NewEvent(eventbus.EventScanFailed, eventbus.ScanFailedData{Error: err}))
	return err
}

func (o *Orchestrator) isScannerDisabled(name string) bool {
	for _, d := range o.cfg.Scanners.Disable {
		if d == name {
			return true
		}
	}
	return false
}

func filterBySeverity(findings []models.Finding, minSeverity models.Severity) []models.Finding {
	if minSeverity == models.SeverityInfo {
		return findings
	}
	result := make([]models.Finding, 0, len(findings))
	for _, f := range findings {
		if f.Severity >= minSeverity {
			result = append(result, f)
		}
	}
	return result
}

func (o *Orchestrator) buildReporters() []reporter.Reporter {
	outDir := o.cfg.Output.Dir
	formats := o.cfg.Output.Formats

	// Default to JSON if no formats specified
	if len(formats) == 0 {
		formats = []string{"json"}
	}

	var reporters []reporter.Reporter
	seen := make(map[string]bool)

	for _, f := range formats {
		if seen[f] {
			continue
		}
		seen[f] = true

		switch f {
		case "json":
			reporters = append(reporters, reporter.NewJSONReporter(outDir))
		case "sarif":
			reporters = append(reporters, reporter.NewSARIFReporter(outDir, o.version))
		case "html":
			reporters = append(reporters, reporter.NewHTMLReporter(outDir))
		case "markdown", "md":
			reporters = append(reporters, reporter.NewMarkdownReporter(outDir, o.version))
		case "terminal":
			// terminal output is handled by the headless/TUI path, not a file reporter
			continue
		}
	}

	// Always generate JSON as baseline
	if !seen["json"] {
		reporters = append(reporters, reporter.NewJSONReporter(outDir))
	}

	return reporters
}

func deduplicate(findings []models.Finding) []models.Finding {
	seen := make(map[string]bool)
	result := make([]models.Finding, 0, len(findings))
	for _, f := range findings {
		key := f.ID + "|" + f.File + "|" + fmt.Sprintf("%d", f.Line)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, f)
	}
	return result
}

// filterScanners applies the config enable/disable lists.
func (o *Orchestrator) filterScanners(scanners []models.Scanner) []models.Scanner {
	enable := o.cfg.Scanners.Enable
	disable := o.cfg.Scanners.Disable

	if len(enable) == 0 && len(disable) == 0 {
		return scanners
	}

	enableSet := make(map[string]bool, len(enable))
	for _, name := range enable {
		enableSet[strings.ToLower(name)] = true
	}

	disableSet := make(map[string]bool, len(disable))
	for _, name := range disable {
		disableSet[strings.ToLower(name)] = true
	}

	var filtered []models.Scanner
	for _, s := range scanners {
		name := strings.ToLower(s.Name())

		// If enable list is set, only run scanners in that list
		if len(enableSet) > 0 && !enableSet[name] {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "info", Message: fmt.Sprintf("Skipping %s (not in enable list)", s.Name()),
			}))
			continue
		}

		// If disable list is set, skip scanners in that list
		if disableSet[name] {
			o.bus.Publish(eventbus.NewEvent(eventbus.EventLogMessage, eventbus.LogMessageData{
				Level: "info", Message: fmt.Sprintf("Skipping %s (disabled)", s.Name()),
			}))
			continue
		}

		filtered = append(filtered, s)
	}

	return filtered
}
