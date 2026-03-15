package capture

import (
	"fmt"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	DefaultSampleRate = 44100
	DefaultChannels   = 1
	DefaultFrameSize  = 512
)

// Config holds microphone capture settings.
type Config struct {
	SampleRate float64
	Channels   int
	FrameSize  int
	DeviceName string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		SampleRate: DefaultSampleRate,
		Channels:   DefaultChannels,
		FrameSize:  DefaultFrameSize,
	}
}

// Stream represents an active microphone capture session.
type Stream struct {
	stream *portaudio.Stream
	buffer []float32
	cfg    Config
	mu     sync.Mutex
	closed bool
}

// Init initializes the PortAudio library. Call once at startup.
func Init() error {
	return portaudio.Initialize()
}

// Terminate cleans up PortAudio resources. Call on shutdown.
func Terminate() {
	portaudio.Terminate()
}

// ListDevices prints all available audio input devices to stdout.
func ListDevices() error {
	devices, err := portaudio.Devices()
	if err != nil {
		return fmt.Errorf("list devices: %w", err)
	}

	fmt.Println("Available audio input devices:")
	fmt.Println()

	for i, d := range devices {
		if d.MaxInputChannels > 0 {
			fmt.Printf("  [%d] %s\n", i, d.Name)
			fmt.Printf("      Channels: %d | Default Sample Rate: %.0f Hz\n",
				d.MaxInputChannels, d.DefaultSampleRate)
			fmt.Println()
		}
	}

	return nil
}

// NewStream opens the microphone and begins capturing audio.
func NewStream(cfg Config) (*Stream, error) {
	buf := make([]float32, cfg.FrameSize*cfg.Channels)

	device, err := resolveDevice(cfg.DeviceName)
	if err != nil {
		return nil, err
	}

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: cfg.Channels,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      cfg.SampleRate,
		FramesPerBuffer: cfg.FrameSize,
	}

	stream, err := portaudio.OpenStream(params, buf)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	if err := stream.Start(); err != nil {
		stream.Close()
		return nil, fmt.Errorf("start stream: %w", err)
	}

	return &Stream{
		stream: stream,
		buffer: buf,
		cfg:    cfg,
	}, nil
}

// ReadFrame reads one frame of audio samples from the microphone.
// Returns a copy of the buffer with float32 values in the range [-1.0, 1.0].
func (s *Stream) ReadFrame() ([]float32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, fmt.Errorf("stream is closed")
	}

	if err := s.stream.Read(); err != nil {
		return nil, fmt.Errorf("read frame: %w", err)
	}

	out := make([]float32, len(s.buffer))
	copy(out, s.buffer)
	return out, nil
}

// Close stops and releases the microphone stream.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	s.stream.Stop()
	return s.stream.Close()
}

// resolveDevice finds the requested input device, or returns the system default.
func resolveDevice(name string) (*portaudio.DeviceInfo, error) {
	if name == "" {
		dev, err := portaudio.DefaultInputDevice()
		if err != nil {
			return nil, fmt.Errorf("no default input device found: %w", err)
		}
		return dev, nil
	}

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("enumerate devices: %w", err)
	}

	for _, d := range devices {
		if d.Name == name && d.MaxInputChannels > 0 {
			return d, nil
		}
	}

	return nil, fmt.Errorf("input device %q not found (use --list-devices to see options)", name)
}