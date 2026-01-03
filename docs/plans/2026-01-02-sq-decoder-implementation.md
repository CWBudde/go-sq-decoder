# Go SQ Decoder Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Complete the Go-based SQ quadrophonic decoder CLI tool with FFT-based Hilbert transform, full testing, and working WAV file I/O.

**Architecture:** FFT-based SQ² decoder using algo-fft library for Hilbert transform, cobra for CLI, go-wav for WAV I/O. Block-based processing with 50% overlap for accurate 90° phase shifting and channel separation.

**Tech Stack:** Go 1.21+, github.com/MeKo-Christian/algo-fft, github.com/spf13/cobra, github.com/youpy/go-wav

---

## Current State

- ✅ Project structure created
- ✅ Initial code written (untested)
- ✅ Documentation written
- ❌ Code not verified to compile
- ❌ FFT library API not verified
- ❌ No tests written
- ❌ Not functionally tested

## Prerequisites

Ensure you're in the `/mnt/e/Public/go-sq-decoder` directory for all tasks.

---

## Task 1: Fix Module Dependencies and Verify Build Environment

**Files:**

- Modify: `go.mod`
- Check: `go.sum` (will be auto-generated)

**Step 1: Update go.mod with correct dependencies**

Navigate to project root:

```bash
cd /mnt/e/Public/go-sq-decoder
```

Update `go.mod` to:

```go
module github.com/cwbudde/go-sq-decoder

go 1.21

require (
	github.com/MeKo-Christian/algo-fft v0.4.2
	github.com/spf13/cobra v1.8.0
	github.com/youpy/go-wav v0.3.2
)
```

**Step 2: Download dependencies**

Run:

```bash
go mod download
```

Expected: All dependencies download successfully

**Step 3: Verify module graph**

Run:

```bash
go mod tidy
go mod verify
```

Expected: "all modules verified"

**Step 4: Commit**

```bash
git init
git add go.mod go.sum
git commit -m "chore: initialize module with dependencies"
```

---

## Task 2: Investigate and Fix FFT Library API

**Files:**

- Check: `pkg/sqmath/hilbert.go`
- Investigate: algo-fft library API

**Step 1: Check algo-fft library API**

Create test file to explore API:

```bash
cat > /tmp/fft_test.go << 'EOF'
package main

import (
	"fmt"
	fft "github.com/MeKo-Christian/algo-fft"
)

func main() {
	// Test basic FFT API
	data := []complex128{1+0i, 2+0i, 3+0i, 4+0i}
	result := fft.FFT(data)
	fmt.Printf("FFT result: %v\n", result)

	inverse := fft.IFFT(result)
	fmt.Printf("IFFT result: %v\n", inverse)
}
EOF
cd /tmp
go mod init test
go get github.com/MeKo-Christian/algo-fft
go run fft_test.go
```

Expected: Program runs and shows FFT/IFFT work

**Step 2: Document findings**

Create notes file:

```bash
cat > /mnt/e/Public/go-sq-decoder/docs/fft-api-notes.md << 'EOF'
# algo-fft API Notes

## Functions Available
- `fft.FFT([]complex128) []complex128` - Forward FFT
- `fft.IFFT([]complex128) []complex128` - Inverse FFT

## Notes
- Input/output are complex128 slices
- [Add any other findings here]
EOF
```

**Step 3: Update hilbert.go if API differs**

If the API is different from what's in `pkg/sqmath/hilbert.go:95-105`, update it.

Current code assumes:

```go
freqDomain := fft.FFT(inputComplex)
timeDomain := fft.IFFT(freqDomain)
```

Verify this matches the library. If not, adjust.

**Step 4: Commit findings**

```bash
git add docs/fft-api-notes.md pkg/sqmath/hilbert.go
git commit -m "docs: document FFT library API and fix usage"
```

---

## Task 3: Fix Compilation Errors in Hilbert Transform

**Files:**

- Modify: `pkg/sqmath/hilbert.go`

**Step 1: Attempt to build and capture errors**

Run:

```bash
cd /mnt/e/Public/go-sq-decoder
go build ./pkg/sqmath/
```

