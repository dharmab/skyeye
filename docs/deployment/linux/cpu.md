# Deploy SkyEye on Linux - CPU

This guide is a step-by-step on how to run SkyEye on a Linux computer or server, separate from the computer running DCS, TacView and SRS Server, using the CPU for local speech recognition. This guide is not tied to any particular hosting provider; it works on a spare Linux machine, a rented cloud server, or a container host. If you want a guide tailored to a specific cloud provider, see the [Hetzner Cloud](../cloud-providers/hetzner.md) or [Vultr](../cloud-providers/vultr.md) guides.

You can also deploy SkyEye on Linux using either [the same computer as DCS using GPU speech recognition](gpu.md) or [cloud API speech recognition](api.md).

## Getting Help

See [the admin guide](../../ADMIN.md#getting-help) for how to get help if you have a problem.

## Hardware Requirements

This guide requires a **second** Linux computer or server, separate from the one running DCS, TacView and SRS Server. The computer running SkyEye needs a fast, multithreaded, **dedicated** CPU with support for [AVX2](https://en.wikipedia.org/wiki/Advanced_Vector_Extensions#Advanced_Vector_Extensions_2), 3GB of RAM, and about 2GB of disk space.

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

On the Linux computer running SkyEye, you can install SkyEye either as a container, or as a native binary. The container is easier to set up and upgrade; the native binary is useful if you don't want to run a container runtime.

### Container

Install [Podman](https://podman.io/):

```sh
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install podman

# Fedora
sudo dnf install podman

# Arch Linux
sudo pacman -Syu podman
```

Create a config directory and copy in the [sample config file](../../../config.yaml):

```sh
sudo mkdir -p /etc/skyeye
sudoedit /etc/skyeye/config.yaml
```

Set `recognizer` to `openai-whisper-local` and set `whisper-model` to a path inside the container, such as `/etc/skyeye/ggml-small.en.bin`, along with your Tacview and SRS connection details.

Download a Whisper model:

```sh
curl -sL https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin -o /etc/skyeye/ggml-small.en.bin
```

Create a [Podman Quadlet](https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html) file at `/etc/containers/systemd/skyeye.container`:

```ini
[Unit]
Description=SkyEye GCI Bot
After=network-online.target

[Container]
Image=ghcr.io/dharmab/skyeye:latest
ContainerName=skyeye
Volume=/etc/skyeye:/etc/skyeye:ro

[Service]
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
```

I recommend pinning `Image=` to a specific version instead of `latest`, to avoid unexpected breaking changes when a new version is released. Find version tags on the [releases page](https://github.com/dharmab/skyeye/releases), e.g. `Image=ghcr.io/dharmab/skyeye:v1.9.3`.

Load the Quadlet and start SkyEye, enabling it to start on boot:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now skyeye.service
```

### Native Binary

Install shared libraries for [Opus](https://opus-codec.org/) and [SoX Resampler](https://sourceforge.net/p/soxr/wiki/Home/):

```sh
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install libopus0 libsoxr0

# Fedora
sudo dnf install opus sox

# Arch Linux
sudo pacman -Syu opus soxr
```

Download and install SkyEye, along with a Whisper speech recognition model:

```sh
sudo useradd -G users skyeye
curl -sL https://github.com/dharmab/skyeye/releases/latest/download/skyeye-linux-amd64.tar.gz -o /tmp/skyeye-linux-amd64.tar.gz
tar -xzf /tmp/skyeye-linux-amd64.tar.gz -C /tmp/
sudo mkdir -p /opt/skyeye/bin /opt/skyeye/models
sudo mv /tmp/skyeye-linux-amd64/skyeye /opt/skyeye/bin/skyeye
sudo chmod +x /opt/skyeye/bin/skyeye
curl -sL https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin -o /opt/skyeye/models/ggml-small.en.bin
sudo chown -R skyeye:users /opt/skyeye
sudo mkdir -p /etc/skyeye
sudo mv /tmp/skyeye-linux-amd64/config.yaml /etc/skyeye/config.yaml
sudo chmod 600 /etc/skyeye/config.yaml
sudo chown -R skyeye:users /etc/skyeye
rm -rf /tmp/skyeye-linux-amd64.tar.gz /tmp/skyeye-linux-amd64
```

Edit the config file, setting `recognizer` to `openai-whisper-local` and `whisper-model` to `/opt/skyeye/models/ggml-small.en.bin`, along with your Tacview and SRS connection details:

```sh
sudoedit /etc/skyeye/config.yaml
```

Save this systemd unit to `/etc/systemd/system/skyeye.service`:

```ini
[Unit]
Description=SkyEye GCI Bot
After=network-online.target

[Service]
Type=simple
User=skyeye
WorkingDirectory=/opt/skyeye
ExecStart=/opt/skyeye/bin/skyeye
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
```

To start SkyEye, and enable it to start on boot:

```sh
sudo systemctl daemon-reload
sudo systemctl enable skyeye.service --now
```

## Using SkyEye

Check that SkyEye is running:

```sh
sudo systemctl status skyeye
```

Stream the logs and look for any repeated WARN or ERROR lines that don't go away:

```sh
journalctl -fu skyeye
```

Connect to your DCS game and SRS server. Switch to one of the SkyEye frequencies you configured. Try some test commands like a RADIO CHECK, ALPHA CHECK and PICTURE. (See the [player guide](../../PLAYER.md).)

To stop SkyEye:

```sh
sudo systemctl stop skyeye
```

## Upgrading SkyEye

### Container

If your Quadlet file uses the `latest` tag, force a pull of the newest image and restart SkyEye:

```sh
sudo podman pull ghcr.io/dharmab/skyeye:latest
sudo systemctl restart skyeye.service
```

If you pinned a specific version instead, edit the `Image=` line in `/etc/containers/systemd/skyeye.container` to the version you want, then reload and restart:

```sh
sudo systemctl daemon-reload
sudo systemctl restart skyeye.service
```

### Native Binary

```sh
new_version=v1.9.3
curl -sL https://github.com/dharmab/skyeye/releases/download/$new_version/skyeye-linux-amd64.tar.gz -o /tmp/skyeye-linux-amd64.tar.gz
tar -xzf /tmp/skyeye-linux-amd64.tar.gz -C /tmp/
sudo mv /tmp/skyeye-linux-amd64/skyeye /opt/skyeye/bin/skyeye
sudo chown skyeye:users /opt/skyeye/bin/skyeye
rm -rf /tmp/skyeye-linux-amd64.tar.gz /tmp/skyeye-linux-amd64
sudo systemctl restart skyeye.service
```

## Uninstalling SkyEye

### Container

```sh
sudo systemctl disable --now skyeye.service
sudo rm /etc/containers/systemd/skyeye.container
sudo systemctl daemon-reload
sudo podman rmi ghcr.io/dharmab/skyeye:latest
sudo rm -rf /etc/skyeye
```

### Native Binary

```sh
sudo systemctl disable --now skyeye.service
sudo rm /etc/systemd/system/skyeye.service
sudo systemctl daemon-reload
sudo userdel skyeye
sudo rm -rf /opt/skyeye /etc/skyeye
```

## Advanced Topics

For instructions on autoscaling and running multiple instances, see [the full admin guide](../../ADMIN.md#autoscaling-experimental).
