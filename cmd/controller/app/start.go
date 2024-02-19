package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/dustin/go-humanize"
	"github.com/oursky/pageship/internal/command"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/cron"
	"github.com/oursky/pageship/internal/db"
	_ "github.com/oursky/pageship/internal/db/postgres"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	domaindb "github.com/oursky/pageship/internal/domain/db"
	"github.com/oursky/pageship/internal/handler/controller"
	"github.com/oursky/pageship/internal/handler/site"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/httputil"
	sitedb "github.com/oursky/pageship/internal/site/db"
	"github.com/oursky/pageship/internal/storage"
	"github.com/oursky/pageship/internal/watch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var errUnknownDomain = errors.New("unknown domain")

const defaultControllerHostID = "api"

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().String("database-url", "", "database URL")
	startCmd.MarkPersistentFlagRequired("database-url")

	startCmd.PersistentFlags().String("storage-url", "", "object storage URL")
	startCmd.MarkPersistentFlagRequired("storage-url")

	startCmd.PersistentFlags().Bool("migrate", false, "migrate before starting")
	startCmd.PersistentFlags().String("addr", ":8001", "listen address")

	startCmd.PersistentFlags().Bool("tls", false, "use TLS")
	startCmd.PersistentFlags().String("tls-addr", ":443", "TLS listen address")
	startCmd.PersistentFlags().String("tls-acme-endpoint", "", "TLS ACME directory endpoint")
	startCmd.PersistentFlags().String("tls-acme-email", "", "TLS ACME directory account email")
	startCmd.PersistentFlags().String("tls-protect-key", "", "TLS data protection key")

	startCmd.PersistentFlags().String("max-deployment-size", "200M", "max deployment files size")
	startCmd.PersistentFlags().String("storage-key-prefix", "", "storage key prefix")
	startCmd.PersistentFlags().String("host-pattern", config.DefaultHostPattern, "host match pattern")
	startCmd.PersistentFlags().String("host-id-scheme", string(config.HostIDSchemeDefault), "host ID scheme")
	startCmd.PersistentFlags().StringSlice("reserved-apps", []string{defaultControllerHostID}, "reserved app IDs")
	startCmd.PersistentFlags().String("api-acl", "", "API ACL file")

	startCmd.PersistentFlags().String("token-authority", "pageship", "auth token authority")
	startCmd.PersistentFlags().String("token-signing-key", "", "auth token signing key")

	startCmd.PersistentFlags().String("custom-domain-message", "", "message for custom domain users")

	startCmd.PersistentFlags().String("cleanup-expired-crontab", "", "cleanup expired schedule")
	startCmd.PersistentFlags().Duration("keep-after-expired", time.Hour*24, "keep-after-expired")
	startCmd.PersistentFlags().String("verify-domain-ownership-crontab", "", "verify domain ownership schedule")
	startCmd.PersistentFlags().Bool("domain-verification-enabled", false, "enable/disable domain verification")
	startCmd.PersistentFlags().Duration("domain-verification-interval", time.Hour, "duration before next domain verification start for a verified domain")

	startCmd.PersistentFlags().Bool("controller", true, "run controller server")
	startCmd.PersistentFlags().Bool("cron", true, "run cron jobs")
	startCmd.PersistentFlags().Bool("sites", true, "run sites server")
	startCmd.PersistentFlags().String("controller-domain", "", "controller domain")

	startCmd.PersistentFlags().Int64("content-cache-max-size", int64(1)<<24, "maximum size of server-side content cache in bytes, default is 16MB")
}

type StartConfig struct {
	DatabaseURL string `mapstructure:"database-url" validate:"url"`
	StorageURL  string `mapstructure:"storage-url" validate:"url"`
	Addr        string `mapstructure:"addr" validate:"hostname_port"`

	TLS             bool   `mapstructure:"tls"`
	TLSAddr         string `mapstructure:"tls-addr" validate:"hostname_port"`
	TLSACMEEndpoint string `mapstructure:"tls-acme-endpoint"`
	TLSACMEEmail    string `mapstructure:"tls-acme-email"`
	TLSProtectKey   string `mapstructure:"tls-protect-key"`

	Controller       bool   `mapstructure:"controller"`
	Cron             bool   `mapstructure:"cron"`
	Sites            bool   `mapstructure:"sites"`
	ControllerDomain string `mapstructure:"controller-domain" validate:"omitempty,hostname_rfc1123"`

	StartSitesConfig      `mapstructure:",squash"`
	StartControllerConfig `mapstructure:",squash"`
	StartCronConfig       `mapstructure:",squash"`
}

