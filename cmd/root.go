package cmd

import (
	"fmt"
	"os"

	"github.com/cwbudde/go-sq-tool/internal/decoder"
	"github.com/spf13/cobra"
)

var (
	verbose   bool
	blockSize int
	overlap   int
	float32   bool
	logic     bool
)

var rootCmd = &cobra.Command{
	Use:   "go-sq-tool",
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
	rootCmd.PersistentFlags().BoolVar(&logic, "logic", false, "enable CBS-style logic steering for decoding")
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(encodeCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(generateCmd)
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
