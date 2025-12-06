# SkyEye Development Guide for AI Assistants

SkyEye is an AI-powered GCI bot for DCS World that uses Whisper speech recognition, Tacview telemetry, and TTS to replace in-game AWACS with natural language command processing following real-world aviation brevity codes.

**Stack:** Go 1.25.5 + CGO, whisper.cpp (C++), Piper TTS (Windows/Linux), macOS Speech Synthesis, Tacview ACMI, SRS protocol

## Platform Support

| Platform | Arch | Status | TTS | Linking | Runtime Deps |
|----------|------|--------|-----|---------|--------------|
| **Windows** | AMD64 | ✅ | Piper (embedded) | Static | None - fully portable exe |
| **Linux** | AMD64 | ✅ | Piper (embedded) | Static whisper, dynamic opus/soxr | libopus0, libsoxr0 |
| **macOS** | ARM64 | ✅ | System (Neural Engine) | Static whisper, dynamic opus/soxr | Homebrew opus, libsoxr |
| macOS Intel | AMD64 | ❌ | - | - | No test hardware |

**Key Differences:**
- **Windows:** MUST build in MSYS2 UCRT64 (not cmd/PowerShell), static linking, portable binary
- **Linux:** Standard Unix build, requires runtime libraries, good for containers
- **macOS:** Requires Homebrew LLVM (not Apple Clang), Metal GPU acceleration (`GGML_METAL=1`), fastest speech recognition (0.09-1.5s), `--use-system-voice` flag available
- **AMD64 platforms:** AVX2 CPU required (Intel Haswell 2013+, AMD Excavator 2015+), 4+ dedicated cores recommended
- **Cross-compilation:** Not supported - must build on target platform

## Critical: Use Make, Not Go Commands

**❌ NEVER:** `go build`, `go test`, `go run`  
**✅ ALWAYS:** `make skyeye`, `make test`

**Why:** Requires `CGO_ENABLED=1`, `-tags nolibopusfile`, C include paths, platform-specific compilers/linker flags, and pre-built `libwhisper.a`. Direct `go` commands will fail.

## Make Targets

**Build:** `make skyeye`, `make skyeye-scaler`, `make whisper`, `make generate`  
**Test:** `make test`, `make benchmark-whisper`  
**Quality:** `make lint`, `make vet`, `make format` (all required for PR approval)  
**Clean:** `make mostlyclean`, `make clean`  
**Dependencies:** `make install-{msys2,macos,arch-linux,debian,fedora}-dependencies`

**DO NOT run the application** - it requires DCS server, SRS, human interaction, and multi-GB model files.

## CI Requirements (Must Pass)

```bash
make lint && make vet && make format && go mod tidy
git diff --exit-code  # Must have no changes
```

## Project Structure

```
cmd/skyeye/                    - Main entrypoint
cmd/skyeye-scaler/             - Autoscaler
pkg/                           - Public APIs
  recognizer/                  - Speech recognition interface
  simpleradio/                 - SRS protocol client
  synthesizer/speakers/        - Platform-specific TTS (macos.go, piper.go)
  tacview/                     - Telemetry parsing
  brevity/, parser/, composer/ - GCI command handling
internal/                      - Private packages
  application/                 - Platform detection & glue
  controller/, radar/, conf/   - Core logic
third_party/whisper.cpp/       - C++ dependency → libwhisper.a
```

**Architecture:** Players → SRS → simpleradio.Client → recognizer → parser → controller ← radar ← tacview ← DCS  
controller → composer → synthesizer (platform-specific) → simpleradio.Client → SRS

Platform-specific code isolated to `pkg/synthesizer/speakers/{macos,piper}.go` and Makefile. Runtime detection: `runtime.GOOS` ("darwin"/"windows"/"linux").

## Common Pitfalls

1. Using `go` commands directly instead of `make`
2. Wrong terminal on Windows (must be MSYS2 UCRT64)
3. Forgetting `make format` or `go mod tidy` before committing
4. Attempting to run the application (focus on unit tests)
5. Missing runtime libraries on Linux/macOS

## Quick Reference

```bash
# Setup (platform-specific, run once)
make install-{msys2,macos,debian}-dependencies

# Workflow
make whisper generate skyeye test

# Pre-commit (required)
make lint vet format && go mod tidy && git diff
```

## Resources

- `docs/PLAYER.md` - GCI commands
- `docs/ADMIN.md` - Deployment
- `docs/CONTRIBUTING.md` - Setup details