# Sample Apple Speech recognizer

`NewAppleSpeechRecognizer(callsign, opts...)` implements `Recognizer` using
macOS Speech and AVFoundation, with an unsupported-platform stub for other
platforms and builds without cgo. It uses en-US, mono Float32 PCM at
`rate.Wideband`, and contextual vocabulary from the callsign, `WithLocations`,
and a small aviation brevity list. Hints improve recognition but are not a
grammar or guaranteed output constraints.

The constructor does not request authorization. It verifies a nonempty
`NSSpeechRecognitionUsageDescription` in the main bundle/executable Info.plist
and checks on-device recognition support. The macOS Makefile embeds
`cmd/skyeye/Info.plist` in the executable. Select this backend with
`recognizer: apple-speech` in configuration or `--recognizer=apple-speech`.

The first valid `Recognize` call requests Speech permission if necessary, without
waiting on a semaphore or synchronously dispatching to the main queue. The usage
description is checked again before requesting permission. Buffer recognition
does not capture a microphone and does not request microphone permission.

Recognition requires on-device support and sets `requiresOnDeviceRecognition`;
it never falls back to server transcription. English speech resources must be
installed and the OS recognizer must be available. macOS 10.15 is the minimum
API version, not a guarantee that a particular machine has suitable local
resources. Authorization and local model provisioning are managed by macOS.

Use a context deadline: OS authorization or recognition can take an unbounded
amount of time. Cancellation returns promptly from Go polling and cancels an
active task. It cannot dismiss an OS authorization dialog. Its native callback
may remain alive until macOS answers; it retains only native state, not Go
pointers. PCM is copied before native asynchronous work begins. Each call owns a
separate synchronized native session, although macOS can limit concurrent tasks
and utterance duration. Callers should submit bounded radio transmissions.

Permission identity and prompts for command-line executables can depend on the
launching terminal, signing, and packaging. Test the packaged host manually on
the deployment machine. No custom Cocoa main run loop is installed here.

Automated Apple tests intentionally do not request permission or perform live
recognition. They cover PCM validation, vocabulary construction, and early
context cancellation. Go's race detector does not validate Objective-C locking;
live authorization, in-flight cancellation, concurrency, and local recognition
still need manual testing with the packaged executable and installed resources.
Transcription text is logged only when the per-call logging flag is enabled.
