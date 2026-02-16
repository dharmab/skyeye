# Replace Whisper STT with Parakeet TDT via sherpa-onnx

## Context

SkyEye currently uses whisper.cpp for local speech recognition, with OpenAI API backends as alternatives. The whisper.cpp integration is the most complex part of the build system - it requires building a C++ static library from source, platform-specific compilers (Homebrew LLVM on macOS, MSYS2 UCRT64 on Windows), and careful CGO configuration. This change replaces all STT backends with a single Parakeet TDT recognizer via sherpa-onnx, which provides pre-built shared libraries through Go modules, eliminating the whisper.cpp build complexity while using a newer, high-quality speech recognition model.

## Approach

Use `github.com/k2-fsa/sherpa-onnx-go` with platform-specific packages (`sherpa-onnx-go-linux`, `sherpa-onnx-go-macos`, `sherpa-onnx-go-windows`) that embed pre-built shared libraries. The Parakeet TDT 0.6B v3 INT8 model (encoder/decoder/joiner/tokens) runs via ONNX Runtime through sherpa-onnx's offline recognizer API.

Key tradeoffs:
- Parakeet has no prompt/hint support (aviation brevity hints are lost)
- Model is 4 files (~640MB total) instead of 1 file (~466MB)
- Still requires CGO (for opus codec and sherpa-onnx), but no source compilation of C++ dependencies

## Step 1: Add sherpa-onnx dependency

Add to `go.mod`:
```
github.com/k2-fsa/sherpa-onnx-go v1.12.25
github.com/k2-fsa/sherpa-onnx-go-linux v1.12.25
github.com/k2-fsa/sherpa-onnx-go-macos v1.12.25
github.com/k2-fsa/sherpa-onnx-go-windows v1.12.25
```

Remove:
```
github.com/ggerganov/whisper.cpp/bindings/go
github.com/openai/openai-go
```

Run `go get` and `go mod tidy`.

## Step 2: Create Parakeet recognizer

**New file**: `pkg/recognizer/parakeet.go`

```go
type parakeetRecognizer struct {
    recognizer *sherpa.OfflineRecognizer
}

func NewParakeetRecognizer(modelDir string) (Recognizer, error)
func (r *parakeetRecognizer) Recognize(ctx context.Context, pcm []float32, enableTranscriptionLogging bool) (string, error)
```

- Import: `sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"`
- Config: `OfflineRecognizerConfig` with `FeatConfig{SampleRate: 16000, FeatureDim: 80}`, `ModelConfig.Transducer` pointing to encoder/decoder/joiner ONNX files, `ModelConfig.Tokens` to tokens.txt, `ModelConfig.ModelType: "nemo_transducer"`, `DecodingMethod: "greedy_search"`
- Per-recognition: create `OfflineStream` via `NewOfflineStream(recognizer)`, call `AcceptWaveform(16000, pcm)`, `Decode(stream)`, `GetResult()` for text
- Audio format: `[]float32` at 16kHz, same as current pipeline - no conversion needed
- Delete stream after each recognition

## Step 3: Delete old recognizer implementations

**Delete files**:
- `pkg/recognizer/whisper.go`
- `pkg/recognizer/openai.go`
- `pkg/recognizer/prompt.go`
- `pkg/recognizer/whisper_test.go`

**Keep**: `pkg/recognizer/recognizer.go` (interface unchanged)

## Step 4: Update configuration

**File**: `internal/conf/configuration.go`
- Remove import of `github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper`
- Remove `Recognizer` type and constants (`WhisperLocal`, `WhisperAPI`, `GPT4o`, `GPT4oMini`)
- Remove fields: `Recognizer Recognizer`, `WhisperModel *whisper.Model`, `OpenAIAPIKey string`
- Add field: `ModelDirectory string`

## Step 5: Update CLI

**File**: `cmd/skyeye/main.go`
- Remove imports: `whisper` package, `golang.org/x/sys/cpu`
- Remove variables: `recognizerName`, `whisperModelPath`, `recognizerLockPath`, `openAIAPIKey`
- Add variable: `modelDirectory string`
- Remove flags: `--recognizer`, `--whisper-model`, `--openai-api-key`, `--recognizer-lock-path`, and the `MarkFlagsOneRequired` call
- Add flag: `--model` (required, path to directory with Parakeet model files)
- Delete `loadWhisperModel()` function entirely (lines 246-267)
- In `run()`: remove `whisperModel := loadWhisperModel()`, `recognizerLock := loadLock(recognizerLockPath)`
- Update `conf.Configuration{}` struct literal: remove `Recognizer`, `RecognizerLock`, `WhisperModel`, `OpenAIAPIKey`; add `ModelDirectory: modelDirectory`

