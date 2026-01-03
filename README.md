# SQ Quadrophonic Encoder/Decoder

A high-quality SQ (Stereo Quadrophonic) encoder/decoder written in Go, implementing the FFT-based SQ² algorithm for accurate 90° phase shifting and superior channel separation.

## Overview

**SQ (Stereo Quadraphonic)** is a matrix encoding system developed by CBS in the 1970s for encoding 4-channel quadrophonic audio into a 2-channel stereo signal. This tool can decode SQ-encoded stereo into four channels and encode four-channel quad audio into SQ-compatible stereo.

### Features

- ✅ **FFT-based Hilbert Transform**: Accurate 90° phase shift across all frequencies
- ✅ **High-quality decoding**: Superior channel separation using frequency-domain processing
- ✅ **SQ encoding**: Convert quad audio into SQ-compatible stereo
- ✅ **Simple CLI interface**: Easy to use command-line tool
- ✅ **WAV file support**: Standard WAV file I/O for compatibility
- ✅ **Configurable parameters**: Adjustable block size and overlap for quality/performance tuning

## Algorithm

This implementation is based on the **SQ² decoder** algorithm which uses:

1. **FFT-based Hilbert Transform** for precise 90° phase shifting
2. **SQ Decode Matrix**:
   - LF (Left Front) = LT (pass through)
   - RF (Right Front) = RT (pass through)
   - LB (Left Back) = 0.707 × H(LT) - 0.707 × RT
   - RB (Right Back) = 0.707 × LT - 0.707 × H(RT)

Where `H()` denotes the Hilbert transform (90° phase shift) and 0.707 ≈ √2/2.

See [`sq-decoder-explained.md`](./sq-decoder-explained.md) for comprehensive technical documentation.

## Installation

### Prerequisites

- Go 1.21 or later

### Build from source

```bash
git clone https://github.com/cwbudde/go-sq-decoder.git
cd go-sq-decoder
go mod download
go build -o go-sq-decoder
```

## Usage

### Basic Usage (Decode)

```bash
go-sq-decoder input.wav output.wav
```

**Input**: 2-channel stereo WAV file (SQ-encoded)
**Output**: 4-channel quadrophonic WAV file (LF, RF, LB, RB)

### Decode (Explicit)

```bash
go-sq-decoder decode input.wav output.wav
```

### Encode (Quad to SQ Stereo)

```bash
go-sq-decoder encode quad_input.wav sq_output.wav
```

**Input**: 4-channel quadrophonic WAV file (LF, RF, LB, RB)
**Output**: 2-channel stereo WAV file (LT, RT)

### Verbose Output

```bash
go-sq-decoder -v input.wav output.wav
```

Shows detailed information about processing:

- Input file properties (sample rate, duration)
- Decoder configuration (block size, latency)
- Processing status

### Custom Parameters

```bash
go-sq-decoder -b 2048 -o 1024 input.wav output.wav
```

- `-b, --block-size`: FFT block size (default: 1024, must be power of 2)
- `-o, --overlap`: Overlap in samples (default: 512, typically blockSize/2)

### Help

```bash
go-sq-decoder --help
```

## Examples

### Decode SQ record to quadrophonic

```bash
# Basic decode
go-sq-decoder my_sq_recording.wav quad_output.wav

# Verbose mode to see processing details
go-sq-decoder -v my_sq_recording.wav quad_output.wav

# Higher quality (larger FFT, more latency)
go-sq-decoder -b 2048 -o 1024 input.wav output.wav
```

### Encode quad to SQ stereo

```bash
# Encode 4-channel WAV to SQ stereo
go-sq-decoder encode quad_input.wav sq_output.wav

# Float output for headroom
go-sq-decoder encode --float32 quad_input.wav sq_output.wav
```

## Technical Details

### Decoder Characteristics

