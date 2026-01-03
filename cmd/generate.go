package cmd

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cwbudde/go-sq-tool/internal/wav"
	"github.com/spf13/cobra"
)

var (
	genDuration  float64
	genRate      int
	genToneLevel float64
	genNoise     float64
)

var generateCmd = &cobra.Command{
	Use:   "generate-test [output.wav]",
	Short: "Generate a 4-channel test WAV with tones and noise",
	Args:  cobra.ExactArgs(1),
	RunE:  runGenerate,
}

func init() {
	generateCmd.Flags().Float64Var(&genDuration, "duration", 5.0, "duration in seconds")
	generateCmd.Flags().IntVar(&genRate, "rate", 44100, "sample rate in Hz")
	generateCmd.Flags().Float64Var(&genToneLevel, "tone-level", 0.6, "tone amplitude (0-1)")
	generateCmd.Flags().Float64Var(&genNoise, "noise-level", 0.05, "white noise amplitude (0-1)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	outputFile := args[0]
	if genDuration <= 0 {
		return fmt.Errorf("duration must be > 0")
	}
	if genRate <= 0 {
		return fmt.Errorf("rate must be > 0")
	}
	if genToneLevel < 0 || genToneLevel > 1 {
		return fmt.Errorf("tone-level must be between 0 and 1")
	}
	if genNoise < 0 || genNoise > 1 {
		return fmt.Errorf("noise-level must be between 0 and 1")
	}

	numSamples := int(genDuration * float64(genRate))
	if numSamples <= 0 {
		return fmt.Errorf("duration too short for sample rate")
	}

	freqs := []float64{100.0, 200.0, 400.0, 800.0}
	samples := make([][]float64, 4)
	for ch := range 4 {
		samples[ch] = make([]float64, numSamples)
	}

	rng := rand.New(rand.NewSource(1))
	for i := range numSamples {
		t := float64(i) / float64(genRate)
		for ch := range 4 {
			tone := genToneLevel * math.Sin(2.0*math.Pi*freqs[ch]*t)
			noise := genNoise * (rng.Float64()*2.0 - 1.0)
			samples[ch][i] = tone + noise
		}
	}

	audioData := &wav.AudioData{
		SampleRate: uint32(genRate),
		Samples:    samples,
		NumSamples: numSamples,
	}

	if float32 {
		return wav.WriteFloat32WAV(outputFile, audioData)
	}
	return wav.WriteWAV(outputFile, audioData)
}