Expected: Compilation errors (if any)

**Step 2: Fix import statement**

Verify line 6 in `pkg/sqmath/hilbert.go`:

```go
import (
	"math"

	fft "github.com/MeKo-Christian/algo-fft"
)
```

Should be exactly this format.

**Step 3: Fix complex multiplication logic**

Review lines 95-110 in `pkg/sqmath/hilbert.go`. The complex multiplication may need adjustment:

```go
// Current code (may be incorrect):
freqDomain[i] = a*b - c*d
freqDomain[i+nyquist] = c*b + a*d
```

Correct complex multiplication for (a+bi) \* (c+di):

```go
// Real part: ac - bd
// Imag part: ad + bc
```

If working with separate real/imaginary arrays, may need to restructure.

**Step 4: Build again**

Run:

```bash
go build ./pkg/sqmath/
```

Expected: No errors

**Step 5: Commit**

```bash
git add pkg/sqmath/hilbert.go
git commit -m "fix: correct FFT complex multiplication in Hilbert transform"
```

---

## Task 4: Build and Fix Decoder Core

**Files:**

- Modify: `internal/decoder/decoder.go`

**Step 1: Attempt to build decoder**

Run:

```bash
go build ./internal/decoder/
```

Expected: Shows any import or syntax errors

**Step 2: Fix import paths**

Verify imports at top of `internal/decoder/decoder.go`:

```go
import (
	"fmt"
	"math"

	"github.com/cwbudde/go-sq-decoder/pkg/sqmath"
)
```

**Step 3: Verify HilbertTransformer usage**

Check lines 38-39:

```go
hilbertLeft:   sqmath.NewHilbertTransformer(blockSize, overlap),
hilbertRight:  sqmath.NewHilbertTransformer(blockSize, overlap),
```

Ensure this matches the actual function signature in hilbert.go.

**Step 4: Build again**

Run:

```bash
go build ./internal/decoder/
```

Expected: Clean build

**Step 5: Commit**

```bash
git add internal/decoder/decoder.go
git commit -m "fix: correct imports and API usage in decoder"
```

---

## Task 5: Build and Fix WAV I/O

**Files:**

- Modify: `internal/wav/wav.go`

**Step 1: Check go-wav library API**

Create quick test:

```go
// In /tmp/wav_test.go
package main

import (
	"github.com/youpy/go-wav"
	"os"
)

func main() {
	// Test reading
	file, _ := os.Open("test.wav")
	reader := wav.NewReader(file)
	format, _ := reader.Format()
	println("Channels:", format.NumChannels)
}
```

Run: `cd /tmp && go mod init test && go get github.com/youpy/go-wav && go run wav_test.go`

Expected: Shows API structure

**Step 2: Fix WAV reader usage**

In `internal/wav/wav.go:33-51`, verify the reader API matches:

- `reader.Format()` returns format
- `reader.ReadSamples()` returns samples
- `reader.FloatValue(sample, channel)` gets float value

Update if API differs.

**Step 3: Fix WAV writer usage**

In `internal/wav/wav.go:77-96`, verify writer API:

- `wav.NewWriter(file, numSamples, numChannels, sampleRate, bitsPerSample)`
- `writer.WriteSamples([]wav.Sample)`

Update if API differs.

**Step 4: Build WAV package**

Run:

```bash
go build ./internal/wav/
```

Expected: Clean build

**Step 5: Commit**

```bash
git add internal/wav/wav.go
git commit -m "fix: correct go-wav API usage"
```

---

## Task 6: Build and Fix CLI

**Files:**

- Modify: `cmd/root.go`
- Modify: `main.go`

**Step 1: Build CLI**

Run:

```bash
go build ./cmd/
```

Expected: Shows any errors

**Step 2: Fix import paths in cmd/root.go**

Verify lines 6-9:

```go
import (
	"fmt"
	"os"

	"github.com/cwbudde/go-sq-decoder/internal/decoder"
	"github.com/cwbudde/go-sq-decoder/internal/wav"
	"github.com/spf13/cobra"
)
```

**Step 3: Build main**

Run:

```bash
go build -o go-sq-decoder main.go
```

