```
 ██╗    ██╗ █████╗ ██████╗ ██████╗
 ██║    ██║██╔══██╗██╔══██╗██╔══██╗
 ██║ █╗ ██║███████║██████╔╝██║  ██║
 ██║███╗██║██╔══██║██╔══██╗██║  ██║
 ╚███╔███╔╝██║  ██║██║  ██║██████╔╝
  ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝
```

# Ward

**A security scanner built specifically for Laravel.**

Ward understands your Laravel application — its routes, models, controllers, middleware, Blade templates, config files, `.env` secrets, Composer dependencies, and more. It doesn't just grep for patterns. It resolves your project's structure first, then runs targeted security checks against it.

> **Status:** Early development. The TUI, event system, config layer, and core architecture are in place. Scanner implementations are coming next.

---

## Why Ward?

Laravel gives you a lot out of the box — CSRF protection, Eloquent's mass assignment guards, Bcrypt hashing, encrypted cookies. But it's easy to misconfigure things or leave gaps that standard linters won't catch:

- `APP_DEBUG=true` shipping to production
- A controller action with no authorization check
- `$guarded = []` on a model that handles payments
- `DB::raw()` with interpolated user input
- Session cookies without the `Secure` flag
- An API route group missing `auth:sanctum`
- Outdated Composer packages with known CVEs
- Blade templates using `{!! !!}` on user data

Ward checks for all of these and more. It's designed to fit into the workflow you already have — run it locally during development, or wire it into CI to gate deployments.

---

## How It Works

Ward scans your project in a pipeline of five stages:

```
 ✓ Provider  →  ✓ Resolvers  →  ● Scanners  →  ○ Post-Process  →  ○ Report
```

**1. Provider** — Locates and prepares your project source (local path, git clone, or archive).

**2. Resolvers** — Parses your Laravel project and builds a structured context: routes with their middleware, models with `$fillable`/`$guarded`/`$casts`, controllers with authorization calls, config values, `.env` variables, Blade templates, service providers, and more. This happens once and every scanner benefits from it.

**3. Scanners** — Independent security checks run against the resolved context. Each scanner focuses on a category (auth, injection, config, dependencies, etc.) and produces findings. Scanners run in parallel where possible.

**4. Post-Process** — Deduplicates findings, applies your ignore rules, diffs against a previous baseline, and scores results.

**5. Report** — Generates output in your chosen format (terminal, JSON, SARIF, HTML, Markdown) and persists results for trending.

The key insight: scanners don't parse PHP themselves. When the auth scanner wants to know "does this controller method have authorization?", it reads the already-resolved context. The parsing happened upstream, once, and is shared.

---

## What Ward Checks

| Category | Examples |
|----------|---------|
| **Environment & Secrets** | `APP_DEBUG` in production, weak `APP_KEY`, credentials in `.env.example`, exposed API keys |
| **Configuration** | Session cookies missing `Secure`/`HttpOnly`, permissive CORS, weak bcrypt rounds, insecure mail config |
| **Authentication & Authorization** | Controller methods without authorization, missing `auth` middleware on API routes, no rate limiting on login |
| **Mass Assignment** | Models with `$guarded = []`, sensitive fields not in `$hidden`, unprotected `$fillable` |
| **Injection** | SQL injection via `DB::raw()`, command injection via `exec()`/`system()`, open redirects |
| **Cross-Site Scripting** | `{!! !!}` on user data in Blade, unescaped JS injection, missing CSP headers |
| **Dependencies** | Known CVEs in Composer and npm packages, outdated frameworks |
| **Cryptography** | Weak hashing, insecure encryption config, hardcoded secrets |
| **Infrastructure** | Unsafe scheduled tasks, debug routes in production, missing security headers |

---

## Quick Start

### Install

```bash
go install github.com/eljakani/ward@latest
```

Or build from source:

```bash
git clone https://github.com/eljakani/ward.git
cd ward
go build -o ward .
```

### Initialize

```bash
ward init
```

This creates `~/.ward/` with your configuration:

```
~/.ward/
├── config.yaml          # Main configuration
├── rules/               # Custom rule definitions (YAML)
│   └── example.yaml     # Documented example rule
├── reports/             # Scan report output
└── store/               # Result persistence for trending
```

### Scan

```bash
ward scan /path/to/your/laravel-project
```

Ward opens an interactive terminal UI showing real-time progress, scanner status, and findings as they're discovered. When the scan completes, it transitions to a results view where you can browse, sort, and inspect each finding.

