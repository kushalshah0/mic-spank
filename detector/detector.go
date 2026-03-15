package detector

import (
	"math"
)

// Event describes a detected slap impact.
type Event struct {
	RMS       float64 // RMS amplitude of the triggering frame
	Peak      float64 // Peak absolute sample value in the frame
	Ratio     float64 // STA/LTA ratio at trigger time
	Intensity float64 // Normalized intensity from 0.0 to 1.0
}

// Config holds all detector tuning parameters.
type Config struct {
	// RMSFloor is the minimum RMS amplitude to consider.
	// Frames below this are treated as silence.
	RMSFloor float64

	// STALTAThreshold is the short-term/long-term average ratio
	// required to trigger detection. Lower values are more sensitive.
	STALTAThreshold float64

	// CrestThreshold is the minimum peak-to-RMS ratio.
	// Impulsive sounds like slaps typically exceed 3.0.
	CrestThreshold float64

	// HFRatioThreshold is the minimum high-frequency energy ratio.
	// Slaps produce broadband noise with high zero-crossing rates.
	HFRatioThreshold float64

	// STAWindow is the number of frames in the short-term average window.
	STAWindow int

	// LTAWindow is the number of frames in the long-term average window.
	LTAWindow int
}

// DefaultConfig returns a configuration tuned for laptop slap detection.
func DefaultConfig() Config {
    return Config{
        RMSFloor:         0.02,
        STALTAThreshold:  1.5,   // was 3.0
        CrestThreshold:   1.5,   // was 2.5
        HFRatioThreshold: 0.05,  // was 0.25 — this was killing it
        STAWindow:        3,
        LTAWindow:        80,
    }
}

// Detector performs multi-stage slap detection on audio frames.
type Detector struct {
	cfg Config
	sta *RingBuffer
	lta *RingBuffer
}

// New creates a Detector with the given configuration.
func New(cfg Config) *Detector {
	return &Detector{
		cfg: cfg,
		sta: NewRingBuffer(cfg.STAWindow),
		lta: NewRingBuffer(cfg.LTAWindow),
	}
}

// Analyze processes one audio frame and returns an Event if a slap
// was detected, or nil otherwise.
func (d *Detector) Analyze(samples []float32) *Event {
	if len(samples) == 0 {
		return nil
	}

	// Compute frame statistics.
	rms, peak := frameStats(samples)

	// Stage 1: RMS floor gate.
	// Reject frames that are too quiet to be a slap.
	if rms < d.cfg.RMSFloor {
		d.sta.Push(rms)
		d.lta.Push(rms)
		return nil
	}

	// Update running averages.
	d.sta.Push(rms)
	d.lta.Push(rms)

	// Wait until the long-term buffer is fully populated.
	if !d.lta.Full() {
		return nil
	}

	// Stage 2: STA/LTA ratio.
	// A sudden energy spike produces a high ratio.
	ltaAvg := d.lta.Average()
	if ltaAvg == 0 {
		return nil
	}

	ratio := d.sta.Average() / ltaAvg
	if ratio < d.cfg.STALTAThreshold {
		return nil
	}

	// Stage 3: Crest factor.
	// Impulsive sounds have a high peak relative to their RMS.
	crest := peak / rms
	if crest < d.cfg.CrestThreshold {
		return nil
	}

	// Stage 4: High-frequency energy via zero-crossing rate.
	// Slaps are broadband noise; speech and music are more tonal.
	zcr := zeroCrossingRate(samples)
	if zcr < d.cfg.HFRatioThreshold {
		return nil
	}

	// Compute a normalized intensity value from 0.0 to 1.0.
	intensity := (ratio - d.cfg.STALTAThreshold) / (d.cfg.STALTAThreshold * 3.0)
	if intensity > 1.0 {
		intensity = 1.0
	}
	if intensity < 0.0 {
		intensity = 0.0
	}

	return &Event{
		RMS:       rms,
		Peak:      peak,
		Ratio:     ratio,
		Intensity: intensity,
	}
}

// frameStats computes the RMS and peak absolute value of a sample buffer.
func frameStats(samples []float32) (rms float64, peak float64) {
	var sumSq float64

	for _, s := range samples {
		v := float64(s)
		sumSq += v * v

		if a := math.Abs(v); a > peak {
			peak = a
		}
	}

	rms = math.Sqrt(sumSq / float64(len(samples)))
	return
}

// zeroCrossingRate returns the fraction of consecutive sample pairs
// that cross the zero axis. Values closer to 1.0 indicate noise-like
// high-frequency content.
func zeroCrossingRate(samples []float32) float64 {
	if len(samples) < 2 {
		return 0
	}

	crossings := 0
	for i := 1; i < len(samples); i++ {
		if (samples[i] >= 0 && samples[i-1] < 0) ||
			(samples[i] < 0 && samples[i-1] >= 0) {
			crossings++
		}
	}

	return float64(crossings) / float64(len(samples)-1)
}