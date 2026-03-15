package modes

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Mode represents the audio response behavior.
type Mode int

const (
	// Pain plays a random clip on each slap.
	Pain Mode = iota

	// Sexy escalates through clips based on slap frequency in a rolling window.
	Sexy

	// Halo plays clips based on slap intensity.
	Halo

	// Custom plays random clips from a user-provided directory.
	Custom
)

// String returns the display name of the mode.
func (m Mode) String() string {
	switch m {
	case Pain:
		return "pain"
	case Sexy:
		return "sexy"
	case Halo:
		return "halo"
	case Custom:
		return "custom"
	default:
		return "unknown"
	}
}

// ParseMode converts a string to a Mode value.
func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(s) {
	case "pain":
		return Pain, nil
	case "sexy":
		return Sexy, nil
	case "halo":
		return Halo, nil
	case "custom":
		return Custom, nil
	default:
		return Pain, fmt.Errorf("unknown mode %q (options: pain, sexy, halo, custom)", s)
	}
}

// Manager handles clip selection based on the active mode.
type Manager struct {
	mode  Mode
	clips []string

	// Sexy mode state.
	mu         sync.Mutex
	hitTimes   []time.Time
	sexyWindow time.Duration
}

// NewManager creates a Manager and loads MP3 clips from the given directory.
func NewManager(mode Mode, audioDir string) (*Manager, error) {
	clips, err := loadMP3s(audioDir)
	if err != nil {
		return nil, err
	}
	if len(clips) == 0 {
		return nil, fmt.Errorf("no .mp3 files found in %s", audioDir)
	}

	sort.Strings(clips)

	return &Manager{
		mode:       mode,
		clips:      clips,
		sexyWindow: 5 * time.Minute,
	}, nil
}

// PickClip selects the next audio clip to play based on the current mode.
func (m *Manager) PickClip(intensity float64) string {
	switch m.mode {
	case Sexy:
		return m.pickSexy()
	case Halo:
		return m.pickHalo(intensity)
	default:
		return m.clips[rand.Intn(len(m.clips))]
	}
}

// ClipCount returns the number of loaded clips.
func (m *Manager) ClipCount() int {
	return len(m.clips)
}

// pickSexy tracks hits in a rolling time window and escalates the clip index.
func (m *Manager) pickSexy() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.hitTimes = append(m.hitTimes, now)

	// Remove hits that have fallen outside the rolling window.
	cutoff := now.Add(-m.sexyWindow)
	filtered := m.hitTimes[:0]
	for _, t := range m.hitTimes {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	m.hitTimes = filtered

	// Map hit count to a clip index for escalation.
	level := len(m.hitTimes) - 1
	if level >= len(m.clips) {
		level = len(m.clips) - 1
	}

	return m.clips[level]
}

// pickHalo selects a clip based on slap intensity (0.0 to 1.0).
func (m *Manager) pickHalo(intensity float64) string {
	level := int(intensity * float64(len(m.clips)-1))
	if level >= len(m.clips) {
		level = len(m.clips) - 1
	}
	return m.clips[level]
}

// loadMP3s recursively finds all .mp3 files in a directory.
func loadMP3s(dir string) ([]string, error) {
	var clips []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			clips = append(clips, path)
		}
		return nil
	})

	return clips, err
}
