package decoder

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

// SQDecoder implements the SQ² (FFT-based) quadrophonic decoder
type SQDecoder struct {
	blockSize     int
	overlap       int
	initialDelay  int
	sqrt2         float64
	hilbertLeft   *sqmath.HilbertTransformer
	hilbertRight  *sqmath.HilbertTransformer
	inputBufferL  []float64
	inputBufferR  []float64
	outputBuffers [4][]float64
	bufferPos     int
}

// NewSQDecoder creates a new SQ decoder with FFT-based Hilbert transform
func NewSQDecoder() *SQDecoder {
	return NewSQDecoderWithParams(DefaultBlockSize, DefaultOverlap)
}

// NewSQDecoderWithParams creates a new SQ decoder with custom parameters
func NewSQDecoderWithParams(blockSize, overlap int) *SQDecoder {
	// Initial delay calculation from SQ² implementation
	initialDelay := overlap + overlap/2

	decoder := &SQDecoder{
		blockSize:    blockSize,
		overlap:      overlap,
		initialDelay: initialDelay,
		sqrt2:        math.Sqrt(2.0) / 2.0, // ≈ 0.707
		hilbertLeft:  sqmath.NewHilbertTransformer(blockSize, overlap),
		hilbertRight: sqmath.NewHilbertTransformer(blockSize, overlap),
		inputBufferL: make([]float64, blockSize),
		inputBufferR: make([]float64, blockSize),
		bufferPos:    0,
	}

	// Initialize output buffers
	for i := 0; i < 4; i++ {
		decoder.outputBuffers[i] = make([]float64, blockSize)
	}

	return decoder
}

// Process decodes stereo SQ-encoded audio to 4-channel quadrophonic
// Input: [2][numSamples] - LT, RT (Left Total, Right Total)
// Output: [4][numSamples] - LF, RF, LB, RB (Left Front, Right Front, Left Back, Right Back)
func (d *SQDecoder) Process(input [][]float64) ([][]float64, error) {
	if len(input) != 2 {
		return nil, fmt.Errorf("input must have 2 channels, got %d", len(input))
	}

	numSamples := len(input[0])
	if len(input[1]) != numSamples {
		return nil, fmt.Errorf("input channels must have same length")
	}

	// Pad input to block boundaries
	numBlocks := (numSamples + d.overlap - 1) / d.overlap

	// Initialize output
	output := make([][]float64, 4)
	for i := 0; i < 4; i++ {
		output[i] = make([]float64, numSamples)
	}

	// Process in blocks with overlap
	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		startIdx := blockIdx * d.overlap

		// Prepare input block (with zero padding if needed)
		blockL := make([]float64, d.blockSize)
		blockR := make([]float64, d.blockSize)

		for i := 0; i < d.blockSize; i++ {
			srcIdx := startIdx + i
			if srcIdx < numSamples {
				blockL[i] = input[0][srcIdx]
				blockR[i] = input[1][srcIdx]
			}
			// else remains 0 (zero padding)
		}

		// Apply Hilbert transform
		phaseShiftedL := d.hilbertLeft.ProcessBlock(blockL)
		phaseShiftedR := d.hilbertRight.ProcessBlock(blockR)

		// Apply SQ decode matrix
		// Based on SQ² VSTDataModule.pas V2M_Process
		outputOffset := d.overlap / 2
		inputOffset := d.overlap / 4

		for i := 0; i < d.overlap; i++ {
			outIdx := startIdx + i
			if outIdx >= numSamples {
				break
			}

			inIdx := inputOffset + i
			if inIdx >= d.blockSize {
				break
			}

			phaseIdx := outputOffset + i
			if phaseIdx >= d.blockSize {
				break
			}

			// SQ Decode Matrix:
			// LF = LT (pass through)
			// RF = RT (pass through)
			// LB = sqrt(2)/2 * H(LT) - sqrt(2)/2 * RT
			// RB = sqrt(2)/2 * LT - sqrt(2)/2 * H(RT)

			lt := blockL[inIdx]
			rt := blockR[inIdx]
			hlt := phaseShiftedL[phaseIdx]
			hrt := phaseShiftedR[phaseIdx]

			output[0][outIdx] = lt                       // LF = LT
			output[1][outIdx] = rt                       // RF = RT
			output[2][outIdx] = d.sqrt2*hlt - d.sqrt2*rt // LB
			output[3][outIdx] = d.sqrt2*lt - d.sqrt2*hrt // RB
		}
	}

	return output, nil
}

// GetLatency returns the decoder latency in samples
func (d *SQDecoder) GetLatency() int {
	return d.initialDelay
}

// GetInfo returns information about the decoder configuration
func (d *SQDecoder) GetInfo() string {
	return fmt.Sprintf("SQ² Decoder (FFT-based)\n"+
		"Block Size: %d samples\n"+
		"Overlap: %d samples\n"+
		"Latency: %d samples (%.2f ms @ 44.1kHz)",
		d.blockSize, d.overlap, d.initialDelay,
		float64(d.initialDelay)/44100.0*1000.0)
}
