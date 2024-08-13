# SkyEye: AI Powered GCI Bot for DCS

SkyEye is a [Ground Controlled Intercept](https://en.wikipedia.org/wiki/Ground-controlled_interception) (GCI) bot for the flight simulator [Digital Combat Simulator](https://www.digitalcombatsimulator.com) (DCS). A GCI bot allows players to request information about the airspace in English using either voice commands or text entry, and to receive answers via verbal speech and text messages

SkyEye uses Speech-To-Text and Text-To-Speech technology which runs locally on the same computer as SkyEye. No cloud APIs are required. It works with any DCS mission, singleplayer or multiplayer. No special scripting or mission editor setup is required. You can even run SkyEye on your own PC to provide GCI service on a remote multiplayer server.

SkyEye is under active development. Most types of radio calls are functional running against live multiplayer servers. Howevever, there's still plenty to do before this is ready for widespread use. To see what I'm working on, check out the [milestones](https://github.com/dharmab/skyeye/milestones?direction=asc&sort=due_date&state=open)!

## Goals

* Implement `ALPHA CHECK`, `BOGEY DOPE`, `DECLARE`, `FADED`, `PICTURE`, `RADIO CHECK`, `SNAPLOCK`, `SPIKED` and `THREAT` calls
* Run entirely locally on reasonable consumer hardware
* Use modern speech synthesis that sounds like a human (Goodbye, Microsoft SAM! Hello, [Piper](https://rhasspy.github.io/piper-samples)!)
* Hybridize real-world [air control communication](https://www.alsa.mil/Portals/9/Documents/mttps/acc_2021.pdf) and [brevity](https://rdl.train.army.mil/catalog-ws/view/100.ATSC/5773E259-8F90-4694-97AD-81EFE6B73E63-1414757496033/atp1-02x1.pdf) with pragmatism
* Proactively inform and update players instead of using static tripwire rules
* Support accessible interfaces in addition to voice and audio, including keyboard based input and in-game subtitles
* Excellent documentation for developers, server administrators and players
* Be easy for a beginner programmer to customize
* Have useful test coverage, especially of controller logic
* Support Windows x86-64, Linux x86-64 and Linux ARM. Experimental functionality on macOS with Apple Sillicon.
* Allow multiple GCI bots to run on the same DCS and SRS instance with different callsigns and frequencies
* Minimize maintenance burden. Ship a static binary with as many pinned dependencies as possible, so this software continues to function with reduced maintainer activity

## Anti-Goals

* Follow [grug-brained principles](https://grugbrain.dev/). Avoid unecessary design patterns. Keep it simple!
* Focused feature set. Don't try to match other bots 1:1 on feature set.
* [Say "no" to complex features.](https://grugbrain.dev/#grug-on-saying-no) Provide the basics, and sufficient documentation for others to fork and customize for their use case.

## Getting Started

* Developers: See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for instructions on building, running and modifying the bot.
* Server admins: Documentation coming Soon™
* Players: See [the user guide](docs/PLAYER.md) (work in progress) for instructions on using the bot.
* Please also see [the privacy statement](docs/PRIVACY.md) to understand how SkyEye uses your voice and gameplay data to function.

## Technology

Skyeye would not be possible without these people and projects, for whom I am deeply appreciative:

* [DCS-SRS](https://github.com/ciribob/DCS-SimpleRadioStandalone) by @ciribob. Ciribob also patiently answered many of my questions on SRS internals and provided helpful debugging tips whenever I ran into a block in the SRS integration.
* [Tacview](https://www.tacview.net/) - specifically, [ACMI real time telemetry](https://www.tacview.net/documentation/realtime/en/) - provides the data feed from DCS World.
* @rurounijones's [OverlordBot](https://gitlab.com/overlordbot) was a useful reference against SkyEye during early development, and Jones himself was also patient with my questions on Discord.
* @ggerganov's [whisper.cpp](https://github.com/ggerganov/whisper.cpp) models provides text-to-speech.
* @rodaine's [numwords](https://github.com/rodaine/numwords) module is invaluable for parsing numeric quantities from voice input.
* [Piper](https://github.com/rhasspy/piper) by the [Rhasspy](https://rhasspy.readthedocs.io/en/latest/) voice assistant project is used for speech-to-text.
* The [Jenny dataset by Dioco](https://github.com/dioco-group/jenny-tts-dataset) provides the feminine voice for SkyEye.
* @popey's dataset provides the masculine voice for SkyEye.
* @amitybell's [embedded Piper module](https://github.com/amitybell/piper) makes distribution and implementation of Piper a breeze. @nabbl improved this module by adding support for macOS.
* The [Opus codec](https://opus-codec.org) and the [`hraban/opus`](https://github.com/hraban/opus) module provides audio compression for the SRS protocol.
* @hbollon's [go-edlib](github.com/hbollon/go-edlib) module provides algorithms to help SkyEye understand when it slightly mishears/the user slightly misspeaks a callsign or command over the radio.
* @lithammer's [shortuuid](https://github.com/lithammer/shortuuid) module provides a GUID implementation compatible with the SRS protocols.
* @zaf's [resample](https://github.com/zaf/resample) module helps with audio format conversion between Piper and SRS.
* @martinlindhe's [unit](https://github.com/martinlindhe/unit) module provides easy angular, length, speed and frequency unit conversion.
* @paulmach's [orb](https://github.com/paulmach/orb) module provides a simple, flexible GIS library for analyzing the geometric relationships between aircraft.
* @proway's [go-igrf](github.com/proway2/go-igrf) module implements the [Internation Geomagnetic Reference Field](https://www.ngdc.noaa.gov/IAGA/vmod/igrf.html) used to correct for magnetic declination.
* [Cobra](https://cobra.dev) is used for the CLI frontend, including configuration, help and examples.
* [MSYS2](https://www.msys2.org/) provides a Windows build environment.
* [Oto](https://github.com/ebitengine/oto) was helpful for debugging audio format conversion problems.
* [zerolog](https://github.com/rs/zerolog) is helpful for general logging and printf debugging.
* [testify](https://github.com/stretchr/testify) is used in unit tests.
* Multiple DCS communities provide invaluable feedback and morale-booster energy:
  * [Team Lima Kilo](https://github.com/team-limakilo/) and the Flashpoint Levant community 
  * The Hoggit Discord server
  * [Digital Controllers](https://digital-controllers.com/)
  * [1VSC](https://1stvsc.com/wing/)
  * [CVW8](https://virtualcvw8.com/)
  * @Frosty-nee
* The _Ace Combat_ series by PROJECT ACES/Bandai Namco and _Project Wingman_ by Sector D2 are _massive_ influences on my interest in GCI/AWACS, and aviation in general. This project would not exist without the impact of _Ace Combat 04: Shattered Skies_.
* And of course, [_DCS World_](https://www.digitalcombatsimulator.com/en/) is produced by Eagle Dynamics.

## FAQ

### Is this ready?

This project is close to a Limited Availability release by early fall 2024. A General Availability release is expected during winter 2024-2025.

You can check current progress [here](https://github.com/dharmab/skyeye/milestones)!

### What kind of hardware does it require?

CPU: SkyEye is currently highly sensitive to CPU performance. On my system with an AMD 5900X, it takes 1-3 seconds to recognize a voice command, and starts responding 1 second after that. However, SkyEye is extremely sensitive to CPU latency. It does not run well when sharing a CPU with other intensive software.

* Avoid running SkyEye on the same physical machine as another intensive app like DCS or TacView client. Ideally, run it on a separate computer.
* If you're running SkyEye on a cloud provider, ensure your virtual machine has dedicated CPU cores instead of shared CPU cores.
* SkyEye is heavily multi-threaded and benefits from multi-core performance.

Memory: SkyEye uses about 2.5-3.0GB of RAM when using the `ggml-small.en.bin` model.

Disk: SkyEye requires around 1-2GB of disk space depending on the selected Whisper model.

There is some room for improvement:

* I'm using an off the shelf general purpose Whisper model in my development environment. There's some exciting research into faster [distilled models](https://github.com/huggingface/distil-whisper) and custom trained models that will be revisited in a few months. I also strongly suspect a combination of advances in AI and Moore's Law will significantly improve Speech-To-Text performance within the next year or so.
* I need to investigate hardware acceleration using [CUDA](https://developer.nvidia.com/cuda-toolkit), [OpenVINO](https://docs.openvino.ai) and [Core ML](https://developer.apple.com/machine-learning/core-ml/). This is challenging because I have limited hardware - if you're interested in this and have hardware please get in touch!

### Why not update OverlordBot?

It would probably be less effort to update OverlordBot to use OpenAI Whisper speech recognition. I certainly wouldn't have had to reimplement the SRS wire protocol from scratch! If you are willing and capable, I encourage you to contribute that change to OverlordBot.

I have some personal, selfish reasons for writing a new bot:

1. I like programming in Go and *nix more than I like C#/.NET. Instrinic motivation is extremely important for hobby developers
1. I use Go, Python and Linux professionally so this is more relevant to my career development than .NET development
1. I want to learn more about practical network programming with coroutine-based concurrency
1. I believe the TRIPWIRE functionality in OverlordBot is damaging to the community and want to eradicate it.
1. I want to innovate and deliver new features that would be breaking changes to the OverlordBot community.
1. Given my lack of .NET development skills, it is faster for me to write new software using technologies to which I am "native" rather than contribute to OverlordBot.

### Why didn't you implement TRIPWIRE?

TRIPWIRE encourages players to think about themselves in a small bubble. It also clutters the channel with information in a format only useful to a specific player. It encourages players to act as lone wolves rather than as members of a team.

Instead, I have implemented THREAT monitoring. THREAT monitoring warns you when a hostile aircraft is a danger to your coalition. The advantages:

- THREAT calls do not require you to individually register with the bot. The bot automatically monitors all friendly aircraft which tune to the SRS frequency.
- Locations in THREAT calls are given in either BRAA or BULLSEYE format, depending on whether the call is relevant to a single aircraft or multiple aircraft TRIPWIRE calls only provide BRAA format.
- THREAT monitoring provides continual updates on the threat as long as threat criteria are met, all the way til the merge. A TRIPWIRE call is only given once, at a single requested threat range. 
- THREAT monitoring considers the bandit group's distance, platform (aircraft & weapons) and aspect (Hot, Flank, Beam, Drag). THREAT calls are broadcast earlier in the BVR timeline for bandits that present a higher relative threat. TRIPWIRE calls only consider threat range, and are broadcast at the same range regardless of other factors. 
- THREAT monitoring deduplicates calls to multiple friendly aircraft about the same bandit group.

### Can I train the speech recognition on my voice/accent?

Since the software runs 100% locally, the speech recognition model is a local file. Server oprators can provide a trained model as an alternative to the off-the-shelf model. See [this blog post](https://huggingface.co/blog/fine-tune-whisper) for an example.

I don't plan to provide a mechanism for players to submit their voice recordings to the main repostitory due to data privacy concerns.

### Does this use Line-Of-Sight restrictions?

No. Excluding this feature was an explicit choice in order to avoid [the complexity demon](https://grugbrain.dev/#grug-on-complexity).

If this is a critical feature for you, consider using [MOOSE's AWACS module](https://flightcontrol-master.github.io/MOOSE_DOCS_DEVELOP/Documentation/Ops.AWACS.html) instead. It supports Line-Of-Sight and datalink simulation, at the tradeoff of requiring some special setup in the Mission Editor.

OverlordBot also optionally supports this feature, although less than 1% of users used it.

### Will this work with DCS's built-in VoIP?

Hopefully in the future Eagle Dynamics will add support for external GCI bots. If anyone at ED is reading this, access to any relevant preview builds would be really helpful!

### Could this use a Large Language Model? (llama, mistral, etc.)

This deserves a longer answer, for now see [this issue](https://github.com/dharmab/skyeye/issues/57)

TL;DR most of the controller logic is simple geometry that completes in about a millisecond. An LLM is several orders of magnitude slower, less accurate and a more difficult user experience.

We use AI for the "squishy" problems - understanding human speech, and synthesizing human-like speech. We use traditional code for the algorithmic problems.

### Could this provide ATC services?

This deserves a longer answer, for now see [this issue](https://github.com/dharmab/skyeye/issues/56)

TL;DR I have no plans to attempt an ATC bot.

### When is SkyEye's birthday?

October 12th. At some point I'll put an Ace Combat 04 easter egg in there.
