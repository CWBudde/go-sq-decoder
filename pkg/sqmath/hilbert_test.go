package sqmath_test

import (
	"math"
	"testing"

	"github.com/cwbudde/go-sq-decoder/pkg/sqmath"
)

func TestHilbertTransformer_ProcessBlock_PanicsOnWrongBlockSize(t *testing.T) {
	t.Parallel()

	ht := sqmath.NewHilbertTransformer(1024, 512)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on wrong input length")
		}
	}()

	_ = ht.ProcessBlock(make([]float64, 1023))
}

func TestHilbertTransformer_ProcessBlock_SineBecomesApproximatelyCosine(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		k         = 37 // bin index; avoid DC/Nyquist
	)

	ht := sqmath.NewHilbertTransformer(blockSize, overlap)

	in := make([]float64, blockSize)
	refSin := make([]float64, blockSize)
	refCos := make([]float64, blockSize)
	for n := 0; n < blockSize; n++ {
		phi := 2.0 * math.Pi * float64(k) * float64(n) / float64(blockSize)
		refSin[n] = math.Sin(phi)
		refCos[n] = math.Cos(phi)
		in[n] = refSin[n]
	}

	out := ht.ProcessBlock(in)
	if len(out) != blockSize {
		t.Fatalf("len(out)=%d, want %d", len(out), blockSize)
	}

	// We don’t assert absolute gain because the implementation windows/scales.
	// Also, the SQ²-style processing intentionally uses offsets when consuming
	// the Hilbert output. Mirror that here to avoid failing due to expected delay.
	inputOffset := overlap / 4
	outputOffset := overlap / 2
	windowLen := overlap

	outWin := out[outputOffset : outputOffset+windowLen]
	cosWin := refCos[inputOffset : inputOffset+windowLen]
	sinWin := refSin[inputOffset : inputOffset+windowLen]

	// Quadrature check: strong correlation with cos (sign depends on convention)
	// and weak correlation with sin.
	corrCos := math.Abs(normalizedDot(outWin, cosWin))
	corrSin := normalizedDot(outWin, sinWin)

	if corrCos < 0.80 {
		t.Fatalf("|corr(outWin, cosWin)|=%.3f, want >= 0.80", corrCos)
	}
	if math.Abs(corrSin) > 0.20 {
		t.Fatalf("|corr(outWin, sinWin)|=%.3f, want <= 0.20", math.Abs(corrSin))
	}

	// Sanity: finite outputs.
	for i := range out {
		if math.IsNaN(out[i]) || math.IsInf(out[i], 0) {
			t.Fatalf("out[%d] is not finite: %v", i, out[i])
		}
	}
}

func normalizedDot(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("length mismatch")
	}

	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / math.Sqrt(na*nb)
}
