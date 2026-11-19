package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/mysterion/avrp/internal/dist"
	"github.com/mysterion/avrp/internal/screenshare"
	"github.com/mysterion/avrp/internal/server"
	"github.com/mysterion/avrp/internal/thumbnails"
	"github.com/mysterion/avrp/internal/utils"
)

func isPathValid(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}

func getUserInput(prompt string) string {
	fmt.Println(prompt + ": ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func askForPath() string {
	for {
		path := getUserInput("Please enter path to serve")
		if isPathValid(path) {
			return path
		} else {
			fmt.Println("Invalid Path")
		}
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var servDir string
	var sha string
	var update bool
	var reset bool
	var port int
	var getFfmpeg bool
	var noThumb bool
	var toggleThumb bool
	var screenShare bool
	var screenPreset string
	var screenWidth int
	var screenHeight int
	var screenFps int
	var screenBitrate int
	var screenSource string
	var screenWindow string
	var screenAudioDevice string

	flag.BoolVar(&utils.DEV, "dev", false, "starts in dev mode, serves 'index.html' from current directory")
	flag.BoolVar(&update, "update", false, "checks & downloads the latest version(commit) of 'aframe-vr-player'")
	flag.StringVar(&sha, "sha", "latest", "Optional - download a specific commit of aframe-vr-player")
	flag.StringVar(&servDir, "dir", "", "path to video files")
	flag.BoolVar(&reset, "reset", false, "removes all configs, thumbnails & 'aframe-vr-player' files")
	flag.IntVar(&port, "port", 5000, "port to serve on (default 5000)")
	flag.BoolVar(&getFfmpeg, "get-ffmpeg", false, "downloads ffmpeg")
	flag.BoolVar(&noThumb, "no-thumb", false, "disable thumbnail generation for current session")
	flag.BoolVar(&toggleThumb, "toggle-thumb", false, "toggles disable/enable thumbnail generation")
	flag.BoolVar(&screenShare, "screenshare", false, "captures the Windows desktop and exposes it as a video source")
	flag.StringVar(&screenPreset, "screen-preset", "medium", "screenshare preset: low, medium, high or ultra")
	flag.IntVar(&screenWidth, "screen-width", 0, "screenshare capture width (overrides preset)")
	flag.IntVar(&screenHeight, "screen-height", 0, "screenshare capture height (overrides preset)")
	flag.IntVar(&screenFps, "screen-fps", 0, "screenshare capture fps (overrides preset)")
	flag.IntVar(&screenBitrate, "screen-bitrate", 0, "screenshare video bitrate in Mbps (overrides preset)")
	flag.StringVar(&screenSource, "screen-source", "desktop", "screenshare source: desktop or window")
	flag.StringVar(&screenWindow, "screen-window", "", "window title to capture when --screen-source=window")
	flag.StringVar(&screenAudioDevice, "screen-audio-device", "", "DirectShow audio loopback device for screenshare")

	flag.Parse()

	utils.Init()

	// commands 👇
	if getFfmpeg {
		utils.Panic(thumbnails.DownloadFfmpeg())
		return
	}

	if toggleThumb {
		if thumbnails.ThumbEnabled() {
			thumbnails.ThumbDisable()
			log.Println("thumbnail generation disabled👎")
		} else {
			thumbnails.ThumbEnable()
			log.Println("thumbnail generation enabled👍")
		}
		return
	}

	if !noThumb {
		thumbnails.Init()
	}

	if reset {
		err := os.RemoveAll(utils.ConfigDir)
		if err != nil {
			log.Panicln(err)
		} else {
			log.Printf("Success! removed dir : %s\n", utils.ConfigDir)
		}
		return
	}

	if update {
		dist.Update(sha)
		return
	}
	// commands 👆

	// running for the first time
	if !dist.Valid() && !utils.DEV {
		dist.Update("latest")
	}

	if len(servDir) == 0 {
		servDir = askForPath()
	} else {
		if !isPathValid(servDir) {
			servDir = askForPath()
		}
	}

	if screenShare {
		cfg, err := screenshare.Build(screenshare.Options{
			Enabled:     true,
			Preset:      screenPreset,
			Width:       screenWidth,
			Height:      screenHeight,
			FPS:         screenFps,
			BitrateMbps: screenBitrate,
			Source:      screenSource,
			Window:      screenWindow,
			AudioDevice: screenAudioDevice,
		})
		if err != nil {
			log.Fatalf("screenshare config: %v", err)
		}

		ffmpeg, err := screenshare.FindFFmpeg()
		if err != nil {
			log.Fatalf("screenshare: %v", err)
		}

		cfg.AudioDevice, cfg.Audio = screenshare.PickAudioDevice(ffmpeg, cfg.AudioDevice)

		capture, err := screenshare.Start(cfg, ffmpeg)
		if err != nil {
			log.Fatalf("screenshare: %v", err)
		}
		defer capture.Stop()

		server.SetScreenshare(cfg.Info(), capture.Dir())
	}

	server.Init(servDir)
	server.Start(port)
}
