# Deploy SkyEye on macOS - GPU

This guide is a step-by-step on how to run SkyEye on an Apple Silicon computer, separate from the computer running DCS, TacView and SRS Server, using the Apple Silicon GPU for local speech recognition.

## Getting Help

See [the admin guide](../../ADMIN.md#getting-help) for how to get help if you have a problem.

## Hardware Requirements

You need an Apple Silicon Mac, such as a Mac Mini or MacBook Air/Pro. SkyEye requires around 3GB of RAM and about 2GB of disk space.

Intel Macs are not supported.

## Set Up DCS, TacView, and SRS

Install DCS on a Windows computer (either the client for singleplayer/hosted multiplayer use, or a dedicated server for multiplayer use).

Install the TacView Exporter on the DCS machine. Within DCS, go to OPTIONS → SPECIAL → Tacview and enable Real-Time Telemetry. (See https://www.tacview.net/documentation/dcs/en/)

Install SRS on a Windows computer. Start and configure the SRS Server. Ensure EAM mode is enabled and an EAM password is set.

Make sure DCS is running an unpaused mission with both friendly and hostile air units so you can do some tests.

## Configure System Voice (Optional, Strongly Recommended)

SkyEye uses AI generated voices built into macOS. By default, the "Samantha" voice is used. This is the version of Siri's voice from the iPhone 4s, iPhone 5 and iPhone 6, based on [Susan Bennett](https://susancbennett.com/).

It is also possible to use one of the newer Siri voices. **I strongly recommend enabling one of the newer voices**, because they provide excellent quality, nearly indistinguishable from a human voice.

I've validated one Siri voice for each version of macOS that correctly pronounces aviation brevity and terminology:

### macOS 26 Tahoe

On macOS 26 Tahoe, the best voice is **Siri Voice 2**.

1. Open System Settings
2. Click on "Accessibility"
3. Click on "Siri"
4. If the system language is not English, set the system speech language to English
5. Next to "System Voice", click the "i" button
6. In the list of languages, make sure "English" is selected
7. Click on "Voice"
8. Scroll down to "Siri".
9. Download Siri Voice 2.
10. Click "Done"
11. Set the system voice to Siri Voice 2.

### macOS 15 Sequoia

On macOS 15 Sequoia, the best voice is **Siri Voice 5**.

1. Open System Settings
2. Click on "Accessibility"
3. Click on "Spoken Content"
4. If the system language is not English, set the system speech language to English
5. Next to "System Voice", click the "i" button
6. In the list of languages, make sure "English" is selected
7. Click on "Voice"
8. Scroll down to "Siri".
9. Download the English (United States) Siri Voice 5.
10. Click "Done"
11. Set the system voice to Siri Voice 5.

### Testing the System Voice

To test your change, open Terminal and run this command:

```sh
say "Hello! This should be read in the voice you chose."
```

Finally, to use the selected voice instead of Samantha, set `use-system-voice: true` in your config file, as shown below.

## Install Homebrew

Follow the instructions at https://brew.sh to install Homebrew.

## Install SkyEye

Run the following commands in a terminal:

```sh
brew tap dharmab/skyeye
brew trust dharmab/skyeye/skyeye
brew install dharmab/skyeye/skyeye
```

Open `$(brew --prefix)/etc/skyeye/config.yaml` with a text editor. It's probably at `/opt/homebrew/etc/skyeye/config.yaml`. (If you don't have a text editor, download [Visual Studio Code](https://code.visualstudio.com). I don't recommend trying to edit YAML with TextEdit because it's easy to make an indentation error.) Edit the file as required and save your changes.

A minimal sample config file might look like:

```yaml
callsign: Focus
recognizer: openai-whisper-local
whisper-model: /opt/homebrew/share/skyeye/models/ggml-small.en.bin
telemetry-address: tacview.example.com:42674
telemetry-password: yourtacviewpasswordhere
srs-server-address: srs.example.com:5002
srs-eam-password: yoursrspasswordhere
use-system-voice: true  # Set to false if you didn't configure Siri Voice above
```

Run this command to start SkyEye:

```sh
brew services run dharmab/skyeye/skyeye
```

Confirm SkyEye is running:

```sh
brew services info dharmab/skyeye/skyeye
```

The output should indicate that SkyEye is running.

## Using SkyEye

Take a look at the logs:

```sh
tail -f "$(brew --prefix)/var/log/skyeye.log"
```

If you see a lot of repeated WARN or ERROR lines that don't go away, something may be wrong.

If you see an error message containing `connect: no route to host`, see [this issue](https://github.com/dharmab/skyeye/issues/566) for a possible cause and workarounds. This issue specifically affects macOS 15+ when SkyEye runs as a background service and connects to a SRS, Tacview, or DCS-gRPC server on the same LAN.

On the Windows computer, connect to your DCS and SRS servers. Switch to one of the SkyEye frequencies you configured. Try some test commands like a RADIO CHECK, ALPHA CHECK and PICTURE. (See the [player guide](../../PLAYER.md).)

To stop SkyEye, run this command:

```sh
brew services kill dharmab/skyeye/skyeye
```

## Automatically Starting SkyEye on Login

To start SkyEye and enable it to automatically start on login:

```sh
brew services start dharmab/skyeye/skyeye
```

To stop SkyEye and disable autostart:

```sh
brew services stop dharmab/skyeye/skyeye
```

## Upgrading SkyEye

```sh
brew update
brew upgrade dharmab/skyeye/skyeye
brew services restart dharmab/skyeye/skyeye
```

## Uninstalling SkyEye

```sh
brew uninstall dharmab/skyeye/skyeye
```
