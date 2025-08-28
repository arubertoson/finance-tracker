package cmd

import (
	"database/sql"
	"finance-tracker/internal/transaction"
	"fmt"

	"github.com/spf13/cobra"
)

var ruleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Manage categorization rules",
}

var addRuleCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new categorization rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern, _ := cmd.Flags().GetString("pattern")
		categoryID, _ := cmd.Flags().GetInt64("category-id")
		priority, _ := cmd.Flags().GetInt("priority")

		return withDB(func(db *sql.DB) error {
			service := transaction.NewTransactionService(db)
			if err := service.CreateRule(pattern, categoryID, priority); err != nil {
				return fmt.Errorf("error creating rule: %w", err)
			}
			return nil
		})
	},
}

var generateRulesCmd = &cobra.Command{
	Use:   "generate",
	Short: "Analyze uncategorized transactions for patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		return withDB(func(db *sql.DB) error {
			service := transaction.NewTransactionService(db)

			patterns, err := service.GenerateRules(limit)
			if err != nil {
				return fmt.Errorf("error analyzing transactions: %w", err)
			}

			fmt.Println("\nFound patterns in uncategorized transactions:")
			fmt.Println("--------------------------------------------")
			for _, p := range patterns {
				fmt.Printf("\n%s\n", p.Pattern)
			}
			fmt.Println("\nTo create a rule for a pattern:")
			fmt.Println("finance-tracker rule add --pattern \"PATTERN\" --category-id ID")
			return nil
		})
	},
}

var applyRulesCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply rules to all transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *sql.DB) error {
			service := transaction.NewTransactionService(db)
			if err := service.ApplyRulesToAll(); err != nil {
				return fmt.Errorf("error applying rules to all transactions: %w", err)
			}

			fmt.Println("Successfully applied rules to all transactions")
			return nil
		})
	},
}

func init() {
	addRuleCmd.Flags().String("pattern", "", "Pattern to match (supports * wildcards)")
	addRuleCmd.Flags().Int64("category-id", 0, "Category ID for the rule")
	addRuleCmd.Flags().Int("priority", 0, "Rule priority (higher = checked first)")

	addRuleCmd.MarkFlagRequired("pattern")
	addRuleCmd.MarkFlagRequired("category-id")

	generateRulesCmd.Flags().Int("limit", 10, "Limit the number of patterns to generate")

	ruleCmd.AddCommand(addRuleCmd)
	ruleCmd.AddCommand(generateRulesCmd)
	ruleCmd.AddCommand(applyRulesCmd)
}
