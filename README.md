# SkyEye: AI Powered GCI Bot for DCS

SkyEye is a concept for a new [Ground Controlled Intercept](https://en.wikipedia.org/wiki/Ground-controlled_interception) (GCI) bot for the flight simulator [Digital Combat Simulator](https://www.digitalcombatsimulator.com) (DCS). A GCI bot allows players to request information about the airspace in English using either voice commands or text entry, and to receive answers via verbal speech and text messages

Unlike previous GCI bots, SkyEye uses Speech-To-Text and Text-To-Speech technology which runs locally on the same server as SkyEye. No cloud APIs are required.

## Goals

* Run entirely locally on reasonable consumer hardware
* Use modern speech synthesis that sounds like a human (Goodbye, Microsoft SAM! Hello, [Piper](https://rhasspy.github.io/piper-samples)!)
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

* Follow [grug-brained principles](https://grugbrain.dev/). Avoid unecessary design patterns like factories, dependency injection frameworks, and large inheritance hierarchies
* Focused feature set. Don't try to match other bots 1:1 on features like ATC or DCT integration
* [Say "no" to complex features.](https://grugbrain.dev/#grug-on-saying-no) Provide the basics, and sufficient documentation for others to fork and customize for their use case

## Getting Started

* Developers: See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for instructions on building, running and modifying the bot.
* Server admins: Documentation coming Soon™
* Players: Guides coming Soon™

## Technology

Skyeye would not be possible without these people and projects, for whom I am deeply appreciative:

* [DCS-SRS](https://github.com/ciribob/DCS-SimpleRadioStandalone) by @ciribob. Ciribob also patiently answered many of my questions on SRS internals and provided helpful debugging tips whenever I ran into a block in the SRS integration.
* [DCS-gRPC](https://github.com/DCS-gRPC) provides the interface into DCS World. 
* @rurounijones's [OverlordBot](https://gitlab.com/overlordbot) was a useful reference against Skyeye during early development, and Jones himself was also patient with my questions on Discord.
* @ggerganov's [whisper.cpp](https://github.com/ggerganov/whisper.cpp) models provides text-to-speech.
* [Piper](https://github.com/rhasspy/piper) by the [Rhasspy](https://rhasspy.readthedocs.io/en/latest/) voice assistant project is used for speech-to-text.
* The [Jenny dataset by Dioco](https://github.com/dioco-group/jenny-tts-dataset) provides the feminine voice for Skyeye.
* @popey's dataset provides the masculine voice for Skyeye.
* @amitybell's [embedded Piper module](https://github.com/amitybell/piper) makes distribution and implementation of Piper a breeze.
* The [Opus codec](https://opus-codec.org/) and the [`hraban/opus`](https://github.com/hraban/opus) module provides audio compression for the SRS protocol.
* @lithammer's [shortuuid](https://github.com/lithammer/shortuuid) module provides a GUID implementation compatible with the SRS protocols.
* @zaf's [resample](https://github.com/zaf/resample) module helps with audio format conversion between Piper and SRS.
* [Oto](https://github.com/ebitengine/oto) is helpful for debugging audio format conversion problems.
* [Team Lima Kilo](https://github.com/team-limakilo/) and the Flashpoint Levant community provided morale-boosting energy and feedback.
* And of course, [DCS World](https://www.digitalcombatsimulator.com/en/) is produced by Eagle Dynamics.

## FAQ

### Is this ready?

No, it's still in development. I work on it about one night a week. At this rate I hope it will be ready by early 2025.

Current status:

- ✅ SRS integration - bot can listen to and talk on an SRS channel
- ✅ Speech recognition - bot can recognize what humans are saying on SRS and turn it into text
- ✅ Speech synthesis - bot can turn text into human-like speech and say it on SRS
- ✅ DCS-gRPC - Prototyped connection to DCS via DCS-gRPC and reading game world state
- 🚧 Text input and in-game subtitles not yet implemented
- 🚧 GCI controller request parser not yet implemented
- 🚧 Game world state interface not yet implemented
- 🚧 GCI controller logic not yet implemented
- 🚧 GCI controller response composer not yet implemented
- 🚧 Limited test coverage
- 🚧 CI/CD pipeline not built
- 🚧 Documentation not written
- 🚧 Observability is sporadic

### What kind of hardware does it require?

I'm not sure yet but it shouldn't be too bad. Currently the dev build takes about 4GB of RAM and takes ~5s to recognize audio on an AMD 5900X, but I have done essentially no performance optimization yet and I expect those requirements to drop significantly. Some areas to improve:

* I'm making unecessary copies of data all over the place - this is usually the default practice in Go unless you either need the receiving function to mutate the passed object, you need to do so for concurrency safety, or you can provably improve performance. I plan to revisit this when the bot is closer to release.
* I'm using a fairly large, off the shelf general purpose Whisper model in my development environment. There's some exciting research into faster [distilled models](https://github.com/huggingface/distil-whisper) and custom trained models that will be revisited in a few months. I also strongly suspect a combination of advances in AI and Moore's Law will significantly improve Speech-To-Text performance within the next year or so.

### Why not update OverlordBot?

It would probably be less effort to update OverlordBot to use OpenAI Whisper speech recognition. I certainly wouldn't have had to reimplement the SRS wire protocol from scratch! If you are willing and capable, I encourage you to contribute that change to OverlordBot.

I have some personal, selfish reasons for writing a new bot:

1. I like programming in Go and *nix more than I like C#/.NET. Instrinic motivation is extremely important for hobby developers
1. I use Go, Python and Linux professionally so this is more relevant to my career development than .NET development
1. I want to learn more about practical network programming with coroutine-based concurrency
1. I believe the TRIPWIRE functionality in OverlordBot is damaging to the community and want to eradicate it.
1. I want to defeat the sense of fatalism about GCI bot development in the community. It's not enough to band-aid an unmaintained project- I want to prove we can still innovate
1. The requirements to submit a merge request to the official OverlordBot repository required more work than writing my own bot from scratch (sorry Jones)

### Why aren't you implementing TRIPWIRE?

TRIPWIRE encourages players to think about themselves in a small bubble. It also clutters the channel with information in a format only useful to a specific player. It encourages players to act as lone wolves rather than as members of a team.

Instead, I will implement THREAT brevity. THREAT provides similar benefit to a player as a TRIPWIRE- it warns you when a hostile aircraft is a danger to you. The advantages:

- THREAT calls do not require you to individually register with the bot. The bot can see the radar, and it can see which players are currently on the frequency. Therefore, it can automatically make THREAT calls to players on frequency.
- Locations in THREAT calls can be given in either BRAA or BULLSEYE format, depending on whether the call is relevant to a single aircraft or multiple aircraft.
- A TRIPWIRE call is only given once, at a single threat range. THREAT calls can be given at multiple threat ranges, which may be configurable based on mission requirements. For example, ATP 3-52.4 recommends 35nmi and 5nmi by default, regardless of aspect.
- By building trackfiles, the bot can determine the aspect of aircraft and provide calls independent of range. For example, if the bot sees a retreating hostile aircraft change course and turn nose-on to a friendly aircraft 45nmi away, the bot can make a THREAT call immediately for the aircraft under threat.

### Can I train the speech recognition on my voice/accent?

Since the software runs 100% locally, the speech recognition model is a local file. Server oprators can provide a trained model as an alternative to the off-the-shelf model. See [this blog post](https://huggingface.co/blog/fine-tune-whisper) for an example.

I don't plan to provide a mechanism for players to submit their voice recordings to the main repostitory due to data privacy concerns.

### Will this work with DCS's built-in VoIP?

Hopefully in the future Eagle Dynamics will add support for external GCI bots. If anyone at ED is reading this, access to any relevant preview builds would be really helpful!

### When is SkyEye's birthday?

October 12th. At some point I'll put an Ace Combat 04 easter egg in there.