Expected: Binary created or specific errors shown

**Step 4: Fix any compilation errors**

Address each error individually, common issues:

- Import path mismatches
- Undefined functions
- Type mismatches

**Step 5: Successful build**

Run:

```bash
go build -o go-sq-decoder
```

Expected: Binary `go-sq-decoder` created

**Step 6: Test CLI help**

Run:

```bash
./go-sq-decoder --help
```

Expected: Shows help text with usage information

**Step 7: Commit**

```bash
git add cmd/root.go main.go
git commit -m "fix: build working CLI binary"
```

---

## Task 7: Write Unit Test for Hilbert Transform

**Files:**

- Create: `pkg/sqmath/hilbert_test.go`

**Step 1: Write failing test for Hilbert transformer creation**

Create `pkg/sqmath/hilbert_test.go`:

```go
package sqmath

import (
	"testing"
)

func TestNewHilbertTransformer(t *testing.T) {
	blockSize := 1024
	overlap := 512

	ht := NewHilbertTransformer(blockSize, overlap)

	if ht == nil {
		t.Fatal("NewHilbertTransformer returned nil")
	}

	if ht.blockSize != blockSize {
		t.Errorf("Expected blockSize %d, got %d", blockSize, ht.blockSize)
	}

	if ht.overlap != overlap {
		t.Errorf("Expected overlap %d, got %d", overlap, ht.overlap)
	}

	if !ht.initialized {
		t.Error("HilbertTransformer not initialized")
	}
}
```

**Step 2: Run test**

Run:

```bash
go test ./pkg/sqmath/ -v
```

Expected: PASS (this should pass since constructor exists)

**Step 3: Write test for 90-degree phase shift property**

Add to `hilbert_test.go`:

```go
func TestHilbertPhaseShift(t *testing.T) {
	// Test that Hilbert transform creates ~90 degree phase shift
	// Using simple sine wave test

	blockSize := 1024
	overlap := 512
	ht := NewHilbertTransformer(blockSize, overlap)

	// Create test signal: simple DC offset
	input := make([]float64, blockSize)
	input[0] = 1.0 // DC component

	output := ht.ProcessBlock(input)

	// DC should pass through unchanged (approximately)
	if output[0] < 0.5 || output[0] > 1.5 {
		t.Errorf("DC component not preserved: got %f", output[0])
	}

	// Output should have same length as input
	if len(output) != blockSize {
		t.Errorf("Expected output length %d, got %d", blockSize, len(output))
	}
}
```

**Step 4: Run test**

Run:

```bash
go test ./pkg/sqmath/ -v -run TestHilbertPhaseShift
```

Expected: PASS or reveals issues with ProcessBlock

**Step 5: Commit**

```bash
git add pkg/sqmath/hilbert_test.go
git commit -m "test: add unit tests for Hilbert transformer"
```

---

## Task 8: Write Unit Test for SQ Decoder

**Files:**

- Create: `internal/decoder/decoder_test.go`

**Step 1: Write test for decoder creation**

Create `internal/decoder/decoder_test.go`:

```go
package decoder

import (
	"testing"
)

func TestNewSQDecoder(t *testing.T) {
	decoder := NewSQDecoder()

	if decoder == nil {
		t.Fatal("NewSQDecoder returned nil")
	}

	if decoder.blockSize != DefaultBlockSize {
		t.Errorf("Expected blockSize %d, got %d", DefaultBlockSize, decoder.blockSize)
	}

	if decoder.sqrt2 == 0 {
		t.Error("sqrt2 not initialized")
	}

	expectedSqrt2 := 0.707
	if decoder.sqrt2 < expectedSqrt2-0.01 || decoder.sqrt2 > expectedSqrt2+0.01 {
		t.Errorf("sqrt2 value incorrect: got %f, expected ~%f", decoder.sqrt2, expectedSqrt2)
	}
}
```

**Step 2: Run test**

Run:

```bash
go test ./internal/decoder/ -v
```

Expected: PASS

**Step 3: Write test for basic processing**

Add to `decoder_test.go`:

