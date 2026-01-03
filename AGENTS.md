# Repository Guidelines

## Project Structure & Module Organization

- `main.go` is the CLI entry point.
- `cmd/` contains Cobra command wiring (see `cmd/root.go`).
- `internal/` holds non-exported packages:
  - `internal/decoder/` is the SQ decode algorithm.
  - `internal/wav/` is WAV file I/O.
- `pkg/sqmath/` provides reusable math helpers (Hilbert transform).
- `docs/` contains design notes and explanations (see `docs/sq-decoder-explained.md`).

## Build, Test, and Development Commands

- `go mod download` to fetch dependencies.
- `go build -o go-sq-decoder` to build the CLI binary.
- `go run ./ --help` to run the CLI without building.
- `go test ./...` to run all tests (currently minimal/none).

## Coding Style & Naming Conventions

- Follow standard Go formatting: run `gofmt -w .` before committing.
- Package names are lowercase and short (e.g., `decoder`, `wav`, `sqmath`).
- Keep exported identifiers in `pkg/` and unexported ones in `internal/`.

## Testing Guidelines

- Use Go’s built-in `testing` package.
- Name tests `*_test.go` with `TestXxx` functions.
- Prefer small, deterministic unit tests around decoding math and WAV I/O.

## Commit & Pull Request Guidelines

- No commit message convention is enforced (no Git history in this repo).
- Suggested format: short, imperative summary (e.g., “Add overlap validation”).
- PRs should include: a concise summary, how you tested (commands), and
  sample usage when changing CLI behavior.

## Configuration & Usage Notes

- Requires Go 1.21+ (`go.mod`).
- The CLI expects a 2-channel input WAV and outputs a 4-channel WAV:
  `go-sq-decoder input.wav output.wav`.
