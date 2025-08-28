package cmd

import (
	"fmt"
	"time"

	"database/sql"
	"finance-tracker/internal/transaction"

	"github.com/spf13/cobra"
)

var transactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"tx", "t"},
	Short:   "Manage transactions",
	Long:    `Create financial transactions.`,
}

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a new transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		amount, _ := cmd.Flags().GetFloat64("amount")
		description, _ := cmd.Flags().GetString("description")
		categoryID, _ := cmd.Flags().GetInt64("category-id")
		entity, _ := cmd.Flags().GetString("entity")

		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return fmt.Errorf("invalid date format: %v", err)
		}

		t := &transaction.Transaction{
			Date:        parsedDate,
			Amount:      amount,
			Description: description,
			CategoryID:  categoryID,
			Entity:      entity,
		}

		return withDB(func(db *sql.DB) error {
			service := transaction.NewTransactionService(db)
			if err := service.Create(t); err != nil {
				return fmt.Errorf("error creating transaction: %w", err)
			}

			// Apply rules to categorize the transaction
			if err := service.ApplyRules(t); err != nil {
				return fmt.Errorf("error applying rules: %w", err)
			}

			fmt.Printf("Successfully created transaction with ID: %d\n", t.ID)
			return nil
		})
	},
}

var makeRecurringCmd = &cobra.Command{
	Use:     "make-recurring",
	Aliases: []string{"recurring", "rec"},
	Short:   "Convert a transaction into a recurring pattern",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetInt64("id")
		frequency, _ := cmd.Flags().GetString("frequency")
		interval, _ := cmd.Flags().GetInt("interval")

		return withDB(func(db *sql.DB) error {
			service := transaction.NewTransactionService(db)

			// Create recurring pattern
			pattern := &transaction.RecurringPattern{
				Frequency: frequency,
				Interval:  interval,
			}

			if err := service.CreateRecurringPattern(pattern); err != nil {
				return fmt.Errorf("error creating recurring pattern: %w", err)
			}

			// Update the transaction to link it to the pattern
			t, err := service.Get(id)
			if err != nil {
				return fmt.Errorf("error getting transaction: %w", err)
			}

			t.RecurringPatternID = &pattern.ID
			if err := service.Update(t); err != nil {
				return fmt.Errorf("error updating transaction: %w", err)
			}

			fmt.Printf("Successfully created recurring pattern from transaction %d\n", id)
			return nil
		})
	},
}

func init() {
	// Add flags for add command
	addCmd.Flags().String("date", "", "Transaction date (YYYY-MM-DD)")
	addCmd.Flags().Float64("amount", 0.0, "Transaction amount")
	addCmd.Flags().String("description", "", "Transaction description")
	addCmd.Flags().Int64("category-id", 0, "Category ID")
	addCmd.Flags().String("entity", "", "Transaction entity")

	// Mark required flags for add command
	addCmd.MarkFlagRequired("date")
	addCmd.MarkFlagRequired("amount")
	addCmd.MarkFlagRequired("description")
	addCmd.MarkFlagRequired("category-id")

	makeRecurringCmd.Flags().Int64("id", 0, "Transaction ID to convert")
	makeRecurringCmd.Flags().String("frequency", "monthly", "Frequency (daily/weekly/monthly/yearly)")
	makeRecurringCmd.Flags().Int("interval", 1, "Interval between recurrences (e.g., every 2 weeks)")
	makeRecurringCmd.MarkFlagRequired("id")

	// Add subcommands to transaction command
	transactionCmd.AddCommand(addCmd)
	transactionCmd.AddCommand(makeRecurringCmd)
}
