package og

import (
	"strings"
	"unicode"
)

// Content is everything a card renders. For this template that's just the page
// title; resolve derives it from the request slug so any page gets a card with
// no per-page wiring.
type Content struct {
	Title string
}

// canonical folds a request slug onto the cache/render key. The handler has
// already lower-cased it and stripped ".png"; here we only map the empty slug
// onto "home" (layouts.ogImage points "/" at /og/home.png).
func canonical(slug string) string {
	slug = strings.Trim(strings.TrimSpace(slug), "/")
	if slug == "" {
		return "home"
	}
	return slug
}

// resolve turns a canonical slug into card content. With no page registry yet,
// it title-cases the slug ("getting-started" -> "Getting Started") so the card
// shows a readable page title. Swap in a slug->title lookup here when pages need
// bespoke copy.
func resolve(key string) Content {
	return Content{Title: titleFromSlug(key)}
}

// titleFromSlug renders a slug as a display title: split into words on -, _ and
// /, then capitalise each word. Falls back to "Home" for an empty slug.
func titleFromSlug(slug string) string {
	words := strings.FieldsFunc(slug, func(r rune) bool {
		return r == '-' || r == '_' || r == '/'
	})
	if len(words) == 0 {
		return "Home"
	}
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}
