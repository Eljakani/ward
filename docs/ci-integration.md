# Ward — CI/CD Integration Guide

Integrate Ward into your CI/CD pipeline to catch Laravel security issues before they reach production.

---

## Quick Start

```bash
# Install
go install github.com/eljakani/ward@latest

# Initialize (sets up config and security rules)
ward init

# Scan with CI exit codes
ward scan . --output json,sarif --fail-on high
```

Ward exits with code **1** when findings meet or exceed `--fail-on` threshold. Use this to block merges.

---

## Severity Levels

| Level      | `--fail-on` value | Includes                       |
| ---------- | ----------------- | ------------------------------ |
| `critical` | `critical`        | Critical only                  |
| `high`     | `high`            | High + Critical                |
| `medium`   | `medium`          | Medium + High + Critical       |
| `low`      | `low`             | Low + Medium + High + Critical |
| `info`     | `info`            | All findings                   |

> **Recommendation:** Start with `--fail-on high` to catch serious issues without overwhelming your team.

---

## Baseline Workflow

### Problem

Your first scan will find dozens of existing issues. Without a baseline, your CI will always fail.

### Solution

**1. Generate a baseline** (run once, commit the file):

```bash
ward scan . --output json --update-baseline .ward-baseline.json
git add .ward-baseline.json
git commit -m "chore: add ward security baseline"
```

**2. Use the baseline in CI** (only new findings trigger failure):

```bash
ward scan . --output json --fail-on high --baseline .ward-baseline.json
```

**3. Reduce the baseline over time** — as you fix existing issues, re-generate:

```bash
ward scan . --output json --update-baseline .ward-baseline.json
```

> **Tip:** Review your baseline periodically. Treat it like tech debt — the goal is to shrink it to zero.

---

## GitHub Actions

### Basic Workflow

```yaml
# .github/workflows/ward.yml
name: Ward Security Scan

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  security:
    name: Laravel Security Scan
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Ward
        run: go install github.com/eljakani/ward@latest

      - name: Initialize Ward
        run: ward init

      - name: Run Security Scan
        run: |
          ward scan . --output json,sarif \
            --baseline .ward-baseline.json \
            --fail-on high

      - name: Upload SARIF to GitHub Security
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: ward-report.sarif
```

### With Artifact Upload

```yaml
      - name: Upload Reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: ward-security-reports
          path: |
            ward-report.json
            ward-report.sarif
          retention-days: 30
```

### PR Comment with Results

```yaml
      - name: Comment on PR
        if: failure() && github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const report = JSON.parse(fs.readFileSync('ward-report.json', 'utf8'));
            const findings = report.findings || [];
            const critical = findings.filter(f => f.severity === 'Critical').length;
            const high = findings.filter(f => f.severity === 'High').length;

            let body = `## 🛡 Ward Security Scan Failed\n\n`;
            body += `| Severity | Count |\n|----------|-------|\n`;
            if (critical) body += `| 🔴 Critical | ${critical} |\n`;
            if (high) body += `| 🟠 High | ${high} |\n`;
            body += `\nSee the Actions tab for full details.`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body
            });
```

---

## GitLab CI

```yaml
# .gitlab-ci.yml
ward-security:
  stage: test
  image: golang:1.24-alpine
  before_script:
    - go install github.com/eljakani/ward@latest
    - ward init
  script:
    - ward scan . --output json,sarif --baseline .ward-baseline.json --fail-on high
  artifacts:
    paths:
      - ward-report.sarif
      - ward-report.json
    when: always
    expire_in: 30 days
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'
```

---

## Bitbucket Pipelines

```yaml
# bitbucket-pipelines.yml
pipelines:
  default:
    - step:
        name: Ward Security Scan
        image: golang:1.24-alpine
        script:
          - go install github.com/eljakani/ward@latest
          - ward init
          - ward scan . --output json --baseline .ward-baseline.json --fail-on high
        artifacts:
          - ward-report.json
```

---

## Azure DevOps

```yaml
# azure-pipelines.yml
trigger:
  branches:
    include: [main]

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: GoTool@0
    inputs:
      version: '1.24'

  - script: |
      go install github.com/eljakani/ward@latest
      ward init
      ward scan . --output json,sarif --baseline .ward-baseline.json --fail-on high
    displayName: 'Ward Security Scan'

  - task: PublishBuildArtifacts@1
    condition: always()
    inputs:
      pathToPublish: 'ward-report.json'
      artifactName: 'security-reports'
