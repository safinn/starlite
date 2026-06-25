package og

import (
	"bytes"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
)

// Open Graph's recommended card size. Echoed in the og:image:width/height tags
// base.templ emits, so scrapers reserve the box before the image loads.
const (
	cardW  = 1200
	cardH  = 630
	margin = 80.0
	textW  = cardW - 2*margin // wrapping width for the title
)

// palette colours one card. Open Graph serves one static image per URL, so a
// card commits to one theme.
type palette struct {
	bg    string // page background (mist-950)
	title string // headline text (mist-100)
}

// theme mirrors the site's dark "mist" tokens (see static/files/css/global.css),
// so the card reads like the page background with the title on top.
var theme = palette{bg: "#090b0c", title: "#f1f3f3"}

// render draws one share card as PNG bytes: a flat background with the centred
// title, nothing else. Each call parses the embedded font fresh: shared font
// state isn't safe across concurrent renders, and the cost is moot since each
// slug renders once (cached upstream).
func render(c Content, p palette) ([]byte, error) {
	dc := gg.NewContext(cardW, cardH)
	dc.SetHexColor(p.bg)
	dc.Clear()

	// Title, centred on the card. Step down the size when it would wrap past two
	// or three lines, so a long title still fits.
	size := 96.0
	titleFace, err := face(gobold.TTF, size)
	if err != nil {
		return nil, err
	}
	dc.SetFontFace(titleFace)
	switch n := len(dc.WordWrap(c.Title, textW)); {
	case n > 3:
		size = 56
	case n > 2:
		size = 72
	}
	if size != 96 {
		if titleFace, err = face(gobold.TTF, size); err != nil {
			return nil, err
		}
		dc.SetFontFace(titleFace)
	}
	dc.SetHexColor(p.title)
	dc.DrawStringWrapped(c.Title, cardW/2, cardH/2, 0.5, 0.5, textW, 1.15, gg.AlignCenter)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// face builds a font face from embedded TrueType bytes. DPI 72 makes point size
// equal pixels, so the sizes above read directly as pixel heights.
func face(ttf []byte, size float64) (font.Face, error) {
	f, err := opentype.Parse(ttf)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}
