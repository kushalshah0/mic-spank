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
./mic-spank --mode=halo       # intensity-based responses
./mic-spank --mode=custom --audio=/path/to/clips
./mic-spank --quiet           # suppress ALSA warnings
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--mode` | `pain` | Response mode: pain, halo, custom |
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
| **halo** | Selects clip based on detected slap intensity (0-100%) |
| **custom** | Plays random clip from user-provided directory |

## Audio

Place `.mp3` files in `audio/pain/` or `audio/halo/`. Clips are sorted alphabetically and selected based on the active mode.

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

MIT License — see [LICENSE](LICENSE) for details.
 