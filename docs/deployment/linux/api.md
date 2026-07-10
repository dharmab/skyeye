# Deploy SkyEye on Linux - Cloud API

This guide is a step-by-step on how to run SkyEye on a Linux computer or server alongside DCS, TacView and SRS Server, using the OpenAI API for cloud speech recognition.

You can also deploy SkyEye on Linux using either [a separate computer using CPU speech recognition](cpu.md) or [the same computer as DCS using GPU speech recognition](gpu.md).

## Getting Help

See [the admin guide](../../ADMIN.md#getting-help) for how to get help if you have a problem.

## Set up OpenAI API

Go to https://platform.openai.com. If you haven't previously set up the OpenAI API, you'll go through a step-by-step process to set up an organization, buy credits, create a Project and generate an API key.

Otherwise, go to https://platform.openai.com/settings/organization/api-keys and generate a new API key for SkyEye. I recommend adding this API key to a new Project named "SkyEye".

Be sure to review the pricing of the "Whisper" audio model at https://openai.com/api/pricing. You will pay per-second for each second a player transmits on a SkyEye frequency in SRS. (You are not charged for the time SkyEye itself spends transmitting, nor are you charged for players' transmissions on other frequencies.)

## Set Up DCS, TacView, and SRS

Install DCS (either the client for singleplayer/hosted multiplayer use, or a dedicated server for multiplayer use).

Install the TacView Exporter. Within DCS, go to OPTIONS → SPECIAL → Tacview and enable Real-Time Telemetry. (See https://www.tacview.net/documentation/dcs/en/)

Install SRS. Start and configure the SRS Server. Ensure EAM mode is enabled and an EAM password is set.

Make sure DCS is running an unpaused mission with both friendly and hostile air units so you can do some tests.

## Set up SkyEye

You can install SkyEye either as a container, or as a native binary. The container is easier to set up and upgrade; the native binary is useful if you don't want to run a container runtime.

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

Set `recognizer` to `openai-whisper-api` and set `openai-api-key` (or provide it via environment variable; see [the configuration section](../../ADMIN.md#configuration)), along with your Tacview and SRS connection details.

Create a [Podman Quadlet](https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html) file at `/etc/containers/systemd/skyeye.container`:

```ini
[Unit]
Description=SkyEye GCI Bot
After=network-online.target

[Container]
Image=ghcr.io/dharmab/skyeye:latest
ContainerName=skyeye
Volume=/etc/skyeye/config.yaml:/etc/skyeye/config.yaml:ro

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

Download and install SkyEye:

```sh
sudo useradd -G users skyeye
curl -sL https://github.com/dharmab/skyeye/releases/latest/download/skyeye-linux-amd64.tar.gz -o /tmp/skyeye-linux-amd64.tar.gz
tar -xzf /tmp/skyeye-linux-amd64.tar.gz -C /tmp/
sudo mkdir -p /opt/skyeye/bin
sudo mv /tmp/skyeye-linux-amd64/skyeye /opt/skyeye/bin/skyeye
sudo chmod +x /opt/skyeye/bin/skyeye
sudo chown -R skyeye:users /opt/skyeye
sudo mkdir -p /etc/skyeye
sudo mv /tmp/skyeye-linux-amd64/config.yaml /etc/skyeye/config.yaml
sudo chmod 600 /etc/skyeye/config.yaml
sudo chown -R skyeye:users /etc/skyeye
rm -rf /tmp/skyeye-linux-amd64.tar.gz /tmp/skyeye-linux-amd64
```

Edit the config file, setting `recognizer` to `openai-whisper-api` and providing your `openai-api-key`, along with your Tacview and SRS connection details:

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
