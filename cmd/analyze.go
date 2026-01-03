package cmd

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-sq-tool/internal/decoder"
	"github.com/cwbudde/go-sq-tool/internal/encoder"
	"github.com/cwbudde/go-sq-tool/internal/metrics"
	"github.com/cwbudde/go-sq-tool/internal/wav"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [input.wav]",
	Short: "Measure channel separation for a quad input via encode/decode",
	Args:  cobra.ExactArgs(1),
	RunE:  runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringVar(&analyzeLeakMode, "leak-mode", "max", "leakage aggregation: max or avg")
	analyzeCmd.Flags().Float64Var(&analyzeFMin, "fmin", 0, "min frequency for band-limited analysis (Hz)")
	analyzeCmd.Flags().Float64Var(&analyzeFMax, "fmax", 0, "max frequency for band-limited analysis (Hz)")
	analyzeCmd.Flags().StringVar(&analyzePairMode, "pair-mode", "isolated", "pair separation mode: isolated or full")
}

var (
	analyzeLeakMode string
	analyzeFMin     float64
	analyzeFMax     float64
	analyzePairMode string
)

func runAnalyze(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	audioData, err := wav.ReadWAVChannels(inputFile, 4)
	if err != nil {
		return fmt.Errorf("failed to read input WAV: %w", err)
	}

	channelNames := []string{"LF", "RF", "LB", "RB"}
	fmt.Printf("Separation analysis (encode -> decode, isolated channels)\n")
	fmt.Printf("Input: %s\n", inputFile)
	if logic {
		fmt.Printf("Logic steering: enabled\n")
	}
	fmt.Printf("\nChannel  TargetRMS   LeakRMS  Sep(dB)\n")

	switch analyzeLeakMode {
	case string(metrics.LeakModeMax), string(metrics.LeakModeAvg):
	default:
		return fmt.Errorf("invalid leak-mode %q (use max or avg)", analyzeLeakMode)
	}
	switch analyzePairMode {
	case "isolated", "full":
	default:
		return fmt.Errorf("invalid pair-mode %q (use isolated or full)", analyzePairMode)
	}

	options := metrics.SeparationOptions{
		LeakMode:   metrics.LeakMode(analyzeLeakMode),
		SampleRate: int(audioData.SampleRate),
		FMin:       analyzeFMin,
		FMax:       analyzeFMax,
	}
	pairSeps := [4]float64{}

	var decodedFull [][]float64
	if analyzePairMode == "full" {
		fullEncoder := encoder.NewSQEncoderWithParams(blockSize, overlap)
		fullDecoder := decoder.NewSQDecoderWithParams(blockSize, overlap)
		fullDecoder.SetSampleRate(int(audioData.SampleRate))
		if logic {
			fullDecoder.EnableLogicSteering(true)
		}

		encodedFull, err := fullEncoder.Process(audioData.Samples)
		if err != nil {
			return fmt.Errorf("encoding failed: %w", err)
		}
		decodedFull, err = fullDecoder.Process(encodedFull)
		if err != nil {
			return fmt.Errorf("decoding failed: %w", err)
		}
	}

	for ch := 0; ch < 4; ch++ {
		isolated := make([][]float64, 4)
		for i := 0; i < 4; i++ {
			isolated[i] = make([]float64, audioData.NumSamples)
		}
		copy(isolated[ch], audioData.Samples[ch])

		sqEncoder := encoder.NewSQEncoderWithParams(blockSize, overlap)
		sqDecoder := decoder.NewSQDecoderWithParams(blockSize, overlap)
		sqDecoder.SetSampleRate(int(audioData.SampleRate))
		if logic {
			sqDecoder.EnableLogicSteering(true)
		}

		encoded, err := sqEncoder.Process(isolated)
		if err != nil {
			return fmt.Errorf("encoding failed: %w", err)
		}
		decoded, err := sqDecoder.Process(encoded)
		if err != nil {
			return fmt.Errorf("decoding failed: %w", err)
		}

		result := metrics.ChannelSeparation(decoded, ch, options)
		fmt.Printf("%-7s %9.6f %9.6f %7s\n",
			channelNames[ch],
			result.TargetRMS,
			result.LeakRMS,
			formatSeparation(result.SeparationDB),
		)

		if analyzePairMode == "isolated" {
			switch ch {
			case 0:
				pairSeps[ch] = metrics.ChannelPairSeparation(decoded, 0, 1, options).SeparationDB
			case 1:
				pairSeps[ch] = metrics.ChannelPairSeparation(decoded, 1, 0, options).SeparationDB
			case 2:
				pairSeps[ch] = metrics.ChannelPairSeparation(decoded, 2, 3, options).SeparationDB
			case 3:
				pairSeps[ch] = metrics.ChannelPairSeparation(decoded, 3, 2, options).SeparationDB
			}
		}
	}

	if analyzePairMode == "full" && decodedFull != nil {
		pairSeps[0] = metrics.ChannelPairSeparation(decodedFull, 0, 1, options).SeparationDB
		pairSeps[1] = metrics.ChannelPairSeparation(decodedFull, 1, 0, options).SeparationDB
		pairSeps[2] = metrics.ChannelPairSeparation(decodedFull, 2, 3, options).SeparationDB
		pairSeps[3] = metrics.ChannelPairSeparation(decodedFull, 3, 2, options).SeparationDB
	}

	fmt.Printf("\nPair separation (dB)\n")
	fmt.Printf("LF->RF: %s  RF->LF: %s  LB->RB: %s  RB->LB: %s\n",
		formatSeparation(pairSeps[0]),
		formatSeparation(pairSeps[1]),
		formatSeparation(pairSeps[2]),
		formatSeparation(pairSeps[3]),
	)

	return nil
}

func formatSeparation(sep float64) string {
	if math.IsInf(sep, 1) {
		return "+Inf"
	}
	if math.IsNaN(sep) {
		return "NaN"
	}
	return fmt.Sprintf("%.2f", sep)
}
