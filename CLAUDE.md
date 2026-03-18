# SkyEye Development Guide for AI Assistants

SkyEye is an AI-powered GCI bot for DCS World that uses Parakeet TDT speech recognition (via sherpa-onnx), Tacview telemetry, and TTS to replace in-game AWACS with natural language command processing following real-world aviation brevity codes.

**Stack:** Go 1.26 + CGO, sherpa-onnx (Parakeet TDT + Pocket TTS), Tacview ACMI, SRS protocol

## Platform Support

| Platform | Arch | Status | TTS | Linking | Runtime Deps |
|----------|------|--------|-----|---------|--------------|
| **Windows** | AMD64 | ✅ | Pocket TTS (sherpa-onnx) | Static | None - fully portable exe |
| **Linux** | AMD64 | ✅ | Pocket TTS (sherpa-onnx) | Dynamic opus/soxr | libopus0, libsoxr0 |
| **macOS** | ARM64 | ✅ | Pocket TTS (sherpa-onnx) | Dynamic opus/soxr | Homebrew opus, libsoxr |
| macOS Intel | AMD64 | ❌ | - | - | No test hardware |

**Key Differences:**
- **Windows:** MUST build in MSYS2 UCRT64 (not cmd/PowerShell), static linking, portable binary
- **Linux:** Standard Unix build, requires runtime libraries, good for containers
- **macOS:** Uses Apple Clang (system compiler)
- **Cross-compilation:** Not supported - must build on target platform

## Critical: Use Make, Not Go Commands

**❌ NEVER:** `go build`, `go test`, `go run`
**✅ ALWAYS:** `make skyeye`, `make test`

**Why:** Requires `CGO_ENABLED=1`, `-tags nolibopusfile`, and platform-specific compilers/linker flags. Direct `go` commands will fail.

## Make Targets

**Build:** `make skyeye`, `make skyeye-scaler`, `make generate`
**Test:** `make test`, `make benchmark-parakeet`
**Quality:** `make lint`, `make vet`, `make fix`, `make format` (all required for PR approval)
**Clean:** `make mostlyclean`, `make clean`
**Dependencies:** `make install-{msys2,macos,arch-linux,debian,fedora}-dependencies`

**DO NOT run the application** - it requires DCS server, SRS, human interaction, and the Parakeet model files in pkg/recognizer/model/.

## CI Requirements (Must Pass)

```bash
make lint && make vet && make fix && make format && go mod tidy
git diff --exit-code  # Must have no changes
```

## Project Structure

```
cmd/skyeye/                    - Main entrypoint
cmd/skyeye-scaler/             - Autoscaler
pkg/                           - Public APIs
  recognizer/                  - Speech recognition (Parakeet TDT via sherpa-onnx)
  recognizer/model/            - Embedded model files (encoder/decoder/joiner ONNX + tokens.txt)
  simpleradio/                 - SRS protocol client
  synthesizer/pocket/          - Pocket TTS speaker (sherpa-onnx)
  synthesizer/pocket/model/    - TTS model download/verify (no CGO)
  synthesizer/pocket/voice/    - Embedded default reference voice (no CGO)
  synthesizer/speakers/        - Speaker interface + resampling helpers
  tacview/                     - Telemetry parsing
  brevity/, parser/, composer/ - GCI command handling
internal/                      - Private packages
  application/                 - Application glue
  controller/, radar/, conf/   - Core logic
```

**Architecture:** Players → SRS → simpleradio.Client → recognizer → parser → controller ← radar ← tacview ← DCS
controller → composer → synthesizer → simpleradio.Client → SRS

## Common Pitfalls

1. Using `go` commands directly instead of `make`
2. Wrong terminal on Windows (must be MSYS2 UCRT64)
3. Forgetting `make fix`, `make format`, or `go mod tidy` before committing
4. Attempting to run the application (focus on unit tests)
5. Missing runtime libraries on Linux/macOS
6. Missing model files in pkg/recognizer/model/ (required for building)

## Quick Reference

```bash
# Setup (platform-specific, run once)
make install-{msys2,macos,debian}-dependencies

# Workflow
make generate skyeye test

# Pre-commit (required)
make lint vet fix format && go mod tidy && git diff
```

## Resources

- `docs/PLAYER.md` - GCI commands
- `docs/ADMIN.md` - Deployment
- `docs/CONTRIBUTING.md` - Setup details
