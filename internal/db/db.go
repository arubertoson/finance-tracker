// Package db provides database operations and migration management for SQLite databases.
// It handles connection management, schema versioning, and automated migrations.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB opens a database connection and configures it for SQLite-specific optimal usage.
// It ensures the database directory exists and validates the connection, but does not
// apply any migrations. This is useful when you need a raw database connection,
// typically for testing or specialized initialization.
//
// The dataSourceName parameter should be a valid SQLite connection string (file path).
// Returns the configured database connection and any error encountered.
func OpenDB(dataSourceName string) (*sql.DB, error) {
	dir := filepath.Dir(dataSourceName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite works best with minimal connection pooling
	// since it's a file-based database
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %s, with error: %w", dataSourceName, err)
	}

	return db, nil
}

// InitDB initializes a new database connection and ensures it's up-to-date by
// applying any pending migrations. This is the preferred method for normal application
// database initialization.
//
// The dataSourceName parameter should be a valid SQLite connection string (file path).
// Returns the initialized database connection and any error encountered.
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := OpenDB(dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := ApplyMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// GetCurrentVersion retrieves the highest applied migration version from the database.
// Returns -1 if no migrations have been applied or if the schema_versions table
// doesn't exist yet.
func GetCurrentVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), -1) FROM schema_versions").Scan(&version)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return -1, nil
		}
		return -1, fmt.Errorf("failed to get current schema version: %w", err)
	}
	return version, nil
}

// ApplyMigrations executes all pending database migrations in version order.
// Migrations are SQL files in the "migrations" directory, named in the format:
// "NNN_description.sql" where NNN is a three-digit version number.
//
// The function tracks applied migrations in the schema_versions table and only
// executes migrations with version numbers higher than the current database version.
func ApplyMigrations(db *sql.DB) error {
	currentVersion, err := GetCurrentVersion(db)
	if err != nil {
		return err
	}

	migrations, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to list migrations: %w", err)
	}
	sort.Strings(migrations)

	for _, migration := range migrations {
		basename := filepath.Base(migration)
		var version int
		_, err := fmt.Sscanf(basename, "%d_", &version)
		if err != nil {
			return fmt.Errorf("invalid migration filename format %s: %w", basename, err)
		}

		if version <= currentVersion {
			continue
		}

		if err := executeMigration(db, migration, version); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration, err)
		}
	}

	return nil
}

// executeMigration applies a single migration file to the database and records its
// version in the schema_versions table. The entire migration is executed as a single
// operation, ensuring consistency of the migration process.
func executeMigration(db *sql.DB, migrationPath string, version int) error {
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration %s: %w", migrationPath, err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if _, err := db.Exec("INSERT INTO schema_versions (version) VALUES (?)", version); err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	return nil
}
