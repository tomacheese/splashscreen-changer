# Copilot code review instructions

See [AGENTS.md](../AGENTS.md) for the project overview, commands and conventions. Review in English. Focus on the points below.

## Formatting and tooling

- Code must be `go fmt`-clean (CI enforces it). Flag obviously unformatted code, but do not re-review whitespace CI already gates.
- There is no `golangci-lint`. Do not request lint rules or config that the repository does not use.

## Error handling

- Fallible functions must return an `error` rather than calling `panic`. In `main`, errors are logged with `log.Println` and the function returns; `log.Fatal` is acceptable only for unrecoverable startup failures (log-file setup).
- Error strings returned to users must be in English, with a lowercase initial unless it begins with a proper noun (e.g. `EasyAntiCheat folder not found ...`). Flag Japanese or needlessly capitalized error strings.
- Wrap or annotate errors with enough context to locate the failing path; avoid discarding errors with `_`.

## Cross-platform build tags

- Platform-specific logic is split by `//go:build windows` / `!windows` (`steam_*.go`, `specialfolder_*.go`). When an exported function is added to a `_windows.go` file, verify a matching stub exists in the `_other.go` file, otherwise non-Windows builds break.

## Configuration changes

- `Config` in `config.go` is loaded by reflection over struct tags. A new field must carry the appropriate `yaml`, `help`, and (where relevant) `default` / `required` tags, otherwise env override, defaults, validation and `-help` silently skip it.
- New options should also be reflected in `config.yaml.sample` and `docs/settings/`.

## Security

- No secrets, tokens, or credentials in code, logs, or fixtures.
- The tool writes to paths derived from config/env and decodes PNGs from a user-specified directory. Confirm the destination is validated (the `EasyAntiCheat` directory check) before writing, and that paths are built with `filepath.Join`.

## Testing

- New logic in `config.go`, `main.go`, or `versioning.go` should come with tests in the matching `*_test.go` file. Flag substantive behavior changes without coverage.

## Do not flag

- Japanese text in `README-ja.md`, `docs/` and `mkdocs.yml` (localized documentation).
- Emoji prefixes in GitHub Actions step names.
- Use of `golang.org/x/exp/rand` seeded per run — intentional.
