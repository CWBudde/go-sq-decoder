# Copilot Instructions (go-sq-tool)

## Big picture

- This repo is a Go CLI that decodes **SQ (Stereo Quadraphonic)** matrix-encoded stereo WAV (LT/RT) into **4-channel** quad WAV (LF/RF/LB/RB).
- Data flow: CLI → WAV read → SQ² decoder (block/overlap) → WAV write.
  - Entry point: `cmd.Execute()` from main ([main.go](../main.go)).
  - CLI command: [cmd/root.go](../cmd/root.go) (Cobra).
  - WAV I/O: [internal/wav/wav.go](../internal/wav/wav.go).
  - Decoder core: [internal/decoder/decoder.go](../internal/decoder/decoder.go).
  - Hilbert (90° phase shift): [pkg/sqmath/hilbert.go](../pkg/sqmath/hilbert.go) using `github.com/MeKo-Christian/algo-fft` (plan-based API).

## Key DSP conventions in this codebase

- Audio buffers are `[][]float64` shaped as **[channel][sample]**.
  - Input must be exactly 2 channels (LT, RT). Output must be exactly 4 channels in this order: **LF, RF, LB, RB**.
- Samples are normalized floats (roughly in [-1, 1]). When writing WAV, values are clamped to [-1, 1] and written as **16-bit PCM**.
- SQ decode matrix implemented in [internal/decoder/decoder.go](../internal/decoder/decoder.go):
  - `LF = LT`
  - `RF = RT`
  - `LB = (√2/2)·H(LT) - (√2/2)·RT`
  - `RB = (√2/2)·LT - (√2/2)·H(RT)`
- Block processing details matter:
  - Defaults: `blockSize=1024`, `overlap=512`.
  - The decoder uses indexing offsets (`inputOffset := overlap/4`, `outputOffset := overlap/2`) to match the SQ² implementation; preserve this behavior when refactoring.
  - `HilbertTransformer.ProcessBlock` expects a slice of length `blockSize` and **panics** otherwise.

## Developer workflows

- Format: `gofmt -w .`
- Build: `go build ./...`
- Run (local): `go run . --help` or `go run . -v input.wav output.wav`
- Dependencies: `go mod tidy` (generates/updates `go.sum`)
- Typical tuning flags: `-b/--block-size` (power of 2), `-o/--overlap` (often blockSize/2).
- Tests: none in the repo currently; if you add tests, start with deterministic unit tests around the decode matrix and block offset behavior.

## Repo-specific contribution patterns

- CLI surface is implemented with Cobra in [cmd/root.go](../cmd/root.go). Add flags/subcommands there.
- Keep audio I/O inside `internal/wav` (stays internal to the module). Keep DSP primitives in `pkg/sqmath`.
- Prefer returning errors with context (this repo wraps errors with `%w` in CLI and WAV I/O).

## Output format notes

- Current WAV writer outputs **16-bit PCM only** (via `github.com/youpy/go-wav`).
- Internal processing uses float64; samples are clamped to [-1, 1] before writing.
- **Future**: float32 WAV output planned—see [docs/float32-wav-libraries.md](../docs/float32-wav-libraries.md) for library research.

## References

- Technical background and the rationale for SQ²/Hilbert: [docs/sq-decoder-explained.md](../docs/sq-decoder-explained.md).
- Float32 WAV library options: [docs/float32-wav-libraries.md](../docs/float32-wav-libraries.md).
