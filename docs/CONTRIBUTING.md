# Knowledge

Requirements to develop SkyEye:

- Windows or Linux PC
  - If on Windows, willing to learn to use Visual Studio Code
  - This could probably work on macOS similar to how it works on Linux. Cross-compilation may or may not work on Apple Silicon- I'm not sure.
- Beginner level skills in the Go programming language. If you already know another programming language, [A Tour of Go](https://go.dev/tour) can get you up to speed in an afternoon.
- Comfortable with Git
- Familiar with *nix command line basics (not much, mostly `cd` and `make`)
- Familiar with building C/C++ projects is a plus, but not required

# Setup

## Build

### Windows

[Install MSYS2](https://www.msys2.org/#installation).

Run the MSYS2 UCRT application from the start menu.

Run `pacman -Syu --needed git base-devel`. If prompted to select a package from a list, accept the defaults. If the application prompts you to restart, restart and run the command again.

Clone this Git repository somewhere, and navigate to it in the MSYS2 UCRT terminal. Your `C:\` is available at `/c`, so you can access your Documents folder with `cd '/c/Documents and Settings/yourusername/Documents/'`. Similarly, your `D:\` will be at `/d` if present, and so on.

Run `make install-dependencies` to install the C++ and Go compilers as well as some build dependencies.

Run `make` to build `SkyEye.exe`.

### Arch Linux

Full guide/Makefile updates TODO

Basically run this:

```sh
pacman -Syu base-devel go opus libsoxr
make whisper
CGO_ENABLED=1 C_INCLUDE_PATH=third_party/whisper.cpp LIBRARY_PATH=third_party/whisper.cpp go build ./cmd/skyeye
```

And everywhere this guide mentions `skyeye.exe`, remove `.exe`

## Run

Install the [DCS World Dedicated Server](https://www.digitalcombatsimulator.com/en/downloads/world/server/). This can be on a different computer.

Install [DCS gRPC Server](https://github.com/DCS-gRPC/rust-server) on the same machine as your DCS server and configure your DCS server to run DCS gRPC.

Install [DCS-SRS](http://dcssimpleradio.com/). This can be on a different computer.

Launch the DCS server and SRS server. Load a mission on the DCS server. TODO better guide for this stuff.

You will need to download an OpenAI Whisper model. The main source of these models is [Hugging Face](https://huggingface.co/ggerganov/whisper.cpp/tree/main)] The larger models have better accuracy but higher memory consumption and take longer to recognize text. There are also faster distilled models available [here](https://huggingface.co/distil-whisper/distil-medium.en#whispercpp), [although these have some quality trade-offs with the library used in this software.](https://github.com/ggerganov/whisper.cpp/tree/master/models#distilled-models). Whichever model you choose, put the next to `skyeye.exe`.

Run SkyEye by passing command line flags to `skyeye.exe`. You can run `./skyeye.exe -help` for some hints. A simple example:

```sh
./skyeye.exe \
  -dcs-grpc-server-address=your-dcs-grpc-server-ip:50051 \
  -srs-server-address=your-srs-server-ip:5002 \
  -srs-eam-password=yourSRSEAMpassword \
  -whisper-model=ggml-medium.en.bin
```

If all goes well, you should see the SkyEye software start up and start logging JSON lines to the console.

## Develop

### Windows

Install [Visual Studio Code](https://code.visualstudio.com/).

Configure Visual Studio Code for [Go development](https://learn.microsoft.com/en-us/azure/developer/go/configure-visual-studio-code) and [GCC with MinGW](https://code.visualstudio.com/docs/cpp/config-mingw).

For convenience, add MSYS2 to Visual Studio Code's integrated terminal. Open your User `settings.json`, use IntelliSense to complete `terminal.integrated.profiles.windows`, and add this object to the array:

```json
"MSYS2": {
    "path": "C:\\msys64\\usr\\bin\\bash.exe",
    "args": [
        "--login",
        "-i"
    ],
    "env": {
        "MSYSTEM": "MINGW64",
        "CHERE_INVOKING": "1"
    }
}
```

I don't have this project set up to build/run/debug through VSC yet.

### Linux

🐧 Use your favorite editor.


TODO guide to project architecture, file and package layout, entrypoints