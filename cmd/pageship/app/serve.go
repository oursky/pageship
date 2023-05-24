package app

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	handler "github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/site"
	"github.com/oursky/pageship/internal/site/local"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("addr", ":8000", "listen address")
}

type siteLogger struct{}

func (siteLogger) Debug(format string, args ...any) {
	Debug(format, args...)
}

func (siteLogger) Error(msg string, err error) {
	Error(msg+": %s", err)
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

	var resolver site.Resolver
	if sitesConf != nil {
		resolver = local.NewMultiSiteResolver(fsys, sitesConf)
	} else {
		resolver = local.NewSingleSiteResolver(fsys)
		sitesConf = &config.SitesConfig{
			DefaultSite: config.DefaultSite,
			HostPattern: "",
		}

		// Check site on startup.
		_, err = resolver.Resolve(context.Background(), sitesConf.DefaultSite)
		if err != nil {
			return nil, err
		}
	}
	Info("site resolution mode: %s", resolver.Kind())

	handler, err := handler.NewHandler(siteLogger{}, resolver, handler.HandlerConfig{
		HostPattern: sitesConf.HostPattern,
		Middlewares: middleware.Default,
	})
	if err != nil {
		return nil, err
	}

	return handler, nil
}

var serveCmd = &cobra.Command{
	Use:   "serve directory",
	Short: "Start local server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("addr")

		handler, err := makeHandler(args[0])
		if err != nil {
			Error("Failed to setup server: %s", err)
			return
		}

		server := &command.HTTPServer{
			Addr:    addr,
			Handler: handler,
		}
		command.Run([]command.WorkFunc{server.Run})
	},
}
