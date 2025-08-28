package cmd

import (
	"database/sql"
	"finance-tracker/internal/budget"
	"fmt"

	"github.com/spf13/cobra"
)

var budgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Manage budgets",
}

var setBudgetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a budget for a category",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		amount, _ := cmd.Flags().GetFloat64("amount")
		period, _ := cmd.Flags().GetString("period")
		rollover, _ := cmd.Flags().GetBool("rollover")

		return withDB(func(db *sql.DB) error {
			service := budget.NewBudgetService(db)
			if err := service.SetBudget(category, amount, period, rollover); err != nil {
				return fmt.Errorf("error setting budget: %w", err)
			}

			fmt.Println("Successfully set budget")
			return nil
		})
	},
}

var analyzeBudgetCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze budget vs actual spending",
	RunE: func(cmd *cobra.Command, args []string) error {
		period, _ := cmd.Flags().GetString("period")
		charts, _ := cmd.Flags().GetBool("charts")

		return withDB(func(db *sql.DB) error {
			service := budget.NewBudgetService(db)
			
			var report string
			var err error
			if charts {
				report, err = service.AnalyzeBudgetWithCharts(period)
			} else {
				report, err = service.AnalyzeBudget(period)
			}
			
			if err != nil {
				return fmt.Errorf("error analyzing budget: %w", err)
			}

			fmt.Println(report)
			return nil
		})
	},
}

func init() {
	setBudgetCmd.Flags().String("category", "", "Category for the budget")
	setBudgetCmd.Flags().Float64("amount", 0, "Budget amount")
	setBudgetCmd.Flags().String("period", "1m", "Budget period (e.g., '1m', '3m', '1y', 'this-month', 'last-month')")
	setBudgetCmd.Flags().Bool("rollover", false, "Enable rollover of unused amounts")

	analyzeBudgetCmd.Flags().String("period", "3m", "Analysis period (e.g., '3m', '6m', '1y')")
	analyzeBudgetCmd.Flags().Bool("charts", false, "Generate visual charts")

	// Add subcommands to the budget command
	budgetCmd.AddCommand(setBudgetCmd)
	budgetCmd.AddCommand(analyzeBudgetCmd)
}
