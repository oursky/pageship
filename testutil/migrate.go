package testutil

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	migratefs "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/oursky/pageship/migrations"
)

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
