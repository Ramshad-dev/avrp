package screenshare

import (
	"errors"
	"fmt"
	"strings"
)

// Options mirrors the command-line flags for screenshare.
type Options struct {
	Enabled     bool
	Preset      string
	Width       int
	Height      int
	FPS         int
	BitrateMbps int
	Source      string
	Window      string
	AudioDevice string
}

// Config is the validated, runtime configuration used to build the ffmpeg command.
type Config struct {
	Enabled     bool
	Preset      string
	Source      string
	Window      string
	Width       int
	Height      int
	FPS         int
	BitrateMbps int
	AudioDevice string
	Audio       bool

	SegmentSeconds  int
	PlaylistSize    int
	KeyframeSeconds int
	H264Level       string
}

// Info is the JSON representation exposed to the player.
type Info struct {
	Enabled        bool   `json:"enabled"`
	Preset         string `json:"preset"`
	Source         string `json:"source"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	FPS            int    `json:"fps"`
	BitrateKbps    int    `json:"bitrateKbps"`
	Audio          bool   `json:"audioEnabled"`
	AudioDevice    string `json:"audioDevice,omitempty"`
	Transport      string `json:"transport"`
	Path           string `json:"path"`
	SegmentSeconds int    `json:"segmentSeconds"`
}

type preset struct {
	width  int
	height int
	fps    int
	bps    int
}

var presets = map[string]preset{
	"low":    {width: 960, height: 540, fps: 15, bps: 2},
	"medium": {width: 1280, height: 720, fps: 30, bps: 5},
	"high":   {width: 1920, height: 1080, fps: 30, bps: 8},
	"ultra":  {width: 1920, height: 1080, fps: 60, bps: 16},
}

// PresetNames returns the supported preset names in a stable order.
func PresetNames() []string {
	return []string{"low", "medium", "high", "ultra"}
}

// Build validates the options and fills in preset defaults.
func Build(o Options) (Config, error) {
	var c Config
	if !o.Enabled {
		return c, nil
	}

	name := strings.ToLower(strings.TrimSpace(o.Preset))
	if name == "" {
		name = "medium"
	}
	p, ok := presets[name]
	if !ok {
		return c, fmt.Errorf("unknown screen preset %q (valid: %s)", o.Preset, strings.Join(PresetNames(), ", "))
	}

	c.Enabled = true
	c.Preset = name
	c.Width = o.Width
	if c.Width == 0 {
		c.Width = p.width
	}
	c.Height = o.Height
	if c.Height == 0 {
		c.Height = p.height
	}
	c.FPS = o.FPS
	if c.FPS == 0 {
		c.FPS = p.fps
	}
	c.BitrateMbps = o.BitrateMbps
	if c.BitrateMbps == 0 {
		c.BitrateMbps = p.bps
	}

	c.Source = strings.ToLower(strings.TrimSpace(o.Source))
	if c.Source == "" {
		c.Source = "desktop"
	}
	if c.Source != "desktop" && c.Source != "window" {
		return c, fmt.Errorf("invalid screen source %q (valid: desktop, window)", o.Source)
	}
	c.Window = strings.TrimSpace(o.Window)
	if c.Source == "window" && c.Window == "" {
		return c, errors.New("--screen-window is required when --screen-source=window")
	}
	c.AudioDevice = strings.TrimSpace(o.AudioDevice)

	if c.Width < 320 || c.Width > 7680 || c.Height < 240 || c.Height > 4320 {
		return c, fmt.Errorf("invalid screen size %dx%d", c.Width, c.Height)
	}
	if c.Width%2 != 0 || c.Height%2 != 0 {
		return c, errors.New("screen width and height must be even numbers")
	}
	if c.FPS < 1 || c.FPS > 60 {
		return c, fmt.Errorf("invalid screen fps %d (valid: 1-60)", c.FPS)
	}
	if c.BitrateMbps < 1 || c.BitrateMbps > 100 {
		return c, fmt.Errorf("invalid screen bitrate %d Mbps (valid: 1-100)", c.BitrateMbps)
	}

	c.SegmentSeconds = 2
	c.PlaylistSize = 3
	c.KeyframeSeconds = 2
	if c.FPS > 30 || c.Width*c.Height > 1920*1080 {
		c.H264Level = "4.2"
	} else {
		c.H264Level = "4.1"
	}

	return c, nil
}

// Info returns the JSON-safe view of the active configuration.
func (c Config) Info() Info {
	return Info{
		Enabled:        c.Enabled,
		Preset:         c.Preset,
		Source:         c.Source,
		Width:          c.Width,
		Height:         c.Height,
		FPS:            c.FPS,
		BitrateKbps:    c.BitrateMbps * 1000,
		Audio:          c.Audio,
		AudioDevice:    c.AudioDevice,
		Transport:      "hls",
		Path:           "/file/screenshare.m3u8",
		SegmentSeconds: c.SegmentSeconds,
	}
}
