package main

import (
	"URLRotatorGo/infra/config"
	"database/sql"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"strings"
	"time"
)

func main() {
	startTime := time.Now()

	cfg := config.InitConfig("config.json", "../../")

	var direction string
	flag.StringVar(&direction, "d", "up", "example: -d up/down")
	flag.Parse()

	migrationsPath := cfg.GetString("database.postgres.migrations_path")
	if migrationsPath == "" {
		log.Fatal("database migrations path not set")
	}

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.GetString("database.postgres.user"),
		cfg.GetString("database.postgres.pass"),
		cfg.GetString("database.postgres.host"),
		cfg.GetString("database.postgres.port"),
		cfg.GetString("database.postgres.dbname"),
	)

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("failed to connect to database: ", err.Error())
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("failed to create postgres driver: ", err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver,
	)
	if err != nil {
		log.Fatal("failed to create migration instance: ", err.Error())
	}

	switch strings.ToLower(direction) {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("migration up failed: ", err)
		}
		log.Println("Migration up completed successfully.")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("migration down failed:", err)
		}
		log.Println("Migration down completed successfully.")
	default:
		log.Fatal("invalid direction: ", direction)
	}

	log.Println("Time elapsed: ", time.Since(startTime))
}
