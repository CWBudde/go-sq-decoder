package sqmath

import (
	"math"

	algofft "github.com/MeKo-Christian/algo-fft"
)

type WindowType string

const (
	WindowHann        WindowType = "hann"
	WindowHanning     WindowType = "hanning" // alias
	WindowHamming     WindowType = "hamming"
	WindowBlackman    WindowType = "blackman"
	WindowRectangular WindowType = "rect"
)

// HilbertTransformer performs 90-degree phase shift using FFT
type HilbertTransformer struct {
	blockSize   int
	overlap     int
	fftSize     int
	fftPlan     *algofft.Plan[complex128]
	windowType  WindowType
	window      []float64
	transferFn  []complex128
	inputBuffer []float64
	initialized bool
}

// NewHilbertTransformer creates a new Hilbert transformer
// blockSize: FFT block size (should be power of 2)
// overlap: overlap in samples (typically blockSize/2)
func NewHilbertTransformer(blockSize, overlap int) *HilbertTransformer {
	return NewHilbertTransformerWithWindow(blockSize, overlap, WindowHann)
}

// NewHilbertTransformerWithWindow creates a new Hilbert transformer with a selectable window.
// windowType: one of WindowHann/WindowHamming/WindowBlackman/WindowRectangular.
func NewHilbertTransformerWithWindow(blockSize, overlap int, windowType WindowType) *HilbertTransformer {
	plan, err := algofft.NewPlan64(blockSize)
	if err != nil {
		panic(err)
	}

	ht := &HilbertTransformer{
		blockSize:   blockSize,
		overlap:     overlap,
		fftSize:     blockSize,
		fftPlan:     plan,
		windowType:  windowType,
		inputBuffer: make([]float64, blockSize),
	}

	ht.makeFilter()
	return ht
}

// makeFilter constructs the Hilbert transform transfer function
// Based on SQ² decoder implementation from VSTDataModule.pas
func (ht *HilbertTransformer) makeFilter() {
	// Create impulse response: h[n] = 2/(π·n) for odd n, 0 for even
	impulse := make([]float64, ht.blockSize)
	center := ht.overlap / 2

	for i := range center {
		if i%2 == 1 {
			impulse[center+i] = 2.0 / (math.Pi * float64(i))
			impulse[center-i] = -2.0 / (math.Pi * float64(i))
		}
		// Even indices remain 0
	}

	// Apply window
	ht.window = makeWindow(ht.windowType, ht.overlap)
	for i := 0; i < ht.overlap; i++ {
		impulse[i] *= ht.window[i]
	}

	// Scale by 1.8 (from original implementation)
	for i := 0; i < ht.overlap; i++ {
		impulse[i] *= 1.8
	}

	// Convert to complex for FFT
	impulseComplex := make([]complex128, ht.fftSize)
	for i := range impulse {
		impulseComplex[i] = complex(impulse[i], 0)
	}

	// FFT to get transfer function
	ht.transferFn = make([]complex128, ht.fftSize)
	if err := ht.fftPlan.Forward(ht.transferFn, impulseComplex); err != nil {
		panic(err)
	}
	ht.initialized = true
}

func makeWindow(windowType WindowType, size int) []float64 {
	switch windowType {
	case WindowHann, WindowHanning:
		return hannWindow(size)
	case WindowHamming:
		return hammingWindow(size)
	case WindowBlackman:
		return blackmanWindow(size)
	case WindowRectangular:
		return rectangularWindow(size)
	default:
		panic("unknown window type")
	}
}

// hannWindow creates a Hann window (often called "Hanning").
func hannWindow(size int) []float64 {
	window := make([]float64, size)
	if size <= 1 {
		for i := range window {
			window[i] = 1
		}
		return window
	}
	for i := 0; i < size; i++ {
		window[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(size-1)))
	}
	return window
}

func hammingWindow(size int) []float64 {
	window := make([]float64, size)
	if size <= 1 {
		for i := range window {
			window[i] = 1
		}
		return window
	}
	for i := 0; i < size; i++ {
		window[i] = 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(i)/float64(size-1))
	}
	return window
}

func blackmanWindow(size int) []float64 {
	window := make([]float64, size)
	if size <= 1 {
		for i := range window {
			window[i] = 1
		}
		return window
	}
	for i := 0; i < size; i++ {
		x := 2.0 * math.Pi * float64(i) / float64(size-1)
		window[i] = 0.42 - 0.5*math.Cos(x) + 0.08*math.Cos(2*x)
	}
	return window
}

func rectangularWindow(size int) []float64 {
	window := make([]float64, size)
	for i := range window {
		window[i] = 1
	}
	return window
}

// ProcessBlock applies Hilbert transform to a block of samples
func (ht *HilbertTransformer) ProcessBlock(input []float64) []float64 {
	if len(input) != ht.blockSize {
		panic("input size must match block size")
	}

	// Convert to complex
	inputComplex := make([]complex128, ht.fftSize)
	for i := 0; i < len(input); i++ {
		inputComplex[i] = complex(input[i], 0)
	}

	// FFT
	freqDomain := make([]complex128, ht.fftSize)
	if err := ht.fftPlan.Forward(freqDomain, inputComplex); err != nil {
		panic(err)
	}

	// Apply transfer function (complex multiplication per bin)
	for i := 0; i < ht.fftSize; i++ {
		freqDomain[i] *= ht.transferFn[i]
	}

	// Inverse FFT
	timeDomain := make([]complex128, ht.fftSize)
	if err := ht.fftPlan.Inverse(timeDomain, freqDomain); err != nil {
		panic(err)
	}

	// Extract real part and rescale
	output := make([]float64, ht.blockSize)
	scale := 1.0 / float64(ht.fftSize)
	for i := 0; i < ht.blockSize; i++ {
		output[i] = real(timeDomain[i]) * scale
	}

	return output
}
