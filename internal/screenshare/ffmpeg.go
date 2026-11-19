package screenshare

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/mysterion/avrp/internal/utils"
)

var dshowAudioRE = regexp.MustCompile(`"([^"]+)"\s+\(audio\)`)

// FindFFmpeg locates ffmpeg without depending on the thumbnail package.
func FindFFmpeg() (string, error) {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}

	p := filepath.Join(utils.FfmpegDir, "bin", name)
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p, nil
	}

	return "", errors.New("ffmpeg not found; run `avrp --get-ffmpeg` or add ffmpeg to PATH")
}

// ListDshowAudioDevices returns DirectShow audio capture device names.
func ListDshowAudioDevices(ffmpeg string) []string {
	cmd := exec.Command(ffmpeg, "-hide_banner", "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	out, _ := cmd.CombinedOutput()

	matches := dshowAudioRE.FindAllStringSubmatch(string(out), -1)
	devices := make([]string, 0, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name != "" {
			devices = append(devices, name)
		}
	}
	return devices
}

var loopbackHints = []string{
	"virtual-audio-capturer",
	"virtual cable",
	"cable output",
	"stereo mix",
	"loopback",
	"what u hear",
}

// PickAudioDevice resolves the configured audio device. When none is available
// it returns false so the caller can fall back to video-only capture.
func PickAudioDevice(ffmpeg, requested string) (string, bool) {
	devices := ListDshowAudioDevices(ffmpeg)

	if requested != "" {
		for _, d := range devices {
			if strings.EqualFold(d, requested) {
				return d, true
			}
		}
		log.Printf("screenshare: audio device %q not found, falling back to video-only", requested)
		return "", false
	}

	for _, hint := range loopbackHints {
		for _, d := range devices {
			if strings.Contains(strings.ToLower(d), hint) {
				log.Printf("screenshare: using audio device %q", d)
				return d, true
			}
		}
	}

	log.Println("screenshare: no system-audio loopback device found, falling back to video-only")
	return "", false
}
