# mic-spank

Slap your laptop. It yells back.

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
| `--sensitivity` | 1.5 | Detection threshold (lower = more sensitive) |
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

- **pain**: Random pain sound on each slap
- **sexy**: Escalates through clips based on slap frequency
- **halo**: Selects clip based on slap intensity
- **custom**: Random from user-provided directory

## Audio

Place `.mp3` files in `audio/pain/`, `audio/sexy/`, or `audio/halo/`. Clips are sorted alphabetically and selected based on the active mode.

## Troubleshooting

### No microphone input
- Check available devices: `./mic-spank --list-devices`
- Specify device: `./mic-spank --device="Your Microphone Name"`

### Detection not working
- Run with `--debug` to see real-time RMS values
- Lower sensitivity: `--sensitivity=1.0`
- Adjust thresholds as needed

### Audio playback issues
- Ensure PulseAudio or PipeWire is running
- Check system volume settings
