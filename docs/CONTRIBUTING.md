# Knowledge

Requirements to develop SkyEye:

- Windows or Linux PC
  - If on Windows, willing to learn to use Visual Studio Code
  - This could probably work on macOS similar to how it works on Linux. Cross-compilation may or may not work on Apple Silicon- I'm not sure.
- Beginner level skills in the Go programming language. If you already know another programming language, [A Tour of Go](https://go.dev/tour) can get you up to speed in an afternoon.
- Comfortable with Git
- Familiar with *nix command line basics (not much, mostly `cd` and `make`)
- Familiar with building C/C++ projects is a plus, but not required

# Run and Debug

## Windows

Install the [DCS World Dedicated Server](https://www.digitalcombatsimulator.com/en/downloads/world/server/). This can be on a different computer.

Install [DCS gRPC Server](https://github.com/DCS-gRPC/rust-server) on the same machine as your DCS server and configure your DCS server to run DCS gRPC.

Install [DCS-SRS](http://dcssimpleradio.com/). This can be on a different computer.

You will need to download an OpenAI Whisper model. The main source of these models is [Hugging Face](https://huggingface.co/ggerganov/whisper.cpp/tree/main)] The larger models have better accuracy but higher memory consumption and take longer to recognize text. There are also faster distilled models available [here](https://huggingface.co/distil-whisper/distil-medium.en#whispercpp), [although these have some quality trade-offs with the library used in this software.](https://github.com/ggerganov/whisper.cpp/tree/master/models#distilled-models). Whichever model you choose stick it in `models/`.

Install:

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Visual Studio Code](https://code.visualstudio.com/
- Visual Studio Code Extensions:
  - Required: [Go](https://marketplace.visualstudio.com/items?itemName=golang.Go)
  - Required: [Docker](https://marketplace.visualstudio.com/items?itemName=ms-azuretools.vscode-docker)
  - Recommended: [Log Viewer](https://marketplace.visualstudio.com/items?itemName=berublan.vscode-log-viewer)

Launch the DCS server and SRS server. Load a mission on the DCS server. TODO better guide for this stuff.

Edit `docker-compose.yaml` with the addresses and passwords for your DCS and SRS server. If they're running locally, use `host.docker.internal` here instead of `localhost` or `127.0.0.1`. Also use the filename of whatever model you picked.

The launch configuration "Debug in Docker Compose" will launch a Docker container running a remote debugger, attach the Visual Studio Code debugger to the container, and run Skyeye in the background. When the debugging session ends, the container will stop.

What this means to you is that you _should_ be able to just hit F5 to start Skyeye and Shift+F5 to stop Skyeye.

You can view the Skyeye logs opening the Docker view in Visual Studio Code, right-clicking on "Skyeye" and clicking "Compose Logs" 

## Arch Linux

Full guide TODO - use the Dockerfile as a guide for now

