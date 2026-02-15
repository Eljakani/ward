# Laravel Ward

A comprehensive security scanner for Laravel applications, built in Go with a rich terminal UI.

> **Status:** Early development. The TUI foundation and core architecture are in place. Scanner implementations, resolvers, and reporting are coming next.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                              │
│                     (cobra commands, flags, TUI)                    │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Orchestrator                               │
│              (scan lifecycle, pipeline coordination)                 │
└──────┬──────────┬───────────┬───────────┬──────────┬───────────────┘
       │          │           │           │          │
       ▼          ▼           ▼           ▼          ▼
┌──────────┐┌──────────┐┌──────────┐┌──────────┐┌──────────┐
│ Provider ││ Resolver ││ Scanner  ││ Reporter ││  Store   │
│ (source) ││(project  ││ Engine   ││ Engine   ││ (state)  │
│          ││ context) ││          ││          ││          │
└──────────┘└──────────┘└──────────┘└──────────┘└──────────┘
```

The project follows a fully decoupled, plugin-style architecture:

- **Interface-first** — every major component is a Go interface
- **Scanner isolation** — each scanner is self-contained and communicates only through shared types
- **Event-driven** — the TUI subscribes to an event bus; scanners never import TUI code
- **Pipeline, not hierarchy** — the scan is a pipeline of stages, not a tree of calls

## What's Built

### Contract Layer (`internal/models/`)

Shared types that form the API boundary between all components:

- `Severity` — Info / Low / Medium / High / Critical
- `Finding` — a single security issue with title, description, severity, file location, code snippet, remediation, and references
- `ProjectContext` — resolved project metadata consumed by scanners
- `ScanReport` — aggregate scan result with findings grouped by severity and category
- `Scanner` — the interface all scanners will implement
- `PipelineStage` — Provider → Resolvers → Scanners → Post-Process → Report

### Event Bus (`internal/eventbus/`)

A thread-safe publish/subscribe system with 13 event types covering the full scan lifecycle. A bridge adapter converts bus events into Bubble Tea messages via `program.Send()`, keeping the TUI completely decoupled from scan logic.

### TUI (`internal/tui/`)

Built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles).

**Scan View** — displayed during scanning:
- Horizontal pipeline stage progress with animated indicators
- Scanner panel with live spinners and per-scanner finding counts
- Severity count badges updating in real-time
- Scrollable event log

**Results View** — displayed after scan completion:
- Summary dashboard with severity breakdown
- Sortable findings table (by severity, category, or file)
- Detail panel with description, code snippet, remediation, and references
- Keyboard-driven navigation between table and detail panels

**Theme system** with `AdaptiveColor` for automatic light/dark terminal support.

### CLI (`cmd/`)

- `ward scan <path>` — scan a Laravel project (scanner implementations pending)
- `ward version` — print version info

## Project Structure

```
laravel-ward/
├── main.go
├── cmd/
│   ├── root.go              # Cobra root command, global flags
│   ├── scan.go              # ward scan <path>
│   └── version.go           # ward version
└── internal/
    ├── models/              # Shared types (the contract layer)
    │   ├── severity.go
    │   ├── finding.go
    │   ├── context.go
    │   ├── report.go
    │   ├── scanner.go
    │   └── pipeline.go
    ├── eventbus/            # Decoupled event system
    │   ├── events.go        # Event types and payloads
    │   ├── bus.go           # Pub/sub implementation
    │   └── bridge.go        # EventBus → Bubble Tea adapter
    └── tui/                 # Terminal UI
        ├── app.go           # Root Bubble Tea model
        ├── messages.go      # Internal TUI messages
        ├── keymap.go        # Key bindings
        ├── theme/
        │   └── theme.go     # Colors, styles, adaptive theming
        ├── components/
        │   ├── header.go
        │   ├── footer.go
        │   ├── stageprogress.go
        │   ├── scannerpanel.go
        │   ├── livestats.go
        │   ├── eventlog.go
        │   ├── severitybadge.go
        │   └── findingdetail.go
        └── views/
            ├── scan.go      # Scanning-in-progress view
            └── results.go   # Post-scan results view
```

## Requirements

- Go 1.25+

## Build

```bash
go build -o ward .
```

## Usage

```bash
# Print version
ward version

# Scan a Laravel project (scanner implementations pending)
ward scan /path/to/laravel-project
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `?` | Toggle help |
| `Tab` | Switch view / panel |
| `j` / `k` / `Up` / `Down` | Navigate |
| `s` | Cycle sort column (results view) |
| `Esc` | Back to scan view |

## Roadmap

- [ ] Source providers (local, git, archive)
- [ ] Context resolvers (Laravel version, routes, models, controllers, middleware, config, Blade templates)
- [ ] Scanner implementations (env, config, auth, injection, XSS, dependencies, cryptography, infrastructure)
- [ ] Rule engine with YAML-based rule definitions
- [ ] Report generation (JSON, SARIF, HTML, Markdown)
- [ ] Result store (SQLite) for trending and baseline diffs
- [ ] Policy engine for CI pass/fail thresholds
- [ ] AI-assisted scanning

## License

TBD
