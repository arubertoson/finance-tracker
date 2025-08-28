package cmd

import (
	"database/sql"
	"finance-tracker/internal/transaction"
	"fmt"

	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage categories",
}

var addCategoryCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new category",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		categoryType, _ := cmd.Flags().GetString("type")
		description, _ := cmd.Flags().GetString("description")

		return withDB(func(db *sql.DB) error {
			service := transaction.NewTransactionService(db)
			if err := service.CreateCategory(name, categoryType, description); err != nil {
				return fmt.Errorf("error creating category: %w", err)
			}

			fmt.Println("Successfully created category")
			return nil
		})
	},
}

func init() {
	// Add flags for addCategory command
	addCategoryCmd.Flags().String("name", "", "Name of the category")
	addCategoryCmd.Flags().String("type", "", "Type of the category (income/expense)")
	addCategoryCmd.Flags().String("description", "", "Description of the category")

	// Mark required flags for addCategory command
	addCategoryCmd.MarkFlagRequired("name")
	addCategoryCmd.MarkFlagRequired("type")

	// Add subcommands to category command
	categoryCmd.AddCommand(addCategoryCmd)
}
