# Contributing to stacklane

## Prerequisites

- Go 1.22 or newer (matches `.go-version`).
- `make` (any flavour).
- Docker Desktop or Engine for end-to-end testing (not required for `go test -short`).

## Common workflows

```sh
make fmt        # gofmt
make vet        # go vet
make test-short # unit tests, no Docker required
make test       # full test suite (currently equivalent to test-short)
make build      # builds ./stacklane-bin
make tidy       # go mod tidy
```

## Tests and mocks

- All cross-module collaborators are hidden behind interfaces.
- Hand-rolled mocks for those interfaces live in `internal/mocks/`.
- New modules must add their mock to `internal/mocks/mocks.go` so consumers can keep `go test -short ./...` Docker-free.
- Use `t.TempDir()` for any filesystem-touching test.
- Golden files live under `<package>/testdata/`. The gateway template tests write `*.actual` files when output diverges so you can `diff` them before promoting.

## Style

- Run `make fmt` before sending a PR.
- Run `make vet` and `make test-short`.
- Avoid adding fields, options, or helpers that are not actually used by a caller. The Bash reference accumulated configuration knobs that nobody set; do not repeat that in Go.

## Pull requests

Branch naming follows `feature/<short-description>` or `fix/<short-description>`.
Reference the relevant spec under `specs/` in the PR description. The default
branch for the rewrite work is `003-rewrite-language-choices`.
