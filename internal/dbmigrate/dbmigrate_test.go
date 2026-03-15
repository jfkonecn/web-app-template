package dbmigrate_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jfkonecn/web-app-template/internal/dbmigrate"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestOpenRunsMigrationsUpDownUp(t *testing.T) {
	ctx := context.Background()

	const (
		dbName  = "web_app_template_test"
		dbUser  = "postgres"
		dbPass  = "postgres"
		dbImage = "postgres:16-alpine"
		dbPort  = "5432/tcp"
		sslMode = "disable"
	)

	container, err := tcpostgres.Run(
		ctx,
		dbImage,
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPass),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("terminate postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("get postgres host: %v", err)
	}

	port, err := container.MappedPort(ctx, dbPort)
	if err != nil {
		t.Fatalf("get postgres mapped port: %v", err)
	}

	t.Setenv("POSTGRES_HOST", host)
	t.Setenv("POSTGRES_PORT", port.Port())
	t.Setenv("POSTGRES_USER", dbUser)
	t.Setenv("POSTGRES_PASSWORD", dbPass)
	t.Setenv("POSTGRES_DB", dbName)
	t.Setenv("POSTGRES_SSLMODE", sslMode)

	migrationsDir := repoMigrationsDir(t)

	runUp(t, migrationsDir)
	runDown(t, migrationsDir)
	runUp(t, migrationsDir)
}

func runUp(t *testing.T, migrationsDir string) {
	t.Helper()

	m, closeFn, err := dbmigrate.Open(migrationsDir)
	if err != nil {
		t.Fatalf("open migrator for up: %v", err)
	}
	defer closeFn()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("run up migrations: %v", err)
	}
}

func runDown(t *testing.T, migrationsDir string) {
	t.Helper()

	m, closeFn, err := dbmigrate.Open(migrationsDir)
	if err != nil {
		t.Fatalf("open migrator for down: %v", err)
	}
	defer closeFn()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("run down migration: %v", err)
	}
}

func repoMigrationsDir(t *testing.T) string {
	t.Helper()

	migrationsDir := filepath.Join("..", "..", "db", "migrations")
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}

	return absPath
}
