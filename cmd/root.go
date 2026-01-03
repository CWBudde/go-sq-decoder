package cmd

import (
	"fmt"
	"os"

	"github.com/cwbudde/go-sq-decoder/internal/decoder"
	"github.com/cwbudde/go-sq-decoder/internal/encoder"
	"github.com/cwbudde/go-sq-decoder/internal/wav"
	"github.com/spf13/cobra"
)

var (
	verbose   bool
	blockSize int
	overlap   int
	float32   bool
)

var rootCmd = &cobra.Command{
	Use:   "go-sq-decoder",
	Short: "SQ Quadrophonic Encoder/Decoder - Convert between SQ stereo and quad audio",
	Long: `SQ Quadrophonic Encoder/Decoder (FFT-based)

Decodes SQ (Stereo Quadrophonic) matrix-encoded stereo audio into 4-channel
quadrophonic audio, or encodes 4-channel quad audio into SQ-compatible stereo.

Decode Input:  2-channel WAV file (LT, RT - Left Total, Right Total)
Decode Output: 4-channel WAV file (LF, RF, LB, RB - Left Front, Right Front, Left Back, Right Back)

Encode Input:  4-channel WAV file (LF, RF, LB, RB - Left Front, Right Front, Left Back, Right Back)
Encode Output: 2-channel WAV file (LT, RT - Left Total, Right Total)

Based on the SQ² decoder implementation with FFT-based Hilbert transformer
for superior channel separation compared to simple recursive filters.`,
	RunE: runRoot,
}

var decodeCmd = &cobra.Command{
	Use:   "decode [input.wav] [output.wav]",
	Short: "Decode SQ-encoded stereo to quadrophonic WAV",
	Args:  cobra.ExactArgs(2),
	RunE:  runDecode,
}

var encodeCmd = &cobra.Command{
	Use:   "encode [input.wav] [output.wav]",
	Short: "Encode quadrophonic WAV to SQ-encoded stereo",
	Args:  cobra.ExactArgs(2),
	RunE:  runEncode,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().IntVarP(&blockSize, "block-size", "b", decoder.DefaultBlockSize, "FFT block size (power of 2)")
	rootCmd.PersistentFlags().IntVarP(&overlap, "overlap", "o", decoder.DefaultOverlap, "overlap in samples")
	rootCmd.PersistentFlags().BoolVar(&float32, "float32", false, "output 32-bit IEEE float WAV instead of 16-bit PCM")
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(encodeCmd)
}

func runRoot(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	if len(args) != 2 {
		return cobra.ExactArgs(2)(cmd, args)
	}
	return runDecode(cmd, args)
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

	if verbose {
		fmt.Printf("Decoder configuration:\n")
		fmt.Printf("  Block size: %d samples\n", blockSize)
		fmt.Printf("  Overlap: %d samples\n", overlap)
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
