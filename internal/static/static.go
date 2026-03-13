package static

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"starlite/internal/config"

	"github.com/benbjohnson/hashfs"
)

var (
	FilesDirectoryPath = "internal/static/files"
	//go:embed files
	FilesDirectory embed.FS
	filesSubFS, _  = fs.Sub(FilesDirectory, "files")
	FilesSys       = hashfs.NewFS(filesSubFS)
)

type Config interface {
	IsProduction() bool
}

func Handler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isProduction := config.Global.Env == config.Prod
		isFont := strings.HasPrefix(r.URL.Path, "/static/fonts/")
		serveDirect := !isProduction || isFont

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

func StaticPath(path string) string {
	if config.Global.Env == config.Prod {
		return "/static/" + FilesSys.HashName(path)
	}
	return "/static/" + path
}
