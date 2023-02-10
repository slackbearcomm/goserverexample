package main

import (
	"fmt"
	"log"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	// these are constants because the db credentials should not change
	DB_HOST     string = "localhost"
	DB_PORT     int    = 5432
	DB_USERNAME string = "root"
	DB_PASSWORD string = "secret"
	DB_NAME     string = "nextcrm"
)

func main() {
	// make db connection
	dbSource := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_NAME)
	runMigrateDown(dbSource)
}

func runMigrateDown(dbSource string) {
	db, err := sql.Open("postgres", dbSource)
	if err != nil {
		log.Fatal("error when trying to connect:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("db instance is nil:", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://sql/migrations", "postgres", driver)
	if err != nil {
		log.Fatal("migration files not found:", err)
	}

	// migrate up
	downErr := m.Down()
	if downErr != nil {
		if downErr == migrate.ErrNoChange {
			log.Println("migration is up to date:", downErr)
		} else {
			log.Fatal("migration up error:", downErr)
		}
	}

	if downErr == nil {
		log.Println("database migrated down successfully")
	}
}
