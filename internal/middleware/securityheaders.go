package middleware

import (
	"net/http"
	"strings"
)

const defaultPermissionsPolicy = "accelerometer=(), autoplay=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"

const baseCSP = "default-src 'self'; " +
	"base-uri 'self'; " +
	"object-src 'none'; " +
	"frame-ancestors 'none'; " +
	"script-src 'self' 'unsafe-eval'; " +
	"script-src-elem 'self' 'unsafe-inline'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; " +
	"font-src 'self' data:; " +
	"connect-src 'self';" +
	"manifest-src 'self'"

// SecurityHeadersMiddleware sets common security headers on each response.
func SecurityHeadersMiddleware(isDev bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()

			setIfMissing(h, "X-Content-Type-Options", "nosniff")
			setIfMissing(h, "X-Frame-Options", "DENY")
			setIfMissing(h, "Referrer-Policy", "strict-origin-when-cross-origin")
			setIfMissing(h, "Permissions-Policy", defaultPermissionsPolicy)
			setIfMissing(h, "X-DNS-Prefetch-Control", "off")
			setIfMissing(h, "X-Permitted-Cross-Domain-Policies", "none")
			setIfMissing(h, "Content-Security-Policy", baseCSP)

			// Default everything that doesn't opt into caching to revalidate.
			// This catches HTML pages (which set no Cache-Control of their own),
			// so a deploy's new fingerprinted asset URLs are always picked up and
			// no intermediary serves stale markup pointing at asset hashes that no
			// longer exist. Handlers that want real caching -- static assets, OG
			// images, feeds, the sitemap -- Set their own Cache-Control, which
			// overrides this default.
			setIfMissing(h, "Cache-Control", "no-cache")

			if isSecureRequest(r) {
				setIfMissing(h, "Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func setIfMissing(h http.Header, key, value string) {
	if h.Get(key) == "" {
		h.Set(key, value)
	}
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}

	if strings.EqualFold(r.Header.Get("X-Forwarded-Ssl"), "on") {
		return true
	}

	return false
}
