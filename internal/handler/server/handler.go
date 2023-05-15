package server

import (
	"io/fs"
	"net/http"

	"github.com/oursky/pageship/internal/config"
)

func NewHandler(conf *config.ServerConfig, fsys fs.FS) (http.Handler, error) {
	publicFS, err := fs.Sub(fsys, conf.Site.Public)
	if err != nil {
		return nil, err
	}

	return http.FileServer(http.FS(publicFS)), nil
}
