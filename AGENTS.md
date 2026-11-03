# AGENTS.md

## Language

- Deliverables are written in **English**: commit descriptions, PR/Issue titles and bodies, review comments, code comments and doc comments, and documentation.
- Exceptions: localized content (`README-ja.md` and the Japanese documentation site under `docs/` plus the navigation titles in `mkdocs.yml`), user-facing runtime strings that are already localized or parsed (e.g. the `Environment 承認待ち` marker in `build-release.yml`), quotations, technical identifiers, external text, and generated files.
- Error messages returned to users: English.
- Conversation with the maintainer: Japanese, with a half-width space between Japanese and alphanumeric characters.

## Project overview

`splashscreen-changer` is a Go CLI that randomly replaces the EasyAntiCheat splash screen shown at application startup (primarily VRChat, but any EasyAntiCheat app works). On each run it picks a random PNG from a source directory, crops/resizes it to the target aspect ratio, and writes it to `<destination>/EasyAntiCheat/SplashScreen.png`. It ships as a single executable (module `github.com/tomacheese/splashscreen-changer`).

## Development commands

- `go mod download` — install dependencies.
- `go build ./cmd/splashscreen-changer` — build the CLI.
- `go test -v ./...` — run unit tests (CI: `build-release.yml` on pull requests and merge groups, `build.yml` on push to `main`/`master`).
- `go fmt ./...` then `git diff --exit-code` — formatting check. CI fails on any diff, so run `go fmt ./...` before committing.
- `go vet ./...` — not run in CI; available locally (the devcontainer disables the `unsafeptr` analyzer because of the Win32 calls).
- `go run ./cmd/splashscreen-changer -help` / `-version` — show flags and env vars / version.

There is no `golangci-lint` configuration; `go fmt` is the only style gate in CI.

## Architecture

All source lives in `cmd/splashscreen-changer/`:

- `main.go` — entry point: config load, PNG listing, random pick, crop/resize, save.
- `config.go` — `Config` struct and YAML loading. Populated by reflection over struct tags (`yaml`, `help`, `default`, `required`): file, then env-var override (`SECTION_FIELD`, e.g. `SOURCE_PATH`, `DESTINATION_WIDTH`), then defaults, then validation (`checkConfig`).
- `args.go` — `-help` output and source/destination path resolution.
- `log.go` — log file path resolution (`logs/yyyy-mm-dd.log`).
- `versioning.go` — `version`/`date` injected by GoReleaser; falls back to `debug.ReadBuildInfo()`, else `"unknown"`.
- `steam_*.go`, `specialfolder_*.go` — Steam library and Windows known-folder (Pictures) lookup, split by `//go:build windows` / `!windows` (the `_other.go` files are stubs).

Path resolution: (1) env var, (2) config file, (3) auto-detect. Destination auto-detect finds the Steam VRChat install (requires an `EasyAntiCheat` subfolder). Source auto-detect is currently non-functional: `getSourcePath` returns `source.path is required` even when `Pictures/VRChat` exists, so source must be set via env var or config file. Config path: `-config` flag or `CONFIG_PATH`, default `data/config.yml` next to the executable (current directory under `go run`).

Other: `docs/` + `mkdocs.yml` + `requirements.txt` (MkDocs Material site, GitHub Pages), `.goreleaser.yaml` (linux/windows/darwin release builds), `config.yaml.sample`.

## Conventions

- Commits: [Conventional Commits](https://www.conventionalcommits.org/); branches: [Conventional Branch](https://conventional-branch.github.io) short form (`feat`, `fix`, ...). Do not commit to Renovate branches.
- New config option: add `yaml`, `help`, and (where relevant) `default`/`required` tags so loading, env override, defaults, validation and `-help` pick it up; also update `config.yaml.sample` and `docs/settings/`.
- Changing CLI flags or path resolution: update `docs/` and this file. User-facing behavior changes: update `README.md` and `README-ja.md` together.
- Tests use standard `testing`, alongside the code (`config_test.go`, `main_test.go`, `versioning_test.go`).
- Never commit secrets.
- Do not commit build artifacts or local data (`dist/`, `data/`, `logs/`, `.env`, `site/` and the root `splashscreen-changer` binary are gitignored; keep it that way).
- Dependencies are managed by Renovate (`renovate.json`).
