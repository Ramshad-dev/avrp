package screenshare

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mysterion/avrp/internal/utils"
)

const PlaylistName = "stream.m3u8"

type Capture struct {
	cmd      *exec.Cmd
	dir      string
	cfg      Config
	done     chan struct{}
	stopOnce sync.Once
}

func (c *Capture) Dir() string {
	return c.dir
}

func (c *Capture) Config() Config {
	return c.cfg
}

// Start launches ffmpeg and waits until the HLS playlist is usable.
func Start(cfg Config, ffmpeg string) (*Capture, error) {
	if !cfg.Enabled {
		return nil, errors.New("screenshare is not enabled")
	}
	if runtime.GOOS != "windows" {
		return nil, errors.New("screenshare currently requires Windows gdigrab")
	}

	dir := filepath.Join(utils.ConfigDir, "screenshare")
	if err := os.RemoveAll(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	args := buildArgs(cfg, dir)
	log.Printf("screenshare: starting %s %dx%d@%d, %d Mbps, audio=%v",
		cfg.Source, cfg.Width, cfg.Height, cfg.FPS, cfg.BitrateMbps, cfg.Audio)

	cmd := exec.Command(ffmpeg, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	c := &Capture{
		cmd:  cmd,
		dir:  dir,
		cfg:  cfg,
		done: make(chan struct{}),
	}
	go func() {
		_ = cmd.Wait()
		close(c.done)
	}()

	if err := waitForPlaylist(dir); err != nil {
		c.Stop()
		return nil, fmt.Errorf("screenshare failed to start: %w", err)
	}

	return c, nil
}

// Stop terminates ffmpeg and removes the generated segments.
func (c *Capture) Stop() {
	if c == nil {
		return
	}
	c.stopOnce.Do(func() {
		if c.cmd != nil && c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
		select {
		case <-c.done:
		case <-time.After(3 * time.Second):
			log.Println("screenshare: ffmpeg did not exit in time")
		}
		if err := os.RemoveAll(c.dir); err != nil {
			log.Printf("screenshare: could not clean %s: %v", c.dir, err)
		}
		log.Println("screenshare: stopped")
	})
}

func buildArgs(cfg Config, dir string) []string {
	input := "desktop"
	if cfg.Source == "window" {
		input = "title=" + cfg.Window
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-y",
		"-thread_queue_size", "4096",
		"-f", "gdigrab",
		"-framerate", fmt.Sprintf("%d", cfg.FPS),
		"-draw_mouse", "1",
		"-i", input,
	}

	if cfg.Audio && cfg.AudioDevice != "" {
		args = append(args,
			"-thread_queue_size", "4096",
			"-f", "dshow",
			"-i", "audio="+cfg.AudioDevice,
		)
	}

	vf := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1",
		cfg.Width, cfg.Height, cfg.Width, cfg.Height,
	)

	gop := cfg.FPS * cfg.KeyframeSeconds
	maxRateKbps := cfg.BitrateMbps * 1200
	bufSizeMb := cfg.BitrateMbps * 2

	args = append(args,
		"-vf", vf,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-bf", "0",
		"-g", fmt.Sprintf("%d", gop),
		"-keyint_min", fmt.Sprintf("%d", gop),
		"-sc_threshold", "0",
		"-pix_fmt", "yuv420p",
		"-profile:v", "high",
		"-level:v", cfg.H264Level,
		"-b:v", fmt.Sprintf("%dM", cfg.BitrateMbps),
		"-maxrate", fmt.Sprintf("%dk", maxRateKbps),
		"-bufsize", fmt.Sprintf("%dM", bufSizeMb),
	)

	if cfg.Audio && cfg.AudioDevice != "" {
		audioBitrate := 128
		if cfg.BitrateMbps >= 16 {
			audioBitrate = 160
		}
		args = append(args,
			"-c:a", "aac",
			"-b:a", fmt.Sprintf("%dk", audioBitrate),
			"-ar", "48000",
			"-ac", "2",
		)
	} else {
		args = append(args, "-an")
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", cfg.SegmentSeconds),
		"-hls_list_size", fmt.Sprintf("%d", cfg.PlaylistSize),
		"-hls_flags", "delete_segments+independent_segments",
		"-hls_segment_filename", filepath.Join(dir, "seg%06d.ts"),
		filepath.Join(dir, PlaylistName),
	)

	return args
}

func waitForPlaylist(dir string) error {
	playlist := filepath.Join(dir, PlaylistName)
	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-time.After(100 * time.Millisecond):
		}

		data, err := os.ReadFile(playlist)
		if err != nil || len(data) == 0 {
			continue
		}

		hasSegment := false
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			hasSegment = true
			break
		}
		if hasSegment {
			return nil
		}
	}

	return errors.New("timed out waiting for the first HLS segment")
}
