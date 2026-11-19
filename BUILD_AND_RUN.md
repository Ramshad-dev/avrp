# Building and Running avrp (Windows)

`avrp` is a Windows launcher/server for `aframe-vr-player`. It downloads the player from GitHub, serves your video folder over HTTPS, and lets a browser or headset play the videos. By default the updater points at the `Ramshad-dev` fork, which includes camera passthrough with chroma keying.

All commands below are Windows PowerShell.

## Quick start

```powershell
$env:Path += ';C:\Program Files\Go\bin'
cd C:\Code\avrp
go build -o avrp.exe .
.\avrp.exe --update
.\avrp.exe --dir "C:\Videos"
```

Open the printed `https://<ip>:5000` URL.

## Prerequisites

- Windows 10/11, 64-bit (or build for another architecture, see below).
- Go 1.23 or newer.
  - Installed at `C:\Program Files\Go\bin\go.exe` on this machine.
  - Not on `PATH` by default in a fresh shell. Either add it for the session:
    ```powershell
    $env:Path += ';C:\Program Files\Go\bin'
    ```
    or call it by full path:
    ```powershell
    & 'C:\Program Files\Go\bin\go.exe' version
    ```
- Optional: `ffmpeg.exe` and `ffprobe.exe` for video thumbnails.
- Internet access for the first `--update` and for downloading ffmpeg if you do not already have it.

No Go installation or extra DLLs are required on the machine that only runs `avrp.exe`.

## Build

From the repository root:

```powershell
$env:Path += ';C:\Program Files\Go\bin'
cd C:\Code\avrp
go build -o avrp.exe .
```

Smaller executable (strips symbol/debug data):

```powershell
go build -trimpath -ldflags "-s -w" -o avrp.exe .
```

The result is `C:\Code\avrp\avrp.exe`. `*.exe` is ignored by Git, so building does not dirty the repository.

### Build for another Windows architecture

The build is pure Go with `CGO_ENABLED=0`, so it runs without a Go toolchain or runtime DLLs on the target PC.

```powershell
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'    # arm64 for Windows on ARM, 386 for 32-bit Windows

go build -trimpath -ldflags "-s -w" -o avrp.exe .
```

