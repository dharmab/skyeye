# SkyEye: AI Powered GCI Bot for DCS

SkyEye is a concept for a new [Ground Controlled Intercept](https://en.wikipedia.org/wiki/Ground-controlled_interception) (GCI) bot for the flight simulator [Digital Combat Simulator](https://www.digitalcombatsimulator.com) (DCS). A GCI bot allows players to request information about the airspace in English using either voice commands or text entry, and to receive answers via verbal speech and text messages

Unlike previous GCI bots, SkyEye uses Speech-To-Text and Text-To-Speech technology which runs locally on the same server as SkyEye. No cloud APIs are required.

## Goals

* Run entirely locally on reasonable consumer hardware
* Use modern speech synthesis that sounds like a human (Goodbye, Microsoft SAM! Hello, [VCTK](https://datashare.ed.ac.uk/handle/10283/3443)!)
* Hybridize real-world [air control communication](https://www.alsa.mil/Portals/9/Documents/mttps/acc_2021.pdf) and [brevity](https://rdl.train.army.mil/catalog-ws/view/100.ATSC/5773E259-8F90-4694-97AD-81EFE6B73E63-1414757496033/atp1-02x1.pdf) with pragmatism
* Proactively inform and update players instead of using static tripwire rules
* Implement PICTURE and DECLARE/SNAPSHOT calls
* Support text-based input and provide subtitles for output
* Allow multiple GCI bots to run on the same DCS and SRS instance with different callsigns and frequencies
* Excellent documentation for developers, server administrators and players
* Be easy for a beginner programmer to customize
* Have useful test coverage, especially of controller logic
* Support Windows x86-64, Linux x86-64 and Linux ARM
* Minimize maintenance burden. Ship a static binary with as many pinned dependencies as possible, so this software continues to function with reduced maintainer activity

## Anti-Goals

* Follow [grug-brained principles](https://grugbrain.dev/). Avoid unecessary design patterns like factories, dependency injection frameworks, and large inheritance hierarchies.
* Focused feature set. Don't try to match other bots 1:1 on features like ATC or DCT integration
* [Say "no" to complex features.](https://grugbrain.dev/#grug-on-saying-no) Provide the basics, and sufficient documentation for others to fork and customize for their use case

## Getting Started

* Developers: See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for instructions on building, running and modifying the bot.
* Server admins: Documentation coming Soon™
* Players: Guides coming Soon™

## FAQ

### Is this ready?

No, it's still in development. I work on it about one night a week. At this rate I hope it will be ready by early 2025.

Current status:

- ✅ Connects to SRS and reads SRS network traffic
- ✅ Connects to DCS via DCS-gRPC and reads game state
- ✅ OpenAPI Whisper speech recognition prototyped and proven viable
- ✅ Speech recognition partially implemented
- ✅ Mimic3 speech output prototyped and proven viable
- 🚧 SRS integration mostly complete, needs more testing and robustness
- 🚧 Speech recognition implementation functional, but needs more work
- 🚧 Text input and outputs not yet implemented
- 🚧 GCI controller logic not yet implemented
- 🚧 Speech output not yet implemented
- 🚧 No test coverage
- 🚧 CI/CD pipeline not built
- 🚧 Documentation not written
- 🚧 Observability is sporadic

### What kind of hardware does it require?

I'm not sure yet but it shouldn't be too bad. My lowest spec test machine has an AMD 3900 CPU, but I expect it will run on something weaker.

### Why not update OverlordBot?

It would probably be less effort to update OverlordBot to use OpenAI Whisper speech recognition. I certainly wouldn't have had to reimplement the SRS wire protocol from scratch! If you are willing and capable, I encourage you to contribute that change to OverlordBot.

I have some personal, selfish reasons for writing a new bot:

1. I like programming in Go and *nix more than I like C#/.NET. Instrinic motivation is extremely important for hobby developers
1. I use Go, Python and Linux professionally so this is more relevant to my career development than .NET development
1. I want to learn more about practical network programming with coroutine-based concurrency
1. I believe the TRIPWIRE functionality in OverlordBot is damaging to the community and want to eradicate it
1. I want to defeat the sense of fatalism about GCI bot development in the community. It's not enough to band-aid an unmaintained project- I want to prove we can still innovate

### Can I train the speech recognition on my voice/accent?

Since the software runs 100% locally, the speech recognition model is a local file. You can provide a trained model as an alternative to the off-the-shelf model. See [this blog post](https://huggingface.co/blog/fine-tune-whisper) for an example.

I don't plan to provide a mechanism for players to submit their voice recordings to the main repostitory due to data privacy concerns.

### Will this work with DCS's built-in VoIP?

Hopefully in the future Eagle Dynamics will add support for external GCI bots. If anyone at ED is reading this, access to any relevant preview builds would be really helpful!

### When is SkyEye's birthday?

October 12th. At some point I'll put an Ace Combat 04 easter egg in there.
