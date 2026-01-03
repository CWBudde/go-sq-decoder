# Float32 WAV Library Research for Go

**Date**: 2026-01-02  
**Context**: Investigating Go libraries for reading/writing WAV files with 32-bit IEEE float format (instead of 16-bit PCM).

---

## Current Library: youpy/go-wav

**Package**: `github.com/youpy/go-wav` v0.3.2  
**Current Usage**: Read 2-channel stereo, write 4-channel PCM

### Supported Formats (from README)

- **Format**: PCM, **IEEE float (read-only)**, G.711 A-law (read-only), G.711 µ-law (read-only)
- **Bits per sample**: 8-bit, 16-bit, 24-bit, 32-bit
- **Channels**: 1 (mono), 2 (stereo)

### Float32 Support

✅ **Read**: Yes (IEEE float read-only via `AudioFormatIEEEFloat = 3` constant)  
❌ **Write**: **No** - only PCM writing is supported

### Current Implementation in go-sq-tool

- Uses `wav.NewWriter(file, numSamples, 4, sampleRate, 16)` → **16-bit PCM**
- Samples are normalized float64 in `[][]float64` buffers, clamped to [-1, 1], then converted to int16

---

## Alternative 1: go-audio/wav

**Package**: `github.com/go-audio/wav` v1.1.0  
**Popularity**: 289 imports (vs. youpy/go-wav: 121 imports)  
**License**: Apache-2.0

### API Overview

```go
// Create encoder
encoder := wav.NewEncoder(w io.WriteSeeker, sampleRate, bitDepth, numChans, audioFormat int)

// Write buffer
encoder.Write(buf *audio.IntBuffer) error
encoder.Close() error
```

### Audio Format Constants (from decoder.go)

```go
const (
    AudioFormatPCM       = 1  // Linear PCM
    AudioFormatIEEEFloat = 3  // IEEE float
    AudioFormatALaw      = 6  // G.711 A-law
    AudioFormatMULaw     = 7  // G.711 µ-law
)
```

### Float32 Support

⚠️ **Unclear** - `NewEncoder` accepts `audioFormat int` parameter (could pass `AudioFormatIEEEFloat = 3`)  
❌ **Encoder implementation**: Only writes **IntBuffer** (int samples), not float samples

- Encoder.go lines 90-135 show only PCM int writing logic (8/16/24/32-bit integers)
- No float32 sample writing code path found in encoder.go

### Verdict

**Likely does NOT support writing IEEE float** despite having the constant. The encoder API uses `audio.IntBuffer`, not float buffers.

---

## Alternative 2: Manual WAV Writing

### Option: Implement Custom IEEE Float WAV Writer

Since WAV is a simple RIFF-based format, we could write a minimal IEEE float encoder:

**Pros**:

- Full control over format
- No dependency changes
- Can write exactly what we need (4-channel 32-bit float)

**Cons**:

- Need to implement RIFF/WAV header writing
- Need to handle endianness, chunk sizes, alignment
- More code to maintain

**Complexity**: ~100-150 lines for basic implementation

### WAV IEEE Float Format Structure

```
RIFF header
  ├─ fmt chunk
  │   ├─ AudioFormat: 3 (IEEE float)
  │   ├─ NumChannels: 4
  │   ├─ SampleRate: 44100
  │   ├─ BitsPerSample: 32
  │   └─ ByteRate, BlockAlign
  └─ data chunk
      └─ Float32 samples (little-endian)
```

**Reference**: http://www-mmsp.ece.mcgill.ca/Documents/AudioFormats/WAVE/WAVE.html

---

## Alternative 3: github.com/go-audio ecosystem

The `go-audio` organization maintains multiple related packages:

- `github.com/go-audio/wav` - WAV codec
- `github.com/go-audio/audio` - Core audio types (IntBuffer, FloatBuffer, Format)
- `github.com/go-audio/riff` - RIFF file format

### Checking for FloatBuffer Support

Looking at `github.com/go-audio/audio`, it provides:

- `IntBuffer` - integer sample buffer (current encoder uses this)
- Possibly `FloatBuffer` or similar?

**Action needed**: Check if `go-audio/audio` has FloatBuffer and if `go-audio/wav` can encode from it.

---

## Recommendations

### Short-term (keep current youpy/go-wav)

1. **Stay with 16-bit PCM** for now
2. Document in copilot-instructions.md that output is 16-bit only
3. Float processing stays internal (float64 buffers), output is quantized

### Medium-term (migrate to float32 output)

**Option A**: Implement custom IEEE float WAV writer

- ~150 lines of code
- Full control
- Based on existing RIFF/WAV specs

**Option B**: Check if `go-audio/wav` + `go-audio/audio.FloatBuffer` works

- Need to verify FloatBuffer exists and is supported by encoder
- May require PR/fork if not implemented

**Option C**: Find or create a library with proper float32 support

- Search for "golang wav ieee float write" on GitHub
- Check audio DSP projects (sox-like tools in Go)

### Long-term

Consider contributing float32 write support to `youpy/go-wav`:

- It already reads IEEE float
- Adding write support would help the community
- Could submit a PR

---

## Next Steps

1. ✅ Document current limitations in copilot-instructions.md
2. ⬜ Test if `go-audio/wav` + manual float buffer works
3. ⬜ Implement minimal custom IEEE float WAV writer if needed
4. ⬜ Add CLI flag `--output-format` (pcm16|float32) for future compatibility

---

## References

- youpy/go-wav: https://github.com/youpy/go-wav
- go-audio/wav: https://github.com/go-audio/wav
- go-audio/audio: https://github.com/go-audio/audio
- WAV Spec: http://www-mmsp.ece.mcgill.ca/Documents/AudioFormats/WAVE/WAVE.html
- IEEE float in WAV: https://en.wikipedia.org/wiki/WAV#Floating-point
