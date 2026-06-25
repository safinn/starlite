package static

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/benbjohnson/hashfs"
)

var (
	FilesDirectoryPath = "internal/static/files"
	//go:embed files
	FilesDirectory embed.FS
	filesSubFS, _  = fs.Sub(FilesDirectory, "files")
	FilesSys       = hashfs.NewFS(filesSubFS)
)

func Handler(logger *slog.Logger, isDev bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isFont := strings.HasPrefix(r.URL.Path, "/static/fonts/")

		// Fonts are referenced from CSS by a stable, unfingerprinted name, so
		// they can't ride the hashed-name immutable cache the other assets use.
		// Cache them immutably anyway -- a font's bytes never change under a
		// given name.
		if isFont {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		if isDev {
			// Serve straight from disk so edits show up without a rebuild. Fonts
			// keep the immutable cache set above; everything else revalidates.
			if !isFont {
				// no-cache (revalidate), not no-store: file server answers a
				// conditional request with 304, so edits show up immediately yet
				// repeat navigations skip re-downloading/re-parsing every css/js.
				w.Header().Set("Cache-Control", "no-cache")
			}
			logger.Debug("static assets are being served directly", "path", FilesDirectoryPath)
			http.StripPrefix("/static/", http.FileServerFS(os.DirFS(FilesDirectoryPath))).ServeHTTP(w, r)
			return
		}

		// Fonts ship inside the embedded FS but are addressed by their plain
		// name, so serve them from the embed directly: hashfs expects a
		// fingerprinted path, and the on-disk tree isn't present in production,
		// so os.DirFS would 404 here.
		if isFont {
			http.StripPrefix("/static/", http.FileServerFS(filesSubFS)).ServeHTTP(w, r)
			return
		}

		logger.Debug("static assets are served with hash names")
		http.StripPrefix("/static/", hashfs.FileServer(FilesSys)).ServeHTTP(w, r)
	})
}
func StaticPath(path string, isDev bool) string {
	if !isDev {
		return "/static/" + FilesSys.HashName(path)
	}
	return "/static/" + path
}
