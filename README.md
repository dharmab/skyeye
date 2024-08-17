# SkyEye: AI Powered GCI Bot for DCS

SkyEye is a [Ground Controlled Intercept](https://en.wikipedia.org/wiki/Ground-controlled_interception) (GCI) bot for the flight simulator [Digital Combat Simulator](https://www.digitalcombatsimulator.com) (DCS). A GCI bot allows players to request information about the airspace in English using either voice commands or text entry, and to receive answers via verbal speech and text messages

SkyEye uses Speech-To-Text and Text-To-Speech technology which runs locally on the same computer as SkyEye. No cloud APIs are required. It works with any DCS mission, singleplayer or multiplayer. No special scripting or mission editor setup is required. You can run it for less than a nickel per hour on a cloud server, or run it on a PC in your home.

SkyEye is under active development. All of the radio calls I planned to support have been implemented - but there is still lots of work to do on performance, quality, accessibility, and additional features. To see what I'm working on, check out the [milestones](https://github.com/dharmab/skyeye/milestones?direction=asc&sort=due_date&state=open)!

## Getting Started

* Players: See [the user guide](docs/PLAYER.md) (work in progress) for instructions on using the bot.
* Server admins: See [ADMIN.md](docs/ADMIN.md) (work in progress) for a technical guide on deploying the bot.
* Developers: See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for instructions on building, running and modifying the bot.
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
* @amitybell's [embedded Piper module](https://github.com/amitybell/piper) makes distribution and implementation of Piper a breeze. @nabbl improved this module by adding support for macOS and variable speeds.
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

This project is currently available in Limited Availability. Anyone can download the software and try it out, but it may contain bugs or have performance or quality issues. **I am currently only providing support for a limited number of personal friends**.

A General Availability release is expected during winter 2024-2025. At that point, I expect the software to be stable with few or no issues, and support will be provided to the general audience.

You can check current progress [here](https://github.com/dharmab/skyeye/milestones)!

### What kind of hardware does it require?

CPU: SkyEye's speech recognition is extremely sensitive to CPU latency. It does not run well when sharing a CPU with other intensive software.

* Avoid running SkyEye on the same physical machine as another intensive app like DCS or TacView client. Ideally, run it on a separate computer.
* If you're running SkyEye on a cloud provider, ensure your virtual machine has dedicated CPU cores instead of shared CPU cores.
* SkyEye is heavily multi-threaded and benefits from multi-core performance.

Memory: SkyEye uses about 2.5-3.0GB of RAM when using the `ggml-small.en.bin` model.

Disk: SkyEye requires around 1-2GB of disk space depending on the selected Whisper model.

Some examples of the performance you can expect:

* My personal rig: AMD 5900X, 64GB DDR4 RAM. Speech recognition takes 1.5-3.0 seconds.
* Hetzner CCX23: AMD EPYC Milan (4 dedicated cores), 16GB RAM. Speech recognition takes around 5-6 seconds.
* Hetzner CCX13: AMD EPYC Milan (2 dedicated cores), 8GB RAM. Speech recognition takes around 13-16 seconds.

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
