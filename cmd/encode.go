package cmd

import (
	"fmt"

	"github.com/cwbudde/go-sq-tool/internal/encoder"
	"github.com/cwbudde/go-sq-tool/internal/wav"
	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode [input.wav] [output.wav]",
	Short: "Encode quadrophonic WAV to SQ-encoded stereo",
	Args:  cobra.ExactArgs(2),
	RunE:  runEncode,
}

func runEncode(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile := args[1]

	if verbose {
		fmt.Printf("SQ Quadrophonic Encoder\n")
		fmt.Printf("=======================\n\n")
	}

	if verbose {
		fmt.Printf("Reading input file: %s\n", inputFile)
	}

	audioData, err := wav.ReadWAVChannels(inputFile, 4)
	if err != nil {
		return fmt.Errorf("failed to read input WAV: %w", err)
	}

	if verbose {
		fmt.Printf("  Sample rate: %d Hz\n", audioData.SampleRate)
		fmt.Printf("  Samples: %d\n", audioData.NumSamples)
		fmt.Printf("  Duration: %.2f seconds\n\n", float64(audioData.NumSamples)/float64(audioData.SampleRate))
	}

	sqEncoder := encoder.NewSQEncoderWithParams(blockSize, overlap)

	if verbose {
		fmt.Printf("Encoder configuration:\n")
		fmt.Printf("  Block size: %d samples\n", blockSize)
		fmt.Printf("  Overlap: %d samples\n", overlap)
		fmt.Printf("  Latency: %d samples (%.2f ms)\n\n",
			sqEncoder.GetLatency(),
			float64(sqEncoder.GetLatency())/float64(audioData.SampleRate)*1000.0)
		fmt.Printf("Processing...\n")
	}

	output, err := sqEncoder.Process(audioData.Samples)
	if err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}

	outputData := &wav.AudioData{
		SampleRate: audioData.SampleRate,
		Samples:    output,
		NumSamples: audioData.NumSamples,
	}

	if verbose {
		fmt.Printf("Writing output file: %s\n", outputFile)
		if float32 {
			fmt.Printf("  Format: 32-bit IEEE float\n")
		} else {
			fmt.Printf("  Format: 16-bit PCM\n")
		}
	}

	if float32 {
		if err := wav.WriteStereoFloat32WAV(outputFile, outputData); err != nil {
			return fmt.Errorf("failed to write output WAV: %w", err)
		}
	} else {
		if err := wav.WriteStereoWAV(outputFile, outputData); err != nil {
			return fmt.Errorf("failed to write output WAV: %w", err)
		}
	}

	if verbose {
		fmt.Printf("\nDone! Encoded to 2-channel SQ stereo audio.\n")
		fmt.Printf("Channels: LT (Left Total), RT (Right Total)\n")
	} else {
		fmt.Printf("Successfully encoded %s -> %s\n", inputFile, outputFile)
	}

	return nil
}
