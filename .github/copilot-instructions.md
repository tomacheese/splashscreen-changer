# Copilot code review instructions

Guidance for reviewing pull requests in `splashscreen-changer`, a Go CLI that
replaces the EasyAntiCheat splash screen. All source is in
`cmd/splashscreen-changer/`. Focus reviews on the points below.

## Formatting and tooling

- Code must be `go fmt`-clean; CI runs `go fmt ./...` and fails on any diff.
  Flag obviously unformatted code, but do not re-review whitespace CI already
  gates.
- There is no `golangci-lint`. Do not request lint rules or config that the
  repository does not use.

## Error handling

- Fallible functions must return an `error` rather than calling `panic`. In
  `main`, errors are logged with `log.Println` and the function returns; keep
  that pattern. `log.Fatal` is only acceptable for unrecoverable startup
  failures (log-file setup), as in existing code.
- Error strings returned to users must be in English and lowercase-initial
  (e.g. `fmt.Errorf("destination width must be greater than 0")`). Flag new
  error strings written in Japanese or with a capitalized first word.
- Wrap or annotate errors with enough context to locate the failing path; avoid
  discarding errors with `_`.

## Cross-platform build tags

- Platform-specific logic is split by `//go:build windows` / `//go:build
  !windows` (`steam_windows.go` / `steam_other.go`,
  `specialfolder_windows.go` / `specialfolder_other.go`). When a new exported
  function is added to a `_windows.go` file, verify a matching stub exists in
  the corresponding `_other.go` file, otherwise non-Windows builds break.

## Configuration changes

- `Config` in `config.go` is loaded by reflection over struct tags. A new
  config field must carry the appropriate `yaml`, `help`, and (where relevant)
  `default` / `required` tags — otherwise env-var override, default-filling,
  validation, and `-help` output silently skip it. Flag new fields missing
  these tags.
- New user-configurable options should also be reflected in
  `config.yaml.sample` and the `docs/settings/` pages.

## Security

- No secrets, tokens, or credentials in code, logs, or fixtures.
- The tool writes to filesystem paths derived from config/env and decodes PNG
  files from a user-specified directory. Confirm destination paths are
  validated (the `EasyAntiCheat` directory existence check) before writing, and
  that file paths are built with `filepath.Join` rather than string
  concatenation.

## Testing

- New logic in `config.go`, `main.go`, or `versioning.go` should come with
  tests in the matching `*_test.go` file using the standard `testing` package.
  Flag substantive behavior changes that ship without test coverage.

## Do not flag

- Japanese comments and doc comments — this is the project's established
  convention.
- Emoji prefixes in GitHub Actions step names.
- Use of `golang.org/x/exp/rand` and seeding it per run — intentional.
