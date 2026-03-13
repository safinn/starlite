package middleware

import (
	"net/http"
	"strings"

	"starlite/internal/config"
)

const defaultPermissionsPolicy = "accelerometer=(), autoplay=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"

const baseCSP = "default-src 'self'; " +
	"base-uri 'self'; " +
	"object-src 'none'; " +
	"frame-ancestors 'none'; " +
	"script-src 'self' 'unsafe-eval' https://cdn.jsdelivr.net; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; " +
	"font-src 'self' data:; " +
	"connect-src 'self' https://cdn.jsdelivr.net;" +
	"manifest-src 'self'"

const devScriptSrcElem = "script-src-elem 'self' https://cdn.jsdelivr.net 'unsafe-inline'"

// SecurityHeadersMiddleware sets common security headers on each response.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		setIfMissing(h, "X-Content-Type-Options", "nosniff")
		setIfMissing(h, "X-Frame-Options", "DENY")
		setIfMissing(h, "Referrer-Policy", "strict-origin-when-cross-origin")
		setIfMissing(h, "Permissions-Policy", defaultPermissionsPolicy)
		setIfMissing(h, "X-DNS-Prefetch-Control", "off")
		setIfMissing(h, "X-Permitted-Cross-Domain-Policies", "none")
		setIfMissing(h, "Content-Security-Policy", cspValue())

		if isSecureRequest(r) {
			setIfMissing(h, "Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		next.ServeHTTP(w, r)
	})
}

func cspValue() string {
	if config.Global.Env != config.Prod {
		return baseCSP + "; " + devScriptSrcElem
	}

	return baseCSP
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
