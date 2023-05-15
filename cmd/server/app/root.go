package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/handler/server"
	"github.com/oursky/pageship/internal/handler/sites"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ConfigFS struct {
	fs.FS
}

func (fs ConfigFS) Open(name string) (fs.File, error) {
	if path.IsAbs(name) {
		// Translate abs path to raw path name to workaround viper AddConfigPath absify the path.
		name = name[1:]
	}
	return fs.FS.Open(name)
}

type ServerHandler struct {
	fs fs.FS
}

func (*ServerHandler) LoadConfig(fsys fs.FS) (*config.ServerConfig, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("/"))

	v.SetConfigName("pageship")
	v.SetFs(afero.FromIOFS{FS: ConfigFS{FS: fsys}})
	v.AddConfigPath("/")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := config.DefaultServerConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fsys := h.fs
	desc, ok := sites.SiteFromContext(r.Context())
	if ok {
		fsys = desc.FS
	}

	serverConf, err := h.LoadConfig(fsys)
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	handler, err := server.NewHandler(serverConf, fsys)
	if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}
	handler.ServeHTTP(w, r)
}

func loadSitesConfig(fsys fs.FS) (*config.SitesConfig, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("/"))

	v.SetConfigName("sites")
	v.SetFs(afero.FromIOFS{FS: ConfigFS{FS: fsys}})
	v.AddConfigPath("/")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := config.DefaultSitesConfig()
	if err := v.Unmarshal(conf); err != nil {
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
	serverHandler := &ServerHandler{fs: fsys}
	sitesConf, err := loadSitesConfig(fsys)
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// If multi-site config not found: continue in single-site mode.
		err = nil
	}
	if err != nil {
		return nil, err
	}

	if sitesConf != nil {
		logger.Info("running in multi-sites mode")
		return sites.NewHandler(logger, fsys, sitesConf, serverHandler)
	}

	logger.Debug("running in single-site mode")
	// Check config on startup.
	_, err = serverHandler.LoadConfig(fsys)
	if err != nil {
		return nil, err
	}

	return serverHandler, nil
}

func start(ctx context.Context) error {
	handler, err := makeHandler(cmdConfig.Prefix)
	if err != nil {
		logger.Error("failed to load config", zap.Error(err))
		return err
	}

	server := http.Server{
		Addr:    ":8000",
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
