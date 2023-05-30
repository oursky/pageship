package app

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/caddyserver/certmagic"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	handler "github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
	"github.com/oursky/pageship/internal/site/local"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("addr", ":8000", "listen address")
	serveCmd.PersistentFlags().Bool("tls", false, "use TLS")
	serveCmd.PersistentFlags().String("tls-domain", "", "TLS certificate domain")
	serveCmd.PersistentFlags().String("tls-addr", ":443", "TLS listen address")
	serveCmd.PersistentFlags().String("tls-acme-endpoint", "", "TLS ACME directory endpoint")
	serveCmd.PersistentFlags().String("tls-acme-email", "", "TLS ACME directory account email")
}

func loadSitesConfig(fsys fs.FS) (*config.SitesConfig, error) {
	loader := config.NewLoader(config.SitesConfigName)

	conf := config.DefaultSitesConfig()
	if err := loader.Load(fsys, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func makeHandler(prefix string) (*handler.Handler, error) {
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

	handler, err := handler.NewHandler(zapLogger, resolver, handler.HandlerConfig{
		HostPattern: sitesConf.HostPattern,
		Middlewares: middleware.Default,
	})
	if err != nil {
		return nil, err
	}

	return handler, nil
}

var serveCmd = &cobra.Command{
	Use:   "serve [site directory]",
	Short: "Start local server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString("addr")
		useTLS := viper.GetBool("tls")
		tlsDomain := viper.GetString("tls-domain")
		tlsAddr := viper.GetString("tls-addr")
		tlsAcmeEndpoint := viper.GetString("tls-acme-endpoint")
		tlsAcmeEmail := viper.GetString("tls-acme-email")

		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		handler, err := makeHandler(dir)
		if err != nil {
			Error("Failed to setup server: %s", err)
			return
		}

		var tls *httputil.ServerTLSConfig
		if useTLS {
			tls = &httputil.ServerTLSConfig{
				Storage:       certmagic.Default.Storage,
				ACMEDirectory: tlsAcmeEndpoint,
				ACMEEmail:     tlsAcmeEmail,
				Addr:          tlsAddr,
				CheckDomain:   handler.CheckValidDomain,
			}

			if len(tlsDomain) > 0 {
				tls.DomainNames = []string{tlsDomain}
			} else if handler.AllowAnyDomain() {
				Error("Must provide domain name via --tls-domain to enable TLS")
				return
			}
		}

		server := &httputil.Server{
			Logger:  zapLogger,
			Addr:    addr,
			Handler: handler,
			TLS:     tls,
		}
		command.Run([]command.WorkFunc{server.Run})
	},
}
