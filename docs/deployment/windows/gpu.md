# Deploy SkyEye on Windows - GPU

This guide is a step-by-step on how to run SkyEye on the same Windows computer as DCS, TacView and SRS Server, using the GPU for local speech recognition via the experimental Vulkan build.

You can also deploy SkyEye on Windows using either [a separate computer using CPU speech recognition](cpu.md) or [cloud API speech recognition](api.md).

> ⚠️ The Vulkan (GPU) build of SkyEye is experimental. Performance and speech recognition quality can vary significantly between GPU models — it might run great on one and perform poorly on another. I can only test against the GPU hardware I personally own, so I have no control over and very limited ability to troubleshoot how well the Vulkan build runs on any particular GPU or driver.

## Getting Help

See [the admin guide](../../ADMIN.md#getting-help) for how to get help if you have a problem.

## Hardware Requirements

Unlike CPU-based local speech recognition, the Vulkan build offloads speech recognition to your GPU, so it's suitable for running on the same computer as DCS. You need any decent multithreaded CPU with support for [AVX2](https://en.wikipedia.org/wiki/Advanced_Vector_Extensions#Advanced_Vector_Extensions_2), 3GB of RAM, about 2GB of VRAM, and about 2GB of disk space.

## Set Up DCS, TacView, and SRS

Install DCS (either the client for singleplayer/hosted multiplayer use, or a dedicated server for multiplayer use).

Install the TacView Exporter. Within DCS, go to OPTIONS → SPECIAL → Tacview and enable Real-Time Telemetry. (See https://www.tacview.net/documentation/dcs/en/)

Install SRS. Start and configure the SRS Server. Ensure EAM mode is enabled and an EAM password is set.

Make sure DCS is running an unpaused mission with both friendly and hostile air units so you can do some tests.

## Set up SkyEye

Make sure your GPU driver is up to date. Download the latest driver from your GPU manufacturer: [NVIDIA](https://www.nvidia.com/en-us/drivers/), [AMD](https://www.amd.com/en/support), or [Intel](https://www.intel.com/content/www/us/en/support/detect.html).

Download the latest Vulkan release of SkyEye from https://github.com/dharmab/skyeye/releases. (Click on the file `skyeye-windows-amd64-vulkan.zip`. Don't use `skyeye-windows-amd64.zip`; that's the CPU-only build.)

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

Download the latest Vulkan release of SkyEye from https://github.com/dharmab/skyeye/releases. (Click on the file `skyeye-windows-amd64-vulkan.zip`.) Extract it, then replace both `skyeye.exe` and `skyeye-service.yml` in your SkyEye folder with the new versions.

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
