package encoder_test

import (
	"math"
	"testing"

	"github.com/cwbudde/go-sq-tool/internal/encoder"
)

func TestSQEncoder_Process_FrontOnlyShifted(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 10 * overlap
	)

	lf := make([]float64, n)
	rf := make([]float64, n)
	for i := 0; i < n; i++ {
		lf[i] = 0.6 * math.Sin(2.0*math.Pi*float64(i)/97.0)
		rf[i] = 0.4 * math.Cos(2.0*math.Pi*float64(i)/131.0)
	}

	quad := [][]float64{
		lf,
		rf,
		make([]float64, n),
		make([]float64, n),
	}

	sqEnc := encoder.NewSQEncoderWithParams(blockSize, overlap)
	stereo, err := sqEnc.Process(quad)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if got := len(stereo); got != 2 {
		t.Fatalf("channels = %d, want 2", got)
	}

	// Encoder maps LT/RT from input using inputOffset=overlap/4.
	shift := overlap / 4
	const tol = 1e-12
	for i := 0; i < n-shift; i++ {
		if math.Abs(stereo[0][i]-lf[i+shift]) > tol {
			t.Fatalf("LT[%d] = %.15f, want %.15f", i, stereo[0][i], lf[i+shift])
		}
		if math.Abs(stereo[1][i]-rf[i+shift]) > tol {
			t.Fatalf("RT[%d] = %.15f, want %.15f", i, stereo[1][i], rf[i+shift])
		}
	}
}

func TestSQEncoder_Process_ZeroInputIsZeroOutput(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 4 * overlap
	)

	quad := [][]float64{
		make([]float64, n),
		make([]float64, n),
		make([]float64, n),
		make([]float64, n),
	}

	sqEnc := encoder.NewSQEncoderWithParams(blockSize, overlap)
	stereo, err := sqEnc.Process(quad)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	const tol = 1e-12
	for ch := 0; ch < 2; ch++ {
		for i := 0; i < n; i++ {
			if math.Abs(stereo[ch][i]) > tol {
				t.Fatalf("out[%d][%d] = %.15f, want 0", ch, i, stereo[ch][i])
			}
		}
	}
}

func TestSQEncoder_Process_Errors(t *testing.T) {
	t.Parallel()

	sqEnc := encoder.NewSQEncoderWithParams(1024, 512)

	if _, err := sqEnc.Process([][]float64{make([]float64, 10)}); err == nil {
		t.Fatalf("expected error for wrong channel count")
	}

	if _, err := sqEnc.Process([][]float64{make([]float64, 10), make([]float64, 10), make([]float64, 9), make([]float64, 10)}); err == nil {
		t.Fatalf("expected error for length mismatch")
	}
}
