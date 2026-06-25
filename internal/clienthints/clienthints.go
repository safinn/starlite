// Package clienthints reads browser client hints from per-request cookies:
// values the server can't infer from headers, set with JS by a small script in
// the base layout before navigating. Each hint is its own ch_-prefixed cookie,
// read independently (one accessor each) rather than sharing one parsed blob.
// For now there's one: the IANA timezone, for rendering timestamps in the
// visitor's local zone.
package clienthints

import (
	"net/http"
	"time"
)

// timeZoneCookie carries the browser's IANA timezone name ("America/New_York"),
// set client-side from Intl.DateTimeFormat().resolvedOptions().timeZone.
const timeZoneCookie = "ch_tz"

// TimeZone resolves the request's timezone hint to a *time.Location, falling
// back to UTC when the cookie is absent (first visit before the setter runs, or
// a client blocking cookies) or names an unknown zone. LoadLocation reads the
// IANA database embedded via time/tzdata (blank import in main), so it resolves
// the same on any host and never touches the filesystem.
func TimeZone(r *http.Request) *time.Location {
	c, err := r.Cookie(timeZoneCookie)
	if err != nil {
		return time.UTC
	}
	loc, err := time.LoadLocation(c.Value)
	if err != nil {
		return time.UTC
	}
	return loc
}
