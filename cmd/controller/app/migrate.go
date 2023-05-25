package app

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	migratefs "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/manifoldco/promptui"
	"github.com/oursky/pageship/migrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.PersistentFlags().String("database-url", "", "database URL")
	migrateCmd.MarkPersistentFlagRequired("database")

	migrateCmd.PersistentFlags().Bool("down", false, "downgrade database")

	database.Register("postgres", &pgx.Postgres{})
}

type migrateLogger struct{ *log.Logger }

func (migrateLogger) Verbose() bool { return false }

func doMigrate(database string, down bool) error {
	u, err := url.Parse(database)
	if err != nil {
		return fmt.Errorf("invalid database URL: %w", err)
	}

	source, err := migratefs.New(migrations.FS, u.Scheme)
	if err != nil {
		return fmt.Errorf("invalid database scheme: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(u.Scheme, source, database)
	if err != nil {
		return fmt.Errorf("cannot init migration: %w", err)
	}

	m.Log = migrateLogger{Logger: zap.NewStdLog(zap.L().Named("migration"))}

	if !down {
		err = m.Up()
	} else {
		err = m.Down()
	}
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database data",
	Run: func(cmd *cobra.Command, args []string) {
		database := viper.GetString("database-url")
		down := viper.GetBool("down")
		if database == "" {
			logger.Fatal("missing database URL")
			return
		}
		if down {
			prompt := promptui.Prompt{
				Label:     "Downgrade database? (DELETE all data)",
				IsConfirm: true,
			}
			_, err := prompt.Run()
			if err != nil {
				logger.Info("cancelled", zap.Error(err))
				return
			}
		}

		err := doMigrate(database, down)
		if err != nil {
			logger.Fatal("failed to migrate", zap.Error(err))
			return
		}

		logger.Info("done")
	},
}
