package cmd

import (
	"database/sql"
	"fmt"

	"finance-tracker/internal/config"
	"finance-tracker/internal/db"
)

// withDB provides a safe way to perform database operations by handling the setup,
// error checking, and cleanup of database connections. It accepts a function that
// defines the database operations to perform.
//
// The function will:
// 1. Load the application config
// 2. Establish a database connection
// 3. Execute the provided operation
// 4. Automatically close the connection when done
//
// Example usage:
//
//	err := withDB(func(db *sql.DB) error {
//	    return db.QueryRow("SELECT something FROM table")
//	})
func withDB(fn func(*sql.DB) error) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.DBPath == "" {
		return fmt.Errorf("database path not configured. Please run 'db init' first")
	}

	database, err := db.OpenDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	return fn(database)
} 