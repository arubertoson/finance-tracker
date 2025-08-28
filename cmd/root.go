package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "finance",
	Short: "Finance Tracker CLI",
	Long:  `A command-line tool for managing personal finances.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Unwrap the error to get the original error
		originalErr := errors.Unwrap(err)
		if originalErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", originalErr)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func init() {
	// Setup command groups
	rootCmd.AddGroup(&cobra.Group{
		ID:    "setup",
		Title: "Setup Commands:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core Commands:",
	})

	// Assign commands to groups
	initCmd.GroupID = "setup"

	transactionCmd.GroupID = "core"
	importCmd.GroupID = "core"
	versionCmd.GroupID = "core"
	dbCmd.GroupID = "core"
	categoryCmd.GroupID = "core"
	ruleCmd.GroupID = "core"
	budgetCmd.GroupID = "core"

	// Add commands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(transactionCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(categoryCmd)
	rootCmd.AddCommand(ruleCmd)
	rootCmd.AddCommand(budgetCmd)
}
