# Simple Quickstart on Windows

This guide is a step-by-step on how to run SkyEye on Windows alongside DCS, TacView and SRS Server using the OpenAI API for cloud speech recognition.

## Getting Help

See [the admin guide](ADMIN.md#getting-help) for how to get help if you have a problem.

## Set up OpenAI API

Go to https://platform.openai.com. If you haven't previously set up the OpenAI API, you'll go through a step-by-step process to set up an organization, buy credits, create a Project and generate an API key.

Otherwise, go to https://platform.openai.com/settings/organization/api-keys and generate a new API key for SkyEye. I recommend adding this API key to a new Project named "SkyEye".

Be sure to review the pricing of the "Whisper" audio model at https://openai.com/api/pricing. You will pay per-second for each second a player transmits on a SkyEye frequency in SRS. (You are not charged for the time SkyEye itself spends transmitting, nor are you charged for transmissions on other frequencies.)

## Set Up DCS, Tacview, and SRS

Install DCS (either the client for singleplayer/hosted multiplayer use, or a dedicated server for multiplayer use).

Install the TacView Exporter. Within DCS, go to OPTIONS → SPECIAL → Tacview and enable Real-Time Telemetry.  (See https://www.tacview.net/documentation/dcs/en/)

Install SRS. Start and configure the SRS Server.

Makre sure DCS is running a mission with both friendly and hostile air units so you can do some tests.

## Set up SkyEye

Download the latest release of SkyEye from https://github.com/dharmab/skyeye/releases. (Click on the file `skyeye-windows-amd64.zip`).

Extract the zip file somewhere convenient.

Open `config.yaml` with a text editor (if you don't have one, download [Visual Studio Code](https://code.visualstudio.com). I don't recommend trying to edit YAML with Notepad because it's easy to make an indentation error.) Edit the file as required and save your changes.

## Using SkyEye

Open Windows Powershell and use the `cd` command to navigate to the folder containing `skyeye.exe`. (If you need help with this, ask ChatGPT how to do it.)

Once you're in the correct folder, run the following command to install SkyEye.

```powershell
./skyeye-service.exe install
```

Next, run this command to start SkyEye:

```powershell
./skyeye-service.exe start
```

Confirm SkyEye is running using 

```powershell
./skyeye-service.exe status
```

Also, open the `skyeye-service.err.log` in Visual Studio Code. If you see a lot of repeated WARN or ERROR lines that don't go away, something may be wrong.

Connect to your DCS game and SRS server. Switch to one of the SkyEye frequencies you configured. Try some test commands like a RADIO CHECK, ALPHA CHECK and PICTURE. (See the [player guide](PLAYER.md).)

To stop SkyEye, run this command:

```powershell
./skyeye-service.exe stopwait
```

For instructions on:

- Uninstalling SkyEye
- Upgrading to a newer version of SkyEye
- Automatically starting SkyEye on boot

See [the full admin guide](ADMIN.md#windows).