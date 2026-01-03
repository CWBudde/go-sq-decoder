package decoder_test

import (
	"math"
	"testing"

	"github.com/cwbudde/go-sq-tool/internal/decoder"
)

func TestLogicSteering_IncreasesDominantRatio(t *testing.T) {
	t.Parallel()

	const (
		blockSize = 1024
		overlap   = 512
		n         = 20 * overlap
		skip      = 2 * overlap
	)

	lt := make([]float64, n)
	rt := make([]float64, n)
	for i := 0; i < n; i++ {
		rt[i] = 0.8 * math.Sin(2.0*math.Pi*float64(i)/97.0)
	}

	basic := decoder.NewSQDecoderWithParams(blockSize, overlap)
	basic.SetSampleRate(44100)
	outBasic, err := basic.Process([][]float64{lt, rt})
	if err != nil {
		t.Fatalf("basic Process() error = %v", err)
	}

	logic := decoder.NewSQDecoderWithParams(blockSize, overlap)
	logic.SetSampleRate(44100)
	logic.EnableLogicSteering(true)
	outLogic, err := logic.Process([][]float64{lt, rt})
	if err != nil {
		t.Fatalf("logic Process() error = %v", err)
	}

	ratioBasic := dominantRatio(outBasic, skip)
	ratioLogic := dominantRatio(outLogic, skip)

	if ratioLogic <= ratioBasic*1.05 {
		t.Fatalf("dominant ratio = %.4f, want > %.4f (basic %.4f)", ratioLogic, ratioBasic*1.05, ratioBasic)
	}
}

func dominantRatio(out [][]float64, skip int) float64 {
	const eps = 1e-12
	var rf, sum float64
	for i := skip; i < len(out[0]); i++ {
		rf += math.Abs(out[1][i])
		sum += math.Abs(out[0][i]) + math.Abs(out[2][i]) + math.Abs(out[3][i])
	}
	return rf / (sum + eps)
}
