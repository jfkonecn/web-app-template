package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var migrationNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <migrations-dir> <migration-name>", filepath.Base(os.Args[0]))
	}

	migrationsDir := os.Args[1]
	migrationName := os.Args[2]

	if !migrationNamePattern.MatchString(migrationName) {
		log.Fatalf("invalid migration name %q: use lowercase letters, numbers, and underscores only", migrationName)
	}

	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		log.Fatalf("create migrations directory: %v", err)
	}

	nextSequence, err := nextMigrationSequence(migrationsDir)
	if err != nil {
		log.Fatalf("determine next migration sequence: %v", err)
	}

	upPath := filepath.Join(migrationsDir, fmt.Sprintf("%06d_%s.up.sql", nextSequence, migrationName))
	downPath := filepath.Join(migrationsDir, fmt.Sprintf("%06d_%s.down.sql", nextSequence, migrationName))

	if err := createEmptyFile(upPath); err != nil {
		log.Fatalf("create up migration: %v", err)
	}

	if err := createEmptyFile(downPath); err != nil {
		_ = os.Remove(upPath)
		log.Fatalf("create down migration: %v", err)
	}

	log.Printf("created %s", upPath)
	log.Printf("created %s", downPath)
}

func nextMigrationSequence(migrationsDir string) (int, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, err
	}

	maxSequence := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") && !strings.HasSuffix(name, ".down.sql") {
			continue
		}

		sequence, err := parseMigrationSequence(name)
		if err != nil {
			return 0, err
		}
		if sequence > maxSequence {
			maxSequence = sequence
		}
	}

	return maxSequence + 1, nil
}

func parseMigrationSequence(filename string) (int, error) {
	prefix, _, found := strings.Cut(filename, "_")
	if !found {
		return 0, fmt.Errorf("invalid migration filename %q", filename)
	}

	sequence, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("invalid migration filename %q: %w", filename, err)
	}

	return sequence, nil
}

func createEmptyFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	return file.Close()
}
