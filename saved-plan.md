 Plan to implement                                                                                                                              
                                                                                                                                                
 Replace Piper/macOS TTS with Pocket TTS via sherpa-onnx                                                                                        
                                                                                                                                                
 Context                                                                                                                                        
                                                                                                                                                
 SkyEye currently uses two platform-specific TTS backends: Piper (Windows/Linux) and Apple's say command (macOS). This creates platform         
 divergence, different voice characteristics per OS, and heavy dependencies (piper-voice-* models embedded in the binary, macOS-only code       
 paths).                                                                                                                                        
                                                                                                                                                
 Pocket TTS is a voice-cloning TTS model. sherpa-onnx (already used for Parakeet STT) has native Pocket TTS support via                         
 OfflineTtsPocketModelConfig. This means no new Go dependencies are needed — we use the same sherpa-onnx library for both STT and TTS.          
                                                                                                                                                
 Decisions:                                                                                                                                     
 - Use sherpa-onnx's built-in Pocket TTS API (no yalue/onnxruntime_go, no SentencePiece library)                                                
 - Download models from sherpa-onnx's GitHub releases (same pattern as Parakeet)                                                                
 - Ship one embedded default reference voice WAV + --voice-file flag for custom voices                                                          
 - Drop --voice enum, --use-system-voice, --voice-playback-speed, --voice-playback-pause                                                        
 - Keep --voice-volume and --voice-lock-path (orthogonal to TTS backend)                                                                        
 - Hardcode temperature and lsd_steps at sensible defaults (no user-facing flags)                                                               
 - Fully remove macOS say command and Piper support                                                                                             
 - Use functional options pattern for NewPocketSpeaker                                                                                          
 - Add Close() to Speaker interface for C resource cleanup                                                                                      
                                                                                                                                                
 Model Files                                                                                                                                    
                                                                                                                                                
 Source: https://github.com/k2-fsa/sherpa-onnx/releases/download/tts-models/sherpa-onnx-pocket-tts-int8-2026-01-26.tar.bz2                      
                                                                                                                                                
 Archive contents (needed files):                                                                                                               
 - lm_main.int8.onnx — main language model                                                                                                      
 - lm_flow.int8.onnx — flow network                                                                                                             
 - decoder.int8.onnx — audio decoder                                                                                                            
 - encoder.onnx — voice reference encoder (FP32, no INT8 variant)                                                                               
 - text_conditioner.onnx — text encoder (FP32)                                                                                                  
 - vocab.json — vocabulary                                                                                                                      
 - token_scores.json — token scoring config                                                                                                     
                                                                                                                                                
 Package Structure                                                                                                                              
                                                                                                                                                
 pkg/synthesizer/                                                                                                                               
   speakers/                                                                                                                                    
     speaker.go              -- Keep Speaker interface (add Close()), add DownsampleF32 helper                                                  
     piper.go                -- DELETE                                                                                                          
     macos.go                -- DELETE                                                                                                          
   voices/                                                                                                                                      
     voices.go               -- DELETE (entire package)                                                                                         
   pocket/                                                                                                                                      
     pocket.go               -- NEW: PocketSpeaker (Speaker impl + sherpa-onnx TTS wrapper)                                                     
     model/                                                                                                                                     
       model.go              -- NEW: Download/verify (mirrors parakeet/model pattern, no CGO)                                                   
     voice/                                                                                                                                     
       voice.go              -- NEW: go:embed default reference WAV + WAV decoder (no CGO)                                                      
       default.wav           -- NEW: ~5-10s reference audio (provided by user before impl)                                                      
                                                                                                                                                
 Implementation Plan                                                                                                                            
                                                                                                                                                
 Phase 1: Model Infrastructure                                                                                                                  
                                                                                                                                                
 1.1 Create pkg/synthesizer/pocket/model/model.go                                                                                               
                                                                                                                                                
 Mirror pkg/recognizer/parakeet/model/model.go (model.go:1-185):                                                                                
 - DirName = "pocket"                                                                                                                           
 - Filenames: the 7 files listed above                                                                                                          
 - fileHashes: SHA256 hashes (compute after downloading archive)                                                                                
 - Verify() and Download() — same pattern with tar.bz2 extraction                                                                               
 - modelURL pointing to sherpa-onnx GitHub releases                                                                                             
 - Reuse same error type pattern (FileNotFoundError, CorruptFileError)                                                                          
                                                                                                                                                
 1.2 Wire model download into cmd/skyeye/main.go                                                                                                
 - Add setupPocketModel() function (copy pattern from setupParakeetModel() at main.go:312-334)                                                  
 - Call in run() after setupParakeetModel() call at main.go:374                                                                                 
                                                                                                                                                
 Phase 2: Pocket TTS Speaker                                                                                                                    
                                                                                                                                                
 2.1 Create pkg/synthesizer/pocket/voice/voice.go                                                                                               
 //go:embed default.wav                                                                                                                         
 var DefaultVoice []byte                                                                                                                        
                                                                                                                                                
 func DecodeWAV(data []byte) (samples []float32, sampleRate int, err error)                                                                     
 - Custom WAV decoder (go-audio/wav was found incompatible during earlier development)                                                          
 - Decode WAV bytes to F32 PCM samples + sample rate                                                                                            
 - Reference WAV must be 16-bit PCM mono — document this constraint                                                                             
 - Used for both embedded default voice and user-provided voice files                                                                           
 - No CGO dependencies — keeps voice/ package testable in IDEs without CGO                                                                      
                                                                                                                                                
 2.2 Create pkg/synthesizer/pocket/pocket.go                                                                                                    
                                                                                                                                                
 This is the main implementation — a single package containing the Speaker implementation backed by sherpa-onnx.                                
                                                                                                                                                
 type PocketSpeaker struct {                                                                                                                    
     tts *sherpa.OfflineTts                                                                                                                     
     genConfig *sherpa.GenerationConfig  // cached with reference audio                                                                         
 }                                                                                                                                              
                                                                                                                                                
 // Functional options                                                                                                                          
 type Option func(*options)                                                                                                                     
 func WithVoiceFile(path string) Option    // custom reference WAV                                                                              
 func WithNumSteps(n int) Option           // default 10                                                                                        
 func WithTemperature(t float64) Option    // default 0.7 (via Extra JSON if supported)                                                         
                                                                                                                                                
 func New(modelDir string, opts ...Option) (*PocketSpeaker, error)                                                                              
 func (s *PocketSpeaker) Say(ctx context.Context, text string) ([]float32, error)                                                               
 func (s *PocketSpeaker) Close()                                                                                                                
                                                                                                                                                
 New() implementation:                                                                                                                          
 1. Apply functional options (defaults: no voice file, numSteps=10)                                                                             
 2. Build sherpa.OfflineTtsConfig with Pocket model config pointing to model files in modelDir                                                  
 3. Call sherpa.NewOfflineTts(&config) to initialize                                                                                            
 4. Load reference audio: if voiceFile option set, attempt to read from disk;                                                                   
    if file not found or invalid, log a warning and fall back to embedded voice.DefaultVoice                                                    
 5. Decode WAV to []float32 via voice.DecodeWAV()                                                                                               
 6. Build sherpa.GenerationConfig with ReferenceAudio, ReferenceSampleRate, NumSteps                                                            
 7. Cache the GenerationConfig for reuse across Say() calls                                                                                     
                                                                                                                                                
 Say() implementation:                                                                                                                          
 1. Call tts.GenerateWithConfig(text, &genConfig, nil) — sherpa-onnx handles the entire inference pipeline internally (tokenization, flow       
 matching, decoding)                                                                                                                            
 2. Result is *sherpa.GeneratedAudio with Samples []float32 and SampleRate                                                                      
 3. Use GeneratedAudio.SampleRate dynamically (do NOT hardcode 24kHz) to resample to 16kHz via DownsampleF32()                                  
 4. Return resampled []float32                                                                                                                  
                                                                                                                                                
 Close() implementation:                                                                                                                        
 - Call sherpa.DeleteOfflineTts(s.tts) to free C resources                                                                                      
                                                                                                                                                
 Context cancellation: sherpa-onnx's GenerateWithConfig accepts a progress callback. We can check ctx.Done() in the callback and return false   
 to abort generation.                                                                                                                           
                                                                                                                                                
 Phase 3: Wiring and Config Changes                                                                                                             
                                                                                                                                                
 3.1 Modify pkg/synthesizer/speakers/speaker.go                                                                                                 
 - Add Close() to Speaker interface                                                                                                             
 - Add DownsampleF32(samples []float32, sourceRate unit.Frequency) ([]float32, error) (exported, so pocket package can use it)                  
 - Implementation: use resample.F32 format directly (zaf/resample supports it) — no S16LE round-trip needed                                     
   Convert []float32 ↔ []byte via binary.LittleEndian/math.Float32frombits for the resampler's io.Writer interface                              
                                                                                                                                                
 3.2 Modify internal/conf/configuration.go                                                                                                      
 - Remove: Voice voices.Voice, UseSystemVoice bool, VoiceSpeed float64, VoicePauseLength time.Duration                                          
 - Keep: VoiceVolume float64, VoiceLock *flock.Flock                                                                                            
 - Update: VoiceLock comment (remove "Piper" reference)                                                                                         
 - Add: VoiceFile string                                                                                                                        
 - Remove import of voices package                                                                                                              
                                                                                                                                                
 3.3 Modify cmd/skyeye/main.go                                                                                                                  
 - Remove vars: voiceName, useSystemVoice, voiceSpeed, voicePauseLength                                                                         
 - Keep vars: voiceVolume, voiceLockPath                                                                                                        
 - Add var: voiceFile string                                                                                                                    
 - Remove flags: --voice, --use-system-voice, --voice-playback-speed, --voice-playback-pause                                                    
 - Keep flags: --voice-volume, --voice-lock-path                                                                                                
 - Add flag: --voice-file (default "", help: "Path to WAV file for custom voice cloning. Uses built-in default if not set.")                    
 - Remove: loadVoice() function, runtime.GOOS == "darwin" conditional block for flag registration                                               
 - Add: setupPocketModel(), call in run()                                                                                                       
 - Update conf.Configuration struct literal: remove old fields, add VoiceFile                                                                   
 - Remove imports: voices, reflect; add: pocketmodel                                                                                            
                                                                                                                                                
 3.4 Modify internal/application/app.go                                                                                                         
 - Replace lines 187-196 (platform branching) with:                                                                                             
 pocketDir := filepath.Join(config.ModelsPath, pocketmodel.DirName)                                                                             
 var pocketOpts []pocket.Option                                                                                                                 
 if config.VoiceFile != "" {                                                                                                                    
     pocketOpts = append(pocketOpts, pocket.WithVoiceFile(config.VoiceFile))                                                                    
 }                                                                                                                                              
 synthesizer, err := pocket.New(pocketDir, pocketOpts...)                                                                                       
 if err != nil {                                                                                                                                
     return nil, fmt.Errorf("failed to construct application: %w", err)                                                                         
 }                                                                                                                                              
 defer synthesizer.Close()  // or wire into app shutdown                                                                                        
 - Update imports: remove runtime, speakers (for TTS construction); add pocket, pocketmodel                                                     
 - Note: speakers.Speaker interface is still used for the speaker field type                                                                    
                                                                                                                                                
 Phase 4: Cleanup and Delete Old Files                                                                                                          
                                                                                                                                                
 Note: Deletions should happen in the same commits as the replacements — not as a separate phase.                                               
 Delete when wiring up the new Pocket speaker in Phase 3, so the build never has dead code.                                                     
                                                                                                                                                
 4.1 Delete files                                                                                                                               
 - pkg/synthesizer/speakers/piper.go                                                                                                            
 - pkg/synthesizer/speakers/macos.go                                                                                                            
 - pkg/synthesizer/voices/voices.go (entire voices package directory)                                                                           
                                                                                                                                                
 4.2 Remove Go module dependencies                                                                                                              
 go mod tidy will remove:                                                                                                                       
 - github.com/nabbl/piper                                                                                                                       
 - github.com/amitybell/piper-asset                                                                                                             
 - github.com/amitybell/piper-voice-alan                                                                                                        
 - github.com/amitybell/piper-voice-jenny                                                                                                       
 - github.com/amitybell/piper-bin-windows                                                                                                       
 - github.com/amitybell/piper-bin-linux                                                                                                         
 - github.com/go-audio/aiff                                                                                                                     
                                                                                                                                                
 4.3 Update Makefile                                                                                                                            
 - No new build flags needed (sherpa-onnx already linked)                                                                                       
 - Windows DLL copy for sherpa-onnx continues to work as-is                                                                                     
 - Verify no Piper-specific build logic remains                                                                                                 
                                                                                                                                                
 4.4 Update documentation                                                                                                                       
 - CLAUDE.md: Update platform table (TTS → "Pocket TTS (sherpa-onnx)" on all platforms), remove Piper/macOS references, remove                  
 --use-system-voice mention                                                                                                                     
 - docs/ADMIN.md: Update for new flags, remove old TTS flag docs                                                                                
 - docs/CONTRIBUTING.md: If TTS setup is mentioned, update                                                                                      
                                                                                                                                                
 Phase 5: Round-Trip Test                                                                                                                       
                                                                                                                                                
 5.1 Create pkg/synthesizer/pocket/pocket_test.go                                                                                               
                                                                                                                                                
 Automated round-trip test: TTS → STT to verify synthesized speech is recognizable.                                                             
 - Use PocketSpeaker.Say() to synthesize a known phrase (e.g., "Bravo one one, bogey, bullseye zero nine zero, forty, twenty thousand")          
 - Feed the resulting []float32 audio into the Parakeet recognizer                                                                              
 - Assert the recognized text matches the input (with reasonable fuzzy matching — e.g., normalize whitespace/case, allow minor word diffs)      
 - This requires both Pocket TTS and Parakeet models to be present; skip the test if models are not downloaded                                  
 - Tag with build constraint or test flag so it only runs when models are available (e.g., `go test -run TestRoundTrip` with model dir env var) 
 - This test validates that: the sherpa-onnx TTS pipeline produces intelligible audio, the WAV reference voice works,                           
   resampling doesn't corrupt the audio, and the full TTS→STT pipeline is functional                                                            
                                                                                                                                                
 Key Files Reference                                                                                                                            
                                                                                                                                                
 - pkg/recognizer/parakeet/model/model.go — READ (pattern) — Mirror this for pocket model download/verify
 - pkg/synthesizer/speakers/speaker.go — MODIFY — Add Close() to interface, add DownsampleF32 (F32)
 - pkg/synthesizer/speakers/piper.go — DELETE — Delete with Phase 3 wiring, not separately
 - pkg/synthesizer/speakers/macos.go — DELETE — Delete with Phase 3 wiring, not separately
 - pkg/synthesizer/voices/voices.go — DELETE — Entire package, delete with Phase 3
 - internal/conf/configuration.go — MODIFY — Remove old TTS fields, keep volume/lock, add VoiceFile
 - internal/application/app.go — MODIFY — Replace platform-branching TTS init, add Close()
 - cmd/skyeye/main.go — MODIFY — CLI flag changes, add pocket model setup
 - pkg/synthesizer/pocket/pocket.go — CREATE — Main PocketSpeaker implementation
 - pkg/synthesizer/pocket/pocket_test.go — CREATE — Round-trip TTS→STT test
 - pkg/synthesizer/pocket/model/model.go — CREATE — Model download/verify (no CGO)
 - pkg/synthesizer/pocket/voice/voice.go — CREATE — Embedded default WAV + decoder (no CGO)
                                                                                                                                                
 Verification                                                                                                                                   
                                                                                                                                                
 1. Build: make skyeye — compiles on all platforms without new dependencies                                                                     
 2. Tests: make test — model download/verify tests, WAV decoder tests                                                                           
 3. Round-trip test: TTS→STT round-trip (requires model files)                                                                                  
 4. CI: make lint && make vet && make fix && make format && go mod tidy && git diff --exit-code                                                 
 5. Manual: Cannot run end-to-end (requires DCS/SRS per CLAUDE.md)                                                                              
                                                                                                                                                
 Notes                                                                                                                                          
                                                                                                                                                
 - User will provide a default reference WAV file before implementation begins                                                                  
 - Temperature and lsd_steps (num_steps) are hardcoded at sensible defaults (0.7 and 10 respectively). If tuning is needed later, the           
 functional options are already in place to expose them.                                                                                        
 - sherpa-onnx handles the entire inference pipeline internally: tokenization via vocab.json/token_scores.json, flow matching, Euler            
 integration, audio decoding. No need to port Python inference code.                                                                            
 - sherpa-onnx Go API confirmed to support Pocket TTS at current version (verified during planning).                                            
 - voice/ and model/ sub-packages are intentionally separate from pocket.go to avoid CGO — makes them testable in IDEs without CGO setup.       
 - If --voice-file points to a missing or invalid file, log a warning and fall back to the embedded default voice.                              
 - Use GeneratedAudio.SampleRate dynamically rather than hardcoding output sample rate — verify during implementation.                           
