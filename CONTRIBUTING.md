# Contributing to Ward

Thanks for taking the time to contribute. Ward is a security tool for Laravel developers, and every bug report, fix, and new rule makes it more useful for the community.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Writing Rules](#writing-rules)
- [Submitting a Pull Request](#submitting-a-pull-request)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold it.

---

## Reporting Bugs

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include:

- Ward version (`ward version`)
- How you installed Ward
- The command you ran and its full output
- What you expected vs what happened

For **false positives** (Ward flags something it shouldn't), include the relevant code snippet so the rule can be fixed.

For **security vulnerabilities in Ward itself**, see [SECURITY.md](SECURITY.md) — do not open a public issue.

---

## Suggesting Features

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md). Be specific about the problem you're trying to solve and why the current behaviour falls short.

---

## Development Setup

**Requirements:** Go 1.24+

```bash
git clone https://github.com/eljakani/ward.git
cd ward
go build ./...
go test ./...
```

Common make targets:

| Command        | Description                  |
| -------------- | ---------------------------- |
| `make build`   | Compile `ward` binary        |
| `make install` | Install to `$GOPATH/bin`     |
| `make test`    | Run all tests                |
| `make lint`    | Run `go vet`                 |
| `make clean`   | Remove build artifacts       |

---

## Making Changes

1. Fork the repo and create a branch from `main`:
   ```bash
   git checkout -b fix/my-fix
   ```
2. Make your changes.
3. Add or update tests to cover your change.
4. Ensure all tests pass:
   ```bash
   go test ./...
   ```
5. Ensure `go vet ./...` reports no issues.
6. Commit with a clear message following the convention:
   ```
   type: short description

   # types: feat, fix, docs, refactor, test, chore
   ```

---

## Writing Rules

Ward's built-in rules live in `internal/config/defaults/rules/`. Custom rules go in `~/.ward/rules/`.

See `internal/config/defaults/rules/custom-example.yaml` for a fully documented template.

**Pattern types available:**

| Type           | Use when                                                              |
| -------------- | --------------------------------------------------------------------- |
| `regex`        | Simple line-by-line pattern match                                     |
| `contains`     | Exact substring match                                                 |
| `file-exists`  | Checking for the presence or absence of a file                        |
| `regex-scoped` | Match that should be suppressed inside a scope block (e.g. middleware group) |

When contributing a new built-in rule:

- Assign an ID in the appropriate category prefix (`AUTH-`, `INJECT-`, `XSS-`, etc.)
- Set a realistic severity — avoid over-flagging with `critical` or `high`
- Include a `remediation` block with a concrete code example
- Add at least one `references` link (CWE, OWASP, or Laravel docs)
- Test against both a vulnerable snippet and a safe equivalent to confirm no false positives

---

## Submitting a Pull Request

- Fill in the pull request template completely
- Link the issue your PR addresses (`Closes #123`)
- Keep PRs focused — one concern per PR
- If your PR changes behaviour, update the relevant section of `README.md`
- All CI checks must pass before review

A maintainer will review your PR as soon as possible. Thank you.
