package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the finance tracker",
	Long:  `Initialize a new database and create necessary configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// First initialize the database
		if err := dbInitCmd.RunE(cmd, args); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Then apply migrations
		if err := dbMigrateCmd.RunE(cmd, args); err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}

		fmt.Println("\nFinance tracker initialized successfully!")
		fmt.Println("You can now start adding transactions with:")
		fmt.Println("  finance transaction add --help")
		return nil
	},
}

func init() {
	// Copy the flags from dbInitCmd
	initCmd.Flags().String("path", "", "Path to the database file")
}
