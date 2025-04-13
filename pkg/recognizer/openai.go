package recognizer

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/dharmab/skyeye/pkg/pcm"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
)

type openAIRecognizer struct {
	callsign string
	client   *openai.Client
	model    string
}

var _ Recognizer = &openAIRecognizer{}

func newOpenAIRecognizer(apiKey, model, callsign string) Recognizer {
	return &openAIRecognizer{
		callsign: callsign,
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func NewWhisperAPIRecognizer(apiKey, callsign string) Recognizer {
	return newOpenAIRecognizer(apiKey, "whisper-1", callsign)
}

// NewGPT4oRecognizer creates a new recognizer using OpenAI Platform's GPT-4o model.
func NewGPT4oRecognizer(apiKey, callsign string) Recognizer {
	return newOpenAIRecognizer(apiKey, "gpt-4o-transcribe", callsign)
}

// NewGPT4oMiniRecognizer creates a new recognizer using OpenAI Platform's GPT-4o Mini model.
func NewGPT4oMiniRecognizer(apiKey, callsign string) Recognizer {
	return newOpenAIRecognizer(apiKey, "gpt-4o-mini-transcribe", callsign)
}

// NewOpenAIRecognizer creates a new recognizer using OpenAI Platform.
//
// Deprecated: Use NewWhisperAPIRecognizer, NewGPT4oRecognizer, or NewGPT4oMiniRecognizer instead.
func NewOpenAIRecognizer(apiKey, callsign string) Recognizer { // nolint: revive // Ignore deprecated function
	return NewWhisperAPIRecognizer(apiKey, callsign)
}

// Recognize implements [Recognizer.Recognize] using OpenAI Platform's hosted GPT4 transcription model.
func (r *openAIRecognizer) Recognize(ctx context.Context, sample []float32, _ bool) (string, error) {
	log.Debug().Msg("creating WAV from sample")
	buf, err := createWAV(sample)
	if err != nil {
		return "", fmt.Errorf("error creating WAV: %w", err)
	}

	body := openai.AudioTranscriptionNewParams{
		File:     openai.FileParam(buf, "audio.wav", "audio/wav"),
		Model:    openai.String(r.model),
		Language: openai.String("en"),
		Prompt:   openai.String(prompt(r.callsign)),
	}

	log.Info().Str("model", r.model).Msg("calling OpenAI Audio Transcriptions API")
	transcription, err := r.client.Audio.Transcriptions.New(ctx, body)
	if err != nil {
		return "", fmt.Errorf("error transcribing audio: %w", err)
	}
	return transcription.Text, nil
}

// createWAV creates a RIFF WAV file from a 16KHz mono audio sample.
func createWAV(sample []float32) (*bytes.Buffer, error) {
	const (
		sampleRate     = 16000
		channels       = 1
		bitsPerSample  = 16
		bytesPerSample = bitsPerSample / 8
		bytesPerBlock  = channels * bitsPerSample / 8
		bytesPerSecond = sampleRate * bytesPerBlock
	)

	data := pcm.F32toS16LE(sample)
	dataSize := len(data) * bytesPerSample

	buf := new(bytes.Buffer)
	var writeErr error
	_, err := buf.WriteString("RIFF")
	if err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	// File size (placeholder for now)
	if err := writeBinary(buf, int32(0)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if _, err := buf.WriteString("WAVE"); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if _, err := buf.WriteString("fmt "); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	// Remaining size of the fmt chunk = 16 bytes
	if err := writeBinary(buf, int32(16)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	// Audio format (PCM integer=1)
	if err := writeBinary(buf, int16(1)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if err := writeBinary(buf, int16(channels)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if err := writeBinary(buf, int32(sampleRate)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if err := writeBinary(buf, int32(bytesPerSecond)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if err := writeBinary(buf, int16(bytesPerBlock)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if err := writeBinary(buf, int16(bitsPerSample)); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if _, err := buf.WriteString("data"); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	if err := writeBinary(buf, dataSize); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	for _, d := range data {
		if err := writeBinary(buf, d); err != nil {
			writeErr = errors.Join(writeErr, err)
		}
	}

	// Update file size
	fileSize := buf.Len() - 8
	fileSizeBytes := new(bytes.Buffer)
	if err := writeBinary(fileSizeBytes, fileSize); err != nil {
		writeErr = errors.Join(writeErr, err)
	}
	copy(buf.Bytes()[4:8], fileSizeBytes.Bytes())

	if writeErr != nil {
		return nil, writeErr
	}
	return buf, nil
}

// writeBinary is a helper function to write binary data to a buffer in
// little-endian order.
func writeBinary(w *bytes.Buffer, data any) error {
	return binary.Write(w, binary.LittleEndian, data)
}