---

## Configuration

Ward loads its config from `~/.ward/config.yaml`. Run `ward init` to generate one with documented defaults, or create it manually:

```yaml
# Minimum severity to report: info, low, medium, high, critical
severity: info

output:
  formats:
    - terminal        # interactive TUI
    # - json          # machine-readable
    # - sarif         # IDE/GitHub integration
    # - html          # shareable report
  # dir: ./reports    # where file reports are written

scanners:
  # enable: []        # if empty, all scanners run
  disable: []         # scanner names to skip, e.g. ["dep-scanner"]

rules:
  disable: []         # rule IDs to silence, e.g. ["ENV-001", "AUTH-005"]
  override:           # change severity for specific rules
    # ENV-001:
    #   severity: medium
  # custom_dirs:      # load rules from additional directories
  #   - /path/to/team-rules

ai:
  enabled: false
  provider: openai    # openai, anthropic, ollama
  model: gpt-4o
  # api_key: sk-...   # or set WARD_AI_API_KEY env var
  # endpoint: http://localhost:11434  # for ollama

providers:
  git_depth: 1        # shallow clone depth (0 = full history)
```

### Per-Project Config

You can also place a `.ward.yaml` file in your Laravel project root. Ward merges it on top of the global config, so teams can commit shared settings (disabled rules, severity overrides) alongside their code.

---

## Custom Rules

Drop `.yaml` files into `~/.ward/rules/` and Ward picks them up automatically. Each file contains a list of rules:

```yaml
rules:
  - id: CUSTOM-001
    title: "Hardcoded internal API key"
    description: "Detects hardcoded internal API keys in source files."
    severity: high
    category: secrets
    enabled: true
    tags:
      - secrets
      - cwe-798
    patterns:
      - type: regex
        target: php-files
        pattern: 'INTERNAL_API_KEY\s*=\s*[''"][a-zA-Z0-9]+'
    remediation: |
      Move API keys to environment variables.
      Use .env files or a secrets manager instead of hardcoding keys.
    references:
      - https://cwe.mitre.org/data/definitions/798.html
```

**Pattern types:** `regex`, `contains`, `file-exists`
**Targets:** `php-files`, `blade-files`, `config-files`, `env-files`
**Setting `negative: true`** fires the rule when the pattern is *absent* (useful for "must have X" checks).

---

## Terminal UI

Ward's TUI is built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and adapts to both light and dark terminals automatically.

### Scan View

Displayed while scanning is in progress:

```
┌──────────────────────────────────────────────────────────────────────┐
│  WARD  | Project: myapp | Laravel 11.x | v0.1.0          SCANNING  │
├──────────────────────────────────────────────────────────────────────┤
│     ✓ Provider  →  ✓ Resolvers  →  ● Scanners  →  ○ Post  →  ○ Rpt│
│──────────────────────────────────────────────────────────────────────│
│  Critical: 2    High: 5    Medium: 12    Low: 3    Info: 8          │
│──────────────────────────────────────────────────────────────────────│
│  Scanners              │  Event Log                                  │
│  ✓ env-scanner    (3)  │  15:04:01 ● Stage started: Scanners        │
│  ⠋ auth-scanner        │  15:04:02 ▸ Scanner started: auth-scanner  │
│  ○ injection-scanner   │  15:04:03 ▲ Finding: Weak password policy  │
│  ○ xss-scanner         │  15:04:04 ▲ Finding: Missing CSRF token    │
│  ○ dep-scanner         │  15:04:05 ✓ Scanner completed: csrf-scan   │
├──────────────────────────────────────────────────────────────────────┤
│  ? help  q quit  tab results                                        │
└──────────────────────────────────────────────────────────────────────┘
```

### Results View

Displayed after scan completion:

