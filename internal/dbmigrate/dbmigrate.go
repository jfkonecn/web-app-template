package dbmigrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jfkonecn/web-app-template/internal/config"
	_ "github.com/lib/pq"
)

func Open(dir string) (*migrate.Migrate, func(), error) {
	migrationsSourceURL, err := sourceURL(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid migrations directory: %w", err)
	}

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.PostgresDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("open postgres connection: %w", err)
	}

	closeFn := func() {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "close postgres connection: %v\n", closeErr)
		}
	}

	if err := db.Ping(); err != nil {
		closeFn()
		return nil, nil, fmt.Errorf("ping postgres: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		closeFn()
		return nil, nil, fmt.Errorf("create postgres migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsSourceURL, "postgres", driver)
	if err != nil {
		closeFn()
		return nil, nil, fmt.Errorf("initialize migrations: %w", err)
	}

	return m, closeFn, nil
}

func sourceURL(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("path is empty")
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", absPath)
	}

	return "file://" + filepath.ToSlash(absPath), nil
}
