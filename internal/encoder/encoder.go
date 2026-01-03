package encoder

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-sq-decoder/pkg/sqmath"
)

const (
	// DefaultBlockSize for FFT processing (must be power of 2)
	DefaultBlockSize = 1024
	// DefaultOverlap is 50% overlap
	DefaultOverlap = 512
)

// SQEncoder implements the SQ (FFT-based) quadrophonic encoder
type SQEncoder struct {
	blockSize    int
	overlap      int
	initialDelay int
	sqrt2        float64
	hilbertLB    *sqmath.HilbertTransformer
	hilbertRB    *sqmath.HilbertTransformer
}

// NewSQEncoder creates a new SQ encoder with FFT-based Hilbert transform
func NewSQEncoder() *SQEncoder {
	return NewSQEncoderWithParams(DefaultBlockSize, DefaultOverlap)
}

// NewSQEncoderWithParams creates a new SQ encoder with custom parameters
func NewSQEncoderWithParams(blockSize, overlap int) *SQEncoder {
	initialDelay := overlap + overlap/2

	return &SQEncoder{
		blockSize:    blockSize,
		overlap:      overlap,
		initialDelay: initialDelay,
		sqrt2:        math.Sqrt(2.0) / 2.0, // ≈ 0.707
		hilbertLB:    sqmath.NewHilbertTransformer(blockSize, overlap),
		hilbertRB:    sqmath.NewHilbertTransformer(blockSize, overlap),
	}
}

// Process encodes 4-channel quadrophonic audio to stereo SQ
// Input: [4][numSamples] - LF, RF, LB, RB (Left Front, Right Front, Left Back, Right Back)
// Output: [2][numSamples] - LT, RT (Left Total, Right Total)
func (e *SQEncoder) Process(input [][]float64) ([][]float64, error) {
	if len(input) != 4 {
		return nil, fmt.Errorf("input must have 4 channels, got %d", len(input))
	}

	numSamples := len(input[0])
	for i := 1; i < 4; i++ {
		if len(input[i]) != numSamples {
			return nil, fmt.Errorf("input channels must have same length")
		}
	}

	numBlocks := (numSamples + e.overlap - 1) / e.overlap

	output := make([][]float64, 2)
	for i := 0; i < 2; i++ {
		output[i] = make([]float64, numSamples)
	}

	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		startIdx := blockIdx * e.overlap

		blockLF := make([]float64, e.blockSize)
		blockRF := make([]float64, e.blockSize)
		blockLB := make([]float64, e.blockSize)
		blockRB := make([]float64, e.blockSize)

		for i := 0; i < e.blockSize; i++ {
			srcIdx := startIdx + i
			if srcIdx < numSamples {
				blockLF[i] = input[0][srcIdx]
				blockRF[i] = input[1][srcIdx]
				blockLB[i] = input[2][srcIdx]
				blockRB[i] = input[3][srcIdx]
			}
		}

		phaseShiftedLB := e.hilbertLB.ProcessBlock(blockLB)
		phaseShiftedRB := e.hilbertRB.ProcessBlock(blockRB)

		outputOffset := e.overlap / 2
		inputOffset := e.overlap / 4

		for i := 0; i < e.overlap; i++ {
			outIdx := startIdx + i
			if outIdx >= numSamples {
				break
			}

			inIdx := inputOffset + i
			if inIdx >= e.blockSize {
				break
			}

			phaseIdx := outputOffset + i
			if phaseIdx >= e.blockSize {
				break
			}

			lf := blockLF[inIdx]
			rf := blockRF[inIdx]
			lb := blockLB[inIdx]
			rb := blockRB[inIdx]
			hlb := phaseShiftedLB[phaseIdx]
			hrb := phaseShiftedRB[phaseIdx]

			// SQ Encode Matrix:
			// LT = LF + sqrt(2)/2 * RB - sqrt(2)/2 * H(LB)
			// RT = RF - sqrt(2)/2 * LB + sqrt(2)/2 * H(RB)
			output[0][outIdx] = lf + e.sqrt2*rb - e.sqrt2*hlb
			output[1][outIdx] = rf - e.sqrt2*lb + e.sqrt2*hrb
		}
	}

	return output, nil
}

// GetLatency returns the encoder latency in samples
func (e *SQEncoder) GetLatency() int {
	return e.initialDelay
}

// GetInfo returns information about the encoder configuration
func (e *SQEncoder) GetInfo() string {
	return fmt.Sprintf("SQ Encoder (FFT-based)\n"+
		"Block Size: %d samples\n"+
		"Overlap: %d samples\n"+
		"Latency: %d samples (%.2f ms @ 44.1kHz)",
		e.blockSize, e.overlap, e.initialDelay,
		float64(e.initialDelay)/44100.0*1000.0)
}
