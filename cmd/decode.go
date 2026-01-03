package cmd

import (
	"fmt"

	"github.com/cwbudde/go-sq-tool/internal/decoder"
	"github.com/cwbudde/go-sq-tool/internal/wav"
	"github.com/spf13/cobra"
)

var decodeCmd = &cobra.Command{
	Use:   "decode [input.wav] [output.wav]",
	Short: "Decode SQ-encoded stereo to quadrophonic WAV",
	Args:  cobra.ExactArgs(2),
	RunE:  runDecode,
}

func runDecode(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile := args[1]

	if verbose {
		fmt.Printf("SQ Quadrophonic Decoder\n")
		fmt.Printf("=======================\n\n")
	}

	// Read input WAV
	if verbose {
		fmt.Printf("Reading input file: %s\n", inputFile)
	}

	audioData, err := wav.ReadWAV(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input WAV: %w", err)
	}

	if verbose {
		fmt.Printf("  Sample rate: %d Hz\n", audioData.SampleRate)
		fmt.Printf("  Samples: %d\n", audioData.NumSamples)
		fmt.Printf("  Duration: %.2f seconds\n\n", float64(audioData.NumSamples)/float64(audioData.SampleRate))
	}

	// Create decoder
	sqDecoder := decoder.NewSQDecoderWithParams(blockSize, overlap)
	sqDecoder.SetSampleRate(int(audioData.SampleRate))
	if logic {
		sqDecoder.EnableLogicSteering(true)
	}

	if verbose {
		fmt.Printf("Decoder configuration:\n")
		fmt.Printf("  Block size: %d samples\n", blockSize)
		fmt.Printf("  Overlap: %d samples\n", overlap)
		if logic {
			fmt.Printf("  Logic steering: enabled\n")
		}
		fmt.Printf("  Latency: %d samples (%.2f ms)\n\n",
			sqDecoder.GetLatency(),
			float64(sqDecoder.GetLatency())/float64(audioData.SampleRate)*1000.0)
		fmt.Printf("Processing...\n")
	}

	// Decode
	output, err := sqDecoder.Process(audioData.Samples)
	if err != nil {
		return fmt.Errorf("decoding failed: %w", err)
	}

	// Prepare output data
	outputData := &wav.AudioData{
		SampleRate: audioData.SampleRate,
		Samples:    output,
		NumSamples: audioData.NumSamples,
	}

	// Write output WAV
	if verbose {
		fmt.Printf("Writing output file: %s\n", outputFile)
		if float32 {
			fmt.Printf("  Format: 32-bit IEEE float\n")
		} else {
			fmt.Printf("  Format: 16-bit PCM\n")
		}
	}

	if float32 {
		if err := wav.WriteFloat32WAV(outputFile, outputData); err != nil {
			return fmt.Errorf("failed to write output WAV: %w", err)
		}
	} else {
		if err := wav.WriteWAV(outputFile, outputData); err != nil {
			return fmt.Errorf("failed to write output WAV: %w", err)
		}
	}

	if verbose {
		fmt.Printf("\nDone! Decoded to 4-channel quadrophonic audio.\n")
		fmt.Printf("Channels: LF (Left Front), RF (Right Front), LB (Left Back), RB (Right Back)\n")
	} else {
		fmt.Printf("Successfully decoded %s -> %s\n", inputFile, outputFile)
	}

	return nil
}