type StartSitesConfig struct {
	HostPattern         string              `mapstructure:"host-pattern"`
	HostIDScheme        config.HostIDScheme `mapstructure:"host-id-scheme" validate:"hostidscheme"`
	ContentCacheMaxSize int64               `mapstructure:"content-cache-max-size"`
}

type StartControllerConfig struct {
	MaxDeploymentSize string   `mapstructure:"max-deployment-size" validate:"size"`
	StorageKeyPrefix  string   `mapstructure:"storage-key-prefix"`
	TokenSigningKey   string   `mapstructure:"token-signing-key"`
	TokenAuthority    string   `mapstructure:"token-authority"`
	ReservedApps      []string `mapstructure:"reserved-apps"`
	APIACLFile        string   `mapstructure:"api-acl" validate:"omitempty,filepath"`

	CustomDomainMessage       string `mapstructure:"custom-domain-message"`
	DomainVerificationEnabled bool   `mapstructure:"domain-verification-enabled" validate:"omitempty"`
}

type StartCronConfig struct {
	CleanupExpiredCrontab        string        `mapstructure:"cleanup-expired-crontab" validate:"omitempty,cron"`
	KeepAfterExpired             time.Duration `mapstructure:"keep-after-expired" validate:"min=0"`
	VerifyDomainOwnershipCrontab string        `mapstructure:"verify-domain-ownership-crontab" validate:"omitempty,cron"`
	DomainVerificationEnabled    bool          `mapstructure:"domain-verification-enabled" validate:"omitempty"`
	DomainVerificationInterval   time.Duration `mapstructure:"domain-verification-interval" validate:"min=1"`
}

type setup struct {
	ctx              context.Context
	database         db.DB
	storage          *storage.Storage
	server           *httputil.Server
	mux              *http.ServeMux
	works            []command.WorkFunc
	checkDomainFuncs []func(name string) error
}

func (s *setup) checkDomain(name string) error {
	var err error
	for _, f := range s.checkDomainFuncs {
		if err = f(name); err == nil {
			break
		}
	}
	return err
}

func (s *setup) sites(conf StartSitesConfig) error {
	domainResolver := &domaindb.Resolver{
		HostIDScheme: conf.HostIDScheme,
		DB:           s.database,
	}
	siteResolver := &sitedb.Resolver{
		HostIDScheme: conf.HostIDScheme,
		DB:           s.database,
		Storage:      s.storage,
	}
	handler, err := site.NewHandler(
		s.ctx,
		logger.Named("site"),
		domainResolver,
		siteResolver,
		site.HandlerConfig{
			HostPattern:         conf.HostPattern,
			MiddlewaresFunc:     middleware.Default,
			ContentCacheMaxSize: conf.ContentCacheMaxSize,
		},
	)
	if err != nil {
		return err
	}

	s.mux.Handle("/", handler)
	s.checkDomainFuncs = append(s.checkDomainFuncs, handler.CheckValidDomain)
	return nil
}

func (s *setup) controller(domain string, conf StartControllerConfig, sitesConf StartSitesConfig) error {
	maxDeploymentSize, _ := humanize.ParseBytes(conf.MaxDeploymentSize)
	tokenSigningKey := conf.TokenSigningKey
	if tokenSigningKey == "" {
		logger.Warn("token signing key not specified; a temporary generated key would be used.")
		tokenSigningKey = generateSecret()
	}

	reservedApps := make(map[string]struct{})
	for _, app := range conf.ReservedApps {
		reservedApps[app] = struct{}{}
	}

	controllerConf := controller.Config{
		MaxDeploymentSize:         int64(maxDeploymentSize),
		StorageKeyPrefix:          conf.StorageKeyPrefix,
		HostIDScheme:              sitesConf.HostIDScheme,
		HostPattern:               config.NewHostPattern(sitesConf.HostPattern),
		ReservedApps:              reservedApps,
		TokenSigningKey:           []byte(tokenSigningKey),
		TokenAuthority:            conf.TokenAuthority,
		ServerVersion:             versioninfo.Short(),
		CustomDomainMessage:       conf.CustomDomainMessage,
		DomainVerificationEnabled: conf.DomainVerificationEnabled,
	}

	if conf.APIACLFile != "" {
		aclLog := logger.Named("api-acl")
		acl, err := watch.NewFile(
			aclLog,
			conf.APIACLFile,
			func(path string) (config.ACL, error) {
				f, err := os.Open(path)
				if err != nil {
					return nil, err
				}
				defer f.Close()

				list, err := config.LoadACL(f)
				if err != nil {
					return nil, err
				}

				aclLog.Info("loaded ACL", zap.Int("count", len(list)))
				return list, nil
			},
		)
		if err != nil {
			return err
		}

		controllerConf.ACL = acl
		s.works = append(s.works, func(ctx context.Context) error {
			<-ctx.Done()
			acl.Close()
			return nil
		})
	}

	ctrl := &controller.Controller{
		Context: s.ctx,
		Logger:  logger.Named("controller"),
		Config:  controllerConf,
		Storage: s.storage,
		DB:      s.database,
	}

	s.mux.Handle(domain+"/", ctrl.Handler())
	if s.server.TLS != nil {
		s.server.TLS.DomainNames = append(s.server.TLS.DomainNames, domain)
	}
	s.checkDomainFuncs = append(s.checkDomainFuncs, func(name string) error {
		if name != domain {
			return errUnknownDomain
		}
		return nil
	})

	logger.Info("setup controller", zap.String("domain", domain))

	return nil
}

