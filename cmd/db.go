package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"

	"finance-tracker/internal/config"
	"finance-tracker/internal/db"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new database",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if dbPath, _ := cmd.Flags().GetString("path"); dbPath != "" {
			cfg.DBPath = dbPath
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		database, err := db.InitDB(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer database.Close()

		fmt.Println("Database initialized successfully at:", cfg.DBPath)
		return nil
	},
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply any pending database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(database *sql.DB) error {
			if err := db.ApplyMigrations(database); err != nil {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}

			fmt.Println("Migrations applied successfully")
			return nil
		})
	},
}

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database migration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return withDB(func(database *sql.DB) error {
			version, err := db.GetCurrentVersion(database)
			if err != nil {
				return fmt.Errorf("failed to get database version: %w", err)
			}

			fmt.Printf("Database location: %s\n", cfg.DBPath)
			fmt.Printf("Current database version: %d\n", version)
			return nil
		})
	},
}

func init() {
	dbInitCmd.Flags().String("path", "", "Path to the database file")

	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbStatusCmd)
}
