# mic-spank

Slap your laptop. It yells back.

## How It Works

The app listens to your microphone and runs incoming audio through a 4-stage detection pipeline:

```
Microphone → RMS Floor → STA/LTA Ratio → Crest Factor → Zero-Crossing Rate → Play Response
```

1. **RMS Floor**: Rejects frames that are too quiet
2. **STA/LTA**: Detects sudden energy spikes (Short-Term vs Long-Term average)
3. **Crest Factor**: Identifies impulsive sounds (high peak-to-RMS ratio)
4. **Zero-Crossing Rate**: Distinguishes noise-like slaps from speech/music

On startup, the detector calibrates its noise floor for 2 seconds before enabling detection.

## Prerequisites

### Linux (Ubuntu/Debian)

```bash
sudo apt install portaudio19-dev libasound2-dev
```

### Linux (Fedora)

```bash
sudo dnf install portaudio-devel alsa-lib-devel
```

### macOS

```bash
brew install portaudio
```

### Windows

PortAudio is bundled with the Go package. No extra installation needed.

## Requirements

- Go 1.24 or later

## Install

```bash
# Clone the repository
git clone https://github.com/kushalshah0/mic-spank.git
cd mic-spank

# Download dependencies
go mod download

# Build
go build -o mic-spank
```

## Run

```bash
./mic-spank                    # pain mode (default)
./mic-spank --mode=sexy       # escalating responses
./mic-spank --mode=halo       # intensity-based responses
./mic-spank --mode=custom --audio=/path/to/clips
./mic-spank --quiet           # suppress ALSA warnings
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--mode` | `pain` | Response mode: pain, sexy, halo, custom |
| `--audio` | auto | Custom audio directory |
| `--device` | default | Audio input device name |
| `--sensitivity` | 1.5 | STA/LTA threshold (lower = more sensitive) |
| `--rms-floor` | 0.02 | Minimum RMS to consider |
| `--crest` | 1.5 | Minimum crest factor |
| `--hf-ratio` | 0.05 | High-frequency ratio threshold |
| `--cooldown` | 750ms | Minimum time between responses |
| `--volume` | 0.8 | Playback volume |
| `--volume-scaling` | false | Scale volume by slap intensity |
| `--debug` | false | Show real-time RMS meter |
| `--quiet` | false | Suppress audio library warnings |
| `--list-devices` | false | List audio input devices |

## Modes

| Mode | Behavior |
|------|----------|
| **pain** | Plays a random pain sound on each slap |
| **sexy** | Escalates through clips based on slap frequency within a 5-minute rolling window |
| **halo** | Selects clip based on detected slap intensity (0-100%) |
| **custom** | Plays random clip from user-provided directory |

## Audio

Place `.mp3` files in `audio/pain/`, `audio/sexy/`, or `audio/halo/`. Clips are sorted alphabetically and selected based on the active mode.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌───────────┐     ┌─────────┐
│  capture/   │────▶│   detector/  │────▶│  modes/   │────▶│ player/ │
│  Microphone │     │  4-stage DSP │     │  Selection│     │  MP3    │
│  (PortAudio)│     │  (STA/LTA)   │     │           │     │         │
└─────────────┘     └──────────────┘     └───────────┘     └─────────┘
       │                  │                                       │
       └──────────────────┴───────────────────────────────────────┘
                     main.go (orchestration + CLI)
```

## Troubleshooting

### No microphone input
- Check available devices: `./mic-spank --list-devices`
- Specify a device: `./mic-spank --device="Your Microphone Name"`

### Detection not working
- Run with `--debug` to see real-time RMS values
- Lower sensitivity: `./mic-spank --sensitivity=1.0`
- Adjust `--rms-floor` and `--crest` thresholds as needed

### Audio playback issues
- Ensure PulseAudio or PipeWire is running
- Check system volume settings

### ALSA warnings on Linux
- Use `--quiet` to suppress warnings
- Or run: `echo "pcm.!default { type null }" > ~/.asoundrc`

### Permission denied on microphone
- On Linux, you may need to add yourself to the `audio` group: `sudo usermod -aG audio $USER`

## License

MIT License

Copyright (c) 2026 Kushal Shah

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
