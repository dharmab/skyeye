# Simple Quickstart on macOS

This guide is a step-by-step on how to run SkyEye on macOS, using local speech recognition.

## Getting Help

See [the admin guide](ADMIN.md#getting-help) for how to get help if you have a problem.

## Set Up DCS, TacView, and SRS

Install DCS on a Windows computer (either the client for singleplayer/hosted multiplayer use, or a dedicated server for multiplayer use).

Install the TacView Exporter on the DCS machine. Within DCS, go to OPTIONS → SPECIAL → Tacview and enable Real-Time Telemetry.  (See https://www.tacview.net/documentation/dcs/en/)

Install SRS on a Windows computer. Start and configure the SRS Server. Ensure EAM mode is enabled and an EAM password is set.

Make sure DCS is running an unpaused mission with both friendly and hostile air units so you can do some tests.

## Configure System Voice (Optional, Strongly Recommended)

For the best possible AI voice, set your macOS voice to Siri Voice 5:

1. Open System settings
2. Click on "Accessibility"
3. Click on "Spoken Content"
4. If the system language is not English, set the system speech language to English
5. Next to "System Voice", click the "i" button
6. In the list of languages, make sure "English" is selected
7. Click on "Voice"
8. Scroll down to "Siri".
9. Download the English (United States) Siriv Voice 5.
10. Click "Done"
11. Set the system voice to Siri Voice 5.

## Install Homebrew

Follow the instructions at https://brew.sh to install Homebrew.

## Install SkyEye

Run the following commands in a terminal:

```sh
brew tap skyeye/skyeye
brew install dharmab/skyeye/skyeye
```

Open `$(brew --prefix)/etc/skyeye/config.yaml` with a text editor.  (if you don't have one, download [Visual Studio Code](https://code.visualstudio.com). I don't recommend trying to edit YAML with TextEdit because it's easy to make an indentation error.) Edit the file as required and save your changes.

A minimal sample config file might look like:

```yaml
callsign: Focus
recognizer: openai-whiser-local
whisper-model: /opt/homebrew/share/skyeye/models/ggml-small.en.bin
telemetry-address: tacview.example.com:42674
telemetry-password: yourtacviewpasswordhere
srs-server-address: srs.example.com:5002
srs-eam-password: yoursrspasswordhere
use-system-voice: true  # Set to false if you didn't configure Siri Voice 5 above
```

Run this command to start SkyEye, and automatically start it whenever you log in:

```sh
brew services run dharmab/skyeye/skyeye
```

Confirm SkyEye is running:

```sh
brew services info dharmab/skyeye/skyeye
```

The output should indicate that SkyEye is running.

Also, take a look at the logs:

```sh
tail -f $(brew --prefix)/var/log/skyeye.log
```

if you see a lot of repeated WARN or ERROR lines that don't go away, something may be wrong.

On a Windows computer, connect to your DCS and SRS servers. Switch to one of the SkyEye frequencies you configured. Try some test commands like a RADIO CHECK, ALPHA CHECK and PICTURE. (See the [player guide](PLAYER.md).)

To stop SkyEye, run this command:

```sh
brew services kill dharmab/skyeye/skyeye
```

For instructions on :

- Uninstalling SkyEye
- Upgrading to a newer version of SkyEye
- Automatically starting SkyEye on login
- Using voices other than Siri Voice 5

See [the full admin guide](ADMIN.md).