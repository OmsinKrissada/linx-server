package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/config"
)

type BrowseJSON struct {
	Filename      string `json:"filename"`
	OriginalName  string `json:"original_name,omitzero"`
	Size          string `json:"size"`
	Mimetype      string `json:"mimetype"`
	Expiry        string `json:"expiry"`
	Uploaded      string `json:"uploaded"`
	Protected     bool   `json:"protected,omitzero"`
	BurnAfterRead bool   `json:"burn_after_read,omitzero"`

	modTime time.Time
}

// Browse lists every unexpired upload marked public.
// shortcut: reads the metadata of every upload per request, add an index if listings get slow.
func Browse(w http.ResponseWriter, r *http.Request) {
	lb, ok := config.StorageBackend.(backends.ListBackend)
	if !ok {
		ErrorMsg(w, r, http.StatusNotImplemented, "Listing is not supported by this storage backend")
		return
	}

	files := []BrowseJSON{}
	for name, err := range lb.List(r.Context()) {
		if err != nil {
			slog.Error("Failed to list uploads", "error", err)
			Error(w, r, http.StatusInternalServerError)
			return
		}
		m, err := lb.Head(r.Context(), name)
		if err != nil || !m.Public || m.Expired() {
			continue
		}
		files = append(files, BrowseJSON{
			Filename:      name,
			OriginalName:  m.OriginalName,
			Size:          strconv.FormatInt(m.Size, 10),
			Mimetype:      m.Mimetype,
			Expiry:        strconv.FormatInt(max(m.Expiry.Unix(), 0), 10),
			Uploaded:      strconv.FormatInt(m.ModTime.Unix(), 10),
			Protected:     m.AccessKey != "",
			BurnAfterRead: m.BurnAfterRead,
			modTime:       m.ModTime,
		})
	}
	slices.SortFunc(files, func(a, b BrowseJSON) int {
		if c := b.modTime.Compare(a.modTime); c != 0 {
			return c
		}
		return strings.Compare(a.Filename, b.Filename)
	})

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(files)
}
