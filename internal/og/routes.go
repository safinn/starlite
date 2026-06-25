package og

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

// server renders share cards on demand and caches the PNG bytes per canonical
// key. The key set is bounded (fixed pages, one per country and region, plus one
// fallback), so the cache stays small; a restart (the only time copy changes)
// clears it.
type server struct {
	log   *slog.Logger
	mu    sync.RWMutex
	cache map[string][]byte
}

// SetupRoutes mounts the on-the-fly Open Graph image endpoint. base.templ points
// each page's og:image at /og/<slug>.png (slug mirrors the page's path); this
// handler renders the matching card.
func SetupRoutes(mux *http.ServeMux, log *slog.Logger) error {
	s := &server{log: log, cache: make(map[string][]byte)}
	mux.HandleFunc("GET /og/{slug}", s.handle)
	return nil
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(strings.TrimSuffix(r.PathValue("slug"), ".png"))
	key := canonical(slug)

	png, err := s.image(key)
	if err != nil {
		s.log.ErrorContext(r.Context(), "rendering og image", "key", key, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// The card only changes on deploy and scrapers cache hard anyway, so a long
	// max-age is safe; the ETag lets a warm scraper revalidate cheaply with a 304.
	etag := `"og-` + key + `"`
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=604800") // one week
	w.Header().Set("ETag", etag)
	if _, err := w.Write(png); err != nil {
		s.log.ErrorContext(r.Context(), "writing og image", "slug", slug, "error", err)
	}
}

// image returns the cached card for a canonical key, rendering and caching it
// on first request.
func (s *server) image(key string) ([]byte, error) {
	s.mu.RLock()
	cached, ok := s.cache[key]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	png, err := render(resolve(key), theme)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache[key] = png
	s.mu.Unlock()
	return png, nil
}
