package player

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/ebitengine/oto/v3"
)

var (
	ctx     *oto.Context
	once    sync.Once
	initErr error
)

// Init initializes the audio output context. Must be called once before playback.
// sampleRate is typically 44100.
func Init(sampleRate int) error {
	once.Do(func() {
		var ready chan struct{}

		// oto/v2 API: NewContext(sampleRate, channelCount, bitDepthInBytes)
		//   sampleRate:      44100
		//   channelCount:    2 (stereo, standard for MP3)
		//   bitDepthInBytes: 2 (16-bit signed int = 2 bytes)
		ctx, ready, initErr = oto.NewContext(&oto.NewContextOptions{
    	SampleRate:   sampleRate,
    	ChannelCount: 2,
   		Format:       oto.FormatSignedInt16LE,})
		if initErr != nil {
			return
		}
		<-ready
	})
	return initErr
}

// PlayFile decodes an MP3 file and plays it at the given volume.
// Blocks until playback is complete.
func PlayFile(path string, volume float64) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return playBytes(data, volume)
}

// PlayFileAsync starts playback in a goroutine and returns a channel
// that receives nil on success or an error when playback finishes.
func PlayFileAsync(path string, volume float64) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- PlayFile(path, volume)
		close(ch)
	}()
	return ch
}

// playBytes decodes raw MP3 data and plays it through the audio context.
func playBytes(data []byte, volume float64) error {
	if ctx == nil {
		return fmt.Errorf("audio not initialized: call player.Init() first")
	}

	dec, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("mp3 decode: %w", err)
	}

	// Create a new player from the decoded MP3 stream.
	p := ctx.NewPlayer(dec)

	// Set volume if supported (oto v2.2+).
	// volume: 0.0 = silent, 1.0 = full
	if volume < 0 {
		volume = 0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	p.SetVolume(volume)

	// Start playback.
	p.Play()

	// Block until playback finishes.
	for p.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}

	// Close the player to release resources.
	return p.Close()
}