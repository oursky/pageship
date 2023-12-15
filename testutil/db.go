package testutil

import (
	dbmigrate "github.com/golang-migrate/migrate/v4/database"
	"github.com/oursky/pageship/internal/db"
	// install drivers
	_ "github.com/oursky/pageship/internal/db/postgres"
	_ "github.com/oursky/pageship/internal/db/sqlite"

	// install drivers for migration
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/spf13/viper"
)

func init() {
	dbmigrate.Register("postgres", &pgx.Postgres{})
}

func SetupDB() (database db.DB, resetDB func()) {
	dbUrl := viper.GetString("database-url")
	database, err := db.New(dbUrl)
	if err != nil {
		panic(err)
	}
	err = doMigrate(dbUrl, false)
	resetDB = func() {
		doMigrate(dbUrl, true)
	}
	if err != nil {
		panic(err)
	}
	return
}

func WithTestDB(f func(db.DB)) {
	db, resetDB := SetupDB()
	defer resetDB()
	f(db)
}
