package testutil

import (
	"github.com/oursky/pageship/internal/db"
	// install drivers
	_ "github.com/oursky/pageship/internal/db/postgres"
	_ "github.com/oursky/pageship/internal/db/sqlite"
	// install drivers for migration
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/spf13/viper"
)

func SetupDB() (database db.DB, resetDB func()) {
	db_url := viper.GetString("database-url")
	database, err := db.New(db_url)
	if err != nil {
		panic(err)
	}
	err = doMigrate(db_url, false)
	resetDB = func() {
		doMigrate(db_url, true)
	}
	if err != nil {
		panic(err)
	}
	return
}