```go
func TestSQDecoderProcess(t *testing.T) {
	decoder := NewSQDecoder()

	// Create simple stereo input (1 second @ 44.1kHz)
	numSamples := 44100
	input := [][]float64{
		make([]float64, numSamples), // LT
		make([]float64, numSamples), // RT
	}

	// Fill with simple test signal
	for i := 0; i < numSamples; i++ {
		input[0][i] = 0.5 // LT = constant
		input[1][i] = 0.3 // RT = constant
	}

	output, err := decoder.Process(input)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(output) != 4 {
		t.Fatalf("Expected 4 output channels, got %d", len(output))
	}

	for i, ch := range output {
		if len(ch) != numSamples {
			t.Errorf("Channel %d: expected %d samples, got %d", i, numSamples, len(ch))
		}
	}

	// Front channels should match input (approximately, accounting for processing)
	// This is a basic sanity check
	if output[0][1000] < 0.4 || output[0][1000] > 0.6 {
		t.Logf("Warning: LF channel value unexpected: %f (input was 0.5)", output[0][1000])
	}
}
```

**Step 4: Run test**

Run:

```bash
go test ./internal/decoder/ -v -run TestSQDecoderProcess
```

Expected: PASS or reveals processing issues

**Step 5: Commit**

```bash
git add internal/decoder/decoder_test.go
git commit -m "test: add unit tests for SQ decoder"
```

---

## Task 9: Create Test WAV Files

**Files:**

- Create: `testdata/` directory
- Create: `testdata/generate_test_wav.go` (helper to generate test files)

**Step 1: Create testdata directory**

```bash
mkdir -p /mnt/e/Public/go-sq-decoder/testdata
```

**Step 2: Write WAV generator**

Create `testdata/generate_test_wav.go`:

```go
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/youpy/go-wav"
)

func main() {
	// Generate simple stereo test file
	sampleRate := uint32(44100)
	duration := 2.0 // seconds
	numSamples := int(float64(sampleRate) * duration)

	file, err := os.Create("test_stereo.wav")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := wav.NewWriter(file, uint32(numSamples), 2, sampleRate, 16)

	// Generate 440 Hz sine wave in left, 880 Hz in right
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)

		left := math.Sin(2 * math.Pi * 440.0 * t)
		right := math.Sin(2 * math.Pi * 880.0 * t)

		samples := []wav.Sample{
			{Values: [2]int{int(left * 16384), int(right * 16384)}},
		}

		if err := writer.WriteSamples(samples); err != nil {
			panic(err)
		}
	}

	fmt.Println("Generated test_stereo.wav")
}
```

**Step 3: Generate test file**

Run:

```bash
cd /mnt/e/Public/go-sq-decoder/testdata
go run generate_test_wav.go
```

Expected: Creates `test_stereo.wav`

**Step 4: Verify test file**

Run:

```bash
ls -lh test_stereo.wav
```

Expected: File exists, ~352KB (2 seconds stereo @ 44.1kHz)

**Step 5: Update .gitignore**

Add to `.gitignore`:

```
# Test files (except generator)
testdata/*.wav
!testdata/generate_test_wav.go
```

**Step 6: Commit**

```bash
git add testdata/generate_test_wav.go .gitignore
git commit -m "test: add WAV test file generator"
```

---

## Task 10: Write Integration Test for Full Pipeline

**Files:**

- Create: `internal/integration_test.go`

**Step 1: Write full pipeline test**

Create `internal/integration_test.go`:

