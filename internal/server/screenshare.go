package server

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mysterion/avrp/internal/screenshare"
)

var (
	screenInfo screenshare.Info
	screenDir  string
)

func init() {
	_ = mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	_ = mime.AddExtensionType(".ts", "video/mp2t")
}

// SetScreenshare registers the active capture with the HTTP server.
func SetScreenshare(info screenshare.Info, dir string) {
	screenInfo = info
	screenDir = dir
}

type noStoreHandler struct {
	next http.Handler
}

func (h noStoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	h.next.ServeHTTP(w, r)
}

func screenshareConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(screenInfo)
}

// mountScreenshare exposes the synthetic playlist and its segments inside the
// existing /file/ namespace so the player's file browser can open it directly.
func mountScreenshare(mux *http.ServeMux) {
	mux.HandleFunc("/screen/config", screenshareConfigHandler)

	if !screenInfo.Enabled || screenDir == "" {
		return
	}

	playlist := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(filepath.Join(screenDir, screenshare.PlaylistName))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// The local ffmpeg build has no -hls_baseurl option, so segment URLs are
		// rewritten here to point at the /file/screenshare/ route.
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if !strings.HasPrefix(trimmed, "/") {
				lines[i] = "/file/screenshare/" + trimmed
			}
		}

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte(strings.Join(lines, "\n")))
	})
	mux.Handle("/file/screenshare.m3u8", noStoreHandler{next: playlist})

	segments := http.StripPrefix("/file/screenshare/", http.FileServer(http.Dir(screenDir)))
	mux.Handle("/file/screenshare/", noStoreHandler{next: segments})
}
