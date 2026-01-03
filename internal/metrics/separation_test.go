package metrics_test

import (
	"math"
	"testing"

	"github.com/cwbudde/go-sq-tool/internal/metrics"
)

func TestChannelSeparation(t *testing.T) {
	t.Parallel()

	decoded := [][]float64{
		{1.0, -1.0},
		{0.1, -0.1},
		{0.0, 0.0},
		{0.0, 0.0},
	}

	result := metrics.ChannelSeparation(decoded, 0, metrics.SeparationOptions{
		LeakMode: metrics.LeakModeMax,
	})
	if math.Abs(result.TargetRMS-1.0) > 1e-12 {
		t.Fatalf("TargetRMS = %.12f, want 1.0", result.TargetRMS)
	}
	if math.Abs(result.LeakRMS-0.1) > 1e-12 {
		t.Fatalf("LeakRMS = %.12f, want 0.1", result.LeakRMS)
	}
	if math.Abs(result.SeparationDB-20.0) > 1e-9 {
		t.Fatalf("SeparationDB = %.9f, want 20.0", result.SeparationDB)
	}
}