```

---

## Docker-Based Scanning

If your CI doesn't have Go available, use a multi-stage approach:

```dockerfile
# Dockerfile.ward
```dockerfile
# Use golang:1.24-alpine (or later) to meet the build requirements for Ward
FROM golang:1.24-alpine
RUN apk add --no-cache git
RUN go install github.com/eljakani/ward@latest && ward init
WORKDIR /app
ENTRYPOINT ["ward"]
CMD ["--help"]
```

Build and run against a URL (saving reports to current directory):
```bash
docker build -t ward-scanner .
# Run scan and mount current dir to /reports to get the output files
docker run --rm -v $(pwd):/reports ward-scanner scan https://github.com/username/repo.git --output json,sarif --output-dir /reports
```

### Option 2: CI Pipeline Scanner (Scan the Code Inside)

If you want to scan the code *during the build process* (e.g. in a CI pipeline):

```dockerfile
FROM golang:1.24-alpine AS scanner
RUN go install github.com/eljakani/ward@latest && ward init
WORKDIR /app
COPY . .
RUN ward scan . --output json,sarif --fail-on high
```

Or as a single script in any CI:

```bash
#!/bin/bash
# scripts/security-scan.sh
set -e

if ! command -v ward &> /dev/null; then
  echo "Installing Ward..."
  go install github.com/eljakani/ward@latest
  ward init
fi

ARGS="--output json,sarif"
[ -f .ward-baseline.json ] && ARGS="$ARGS --baseline .ward-baseline.json"
[ -n "$WARD_FAIL_ON" ] && ARGS="$ARGS --fail-on $WARD_FAIL_ON"

ward scan . $ARGS
```

---

## Configuration Tips

### Recommended `.ward-baseline.json` Policy

| Stage                | `--fail-on` | Baseline                            |
| -------------------- | ----------- | ----------------------------------- |
| **New project**      | `high`      | None — fix everything from day 1    |
| **Existing project** | `high`      | Generate baseline, reduce over time |
| **Strict mode**      | `low`       | None — zero tolerance               |
| **Advisory only**    | _(omit)_    | None — scan without blocking        |

### Output Formats for CI

| Format     | Use case                                   |
| ---------- | ------------------------------------------ |
| `json`     | Machine-readable, parse in scripts         |
| `sarif`    | GitHub/GitLab Security dashboards          |
| `markdown` | Paste into PR comments or Slack            |
| `html`     | Attach as build artifact for manual review |

```bash
# Generate multiple formats at once
ward scan . --output json,sarif,markdown --fail-on high
```

### Caching Ward Between Runs

Speed up CI by caching the Go binary and Ward config:

**GitHub Actions:**
```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/go/bin/ward
      ~/.ward
    key: ward-${{ runner.os }}
```

**GitLab CI:**
```yaml
cache:
  paths:
    - /go/bin/ward
    - ~/.ward/
```

---

## Exit Codes

| Code | Meaning                                                             |
| ---- | ------------------------------------------------------------------- |
| `0`  | Scan completed, no findings above threshold (or no `--fail-on` set) |
| `1`  | Findings exceed `--fail-on` threshold                               |
| `1`  | Configuration error (missing config, invalid path, etc.)            |

---

## Troubleshooting

### "ward: command not found"
Ensure `$GOPATH/bin` is in your `PATH`:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Baseline file not working
- The baseline uses fingerprints (hash of rule ID + file + line number)
- If you rename a file or a finding moves to a different line, it won't match the baseline
- Re-generate the baseline when making large refactors: `ward scan . --output json --update-baseline .ward-baseline.json`

### Too many findings on first run
1. Generate a baseline: `ward scan . --output json --update-baseline .ward-baseline.json`
2. Commit it: `git add .ward-baseline.json && git commit -m "chore: ward baseline"`
3. Fix findings over time, periodically regenerating the baseline

### SARIF not showing in GitHub Security tab
- Make sure you're using `github/codeql-action/upload-sarif@v3`
- The repository must have GitHub Advanced Security enabled (free for public repos)
- The SARIF file must be under 10MB