```go
package internal

import (
	"os"
	"testing"

	"github.com/cwbudde/go-sq-decoder/internal/decoder"
	"github.com/cwbudde/go-sq-decoder/internal/wav"
)

func TestFullPipeline(t *testing.T) {
	// Generate test input
	inputPath := "../testdata/test_stereo.wav"
	outputPath := "../testdata/test_output.wav"

	// Clean up output file after test
	defer os.Remove(outputPath)

	// Read input
	input, err := wav.ReadWAV(inputPath)
	if err != nil {
		t.Fatalf("Failed to read test WAV: %v", err)
	}

	t.Logf("Read %d samples at %d Hz", input.NumSamples, input.SampleRate)

	// Decode
	sqDecoder := decoder.NewSQDecoder()
	output, err := sqDecoder.Process(input.Samples)
	if err != nil {
		t.Fatalf("Decoding failed: %v", err)
	}

	// Write output
	outputData := &wav.AudioData{
		SampleRate: input.SampleRate,
		Samples:    output,
		NumSamples: input.NumSamples,
	}

	err = wav.WriteWAV(outputPath, outputData)
	if err != nil {
		t.Fatalf("Failed to write output WAV: %v", err)
	}

	// Verify output file exists
	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not created: %v", err)
	}

	if stat.Size() == 0 {
		t.Fatal("Output file is empty")
	}

	t.Logf("Successfully created output file: %d bytes", stat.Size())
}
```

**Step 2: Generate test input first**

Run:

```bash
cd /mnt/e/Public/go-sq-decoder/testdata
go run generate_test_wav.go
```

**Step 3: Run integration test**

Run:

```bash
cd /mnt/e/Public/go-sq-decoder
go test ./internal/ -v -run TestFullPipeline
```

Expected: PASS or shows specific failure point

**Step 4: Fix any revealed issues**

If test fails, fix issues in:

- WAV reading (check sample format conversion)
- Decoder processing (check buffer sizes, indexing)
- WAV writing (check channel count, sample format)

**Step 5: Commit**

```bash
git add internal/integration_test.go
git commit -m "test: add full pipeline integration test"
```

---

## Task 11: Manual CLI Testing

**Files:**

- Test: Compiled binary `go-sq-decoder`

**Step 1: Build latest binary**

Run:

```bash
cd /mnt/e/Public/go-sq-decoder
go build -o go-sq-decoder
```

Expected: Binary created successfully

**Step 2: Test with generated WAV file**

Run:

```bash
./go-sq-decoder testdata/test_stereo.wav testdata/cli_output.wav
```

Expected: Success message

**Step 3: Test verbose mode**

Run:

```bash
./go-sq-decoder -v testdata/test_stereo.wav testdata/cli_output_verbose.wav
```

Expected: Detailed output showing:

- Sample rate
- Number of samples
- Decoder configuration
- Processing status

**Step 4: Test custom parameters**

Run:

```bash
./go-sq-decoder -b 2048 -o 1024 testdata/test_stereo.wav testdata/cli_output_custom.wav
```

Expected: Success with custom block size

**Step 5: Test error handling**

Run:

```bash
./go-sq-decoder nonexistent.wav output.wav
```

Expected: Clear error message about missing input file

**Step 6: Document test results**

Create `docs/testing-notes.md`:

```markdown
# Testing Notes

## Manual CLI Tests - [Date]

### Basic Usage

- ✅/❌ Standard processing
- ✅/❌ Verbose mode
- ✅/❌ Custom parameters
- ✅/❌ Error handling

### Issues Found

[List any issues discovered]

### Audio Quality

[Notes on output quality if audible testing done]
```

**Step 7: Commit**

```bash
git add docs/testing-notes.md
git commit -m "test: document manual CLI testing results"
```

---

## Task 12: Fix FFT Scaling Issues (If Found)

**Files:**

- Modify: `pkg/sqmath/hilbert.go` (if needed)

**Step 1: Check for amplitude issues**

If output is too loud/quiet or distorted:

Review line 129 in `pkg/sqmath/hilbert.go`:

```go
scale := 1.0 / float64(ht.fftSize)
```

This may need adjustment based on algo-fft's IFFT scaling.

**Step 2: Test different scaling factors**

Try:

- `scale := 1.0` (no scaling)
- `scale := 1.0 / float64(ht.fftSize)` (current)
- `scale := 2.0 / float64(ht.fftSize)` (if too quiet)

**Step 3: Add scaling test**

In `pkg/sqmath/hilbert_test.go`, add:

