//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/cwbudde/go-sq-tool/internal/decoder"
	"github.com/cwbudde/go-sq-tool/internal/wav"
)

type decodeOptions struct {
	BlockSize int
	Overlap   int
	Logic     bool
	Float32   bool
}

var decodeFunc js.Func

func main() {
	decodeFunc = js.FuncOf(decodeWavJS)
	js.Global().Set("sqDecodeWav", decodeFunc)
	select {}
}

func decodeWavJS(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{"error": "missing input wav bytes"}
	}

	inputBytes, err := valueToBytes(args[0])
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	opts := parseOptions(args)
	outputBytes, err := decodeWavBytes(inputBytes, opts)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	outArray := js.Global().Get("Uint8Array").New(len(outputBytes))
	js.CopyBytesToJS(outArray, outputBytes)
	return map[string]interface{}{"data": outArray}
}

func parseOptions(args []js.Value) decodeOptions {
	opts := decodeOptions{
		BlockSize: decoder.DefaultBlockSize,
		Overlap:   decoder.DefaultOverlap,
	}
	if len(args) < 2 {
		return opts
	}
	raw := args[1]
	if raw.Type() != js.TypeObject {
		return opts
	}
	if v := raw.Get("blockSize"); v.Type() == js.TypeNumber && v.Int() > 0 {
		opts.BlockSize = v.Int()
	}
	if v := raw.Get("overlap"); v.Type() == js.TypeNumber && v.Int() > 0 {
		opts.Overlap = v.Int()
	}
	if v := raw.Get("logic"); v.Type() == js.TypeBoolean {
		opts.Logic = v.Bool()
	}
	if v := raw.Get("float32"); v.Type() == js.TypeBoolean {
		opts.Float32 = v.Bool()
	}
	return opts
}

func valueToBytes(v js.Value) ([]byte, error) {
	uint8Array := js.Global().Get("Uint8Array")
	if v.InstanceOf(uint8Array) {
		buf := make([]byte, v.Get("length").Int())
		js.CopyBytesToGo(buf, v)
		return buf, nil
	}

	arrayBuffer := js.Global().Get("ArrayBuffer")
	if v.InstanceOf(arrayBuffer) {
		view := uint8Array.New(v)
		buf := make([]byte, view.Get("length").Int())
		js.CopyBytesToGo(buf, view)
		return buf, nil
	}

	return nil, errors.New("expected Uint8Array or ArrayBuffer input")
}

func decodeWavBytes(input []byte, opts decodeOptions) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty input")
	}

	audioData, err := wav.ReadWAVBytes(input, 2)
	if err != nil {
		return nil, fmt.Errorf("read wav: %w", err)
	}

	sqDecoder := decoder.NewSQDecoderWithParams(opts.BlockSize, opts.Overlap)
	sqDecoder.SetSampleRate(int(audioData.SampleRate))
	if opts.Logic {
		sqDecoder.EnableLogicSteering(true)
	}

	output, err := sqDecoder.Process(audioData.Samples)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	outputData := &wav.AudioData{
		SampleRate: audioData.SampleRate,
		Samples:    output,
		NumSamples: audioData.NumSamples,
	}

	var buf bytes.Buffer
	if opts.Float32 {
		if err := wav.WriteFloat32WAVToWriter(&buf, outputData); err != nil {
			return nil, fmt.Errorf("write wav: %w", err)
		}
	} else {
		if err := wav.WriteWAVToWriter(&buf, outputData); err != nil {
			return nil, fmt.Errorf("write wav: %w", err)
		}
	}

	return buf.Bytes(), nil
}