Copy only `avrp.exe` to the target machine. The target does not need Go, ffmpeg, or the source code, but it does need network access the first time unless you also copy the player files (see [Offline and portable deployment](#offline-and-portable-deployment)).

## Run

### Normal mode (downloads/uses the published player)

```powershell
.\avrp.exe --update
.\avrp.exe --dir "C:\Videos"
```

- `--update` fetches the latest commit from `Ramshad-dev/aframe-vr-player` into `%USERPROFILE%\.avrp\dist`.
- It prints the commit details and waits for **Enter** before downloading.
- If `%USERPROFILE%\.avrp\dist` is missing, the first normal run performs the same update automatically.
- `--dir` must point to the folder you want to browse. If omitted, `avrp` prompts for a path.

Open the URL printed by the server, for example:

```text
https://192.168.1.20:5000
https://127.0.0.1:5000
```

The server uses a self-signed certificate. The first visit shows a certificate warning; choose **Advanced** and proceed, or trust the certificate on that device.

### Dev mode (uses a local player checkout, no download)

`--dev` serves the player files from the current working directory. This is the fastest way to test local player changes.

```powershell
cd C:\Code\aframe-vr-player
C:\Code\avrp\avrp.exe --dev --dir "C:\Code\aframe-vr-player\assets" --no-thumb
```

- Run it from `C:\Code\aframe-vr-player`, not from `C:\Code\avrp`.
- `--no-thumb` skips ffmpeg/thumbnail setup.
- `--dev` still binds the HTTPS server; open the printed URL.

Do not use `go run .` when you need to reach the server from a phone or headset: `go run` restricts the listener to `127.0.0.1`. Build `avrp.exe` first.

### Common flags

| Flag | Meaning |
| --- | --- |
| `--dir "C:\Videos"` | Folder to serve. Prompts if omitted or invalid. |
| `--port 5000` | HTTPS port. Default `5000`. |
| `--update` | Download the latest player commit, then exit. |
| `--update --sha <full-sha>` | Download a specific player commit. |
| `--dev` | Serve player files from the current directory. |
| `--no-thumb` | Disable thumbnails for this session; avoids the ffmpeg prompt. |
| `--toggle-thumb` | Persistently enable/disable thumbnail generation, then exit. |
| `--get-ffmpeg` | Download ffmpeg/ffprobe into `%USERPROFILE%\.avrp\ffmpeg`, then exit. |
| `--reset` | Delete `%USERPROFILE%\.avrp` (player, thumbnails, ffmpeg, settings). Destructive. |

Examples:

```powershell
# Different port
.\avrp.exe --dir "D:\VR" --port 5055

# Update to an exact commit
.\avrp.exe --update --sha 0672e30d1c0f9c56182fc873e9d9e6cf0d9e1178

# Skip ffmpeg entirely
.\avrp.exe --dir "D:\VR" --no-thumb
```

## Using camera passthrough

1. Load a video.
2. Open **Settings**.
3. Set **camera passthrough** to **ON**.
4. Choose **camera quality**:
   - `low`: 640x480 @ 15 fps
   - `medium`: 1280x720 @ 24 fps
   - `high`: 1920x1080 @ 30 fps
5. Choose **camera facing**: `environment` (rear), `user` (front), or `any`.
6. Adjust the background framing:
   - **camera zoom**: `0.5x` to `5x` (`1x` is the natural camera field of view)
   - **camera horizontal** / **camera vertical**: pan and tilt in percent
   - **camera mirror**: flip the feed horizontally
7. Choose **key color**: green, blue, or a custom `#RRGGBB` value.
8. Adjust **similarity** and **smoothness** if the key is too aggressive or leaves a fringe.

The feed is monoscopic and shown to both eyes. It is projected at its natural field of view instead of being stretched around the whole sphere, so `1x` looks like the camera rather than a heavily zoomed image; increase the zoom to fill more of the view. Green/blue screen areas become transparent, revealing the camera feed behind the video; non-keyed video areas remain visible. The **FLAT** preset also hides the virtual background while passthrough is on.

If the camera is unavailable, the player now tries progressively looser constraints (chosen facing, then any camera, then whatever the browser offers) and logs the failure in the browser console.

Camera access requires HTTPS (or `localhost`). Allow camera permission when the browser prompts. On a phone or headset, use the LAN `https://<ip>:<port>` URL and accept the certificate.

## Thumbnails and ffmpeg

With thumbnails enabled, `avrp` looks for `ffmpeg.exe` and `ffprobe.exe` in:

1. `PATH`
2. `%USERPROFILE%\.avrp\ffmpeg\bin`

If neither is found, it offers to download them. To avoid the prompt permanently:

```powershell
.\avrp.exe --get-ffmpeg
```

Or disable thumbnails:

```powershell
.\avrp.exe --dir "C:\Videos" --no-thumb
```

## Offline and portable deployment

`avrp.exe` is self-contained, but the player and ffmpeg are separate runtime assets.

Minimum portable setup:

1. Copy `avrp.exe` to the target PC.
2. Use internet once to populate the player:
   ```powershell
   .\avrp.exe --update
   ```
3. Optionally copy `%USERPROFILE%\.avrp\ffmpeg` for thumbnails.

Fully offline setup:

1. Copy `avrp.exe`.
2. Copy `C:\Users\<source-user>\.avrp\dist` to `C:\Users\<target-user>\.avrp\dist`.
3. Optionally copy `C:\Users\<source-user>\.avrp\ffmpeg` to `C:\Users\<target-user>\.avrp\ffmpeg`.
4. Run:
   ```powershell
   .\avrp.exe --dir "C:\Videos" --no-thumb
   ```

`dist` must contain `index.html` directly (for example `%USERPROFILE%\.avrp\dist\index.html`).

### Configuration locations

| Path | Contents |
| --- | --- |
| `%USERPROFILE%\.avrp\dist` | Downloaded player build. |
| `%USERPROFILE%\.avrp\ffmpeg` | Downloaded ffmpeg/ffprobe. |
| `%USERPROFILE%\.avrp\thumbnails` | Generated thumbnails. |
| `%USERPROFILE%\.avrp\VERSION` | Installed player commit SHA. |
| `%USERPROFILE%\.avrp\nothumb` | Present when thumbnails are persistently disabled. |

Browser-side settings (passthrough, key color, quality, cursor type, etc.) are stored in the browser's local storage for the served site, not in `.avrp`.

## Player source overrides

The updater defaults to:

- Owner: `Ramshad-dev`
- Repository: `aframe-vr-player`

Override them without rebuilding:

```powershell
$env:AVRP_REPO_OWNER = 'mysterion'
$env:AVRP_REPO_NAME = 'aframe-vr-player'
.\avrp.exe --update
```

The environment variables are read when `avrp.exe` starts.

## Verify the build

```powershell
$env:Path += ';C:\Program Files\Go\bin'
cd C:\Code\avrp
go build -o "$env:TEMP\avrp.exe" .
go vet ./...
go test ./...
```

Expected: build and vet succeed; `internal/cache` tests pass.

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `go` is not recognized | Add `C:\Program Files\Go\bin` to `PATH` or call `go.exe` by full path. |
| Camera passthrough does nothing | Use the HTTPS URL, allow camera permission, and make sure no other app owns the camera. |
| Phone cannot open the page | Allow `avrp.exe` through Windows Firewall on private networks; use the LAN IP, not `127.0.0.1`. |
| Certificate warning | Expected with the built-in self-signed certificate. Proceed, or trust the certificate on the device. |
| Port already in use | Use `--port 5055` (or another free port). |
| ffmpeg prompt every run | Run `--get-ffmpeg` once, install ffmpeg on `PATH`, or use `--no-thumb`. |
| No thumbnails / 503 from `/thumb/` | ffmpeg is unavailable or thumbnails are disabled. Check the console output. |
| `avrp` asks for a path | `--dir` was omitted or points to something that does not exist. |
| `--update` says "Already on the latest version" | The installed `VERSION` matches the fork's latest commit. Use `--sha` to force a specific commit. |
| LAN URL only works on the local PC | `go run .` binds to `127.0.0.1`; build and run `avrp.exe` instead. |
| SmartScreen blocks the exe | Unsigned binary: choose **More info > Run anyway**, or sign it for distribution. |
| `--reset` removed everything | It deletes `%USERPROFILE%\.avrp`; re-run `--update` and `--get-ffmpeg` to restore. |