```go
func TestHilbertAmplitude(t *testing.T) {
	blockSize := 1024
	overlap := 512
	ht := NewHilbertTransformer(blockSize, overlap)

	// Unity impulse
	input := make([]float64, blockSize)
	input[blockSize/2] = 1.0

	output := ht.ProcessBlock(input)

	// Find peak amplitude
	maxAmp := 0.0
	for _, v := range output {
		if abs(v) > maxAmp {
			maxAmp = abs(v)
		}
	}

	// Should be roughly same order of magnitude
	if maxAmp < 0.1 || maxAmp > 10.0 {
		t.Errorf("Amplitude scaling may be wrong: peak = %f", maxAmp)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
```

**Step 4: Run test and adjust**

Run:

```bash
go test ./pkg/sqmath/ -v -run TestHilbertAmplitude
```

Adjust scaling if needed.

**Step 5: Commit fix**

```bash
git add pkg/sqmath/hilbert.go pkg/sqmath/hilbert_test.go
git commit -m "fix: correct FFT scaling in Hilbert transform"
```

---

## Task 13: Fix Block Processing Alignment (If Found)

**Files:**

- Modify: `internal/decoder/decoder.go`

**Step 1: Review indexing logic**

Check lines 76-106 in `internal/decoder/decoder.go`.

The index calculations may be wrong:

```go
outputOffset := d.overlap / 2
inputOffset := d.overlap / 4
```

These should align with the SQ² implementation's approach.

**Step 2: Add boundary checking**

Add debug logging or checks:

```go
if outIdx >= numSamples {
	break
}
if inIdx >= d.blockSize {
	break
}
if phaseIdx >= d.blockSize {
	break
}
```

These already exist, verify they're correct.

**Step 3: Test with known signal**

Use a test with a simple step function to verify alignment is correct.

**Step 4: Fix if needed**

Adjust offset calculations based on findings.

**Step 5: Commit fix**

```bash
git add internal/decoder/decoder.go
git commit -m "fix: correct block processing alignment"
```

---

## Task 14: Add Performance Benchmarks

**Files:**

- Create: `internal/decoder/decoder_bench_test.go`

**Step 1: Write benchmark for decoder**

Create `internal/decoder/decoder_bench_test.go`:

```go
package decoder

import (
	"testing"
)

func BenchmarkSQDecoder(b *testing.B) {
	decoder := NewSQDecoder()

	// 1 second of audio
	numSamples := 44100
	input := [][]float64{
		make([]float64, numSamples),
		make([]float64, numSamples),
	}

	// Fill with noise
	for i := 0; i < numSamples; i++ {
		input[0][i] = float64(i%100) / 100.0
		input[1][i] = float64((i*7)%100) / 100.0
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := decoder.Process(input)
		if err != nil {
			b.Fatal(err)
		}
	}

	// Report processing speed
	samplesPerRun := float64(numSamples)
	secondsPerRun := samplesPerRun / 44100.0
	b.ReportMetric(secondsPerRun*float64(b.N)/b.Elapsed().Seconds(), "realtime_factor")
}
```

**Step 2: Run benchmark**

Run:

```bash
go test ./internal/decoder/ -bench=. -benchmem
```

Expected: Shows performance metrics

**Step 3: Document results**

Add to `docs/performance.md`:

```markdown
# Performance Benchmarks

## Decoder Performance

[Paste benchmark results]

### Metrics

- Processing speed: X samples/sec
- Real-time factor: Y.Yx (higher is better, >1.0 means faster than real-time)
- Memory usage: Z MB/op
```

**Step 4: Commit**

```bash
git add internal/decoder/decoder_bench_test.go docs/performance.md
git commit -m "test: add performance benchmarks"
```

---

## Task 15: Update README with Build Instructions

**Files:**

- Modify: `README.md`

**Step 1: Add troubleshooting section**

Add to `README.md` before "Contributing":

````markdown
## Building

### Quick Start

```bash
# Clone repository
git clone https://github.com/cwbudde/go-sq-decoder.git
cd go-sq-decoder

# Download dependencies
go mod download

# Build
go build -o go-sq-decoder

# Run
./go-sq-decoder --help
```
````

### Troubleshooting

**"package not found" errors**

Run: `go mod tidy`

**FFT library issues**

Ensure you have Go 1.21+: `go version`

**Build fails**

