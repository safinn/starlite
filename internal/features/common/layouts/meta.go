package layouts

import (
	"net/url"
	"strings"
)

// Meta is the per-page head metadata Base renders. Title and Description also
// feed the Open Graph tags, so each page writes its copy once and search results
// and link previews can't disagree.
type Meta struct {
	Title       string
	Description string
	// Canonical is the page's one true absolute URL; when set, Base renders
	// <link rel="canonical"> and og:url from it. Empty for pages with no public
	// address (order pages).
	Canonical string
	// NoIndex keeps the page out of search indexes: private pages reachable by
	// unguessable URL (orders) and thin pages not worth ranking (countries with
	// no plans yet).
	NoIndex bool
	// Image overrides the link-preview image. Empty lets Base derive it from
	// Canonical; set it only for a page wanting a bespoke card.
	Image string
}

// ogImage returns the absolute URL of the page's link-preview image. Prefers an
// explicit Image, else derives /og/<slug>.png from Canonical so the card matches
// the page with no per-page wiring ("/" maps to "home"). Returns "" when there's
// nothing to point at (private pages carry no canonical); Base then falls back to
// the plain summary card.
func ogImage(m Meta) string {
	if m.Image != "" {
		return m.Image
	}
	if m.Canonical == "" {
		return ""
	}
	u, err := url.Parse(m.Canonical)
	if err != nil || u.Host == "" {
		return ""
	}
	slug := strings.Trim(u.Path, "/")
	if slug == "" {
		slug = "home"
	}
	if strings.Contains(slug, "/") { // og route matches a single segment only
		return ""
	}
	return u.Scheme + "://" + u.Host + "/og/" + slug + ".png"
}

// brandedTitle can modify the page title to for example
// include the brand name.
func brandedTitle(title string) string {
	return title
}
