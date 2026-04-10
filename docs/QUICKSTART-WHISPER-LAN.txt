# LAN Whisper Server Quickstart (Windows, CPU)

This guide explains how to offload SkyEye's speech recognition to another Windows computer on your local network using whisper.cpp server, so the DCS host machine is not burdened by CPU-intensive speech recognition.

## Overview

```
┌──────────────────────┐          ┌──────────────────────┐
│   DCS Host Machine   │          │  LAN Whisper Server   │
│                      │          │                       │
│  DCS World           │   LAN    │  whisper-server.exe   │
│  SRS Server          │ -------> │  (speech recognition) │
│  SkyEye              │  HTTP    │                       │
│  (--recognizer       │          │  CPU does the heavy   │
│    whisper-lan)      │          │  work here instead    │
└──────────────────────┘          └──────────────────────┘
```

SkyEye sends audio over HTTP to a whisper.cpp server running on a second machine. This avoids the CPU load that local whisper processing places on the DCS host.

## Requirements

- A second Windows PC on the same LAN as the DCS host
- CPU with AVX2 support (Intel Haswell 2013+ or AMD Excavator 2015+)
- At least 4 GB of free RAM (more for larger models)
- 4+ CPU cores recommended

## Step 1: Download whisper.cpp Server

On the **LAN server machine**, download the latest whisper.cpp release from:

https://github.com/ggml-org/whisper.cpp/releases

Download the Windows package (e.g., `whisper-cublas-...` for NVIDIA GPU or `whisper-bin-x64.zip` for CPU-only). Extract the zip to a folder, for example `C:\whisper`.

The file you need is `whisper-server.exe` inside the `bin` folder.

## Step 2: Download a Whisper Model

Download a whisper.cpp-compatible GGML model from:

https://huggingface.co/ggerganov/whisper.cpp/tree/main

Recommended models for SkyEye speech recognition:

| Model | File | RAM Usage | Quality | Speed |
|-------|------|-----------|---------|-------|
| **Small English (recommended)** | `ggml-small.en.bin` | ~1 GB | Good balance | Fast |
| Medium English | `ggml-medium.en.bin` | ~2 GB | Better quality | Slower |
| Tiny English | `ggml-tiny.en.bin` | ~200 MB | Lower quality | Fastest |

Save the model file to the same folder as `whisper-server.exe`, for example `C:\whisper\bin\ggml-small.en.bin`.

## Step 3: Start the Whisper Server

Open PowerShell on the LAN server machine and navigate to the folder containing `whisper-server.exe`:

```powershell
cd C:\whisper\bin
```

Start the server with these flags:

```powershell
.\whisper-server.exe `
  --model ggml-small.en.bin `
  --host 0.0.0.0 `
  --port 8080 `
  --request-path /v1/audio `
  --inference-path /transcriptions `
  --language en `
  --threads 4
```

**Flag explanations:**

| Flag | Purpose |
|------|---------|
| `--model` | Path to the GGML model file |
| `--host 0.0.0.0` | Listen on all network interfaces (required for LAN access) |
| `--port 8080` | Port number (change if needed) |
| `--request-path /v1/audio` | Required — makes the server OpenAI API compatible |
| `--inference-path /transcriptions` | Required — makes the server OpenAI API compatible |
| `--language en` | English only (matches SkyEye's usage) |
| `--threads 4` | Number of CPU threads to use (adjust to your CPU core count) |

> **Important:** The `--request-path /v1/audio` and `--inference-path /transcriptions` flags are required so the server listens on `/v1/audio/transcriptions`, which is the endpoint SkyEye expects. Without these flags, the server uses `/inference` instead and SkyEye will not be able to connect.

You should see output like:

```
whisper server listening at http://0.0.0.0:8080
```

## Step 4: Verify the Server Is Running

From the **DCS host machine**, open a browser and navigate to:

```
http://<LAN-SERVER-IP>:8080/
```

Replace `<LAN-SERVER-IP>` with the IP address of the whisper server machine (e.g., `192.168.1.100`). You can find this by running `ipconfig` in PowerShell on the server machine.

You should see the whisper.cpp server web interface. If you cannot reach it:

- Check Windows Firewall on the server machine — you may need to allow inbound TCP port 8080
- Verify both machines are on the same network
- Try pinging the server machine from the DCS host

### Windows Firewall Rule

If the server is not reachable, create a firewall rule on the **LAN server machine**. Open PowerShell as Administrator and run:

```powershell
New-NetFirewallRule -DisplayName "Whisper Server" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
```

## Step 5: Configure SkyEye

On the **DCS host machine**, edit your SkyEye `config.yaml`:

```yaml
recognizer: whisper-lan
whisper-lan-endpoint: http://192.168.1.100:8080/v1
```

Replace `192.168.1.100` with the actual IP address of your LAN whisper server.

The `whisper-lan-model` option defaults to `whisper-1` and can usually be left as-is — whisper.cpp server ignores the model name since it loads the model at startup.

If your whisper server requires an API key (most LAN setups do not), add:

```yaml
whisper-lan-api-key: your-api-key-here
```

Or, if using command-line flags instead of `config.yaml`:

```powershell
.\skyeye.exe `
  --recognizer whisper-lan `
  --whisper-lan-endpoint http://192.168.1.100:8080/v1 `
  <other flags...>
```

## Step 6: Test

Start SkyEye and check the logs. You should see:

1. SkyEye starts without errors
2. When a player transmits on SRS, the audio is sent to the LAN whisper server for recognition
3. The whisper server console shows incoming requests being processed

Try a RADIO CHECK or ALPHA CHECK command ([player guide](PLAYER.md)).

## Running the Server Automatically on Boot

To run the whisper server as a Windows service that starts on boot, you can use [NSSM](https://nssm.cc/) (the Non-Sucking Service Manager):

1. Download NSSM from https://nssm.cc/download
2. Open PowerShell as Administrator
3. Run:

```powershell
.\nssm.exe install WhisperServer C:\whisper\bin\whisper-server.exe "--model C:\whisper\bin\ggml-small.en.bin --host 0.0.0.0 --port 8080 --request-path /v1/audio --inference-path /transcriptions --language en --threads 4"
.\nssm.exe start WhisperServer
```

To stop or remove the service:

```powershell
.\nssm.exe stop WhisperServer
.\nssm.exe remove WhisperServer confirm
```

## Tuning

### Thread Count

Set `--threads` to the number of physical CPU cores on the server machine. Using more threads than physical cores will not improve performance.

### Model Selection

- Start with `ggml-small.en.bin` — it provides the best balance of speed and accuracy for SkyEye
- If recognition is too slow (players have to wait too long for responses), try `ggml-tiny.en.bin`
- If recognition quality is poor (SkyEye frequently misunderstands commands), try `ggml-medium.en.bin`

### Static IP

Consider assigning a static IP address to the whisper server machine so the `whisper-lan-endpoint` URL does not change.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| SkyEye logs `error transcribing audio` | Verify the server is running and reachable from the DCS host |
| Server not reachable from DCS host | Check firewall, verify IP address, ensure both machines are on the same LAN |
| Slow recognition | Reduce model size, increase `--threads`, or use a more powerful CPU |
| `whisper-server.exe` crashes on start | Ensure your CPU supports AVX2; try the CPU-only build if using the wrong package |
| SkyEye connects but gets no transcription | Verify the `--request-path /v1/audio --inference-path /transcriptions` flags are set on the server |
