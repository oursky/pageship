package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/caddyserver/certmagic"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/domain"
	domainlocal "github.com/oursky/pageship/internal/domain/local"
	handler "github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
	sitelocal "github.com/oursky/pageship/internal/site/local"
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

	serveCmd.PersistentFlags().String("default-site", config.DefaultSite, "default site")
	serveCmd.PersistentFlags().String("host-pattern", config.DefaultHostPattern, "host match pattern")
}

func loadSitesConfig(fsys fs.FS) (*config.SitesConfig, error) {
	loader := config.NewLoader(config.SitesConfigName)

	conf := config.DefaultSitesConfig()
	if err := loader.Load(fsys, conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func makeHandler(prefix string, defaultSite string, hostPattern string) (*handler.Handler, error) {
	dir, err := filepath.Abs(prefix)
	if err != nil {
		return nil, err
	}

	fsys := os.DirFS(dir)

	var siteResolver site.Resolver
	siteResolver = sitelocal.NewSingleSiteResolver(fsys)
	var domainResolver domain.Resolver
	domainResolver = &domain.ResolverNull{}

	// Check site on startup.
	_, err = siteResolver.Resolve(context.Background(), defaultSite)
	if errors.Is(err, config.ErrConfigNotFound) {
		// continue in multi-site mode

		sitesConf, err := loadSitesConfig(fsys)
		if errors.Is(err, config.ErrConfigNotFound) {
			// Treat as no sites
		} else if err != nil {
			return nil, err
		}

		var sites map[string]config.SitesConfigEntry
		if sitesConf != nil {
			sites = sitesConf.Sites
		}
		siteResolver = sitelocal.NewResolver(fsys, defaultSite, sites)
		domainResolver, err = domainlocal.NewResolver(defaultSite, sites)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	Info("site resolution mode: %s", siteResolver.Kind())

	handler, err := handler.NewHandler(context.Background(), zapLogger,
		domainResolver, siteResolver,
		handler.HandlerConfig{
			HostPattern:     hostPattern,
			MiddlewaresFunc: middleware.Default,
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
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := viper.GetString("addr")
		useTLS := viper.GetBool("tls")
		tlsDomain := viper.GetString("tls-domain")
		tlsAddr := viper.GetString("tls-addr")
		tlsAcmeEndpoint := viper.GetString("tls-acme-endpoint")
		tlsAcmeEmail := viper.GetString("tls-acme-email")

		defaultSite := viper.GetString("default-site")
		hostPattern := viper.GetString("host-pattern")

		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		handler, err := makeHandler(dir, defaultSite, hostPattern)
		if err != nil {
			return fmt.Errorf("failed to setup server: %w", err)
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
			} else if handler.AcceptsAllDomain() {
				return fmt.Errorf("must provide domain name via --tls-domain to enable TLS")
			}
		}

		server := &httputil.Server{
			Logger:  zapLogger,
			Addr:    addr,
			Handler: handler,
			TLS:     tls,
		}
		command.Run([]command.WorkFunc{server.Run})
		return nil
	},
}
