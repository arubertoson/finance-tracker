package cmd

import (
	"database/sql"
	"fmt"

	"finance-tracker/internal/importer"
	"finance-tracker/internal/transaction"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import transactions from CSV",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("error getting file flag: %w", err)
		}
		if filename == "" {
			return fmt.Errorf("file path is required")
		}

		// Import and parse the file
		transactions, err := importer.ImportCSV(filename)
		if err != nil {
			return fmt.Errorf("error parsing file: %w", err)
		}

		fmt.Printf("Found %d transactions to import\n", len(transactions))

		return withDB(func(db *sql.DB) error {
			transactionService := transaction.NewTransactionService(db)
			imported := 0

			for _, t := range transactions {
				if err := transactionService.Create(t); err != nil {
					fmt.Printf("Warning: Failed to import transaction from %s: %v\n",
						t.Date.Format("2006-01-02"), err)
					continue
				}
				imported++
			}

			fmt.Printf("Successfully imported %d out of %d transactions\n",
				imported, len(transactions))
			return nil
		})
	},
}

func init() {
	importCmd.Flags().StringP("file", "f", "", "Path to the CSV file")
	importCmd.MarkFlagRequired("file")
}
