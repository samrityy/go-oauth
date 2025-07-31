package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a migration direction: 'up' or 'down'")
	}

	direction := os.Args[1]
	dsn := os.Getenv("POSTGRES_DSN") // Set this in your .env

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatal("Migration init failed:", err)
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err.Error() != "no change" {
			log.Fatal("Migration up failed:", err)
		}
	case "down":
		if err := m.Down(); err != nil && err.Error() != "no change" {
			log.Fatal("Migration down failed:", err)
		}
	default:
		log.Fatal("Invalid direction. Use 'up' or 'down'.")
	}

	log.Println("✅ Migration", direction, "completed.")
}
