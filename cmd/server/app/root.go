package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/server"
	"github.com/oursky/pageship/internal/handler/sites"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

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

	conf := config.DefaultServerConfig()
	if err := loader.Load(desc.FS, conf); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return server.NewHandler(conf, desc.FS)
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
		logger.Error("failed to load site", zap.Error(err))
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
	prefixDir, err := filepath.Abs(prefix)
	if err != nil {
		return nil, err
	}

	fsys := os.DirFS(prefixDir)
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
		logger.Info("running in multi-sites mode")
		handler, err = sites.NewHandler(logger, fsys, sitesConf, NewHandler(nil))
		if err != nil {
			return nil, err
		}
	} else {
		logger.Debug("running in single-site mode")
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

func start(ctx context.Context) error {
	handler, err := makeHandler(cmdConfig.Prefix)
	if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		return err
	}

	server := http.Server{
		Addr:    cmdConfig.Addr,
		Handler: handler,
	}

	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		close(shutdown)
	}()

	logger.Info("server starting", zap.String("addr", server.Addr))

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}
	<-shutdown

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "start server",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		command.Run(logger, []command.WorkFunc{start})
	},
}

func Execute() error {
	return rootCmd.Execute()
}