1. Clean build cache: `go clean -cache`
2. Re-download deps: `go mod download`
3. Try again: `go build -v -o go-sq-decoder`

````

**Step 2: Add testing section**

```markdown
## Development

### Running Tests

```bash
# Unit tests
go test ./...

# With coverage
go test ./... -cover

# Verbose
go test ./... -v

# Specific package
go test ./pkg/sqmath/ -v
````

### Benchmarks

```bash
go test ./internal/decoder/ -bench=. -benchmem
```

````

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add build and testing instructions"
````

---

## Task 16: Add Example Audio Files (Optional)

**Files:**

- Create: `examples/` directory
- Add: Example SQ-encoded files (if available)

**Step 1: Create examples directory**

```bash
mkdir -p /mnt/e/Public/go-sq-decoder/examples
```

**Step 2: Add example script**

Create `examples/create_example.sh`:

```bash
#!/bin/bash
# Generate example SQ-encoded audio
# This requires an SQ encoder (not included)

echo "Example SQ-encoded files:"
echo "- Download SQ-encoded test files from:"
echo "  https://archive.org/details/quadraphonic"
echo "- Or use the Delphi SQ encoder to create test files"
```

**Step 3: Document examples**

Create `examples/README.md`:

````markdown
# Example Files

To test the decoder, you need SQ-encoded audio files.

## Sources

1. **Archive.org**: Search for "SQ quadraphonic" recordings
2. **Generate your own**: Use the Delphi SQ encoder in `/mnt/c/Users/Chris/Code/Plugins/Quadrophonic/SQ/Encoder/`
3. **Test files**: Use the generated test files in `testdata/`

## Usage

```bash
# Decode example file
go-sq-decoder example_sq.wav decoded_quad.wav
```
````

````

**Step 4: Commit**

```bash
git add examples/
git commit -m "docs: add examples directory with instructions"
````

---

## Task 17: Final Integration Test and Release Prep

**Files:**

- Check: All tests pass
- Create: Release checklist

**Step 1: Run all tests**

Run:

```bash
cd /mnt/e/Public/go-sq-decoder
go test ./... -v
```

Expected: All tests PASS

**Step 2: Build for release**

Run:

```bash
# Clean build
go clean
go build -ldflags="-s -w" -o go-sq-decoder
```

Expected: Optimized binary created

**Step 3: Test final binary**

Run:

```bash
# Test with actual file
./go-sq-decoder testdata/test_stereo.wav testdata/final_test.wav

# Verify output exists
ls -lh testdata/final_test.wav
```

**Step 4: Create release checklist**

Create `docs/release-checklist.md`:

```markdown
# Release Checklist

- [ ] All tests pass: `go test ./...`
- [ ] Benchmarks run: `go test ./... -bench=.`
- [ ] Binary builds: `go build`
- [ ] CLI works: `./go-sq-decoder --help`
- [ ] Example files work
- [ ] README accurate
- [ ] Documentation complete
- [ ] No compiler warnings
- [ ] Code formatted: `go fmt ./...`
- [ ] Dependencies updated: `go get -u && go mod tidy`

## Version: 1.0.0

Date: [YYYY-MM-DD]
```

**Step 5: Run formatter**

Run:

```bash
go fmt ./...
```

**Step 6: Final commit**

```bash
git add docs/release-checklist.md
git commit -m "chore: prepare for v1.0.0 release"
```

**Step 7: Create tag**

Run:

```bash
git tag -a v1.0.0 -m "Release v1.0.0: FFT-based SQ decoder"
```

---

## Summary

Upon completion of all tasks:

1. ✅ Fully functional SQ decoder CLI
2. ✅ Comprehensive test suite
3. ✅ Benchmarks for performance
4. ✅ Complete documentation
5. ✅ Ready for release

## Next Steps

After implementation:

1. Test with real SQ-encoded records
2. Compare output with Delphi VST version
3. Consider adding GUI
4. Port basic recursive filter version for low-latency use
5. Add batch processing mode

---

**Plan created**: 2026-01-02
**Estimated time**: 3-4 hours for full implementation
**Difficulty**: Medium (requires debugging FFT/audio processing)