| Parameter              | Value                           |
| ---------------------- | ------------------------------- |
| **Algorithm**          | FFT-based SQ²                   |
| **Default Block Size** | 1024 samples                    |
| **Default Overlap**    | 512 samples (50%)               |
| **Latency**            | 768 samples (~17.4ms @ 44.1kHz) |
| **Phase Shift Method** | Hilbert transform via FFT       |
| **Input Channels**     | 2 (stereo)                      |
| **Output Channels**    | 4 (quadrophonic)                |

### Encoder Characteristics

| Parameter              | Value                           |
| ---------------------- | ------------------------------- |
| **Algorithm**          | FFT-based SQ                    |
| **Default Block Size** | 1024 samples                    |
| **Default Overlap**    | 512 samples (50%)               |
| **Latency**            | 768 samples (~17.4ms @ 44.1kHz) |
| **Phase Shift Method** | Hilbert transform via FFT       |
| **Input Channels**     | 4 (quadrophonic)                |
| **Output Channels**    | 2 (stereo)                      |

### Channel Layout

**Input (SQ-encoded stereo)**:

- Channel 0: LT (Left Total)
- Channel 1: RT (Right Total)

**Output (Quadrophonic)**:

- Channel 0: LF (Left Front)
- Channel 1: RF (Right Front)
- Channel 2: LB (Left Back)
- Channel 3: RB (Right Back)

**Output (SQ-encoded stereo)**:

- Channel 0: LT (Left Total)
- Channel 1: RT (Right Total)

### Dependencies

- [`github.com/MeKo-Christian/algo-fft`](https://github.com/MeKo-Christian/algo-fft) - FFT implementation
- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) - CLI framework
- [`github.com/youpy/go-wav`](https://github.com/youpy/go-wav) - WAV file I/O

## Performance

The FFT-based SQ² decoder provides:

**Advantages**:

- ✅ Accurate 90° phase shift across audio spectrum
- ✅ Excellent channel separation (30-40 dB estimated)
- ✅ Predictable frequency response
- ✅ Superior quality for archival restoration

**Tradeoffs**:

- ⏱️ Higher latency (~17ms) vs. recursive filter approach
- 💻 Moderate CPU usage (FFT operations)
- 📊 Block-based processing

For real-time applications requiring minimal latency, consider implementing the simpler recursive filter variant (not included in this tool).

## Project Structure

```
go-sq-decoder/
├── main.go                      # Entry point
├── cmd/
│   └── root.go                  # CLI implementation (Cobra)
├── internal/
│   ├── encoder/
│   │   └── encoder.go          # SQ encoder core algorithm
│   ├── decoder/
│   │   └── decoder.go          # SQ decoder core algorithm
│   └── wav/
│       └── wav.go              # WAV file I/O
├── pkg/
│   └── sqmath/
│       └── hilbert.go          # Hilbert transform (FFT-based)
├── sq-decoder-explained.md     # Technical documentation
└── README.md
```

## Contributing

Contributions welcome! Please feel free to submit issues or pull requests.

### Areas for Enhancement

- [ ] Add basic recursive filter decoder (low-latency variant)
- [ ] Support for other audio formats (FLAC, MP3, etc.)
- [ ] Real-time processing mode
- [ ] GUI frontend
- [ ] Batch processing
- [ ] Channel separation quality metrics

## References

- **CBS SQ System**: Original matrix quadraphonic system from 1970s
- **Hilbert Transform**: Linear operator for 90° phase shifting
- **FFT**: Fast Fourier Transform for frequency-domain processing

See [`sq-decoder-explained.md`](./sq-decoder-explained.md) for detailed technical documentation including:

- Complete algorithm explanation
- Mathematical foundations
- Historical context
- Implementation comparisons

## License

MIT License - feel free to use for any purpose.

## Acknowledgments

- Based on the SQ² decoder implementation from Delphi/VST plugins
- Algorithm documentation reverse-engineered from working code
- Uses FFT library by MeKo-Christian

---

**Author**: Christian Budde (cwbudde)
**Version**: 1.0.0
**Date**: 2026-01-02
