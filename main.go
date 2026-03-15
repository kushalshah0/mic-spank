package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/kushalshah0/mic-spank/capture"
	"github.com/kushalshah0/mic-spank/detector"
	"github.com/kushalshah0/mic-spank/modes"
	"github.com/kushalshah0/mic-spank/player"
)

const banner = `
 ███╗   ███╗██╗ ██████╗    ███████╗██████╗  █████╗ ███╗   ██╗██╗  ██╗
 ████╗ ████║██║██╔════╝    ██╔════╝██╔══██╗██╔══██╗████╗  ██║██║ ██╔╝
 ██╔████╔██║██║██║         ███████╗██████╔╝███████║██╔██╗ ██║█████╔╝ 
 ██║╚██╔╝██║██║██║         ╚════██║██╔═══╝ ██╔══██║██║╚██╗██║██╔═██╗ 
 ██║ ╚═╝ ██║██║╚██████╗    ███████║██║     ██║  ██║██║ ╚████║██║  ██╗
 ╚═╝     ╚═╝╚═╝ ╚═════╝    ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝
 Slap your laptop. It yells back. 🎤
`

func main() {
	// ── CLI Flags ──────────────────────────────────────────────────────

	audioDir := flag.String("audio", "",
		"Directory containing .mp3 response clips")
	modeName := flag.String("mode", "pain",
		"Playback mode: pain, sexy, custom")
	cooldownFlag := flag.Duration("cooldown", 750*time.Millisecond,
		"Minimum time between responses")
	volumeFlag := flag.Float64("volume", 0.8,
		"Base playback volume (0.0 to 1.0)")
	volumeScaling := flag.Bool("volume-scaling", false,
		"Scale volume proportionally to slap intensity")
	sensitivity := flag.Float64("sensitivity", 1.5,
		"STA/LTA trigger threshold (lower is more sensitive)")
	rmsFloor := flag.Float64("rms-floor", 0.02,
		"Minimum RMS to consider a frame (raise to reduce false triggers)")
	crestThresh := flag.Float64("crest", 1.5,
		"Minimum crest factor, peak divided by RMS")
	hfThresh := flag.Float64("hf-ratio", 0.05,
		"Minimum high-frequency energy ratio")
	deviceName := flag.String("device", "",
		"Audio input device name (empty uses system default)")
	listDevices := flag.Bool("list-devices", false,
		"List available input devices and exit")
	debug := flag.Bool("debug", false,
		"Print real-time detection metrics")
	quiet := flag.Bool("quiet", false,
		"Suppress audio library warnings")
	frameSize := flag.Int("frame-size", 512,
		"Audio frame size in samples")

	flag.Parse()

	fmt.Print(banner)

	// Suppress audio library warnings if --quiet is set.
	if *quiet {
		devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		syscall.Dup2(int(devNull.Fd()), syscall.Stderr)
	}

	// ── PortAudio ──────────────────────────────────────────────────────

	if err := capture.Init(); err != nil {
		log.Fatalf("PortAudio init failed: %v\n"+
			"Run: sudo apt install portaudio19-dev\n", err)
	}
	defer capture.Terminate()

	if *listDevices {
		if err := capture.ListDevices(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// ── Audio Clips ────────────────────────────────────────────────────

	mode, err := modes.ParseMode(*modeName)
	if err != nil {
		log.Fatalf("Invalid mode: %v", err)
	}

	dir := resolveAudioDir(*audioDir, mode.String())
	fmt.Printf("  Audio directory : %s\n", dir)

	mgr, err := modes.NewManager(mode, dir)
	if err != nil {
		log.Fatalf("Failed to load audio clips: %v\n"+
			"Add .mp3 files to %s or use --audio=<dir>\n", err, dir)
	}
	fmt.Printf("  Loaded clips    : %d\n", mgr.ClipCount())
	fmt.Printf("  Mode            : %s\n", mode)

	// ── Audio Output ───────────────────────────────────────────────────

	if err := player.Init(44100); err != nil {
		log.Fatalf("Audio output init failed: %v\n"+
			"Make sure PulseAudio or PipeWire is running.\n", err)
	}

	// ── Microphone ─────────────────────────────────────────────────────

	cfg := capture.Config{
		SampleRate: 44100,
		Channels:   1,
		FrameSize:  *frameSize,
		DeviceName: *deviceName,
	}

	stream, err := capture.NewStream(cfg)
	if err != nil {
		log.Fatalf("Failed to open microphone: %v\n"+
			"Run: pactl list sources short\n"+
			"Or:  ./mic-spank --list-devices\n", err)
	}
	defer stream.Close()

	fmt.Println("  Microphone      : ready")

	// ── Detector ───────────────────────────────────────────────────────

	detCfg := detector.DefaultConfig()
	detCfg.STALTAThreshold = *sensitivity
	detCfg.RMSFloor = *rmsFloor
	detCfg.CrestThreshold = *crestThresh
	detCfg.HFRatioThreshold = *hfThresh
	det := detector.New(detCfg)

	// ── Calibration ────────────────────────────────────────────────────

	fmt.Println()
	fmt.Println("  Calibrating noise floor — hold still for 2 seconds...")

	warmupFrames := int(2.0 * cfg.SampleRate / float64(cfg.FrameSize))
	for i := 0; i < warmupFrames; i++ {
		frame, err := stream.ReadFrame()
		if err != nil {
			log.Fatalf("Mic read error during calibration: %v", err)
		}
		det.Analyze(frame)
	}

	fmt.Println("  Calibration complete.")

	// ── Signal Handling ────────────────────────────────────────────────

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// ── Main Loop ──────────────────────────────────────────────────────

	fmt.Println()
	fmt.Println("  🖐️  Ready — give your laptop a good slap!")
	fmt.Println("  Press Ctrl+C to exit.")
	fmt.Println()

	var (
		lastTrigger time.Time
		slapCount   int
		playing     int32 // atomic flag: 0 = idle, 1 = playing
	)

	for {
		// Check for shutdown signal.
		select {
		case <-sigCh:
			fmt.Printf("\n  Goodbye! Received %d slap(s) this session.\n", slapCount)
			return
		default:
		}

		// Read one audio frame from the microphone.
		frame, err := stream.ReadFrame()
		if err != nil {
			if *debug {
				log.Printf("mic read error: %v", err)
			}
			continue
		}

		// Debug: show a live RMS meter.
		if *debug {
			rms := computeRMS(frame)
			bars := int(rms * 200)
			if bars > 50 {
				bars = 50
			}
			meter := ""
			for i := 0; i < bars; i++ {
				meter += "█"
			}
			fmt.Printf("\r  🎤 RMS: %.4f  %s          ", rms, meter)
		}

		// Run the detection pipeline.
		event := det.Analyze(frame)

		if event == nil {
			continue
		}

		if atomic.LoadInt32(&playing) == 1 {
			continue
		}

		if time.Since(lastTrigger) < *cooldownFlag {
			continue
		}

		// Slap confirmed.
		lastTrigger = time.Now()
		slapCount++

		clip := mgr.PickClip(event.Intensity)

		vol := *volumeFlag
		if *volumeScaling {
			vol *= 0.3 + (0.7 * event.Intensity)
		}
		if vol > 1.0 {
			vol = 1.0
		}

		fmt.Printf(
			"\n  💥 SLAP #%d  ratio=%.2f  rms=%.4f  peak=%.4f  intensity=%.0f%%  → %s\n",
			slapCount,
			event.Ratio,
			event.RMS,
			event.Peak,
			event.Intensity*100,
			filepath.Base(clip),
		)

		atomic.StoreInt32(&playing, 1)
		go func(c string, v float64) {
			if err := <-player.PlayFileAsync(c, v); err != nil {
				log.Printf("  playback error: %v", err)
			}
			atomic.StoreInt32(&playing, 0)
		}(clip, vol)
	}
}

// resolveAudioDir determines which directory to load clips from.
func resolveAudioDir(explicit string, modeName string) string {
	if explicit != "" {
		return explicit
	}

	exe, _ := os.Executable()
	candidates := []string{
		filepath.Join(filepath.Dir(exe), "audio", modeName),
		filepath.Join(".", "audio", modeName),
		filepath.Join(".", "audio", "pain"),
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}

	return filepath.Join(".", "audio", modeName)
}

// computeRMS calculates the root mean square of an audio frame.
func computeRMS(samples []float32) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}

	return math.Sqrt(sum / float64(len(samples)))
}
