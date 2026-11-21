package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	expiry  time.Time
}

// browseTTL catches changes made outside this process.
const browseTTL = 10 * time.Second

//nolint:gochecknoglobals
var (
	browseGen   atomic.Uint64
	browseCache struct {
		sync.Mutex
		files   []BrowseJSON
		gen     uint64
		builtAt time.Time
	}
)

// InvalidateBrowse makes the next listing re-read storage.
func InvalidateBrowse() { browseGen.Add(1) }

// Browse lists every unexpired upload marked public.
// shortcut: a rebuild still reads the metadata of every upload, add an index if listings get slow.
func Browse(w http.ResponseWriter, r *http.Request) {
	lb, ok := config.StorageBackend.(backends.ListBackend)
	if !ok {
		ErrorMsg(w, r, http.StatusNotImplemented, "Listing is not supported by this storage backend")
		return
	}

	browseCache.Lock()
	defer browseCache.Unlock()
	if gen := browseGen.Load(); browseCache.files == nil || browseCache.gen != gen ||
		time.Since(browseCache.builtAt) > browseTTL {
		files, err := buildBrowse(r, lb)
		if err != nil {
			slog.Error("Failed to list uploads", "error", err)
			Error(w, r, http.StatusInternalServerError)
			return
		}
		// gen was read before the build, so an upload during the build forces another one.
		browseCache.files, browseCache.gen, browseCache.builtAt = files, gen, time.Now()
	}

	now := time.Now()
	files := slices.DeleteFunc(slices.Clone(browseCache.files), func(f BrowseJSON) bool {
		return !f.expiry.IsZero() && f.expiry.Before(now)
	})

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(files)
}

func buildBrowse(r *http.Request, lb backends.ListBackend) ([]BrowseJSON, error) {
	files := []BrowseJSON{}
	for name, err := range lb.List(r.Context()) {
		if err != nil {
			return nil, err
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
			expiry:        m.Expiry,
		})
	}
	slices.SortFunc(files, func(a, b BrowseJSON) int {
		if c := b.modTime.Compare(a.modTime); c != 0 {
			return c
		}
		return strings.Compare(a.Filename, b.Filename)
	})
	return files, nil
}
