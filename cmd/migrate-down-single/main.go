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

	if err := m.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Print("no migration to roll back")
			return
		}
		log.Fatalf("roll back migration: %v", err)
	}

	log.Print("rolled back one migration")
}
