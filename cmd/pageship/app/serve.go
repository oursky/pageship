package app

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/server"
	"github.com/oursky/pageship/internal/handler/sites"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("addr", ":8000", "listen address")
	viper.BindPFlags(serveCmd.PersistentFlags())
}

type handler struct {
	defaultSite *sites.Descriptor
}

func NewHandler(defaultSite *sites.Descriptor) *handler {
	return &handler{
		defaultSite: defaultSite,
	}
}

func (h *handler) LoadHandler(desc *sites.Descriptor) (http.Handler, error) {
	loader := config.NewLoader(config.SiteConfigName)

	var conf struct {
		Site config.SiteConfig `json:"site"`
	}
	if err := loader.Load(desc.FS, &conf); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return server.NewHandler(&conf.Site, desc.FS)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	desc, ok := sites.SiteFromContext(r.Context())
	if !ok {
		desc = h.defaultSite
	}

	if desc == nil {
		http.NotFound(w, r)
		return
	}

	handler, err := h.LoadHandler(desc)
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		Error("Failed to load site '%s': %s", desc.Site, err)
		http.Error(w, "failed to load site", http.StatusInternalServerError)
		return
	}

	handler.ServeHTTP(w, r)
}

func loadSitesConfig(fsys fs.FS) (*config.SitesConfig, error) {
	loader := config.NewLoader(config.SitesConfigName)

	conf := config.DefaultSitesConfig()
	if err := loader.Load(fsys, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func makeHandler(prefix string) (http.Handler, error) {
	dir, err := filepath.Abs(prefix)
	if err != nil {
		return nil, err
	}

	fsys := os.DirFS(dir)
	sitesConf, err := loadSitesConfig(fsys)
	if errors.Is(err, config.ErrConfigNotFound) {
		// If multi-site config not found: continue in single-site mode.
		err = nil
	}
	if err != nil {
		return nil, err
	}

	var handler http.Handler
	if sitesConf != nil {
		Info("Running in multi-sites mode")
		handler, err = sites.NewHandler(fsys, sitesConf, NewHandler(nil))
		if err != nil {
			return nil, err
		}
	} else {
		Debug("Running in single-site mode")
		desc := &sites.Descriptor{
			Site: config.DefaultSite,
			FS:   fsys,
		}

		h := NewHandler(desc)

		// Check config on startup.
		_, err = h.LoadHandler(desc)
		if err != nil {
			return nil, err
		}

		handler = h
	}

	return handler, nil
}

var serveCmd = &cobra.Command{
	Use:   "serve directory",
	Short: "Start server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("addr")

		handler, err := makeHandler(args[0])
		if err != nil {
			Error("Failed to load config: %w", err)
			return
		}

		server := &command.HTTPServer{
			Addr:    addr,
			Handler: handler,
		}
		command.Run([]command.WorkFunc{server.Run})
	},
}
