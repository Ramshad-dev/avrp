# avrp
single executable program for aframe-vr-player with update feature

# usage
```
Usage of avrp:
  -dev
        starts in dev mode, serves 'index.html' from current directory
  -dir string
        path to video files
  -get-ffmpeg
        downloads ffmpeg
  -no-thumb
        disable thumbnail generation for current session
  -port int
        port to serve on (default 5000) (default 5000)
  -reset
        removes all configs, thumbnails & 'aframe-vr-player' files
  -sha string
        Optional - download a specific commit of aframe-vr-player (default "latest")
  -toggle-thumb
        toggles disable/enable thumbnail generation
  -screenshare
        captures the Windows desktop and exposes it as a video source
  -screen-preset string
        screenshare preset: low, medium, high or ultra (default "medium")
  -screen-width int
        screenshare capture width (overrides preset)
  -screen-height int
        screenshare capture height (overrides preset)
  -screen-fps int
        screenshare capture fps (overrides preset)
  -screen-bitrate int
        screenshare video bitrate in Mbps (overrides preset)
  -screen-source string
        screenshare source: desktop or window (default "desktop")
  -screen-window string
        window title to capture when --screen-source=window
  -screen-audio-device string
        DirectShow audio loopback device for screenshare
  -update
        checks & downloads the latest version(commit) of 'aframe-vr-player'
```

### Examples
```sh
# -flag
# --flag
# -flag=x

# deletes all configs of avrp and aframe-vr-player / factory reset
avrp --reset

# serve video directory 
avrp --dir "C:\\Users\\User\\Video\\"
# linux path
avrp --dir "/home/user/Videos/"

# downloads the latest version(commit) of aframe-vr-player
avrp --update

# downloads https://github.com/Ramshad-dev/aframe-vr-player/tree/6757b3a6266be91c57e6236505249508b3e7d9af
# https://github.com/Ramshad-dev/aframe-vr-player/commits/main/
avrp --update --sha 6757b3a6266be91c57e6236505249508b3e7d9af

# point the updater at a different fork without rebuilding
AVRP_REPO_OWNER=mysterion AVRP_REPO_NAME=aframe-vr-player avrp --update

# share the desktop to the headset (720p30 by default)
avrp --dir "C:\Users\User\Video\" --screenshare

# 1080p30 with a specific system-audio loopback device
avrp --dir "C:\Users\User\Video\" --screenshare --screen-preset high --screen-audio-device "virtual-audio-capturer"
```

### Screensharing

`--screenshare` captures the Windows desktop with ffmpeg, encodes H.264/AAC, and serves an HLS playlist as a synthetic `screenshare.m3u8` entry in the player's file browser. Capture settings are fixed at startup. If no DirectShow loopback device is found, capture falls back to video-only.

Presets: `low` (960x540@15), `medium` (1280x720@30), `high` (1920x1080@30), `ultra` (1920x1080@60). Individual `--screen-*` flags override the preset. System audio requires a virtual audio device such as `virtual-audio-capturer` or VB-Audio Cable; use `--screen-audio-device` to pick one explicitly.

The player must include HLS support (`libs/hls/hls.min.js` and the `hls.js` hook in `main/components/VideoState.js`). The installed `dist` is already patched, but running `--update` will replace it with the published player; publish those player changes to the fork to keep screensharing after updates.
# installation
## go cli
```sh
# install ffmpeg and add it to path
go install github.com/mysterion/avrp@latest
```
## standalone binary
[https://github.com/mysterion/avrp/releases](https://github.com/mysterion/avrp/releases)

