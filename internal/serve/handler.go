package serve

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/oursky/pageship/internal/config"
)

func NewHandler(conf *config.ServerConfig) http.Handler {
	publicDir := conf.Site.Public
	if !filepath.IsAbs(publicDir) {
		publicDir = filepath.Join(conf.Root, publicDir)
	}
	publicFS := os.DirFS(publicDir)

	return http.FileServer(http.FS(publicFS))
}
