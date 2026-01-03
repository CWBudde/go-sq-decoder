package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/youpy/go-wav"
)

// AudioData represents multi-channel audio data
type AudioData struct {
	SampleRate uint32
	Samples    [][]float64 // [channel][sample]
	NumSamples int
}

// ReadWAV reads a stereo WAV file and returns the audio data
func ReadWAV(filename string) (*AudioData, error) {
	return ReadWAVChannels(filename, 2)
}

// ReadWAVChannels reads a WAV file with a specific channel count
func ReadWAVChannels(filename string, channels int) (*AudioData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}
	defer file.Close()

	reader := wav.NewReader(file)
	format, err := reader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV format: %w", err)
	}

	if format.NumChannels != uint16(channels) {
		return nil, fmt.Errorf("input must have %d channels, got %d channels", channels, format.NumChannels)
	}

	samplesByChannel := make([][]float64, channels)

	for {
		samples, err := reader.ReadSamples()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read samples: %w", err)
		}

		for _, sample := range samples {
			for ch := 0; ch < channels; ch++ {
				val := reader.FloatValue(sample, ch)
				samplesByChannel[ch] = append(samplesByChannel[ch], float64(val))
			}
		}
	}

	return &AudioData{
		SampleRate: format.SampleRate,
		Samples:    samplesByChannel,
		NumSamples: len(samplesByChannel[0]),
	}, nil
}

// WriteWAV writes 4-channel audio data to a WAV file
func WriteWAV(filename string, data *AudioData) error {
	return writeWAVPCM16(filename, data, 4)
}

// WriteStereoWAV writes 2-channel audio data to a WAV file
func WriteStereoWAV(filename string, data *AudioData) error {
	return writeWAVPCM16(filename, data, 2)
}

func writeWAVPCM16(filename string, data *AudioData, channels int) error {
	if len(data.Samples) != channels {
		return fmt.Errorf("output must have %d channels, got %d", channels, len(data.Samples))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	writer := wav.NewWriter(file, uint32(data.NumSamples), uint16(channels), data.SampleRate, 16)

	// Write samples
	for i := 0; i < data.NumSamples; i++ {
		samples := make([]wav.Sample, channels)
		for ch := 0; ch < channels; ch++ {
			// Clamp to [-1.0, 1.0] and convert to int16
			val := data.Samples[ch][i]
			if val > 1.0 {
				val = 1.0
			}
			if val < -1.0 {
				val = -1.0
			}
			samples[ch].Values[0] = int(val * 32767.0)
		}
		if err := writer.WriteSamples(samples); err != nil {
			return fmt.Errorf("failed to write samples: %w", err)
		}
	}

	return nil
}

// WriteFloat32WAV writes 4-channel audio data to a WAV file in 32-bit IEEE float format
func WriteFloat32WAV(filename string, data *AudioData) error {
	return writeWAVFloat32(filename, data, 4)
}

// WriteStereoFloat32WAV writes 2-channel audio data to a WAV file in 32-bit IEEE float format
func WriteStereoFloat32WAV(filename string, data *AudioData) error {
	return writeWAVFloat32(filename, data, 2)
}

func writeWAVFloat32(filename string, data *AudioData, channels int) error {
	if len(data.Samples) != channels {
		return fmt.Errorf("output must have %d channels, got %d", channels, len(data.Samples))
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	numChannels := uint16(channels)
	bitsPerSample := uint16(32)
	byteRate := data.SampleRate * uint32(numChannels) * uint32(bitsPerSample/8)
	blockAlign := numChannels * (bitsPerSample / 8)
	audioFormat := uint16(3) // IEEE float
	dataSize := uint32(data.NumSamples) * uint32(numChannels) * uint32(bitsPerSample/8)

	// Write RIFF header
	if err := writeString(file, "RIFF"); err != nil {
		return fmt.Errorf("failed to write RIFF header: %w", err)
	}
	// File size - 8 (will be updated at the end if needed)
	if err := binary.Write(file, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return fmt.Errorf("failed to write file size: %w", err)
	}
	if err := writeString(file, "WAVE"); err != nil {
		return fmt.Errorf("failed to write WAVE header: %w", err)
	}

	// Write fmt chunk
	if err := writeString(file, "fmt "); err != nil {
		return fmt.Errorf("failed to write fmt chunk ID: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(16)); err != nil { // fmt chunk size
		return fmt.Errorf("failed to write fmt chunk size: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, audioFormat); err != nil {
		return fmt.Errorf("failed to write audio format: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, numChannels); err != nil {
		return fmt.Errorf("failed to write num channels: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, data.SampleRate); err != nil {
		return fmt.Errorf("failed to write sample rate: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, byteRate); err != nil {
		return fmt.Errorf("failed to write byte rate: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, blockAlign); err != nil {
		return fmt.Errorf("failed to write block align: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, bitsPerSample); err != nil {
		return fmt.Errorf("failed to write bits per sample: %w", err)
	}

	// Write data chunk
	if err := writeString(file, "data"); err != nil {
		return fmt.Errorf("failed to write data chunk ID: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, dataSize); err != nil {
		return fmt.Errorf("failed to write data size: %w", err)
	}

	// Write interleaved float32 samples
	for i := 0; i < data.NumSamples; i++ {
		for ch := 0; ch < channels; ch++ {
			val := data.Samples[ch][i]
			// Clamp to [-1.0, 1.0] to prevent invalid float values
			if val > 1.0 {
				val = 1.0
			} else if val < -1.0 {
				val = -1.0
			} else if math.IsNaN(val) || math.IsInf(val, 0) {
				val = 0.0
			}

			if err := binary.Write(file, binary.LittleEndian, float32(val)); err != nil {
				return fmt.Errorf("failed to write sample data: %w", err)
			}
		}
	}

	return nil
}

// writeString writes a string to the writer without a null terminator
func writeString(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}
