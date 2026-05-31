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
		serveDirect := isDev || isFont

		if serveDirect {
			if isFont {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-store")
			}
			logger.Debug("static assets are being served directly", "path", FilesDirectoryPath)
			http.StripPrefix("/static/", http.FileServerFS(os.DirFS(FilesDirectoryPath))).ServeHTTP(w, r)
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
