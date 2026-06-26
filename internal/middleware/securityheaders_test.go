package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSecurityHeadersCacheControl pins the caching contract: responses that set
// no Cache-Control of their own (HTML pages) default to no-cache so deploys are
// always picked up, while handlers that opt into caching (static assets, OG
// images, feeds) keep their own value.
func TestSecurityHeadersCacheControl(t *testing.T) {
	mw := SecurityHeadersMiddleware(false)

	t.Run("defaults to no-cache when the handler sets none", func(t *testing.T) {
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
			t.Errorf("Cache-Control = %q, want no-cache", got)
		}
	})

	t.Run("a handler's own Cache-Control wins", func(t *testing.T) {
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=604800")
			w.WriteHeader(http.StatusOK)
		}))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/og/home.png", nil))
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=604800" {
			t.Errorf("Cache-Control = %q, want public, max-age=604800", got)
		}
	})
}