func (s *setup) cron(conf StartCronConfig) error {
	cronjobs := []command.CronJob{
		&cron.CleanupExpired{
			Schedule:         conf.CleanupExpiredCrontab,
			KeepAfterExpired: conf.KeepAfterExpired,
			DB:               s.database,
		},
	}
	if conf.DomainVerificationEnabled {
		cronjobs = append(cronjobs,
			&cron.VerifyDomainOwnership{
				Schedule:                     conf.VerifyDomainOwnershipCrontab,
				DB:                           s.database,
				MaxConsumeActiveDomainCount:  10,
				MaxConsumePendingDomainCount: 10,
				Resolver:                     net.DefaultResolver,
				VerificationInterval:         conf.DomainVerificationInterval,
			},
		)
	}
	cronr := command.CronRunner{
		Logger: logger.Named("cron"),
		Jobs:   cronjobs,
	}

	s.works = append(s.works, cronr.Run)
	return nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		migrate := viper.GetBool("migrate")
		if migrate {
			logger.Info("migrating database before starting...")
			err := doMigrate(viper.GetString("database-url"), false)
			if err != nil {
				logger.Fatal("failed to migrate", zap.Error(err))
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cmdArgs StartConfig
		if err := viper.Unmarshal(&cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}
		if err := validate.Struct(cmdArgs); err != nil {
			logger.Fatal("invalid config", zap.Error(err))
			return
		}

		database, err := db.New(cmdArgs.DatabaseURL)
		if err != nil {
			logger.Fatal("failed to setup database", zap.Error(err))
			return
		}

		storage, err := storage.New(cmd.Context(), cmdArgs.StorageURL)
		if err != nil {
			logger.Fatal("failed to setup object storage", zap.Error(err))
			return
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		setup := &setup{
			ctx:      ctx,
			database: database,
			storage:  storage,
			mux:      new(http.ServeMux),
			server: &httputil.Server{
				Logger: logger.Named("server"),
				Addr:   cmdArgs.Addr,
			},
		}
		setup.server.Handler = setup.mux
		setup.works = append(setup.works, setup.server.Run)

		if cmdArgs.TLS {
			if cmdArgs.TLSProtectKey == "" {
				logger.Warn("TLS protect key not specified; certificate private keys would be stored in plain text.")
			}

			setup.server.TLS = &httputil.ServerTLSConfig{
				Storage:       db.NewCertStorage(database, cmdArgs.TLSProtectKey),
				ACMEDirectory: cmdArgs.TLSACMEEndpoint,
				ACMEEmail:     cmdArgs.TLSACMEEmail,
				Addr:          cmdArgs.TLSAddr,
				CheckDomain:   setup.checkDomain,
			}
		}

		if cmdArgs.Controller {
			domain := cmdArgs.ControllerDomain
			if domain == "" {
				pattern := config.NewHostPattern(cmdArgs.HostPattern)
				domain = pattern.MakeDomain(cmdArgs.HostIDScheme.Make(defaultControllerHostID, ""))
			}

			if err := setup.controller(domain,
				cmdArgs.StartControllerConfig,
				cmdArgs.StartSitesConfig,
			); err != nil {
				logger.Fatal("failed to setup controller server", zap.Error(err))
				return
			}
		}

		if cmdArgs.Cron {
			if err := setup.cron(cmdArgs.StartCronConfig); err != nil {
				logger.Fatal("failed to setup cron job runner", zap.Error(err))
				return
			}
		}

		if cmdArgs.Sites {
			if err := setup.sites(cmdArgs.StartSitesConfig); err != nil {
				logger.Fatal("failed to setup sites server", zap.Error(err))
				return
			}
		}

		command.Run(setup.works)
	},
}
