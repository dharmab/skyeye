# Deploy SkyEye on Linux - GPU

This guide is a step-by-step on how to run SkyEye on the same Linux computer as DCS, TacView and SRS Server, using the GPU for local speech recognition via the experimental Vulkan build.

You can also deploy SkyEye on Linux using either [a separate computer using CPU speech recognition](cpu.md) or [cloud API speech recognition](api.md).

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

## Install a GPU Driver

Your GPU driver needs to be installed and up to date on the host system, whether you install SkyEye as a container or a native binary. Install it through your distribution's package manager, not a vendor installer. The `amdgpu` and `i915` kernel drivers for AMD and Intel GPUs ship in the mainline Linux kernel, but you still need the Mesa userspace Vulkan driver (and, for AMD, firmware blobs) installed separately. NVIDIA GPUs need the proprietary driver package, which includes both the kernel module and the Vulkan driver.

```sh
# AMD - Debian/Ubuntu
sudo apt-get install mesa-vulkan-drivers firmware-amd-graphics

# AMD - Fedora
sudo dnf install mesa-vulkan-drivers

# AMD - Arch Linux
sudo pacman -Syu vulkan-radeon linux-firmware

# Intel - Debian/Ubuntu
sudo apt-get install mesa-vulkan-drivers

# Intel - Fedora
sudo dnf install mesa-vulkan-drivers

# Intel - Arch Linux
sudo pacman -Syu vulkan-intel

# NVIDIA (proprietary) - Debian/Ubuntu
sudo apt-get install nvidia-driver

# NVIDIA (proprietary) - Fedora (via RPM Fusion)
sudo dnf install akmod-nvidia

# NVIDIA (proprietary) - Arch Linux
sudo pacman -Syu nvidia
```

Package names vary between distributions and releases, and some distributions split Vulkan support across additional packages beyond what's listed here (e.g. separate ICD loader or 32-bit compatibility packages). Consult your distribution's documentation if these don't match, and use `vulkaninfo` to confirm your GPU is detected before proceeding.

## Set up SkyEye

You can install SkyEye either as a Vulkan container image, or as a native Vulkan binary. The container is easier to set up and upgrade; the native binary is useful if you don't want to run a container runtime.

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

If you have an NVIDIA GPU, install the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html).

If you have an AMD or Intel GPU, no extra toolkit is needed. The Vulkan container image bundles the necessary Mesa drivers.

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

Create a [Podman Quadlet](https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html) file at `/etc/containers/systemd/skyeye.container`, using the GPU passthrough appropriate for your vendor:

```ini
# NVIDIA
[Unit]
Description=SkyEye GCI Bot
After=network-online.target

[Container]
Image=ghcr.io/dharmab/skyeye:latest-vulkan
ContainerName=skyeye
Volume=/etc/skyeye:/etc/skyeye:ro
PodmanArgs=--gpus all

[Service]
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
```

```ini
# AMD/Intel
[Unit]
Description=SkyEye GCI Bot
After=network-online.target

[Container]
Image=ghcr.io/dharmab/skyeye:latest-vulkan
ContainerName=skyeye
Volume=/etc/skyeye:/etc/skyeye:ro
AddDevice=/dev/dri

[Service]
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target
```

I recommend pinning `Image=` to a specific version instead of `latest-vulkan`, to avoid unexpected breaking changes when a new version is released. Find version tags on the [releases page](https://github.com/dharmab/skyeye/releases), e.g. `Image=ghcr.io/dharmab/skyeye:v1.9.3-vulkan`.

Load the Quadlet and start SkyEye, enabling it to start on boot:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now skyeye.service
```

### Native Binary

Install the Vulkan loader, in addition to the GPU driver installed in the previous section:

```sh
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install libopus0 libsoxr0 libvulkan1

# Fedora
sudo dnf install opus sox vulkan-loader

# Arch Linux
sudo pacman -Syu opus soxr vulkan-icd-loader
```

Download and install SkyEye, along with a Whisper speech recognition model:

```sh
sudo useradd -G users skyeye
curl -sL https://github.com/dharmab/skyeye/releases/latest/download/skyeye-linux-amd64-vulkan.tar.gz -o /tmp/skyeye-linux-amd64-vulkan.tar.gz
tar -xzf /tmp/skyeye-linux-amd64-vulkan.tar.gz -C /tmp/
sudo mkdir -p /opt/skyeye/bin /opt/skyeye/models
sudo mv /tmp/skyeye-linux-amd64-vulkan/skyeye /opt/skyeye/bin/skyeye
sudo chmod +x /opt/skyeye/bin/skyeye
curl -sL https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin -o /opt/skyeye/models/ggml-small.en.bin
sudo chown -R skyeye:users /opt/skyeye
sudo mkdir -p /etc/skyeye
sudo mv /tmp/skyeye-linux-amd64-vulkan/config.yaml /etc/skyeye/config.yaml
sudo chmod 600 /etc/skyeye/config.yaml
sudo chown -R skyeye:users /etc/skyeye
rm -rf /tmp/skyeye-linux-amd64-vulkan.tar.gz /tmp/skyeye-linux-amd64-vulkan
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

If your Quadlet file uses the `latest-vulkan` tag, force a pull of the newest image and restart SkyEye:

```sh
sudo podman pull ghcr.io/dharmab/skyeye:latest-vulkan
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
curl -sL https://github.com/dharmab/skyeye/releases/download/$new_version/skyeye-linux-amd64-vulkan.tar.gz -o /tmp/skyeye-linux-amd64-vulkan.tar.gz
tar -xzf /tmp/skyeye-linux-amd64-vulkan.tar.gz -C /tmp/
sudo mv /tmp/skyeye-linux-amd64-vulkan/skyeye /opt/skyeye/bin/skyeye
sudo chown skyeye:users /opt/skyeye/bin/skyeye
rm -rf /tmp/skyeye-linux-amd64-vulkan.tar.gz /tmp/skyeye-linux-amd64-vulkan
sudo systemctl restart skyeye.service
```

## Uninstalling SkyEye

### Container

```sh
sudo systemctl disable --now skyeye.service
sudo rm /etc/containers/systemd/skyeye.container
sudo systemctl daemon-reload
sudo podman rmi ghcr.io/dharmab/skyeye:latest-vulkan
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
