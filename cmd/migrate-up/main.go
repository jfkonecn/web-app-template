package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jfkonecn/web-app-template/internal/dbmigrate"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <migrations-dir>", filepath.Base(os.Args[0]))
	}

	m, closeFn, err := dbmigrate.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeFn()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Print("no migrations to apply")
			return
		}
		log.Fatalf("apply migrations: %v", err)
	}

	log.Print("migrations applied")
}
