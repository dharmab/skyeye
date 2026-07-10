# Deploy SkyEye on Windows - CPU

This guide is a step-by-step on how to run SkyEye on a Windows computer, separate from the computer running DCS, TacView and SRS Server, using the CPU for local speech recognition.

You can also deploy SkyEye on Windows using either [the same computer as DCS using GPU speech recognition](gpu.md) or [cloud API speech recognition](api.md).

## Getting Help

See [the admin guide](../../ADMIN.md#getting-help) for how to get help if you have a problem.

## Hardware Requirements

This guide requires a **second** Windows computer, separate from the one running DCS, TacView and SRS Server. The computer running SkyEye needs a fast, multithreaded, **dedicated** CPU with support for [AVX2](https://en.wikipedia.org/wiki/Advanced_Vector_Extensions#Advanced_Vector_Extensions_2), 3GB of RAM, and about 2GB of disk space.

CPU Series|AVX2 Added In
-|-
Intel Core|Haswell (2013)
AMD|Excavator (2015)
Intel Pentium/Celeron|Tiger Lake (2020)

At least 4 dedicated CPU cores are recommended. Shared-core virtual machines are **not supported** and will result in high latency and stuttering audio.

Running SkyEye's local speech recognition on CPU on the same computer as DCS is not supported; only [GPU-based local speech recognition](gpu.md) and [cloud API speech recognition](api.md) are supported on the same computer as DCS. See [the admin guide](../../ADMIN.md#deployment-with-local-speech-recognition-on-cpu) for details on why.

## Set Up DCS, TacView, and SRS

On the computer running DCS:

Install DCS (either the client for singleplayer/hosted multiplayer use, or a dedicated server for multiplayer use).

Install the TacView Exporter. Within DCS, go to OPTIONS → SPECIAL → Tacview and enable Real-Time Telemetry. (See https://www.tacview.net/documentation/dcs/en/)

Install SRS. Start and configure the SRS Server. Ensure EAM mode is enabled and an EAM password is set.

Make sure DCS is running an unpaused mission with both friendly and hostile air units so you can do some tests.

## Set up SkyEye

On the computer running SkyEye:

Download the latest release of SkyEye from https://github.com/dharmab/skyeye/releases. (Click on the file `skyeye-windows-amd64.zip`).

Extract the zip file somewhere convenient.

Download a Whisper speech recognition model, such as [`ggml-small.en.bin`](https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin), and save it in the folder you extracted SkyEye to.

Open `config.yaml` with a text editor (if you don't have one, download [Visual Studio Code](https://code.visualstudio.com). I don't recommend trying to edit YAML with Notepad because it's easy to make an indentation error.) Set `recognizer` to `openai-whisper-local` and edit the rest of the file as required, then save your changes.

```yaml
recognizer: openai-whisper-local
```

Open `skyeye-service.yml` with a text editor and set `whisper-model` to the path of the model file you downloaded. This file, not `config.yaml`, controls which Whisper model SkyEye loads.

```yaml
whisper-model: C:\path\to\skyeye\ggml-small.en.bin
```

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

Connect to your DCS game and SRS server. Switch to one of the SkyEye frequencies you configured. Try some test commands like a RADIO CHECK, ALPHA CHECK and PICTURE. (See the [player guide](../../PLAYER.md).)

To stop SkyEye, run this command:

```powershell
./skyeye-service.exe stopwait
```

## Automatically Starting SkyEye on Boot

By default, SkyEye only starts when you run `./skyeye-service.exe start`. To make it start automatically on boot:

```powershell
./skyeye-service.exe stopwait
./skyeye-service.exe uninstall
```

Open `skyeye-service.yml` with a text editor and change `startmode` from `Manual` to `Automatic`, then reinstall and start SkyEye:

```powershell
./skyeye-service.exe install
./skyeye-service.exe start
```

## Upgrading SkyEye

Stop and uninstall the current version:

```powershell
./skyeye-service.exe stopwait skyeye-service.yml
./skyeye-service.exe uninstall skyeye-service.yml
```

Download the latest release of SkyEye from https://github.com/dharmab/skyeye/releases. (Click on the file `skyeye-windows-amd64.zip`.) Extract it, then replace both `skyeye.exe` and `skyeye-service.yml` in your SkyEye folder with the new versions.

The new `skyeye-service.yml` won't have your `whisper-model` setting, or your `startmode` setting if you enabled autostart on boot. Re-apply both before proceeding.

Install and start the new version:

```powershell
./skyeye-service.exe install skyeye-service.yml
./skyeye-service.exe start skyeye-service.yml
```

## Uninstalling SkyEye

```powershell
./skyeye-service.exe stopwait
./skyeye-service.exe uninstall
```

Delete the SkyEye folder if you no longer need it.