## Step 6: Update application wiring

**File**: `internal/application/app.go`
- Remove `recognizerLock *flock.Flock` from `Application` struct
- Replace recognizer switch (lines 156-169) with:
  ```go
  speechRecognizer, err := recognizer.NewParakeetRecognizer(config.ModelDirectory)
  if err != nil {
      return nil, fmt.Errorf("failed to construct speech recognizer: %w", err)
  }
  ```
- Update `Application` struct initialization to remove `recognizerLock`
- Check for and remove any usage of `recognizerLock` in the recognize flow (`internal/application/recognize.go`)

**File**: `internal/application/recognize.go`
- Remove recognizer lock acquisition/release if present

## Step 7: Update configuration struct

**File**: `internal/conf/configuration.go`
- Remove `RecognizerLock *flock.Flock` field (no longer needed - sherpa-onnx handles thread safety)

## Step 8: Update Makefile

**File**: `Makefile`
- Remove variables: `WHISPER_CPP_PATH`, `LIBWHISPER_PATH`, `WHISPER_H_PATH`, `WHISPER_CPP_REPO`, `WHISPER_CPP_VERSION`, `WHISPER_CPP_BUILD_ENV`, `ABS_WHISPER_CPP_PATH`
- Update `BUILD_VARS`: Remove `C_INCLUDE_PATH` and `LIBRARY_PATH`. Keep `CGO_ENABLED=1`.
- macOS section: Remove `WHISPER_CPP_BUILD_ENV = GGML_METAL=1`. Test whether `CC`/`CXX` LLVM overrides are still needed (sherpa-onnx likely doesn't need OpenMP, but opus might). If not needed, remove.
- Remove targets: `download-whisper-%`, `$(LIBWHISPER_PATH) $(WHISPER_H_PATH)`, `whisper`, `benchmark-whisper`
- Update `$(SKYEYE_BIN)` dependency: remove `$(LIBWHISPER_PATH) $(WHISPER_H_PATH)`
- Update `lint`: remove `whisper` dependency
- Update `clean`: remove `rm -rf "$(WHISPER_CPP_PATH)"`

## Step 9: Update CI

**Delete**: `.github/actions/build-whisper/` directory

**File**: `.github/workflows/skyeye.yaml`
- Remove all "Build whisper.cpp" steps from: lint, test, build-linux-amd64, build-macos-arm64, build-windows-amd64 jobs
- For release dist creation, add sherpa-onnx shared library bundling:
  - Linux: Copy `libsherpa-onnx-c-api.so` and `libonnxruntime.so*` from Go module cache into dist `lib/` directory
  - macOS: Copy `.dylib` equivalents
  - Windows: Copy `.dll` equivalents into dist root (same directory as .exe)

## Step 10: Update Dockerfile

**File**: `Dockerfile`
- Remove `COPY third_party third_party` and `RUN make whisper`
- After build, extract sherpa-onnx shared libraries from Go module cache and copy to runtime image
- Add `RUN ldconfig` in runtime stage after copying libraries

## Step 11: Delete whisper.cpp

**Delete**: `third_party/whisper.cpp/` directory entirely (via `rm -rf` or git rm)

## Step 12: Update benchmark test

**New file**: `pkg/recognizer/parakeet_test.go`
- Port from `whisper_test.go` structure
- Use env var `SKYEYE_PARAKEET_MODEL` for model directory path
- `BenchmarkParakeetRecognizer` - same WAV loading, use `NewParakeetRecognizer`
- Update Makefile `benchmark-whisper` → `benchmark-parakeet` target

## Step 13: Update documentation and config

- `config.yaml`: Replace `whisper-model` and `recognizer` with `model: /path/to/parakeet-models/`
- `cmd/skyeye/main.go` command examples: Update to use `--model` flag
- `docs/ADMIN.md`: Update model setup instructions
- `docs/CONTRIBUTING.md`: Update build instructions (no whisper.cpp step)
- `CLAUDE.md`: Update build workflow, remove whisper references

## Verification

1. `go mod tidy` - ensure dependencies resolve
2. `make lint && make vet && make format` - code quality passes
3. `make test` - all tests pass
4. `make skyeye` - binary builds successfully on local platform
5. `git diff --exit-code` after format/tidy - no uncommitted changes
6. CI: Verify builds pass on all 3 platforms (Linux, macOS, Windows)
