# CLAUDE.md

## Project overview

`splashscreen-changer` is a Go CLI that randomly replaces the EasyAntiCheat
splash screen shown at application startup (primarily VRChat, but any
EasyAntiCheat app works). On each run it picks a random PNG from a source
directory, crops/resizes it to the target aspect ratio, and writes it to
`<destination>/EasyAntiCheat/SplashScreen.png`. It ships as a single
executable (module path `github.com/tomacheese/splashscreen-changer`).

## Development commands

- `go build ./cmd/splashscreen-changer` — build the CLI.
- `go test -v ./...` — run unit tests (run in CI on every push/PR).
- `go mod download` — install dependencies.
- `go fmt ./...` then `git diff --exit-code` — formatting check. CI fails if
  `go fmt` produces any diff, so always run `go fmt ./...` before committing.
- `go run ./cmd/splashscreen-changer -help` — show flags and env vars.
- `go run ./cmd/splashscreen-changer -version` — show version/build date.

There is no `golangci-lint` config; `go fmt` (plus `go vet`) is the only
enforced style gate. The devcontainer sets `go vet` flag `-unsafeptr=false`.

## Architecture / key files

All source lives in `cmd/splashscreen-changer/`:

- `main.go` — entry point: config load, PNG listing, random pick, resize, save.
- `config.go` — `Config` struct + YAML loading. Config is populated by
  reflection over struct tags (`yaml`, `help`, `default`, `required`), then
  overridden by env vars, then defaults applied, then validated (`checkConfig`).
- `args.go` — `-help` output and the source/destination path resolution logic.
- `log.go` — log file path resolution (`logs/yyyy-mm-dd.log`).
- `versioning.go` — `version`/`date` are injected at build time by GoReleaser;
  falls back to `debug.ReadBuildInfo()`, else `"unknown"`.
- `steam_windows.go` / `steam_other.go` — Steam library lookup, split by
  `//go:build windows` / `!windows` build tags. The `_other.go` files are
  non-Windows stubs.
- `specialfolder_windows.go` / `specialfolder_other.go` — Windows known-folder
  (Pictures) lookup, same build-tag split.

Path resolution precedence (see `args.go`): (1) env var, (2) config file, then
(3) auto-detect — Pictures/VRChat for source, Steam library VRChat install for
destination. Config file path itself: `-config` flag or `CONFIG_PATH` env var,
defaulting to `data/config.yml` next to the executable (or the CWD under
`go run`). Config keys map to env vars as `SECTION_FIELD` (e.g. `SOURCE_PATH`,
`DESTINATION_WIDTH`).

Other roots: `docs/` + `mkdocs.yml` + `requirements.txt` (MkDocs Material site,
deployed to GitHub Pages); `.goreleaser.yaml` (release builds for
linux/windows/darwin); `config.yaml.sample` (example config).

## Coding conventions

- Comments: Japanese, matching the existing code (function/type doc comments
  are written in Japanese).
- Error messages and error strings returned to users: English
  (e.g. `fmt.Errorf("source path '%s' does not exist", ...)`).
- Format with `go fmt` — non-negotiable, CI enforces it.
- Config fields are driven by struct tags; when adding a config option, add the
  `yaml`, `help`, and (where relevant) `default`/`required` tags so the
  reflection-based loader, env override, defaults, and `-help` output all pick
  it up automatically.

## Testing

- Standard Go `testing`. Test files: `config_test.go`, `main_test.go`,
  `versioning_test.go`, all in `cmd/splashscreen-changer/`.
- Run with `go test -v ./...`. Add tests alongside the code they cover.

## Repository conventions

- Conversation language: Japanese. Insert a half-width space between Japanese
  and alphanumeric characters.
- Commit messages: [Conventional Commits](https://www.conventionalcommits.org/);
  the `<description>` is written in Japanese (matches the existing human commit
  history; Renovate dependency commits are the exception).
- Branch names: [Conventional Branch](https://conventional-branch.github.io)
  short form (`feat`, `fix`, ...).
- CI (`.github/workflows/`): `build.yml` (test + fmt check on push),
  `build-release.yml` (test + GoReleaser on PR/merge), `doc-build.yml` /
  `doc-deploy.yml` (MkDocs), `add-reviewer.yml`. Dependencies are managed by
  Renovate (`renovate.json`).

## Documentation update rules

- Adding or changing a config option: update the struct tags in `config.go`,
  `config.yaml.sample`, and the relevant page under `docs/settings/`.
- Changing CLI flags or path-resolution behavior: update `docs/` and this file.
- User-facing behavior changes: update `README.md` and `README-ja.md` together.

## Security / prohibitions

- Never commit secrets (tokens, passwords, credentials) or log their values.
- Do not commit build artifacts or local data — `dist/`, `data/`, `logs/`,
  `.env`, and `site/` are gitignored; keep it that way.