```
┌──────────────────────────────────────────────────────────────────────┐
│  WARD  | Project: myapp | Laravel 11.x | v0.1.0         COMPLETE   │
├──────────────────────────────────────────────────────────────────────┤
│              Scan Complete — 30 findings in 4.2s                     │
│  Critical: 2    High: 5    Medium: 12    Low: 3    Info: 8          │
│──────────────────────────────────────────────────────────────────────│
│  Sev      Category     Title          │  CRITICAL  Mass assignment  │
│ ─────────────────────────────────────  │  Category: Authorization   │
│ >CRITICAL Auth         Mass assignm.  │  Scanner: auth-scanner     │
│  CRITICAL Injection    SQL injection.  │                             │
│  HIGH     Auth         Missing auth.  │  Description                │
│  HIGH     XSS          Unescaped ou.  │  The Product model uses     │
│  MEDIUM   Config       Debug mode e.  │  $guarded = [] which        │
│  MEDIUM   Config       Session cook.  │  disables mass assignment   │
│  LOW      Deps         Outdated lod.  │  protection entirely...     │
│  INFO     Headers      Missing CSP .  │                             │
│                                       │  Remediation                │
│                                       │  Use $fillable to whitelist │
│                                       │  assignable fields.         │
├──────────────────────────────────────────────────────────────────────┤
│  ? help  q quit  tab panel  s sort  esc back                        │
└──────────────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `?` | Toggle help |
| `Tab` | Switch view or panel |
| `j` / `k` / arrows | Navigate findings |
| `s` | Cycle sort column (severity, category, file) |
| `Esc` | Back to scan view |

---

## Commands

| Command | Description |
|---------|-------------|
| `ward` | Show banner and usage |
| `ward init` | Create `~/.ward/` with default config and example rules |
| `ward init --force` | Recreate config files (overwrites existing) |
| `ward scan <path>` | Scan a Laravel project |
| `ward version` | Print version |

---

## Architecture

```
CLI (cobra)  →  Orchestrator  →  Provider → Resolvers → Scanners → Post-Process → Report
                     ↕                                       ↕
                 EventBus  ←──────────────────────────── findings
                     ↓
                TUI (Bubble Tea)
```

The codebase follows a fully decoupled, plugin-style design:

- **Interface-first** — every major component (Scanner, Provider, Reporter, Store) is a Go interface. Swap any implementation without touching consumers.
- **Event-driven** — scanners emit findings through an event bus. The TUI subscribes to it. Neither knows about the other.
- **Shared context, private logic** — resolvers build a `ProjectContext` once, scanners read it. No scanner parses PHP on its own.
- **Configuration as data** — rules are YAML, not code. Adding a rule never requires recompilation.

## Project Structure

```
ward/
├── main.go
├── cmd/
│   ├── root.go              # Root command, banner, usage
│   ├── init.go              # ward init
│   ├── scan.go              # ward scan <path>
│   └── version.go           # ward version
└── internal/
    ├── config/              # Configuration system
    │   ├── dirs.go          # ~/.ward/ directory management
    │   ├── config.go        # WardConfig struct, Load(), Save()
    │   ├── rules.go         # Rule loading from YAML, overrides
    │   └── init.go          # Scaffold ~/.ward/ with defaults
    ├── models/              # Shared types (contract layer)
    │   ├── severity.go      # Info → Critical enum
    │   ├── finding.go       # Security finding
    │   ├── context.go       # ProjectContext
    │   ├── report.go        # ScanReport
    │   ├── scanner.go       # Scanner interface
    │   └── pipeline.go      # Pipeline stages
    ├── eventbus/            # Decoupled event system
    │   ├── events.go        # 13 event types + payloads
    │   ├── bus.go           # Thread-safe pub/sub
    │   └── bridge.go        # EventBus → Bubble Tea adapter
    └── tui/                 # Terminal UI
        ├── app.go           # Root Bubble Tea model
        ├── messages.go      # Internal TUI messages
        ├── keymap.go        # Key bindings
        ├── banner/
        │   └── banner.go    # ASCII art logo with gradient
        ├── theme/
        │   └── theme.go     # Adaptive color palette + styles
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

---

## Requirements

- Go 1.25+

---

## Roadmap

- [x] Interactive terminal UI with real-time progress
- [x] Event-driven architecture (scanners decoupled from UI)
- [x] Configuration system (`~/.ward/config.yaml`)
- [x] Custom YAML rules (`~/.ward/rules/*.yaml`)
- [ ] Source providers (local, git, archive)
- [ ] Context resolvers (routes, models, controllers, middleware, Blade, config, .env)
- [ ] Scanner implementations (env, config, auth, injection, XSS, deps, crypto, infra)
- [ ] Built-in rule library (100+ rules)
- [ ] Report generation (JSON, SARIF, HTML, Markdown)
- [ ] Result store (SQLite) for trending and baseline diffs
- [ ] Policy engine for CI pass/fail thresholds
- [ ] Per-project `.ward.yaml` config
- [ ] AI-assisted scanning

---

## License

TBD
