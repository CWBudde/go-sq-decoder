package decoder_test

import (
	"math"
	"testing"

	"github.com/cwbudde/go-sq-tool/internal/decoder"
)

func TestSQDecoder_Process_FrontChannelsShifted(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 10 * overlap
	)

	lt := make([]float64, n)
	rt := make([]float64, n)
	for i := 0; i < n; i++ {
		lt[i] = 0.7 * math.Sin(2.0*math.Pi*float64(i)/97.0)
		rt[i] = 0.3 * math.Cos(2.0*math.Pi*float64(i)/131.0)
	}

	sqDec := decoder.NewSQDecoderWithParams(blockSize, overlap)
	out, err := sqDec.Process([][]float64{lt, rt})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if got := len(out); got != 4 {
		t.Fatalf("channels = %d, want 4", got)
	}

	// Decoder maps LF/RF from input using inputOffset=overlap/4.
	shift := overlap / 4
	const tol = 1e-12
	for i := 0; i < n-shift; i++ {
		if math.Abs(out[0][i]-lt[i+shift]) > tol {
			t.Fatalf("LF[%d] = %.15f, want %.15f", i, out[0][i], lt[i+shift])
		}
		if math.Abs(out[1][i]-rt[i+shift]) > tol {
			t.Fatalf("RF[%d] = %.15f, want %.15f", i, out[1][i], rt[i+shift])
		}
	}
}

func TestSQDecoder_Process_ZeroInputIsZeroOutput(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 4 * overlap
	)

	lt := make([]float64, n)
	rt := make([]float64, n)

	sqDec := decoder.NewSQDecoderWithParams(blockSize, overlap)
	out, err := sqDec.Process([][]float64{lt, rt})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	const tol = 1e-12
	for ch := 0; ch < 4; ch++ {
		for i := 0; i < n; i++ {
			if math.Abs(out[ch][i]) > tol {
				t.Fatalf("out[%d][%d] = %.15f, want 0", ch, i, out[ch][i])
			}
		}
	}
}

func TestSQDecoder_Process_LogicSteeringFinite(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 6 * overlap
	)

	lt := make([]float64, n)
	rt := make([]float64, n)
	for i := 0; i < n; i++ {
		lt[i] = 0.5 * math.Sin(2.0*math.Pi*float64(i)/97.0)
	}

	sqDec := decoder.NewSQDecoderWithParams(blockSize, overlap)
	sqDec.SetSampleRate(44100)
	sqDec.EnableLogicSteering(true)

	out, err := sqDec.Process([][]float64{lt, rt})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	for ch := 0; ch < 4; ch++ {
		for i := 0; i < n; i++ {
			val := out[ch][i]
			if math.IsNaN(val) || math.IsInf(val, 0) {
				t.Fatalf("out[%d][%d] = %v, want finite", ch, i, val)
			}
		}
	}
}

func TestSQDecoder_Process_Errors(t *testing.T) {
	t.Parallel()

	sqDec := decoder.NewSQDecoderWithParams(1024, 512)

	if _, err := sqDec.Process([][]float64{make([]float64, 10)}); err == nil {
		t.Fatalf("expected error for wrong channel count")
	}

	if _, err := sqDec.Process([][]float64{make([]float64, 10), make([]float64, 9)}); err == nil {
		t.Fatalf("expected error for length mismatch")
	}
}
