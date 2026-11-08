package main

import (
	"strings"
	"testing"

	"github.com/dharmab/skyeye/internal/conf"
)

func TestValidateRecognizer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		recognizer conf.Recognizer
		goos       string
		model      string
		key        string
		wantError  string
	}{
		{name: "Apple on macOS", recognizer: conf.AppleSpeech, goos: "darwin"},
		{name: "Apple on Linux", recognizer: conf.AppleSpeech, goos: "linux", wantError: "only supported on macOS"},
		{name: "Apple on Windows", recognizer: conf.AppleSpeech, goos: "windows", wantError: "only supported on macOS"},
		{name: "local Whisper", recognizer: conf.WhisperLocal, model: "whisper.bin"},
		{name: "local Whisper requires model", recognizer: conf.WhisperLocal, key: "key", wantError: "whisper-model is required"},
		{name: "Whisper API", recognizer: conf.WhisperAPI, key: "key"},
		{name: "Whisper API requires key", recognizer: conf.WhisperAPI, model: "whisper.bin", wantError: "openai-api-key is required"},
		{name: "GPT4o", recognizer: conf.GPT4o, key: "key"},
		{name: "GPT4o requires key", recognizer: conf.GPT4o, wantError: "openai-api-key is required"},
		{name: "GPT4o Mini", recognizer: conf.GPT4oMini, key: "key"},
		{name: "GPT4o Mini requires key", recognizer: conf.GPT4oMini, wantError: "openai-api-key is required"},
		{name: "unknown", recognizer: "unknown", wantError: "unknown recognizer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateRecognizer(tt.recognizer, tt.goos, tt.model, tt.key)
			if tt.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}

//nolint:paralleltest // Mutates the package-level recognizer flag and Whisper model path.
func TestAppleSpeechSkipsWhisperModel(t *testing.T) {
	originalRecognizer, originalPath := recognizerName, whisperModelPath
	t.Cleanup(func() {
		recognizerName, whisperModelPath = originalRecognizer, originalPath
	})

	if err := skyeye.Flags().Lookup("recognizer").Value.Set(string(conf.AppleSpeech)); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", "nonexistent-whisper-model.bin"} {
		whisperModelPath = path
		if model := loadWhisperModel(); model != nil {
			t.Fatal("apple-speech must not load a Whisper model")
		}
	}
	if err := skyeye.ValidateFlagGroups(); err != nil {
		t.Fatalf("apple-speech must not require model or API key flags: %v", err)
	}
}
