package metrics

import (
	"math"

	algofft "github.com/MeKo-Christian/algo-fft"
)

const separationEpsilon = 1e-12

// SeparationResult summarizes target vs leak energy for a decoded channel.
type SeparationResult struct {
	TargetRMS    float64
	LeakRMS      float64
	SeparationDB float64
}

type LeakMode string

const (
	LeakModeMax LeakMode = "max"
	LeakModeAvg LeakMode = "avg"
)

// SeparationOptions controls how separation is computed.
type SeparationOptions struct {
	LeakMode   LeakMode
	SampleRate int
	FMin       float64
	FMax       float64
}

// ChannelSeparation computes RMS-based separation for a target channel.
func ChannelSeparation(decoded [][]float64, target int, options SeparationOptions) SeparationResult {
	if target < 0 || target >= len(decoded) {
		return SeparationResult{}
	}

	targetRMS := rmsWithOptions(decoded[target], options)
	var leakRMS float64
	leakCount := 0
	for ch := 0; ch < len(decoded); ch++ {
		if ch == target {
			continue
		}
		r := rmsWithOptions(decoded[ch], options)
		switch options.LeakMode {
		case LeakModeAvg:
			leakRMS += r
			leakCount++
		default:
			if r > leakRMS {
				leakRMS = r
			}
		}
	}
	if options.LeakMode == LeakModeAvg && leakCount > 0 {
		leakRMS /= float64(leakCount)
	}

	return SeparationResult{
		TargetRMS:    targetRMS,
		LeakRMS:      leakRMS,
		SeparationDB: separationDB(targetRMS, leakRMS),
	}
}

// ChannelPairSeparation computes separation for a target/leak pair.
func ChannelPairSeparation(decoded [][]float64, target, leak int, options SeparationOptions) SeparationResult {
	if target < 0 || target >= len(decoded) {
		return SeparationResult{}
	}
	if leak < 0 || leak >= len(decoded) {
		return SeparationResult{}
	}

	targetRMS := rmsWithOptions(decoded[target], options)
	leakRMS := rmsWithOptions(decoded[leak], options)

	return SeparationResult{
		TargetRMS:    targetRMS,
		LeakRMS:      leakRMS,
		SeparationDB: separationDB(targetRMS, leakRMS),
	}
}

func separationDB(targetRMS, leakRMS float64) float64 {
	if leakRMS > separationEpsilon && targetRMS > separationEpsilon {
		return 20.0 * math.Log10(targetRMS/leakRMS)
	}
	if targetRMS > separationEpsilon && leakRMS <= separationEpsilon {
		return math.Inf(1)
	}
	return 0.0
}

func rmsWithOptions(samples []float64, options SeparationOptions) float64 {
	if options.FMin <= 0 && options.FMax <= 0 {
		return rms(samples)
	}
	if options.SampleRate <= 0 {
		return 0
	}
	return bandRMS(samples, options.SampleRate, options.FMin, options.FMax)
}

func rms(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range samples {
		sum += v * v
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func bandRMS(samples []float64, sampleRate int, fmin, fmax float64) float64 {
	n := len(samples)
	if n == 0 || sampleRate <= 0 {
		return 0
	}
	if fmin < 0 {
		fmin = 0
	}
	nyquist := float64(sampleRate) / 2.0
	if fmax <= 0 || fmax > nyquist {
		fmax = nyquist
	}
	if fmin > fmax {
		return 0
	}

	plan, err := algofft.NewPlan64(n)
	if err != nil {
		return 0
	}

	input := make([]complex128, n)
	for i, v := range samples {
		input[i] = complex(v, 0)
	}
	freq := make([]complex128, n)
	if err := plan.Forward(freq, input); err != nil {
		return 0
	}

	sumPow := 0.0
	nFloat := float64(n)
	for k := 0; k <= n/2; k++ {
		freqHz := float64(k) * float64(sampleRate) / nFloat
		if freqHz < fmin || freqHz > fmax {
			continue
		}
		power := real(freq[k])*real(freq[k]) + imag(freq[k])*imag(freq[k])
		if k == 0 || k == n/2 {
			sumPow += power
		} else {
			sumPow += 2.0 * power
		}
	}

	return math.Sqrt(sumPow / (nFloat * nFloat))
}
