package budget

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type BudgetService struct {
	db *sql.DB
}

func NewBudgetService(db *sql.DB) *BudgetService {
	return &BudgetService{db: db}
}

// ParsePeriod parses time period strings like "1m", "3m", "2w", "1y"
// Returns the start and end dates for the period
func ParsePeriod(period string) (time.Time, time.Time, error) {
	period = strings.ToLower(strings.TrimSpace(period))

	// Special case for "this-month", "last-month"
	switch period {
	case "this-month":
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return start, end, nil
	case "last-month":
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
		end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)
		return start, end, nil
	}

	// Parse period strings like "3m", "2w", "1y"
	period = strings.TrimSpace(period)
	if len(period) < 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period format: %s (expected format: '3m', '2w', '1y', 'this-month', 'last-month')", period)
	}

	// Split into number and unit
	num, err := strconv.Atoi(period[:len(period)-1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid number in period: %w", err)
	}

	unit := period[len(period)-1:]

	if num <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("period must be positive")
	}

	endDate := time.Now()
	var startDate time.Time

	switch unit {
	case "w":
		startDate = endDate.AddDate(0, 0, -7*num)
	case "m":
		startDate = endDate.AddDate(0, -num, 0)
	case "y":
		startDate = endDate.AddDate(-num, 0, 0)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period unit: %s", unit)
	}

	// Normalize dates to start/end of day
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	return startDate, endDate, nil
}

func (s *BudgetService) SetBudget(category string, amount float64, period string, rollover bool) error {
	// Parse the period string
	startDate, _, err := ParsePeriod(period)
	if err != nil {
		return fmt.Errorf("invalid period: %w", err)
	}

	// Get category ID
	var categoryID int
	err = s.db.QueryRow("SELECT id FROM categories WHERE name = ?", category).Scan(&categoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Insert or update the budget plan
	query := `
		INSERT INTO budget_plans (category_id, month, planned_amount, rollover)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(category_id, month) DO UPDATE SET
		planned_amount = excluded.planned_amount,
		rollover = excluded.rollover
	`
	_, err = s.db.Exec(query, categoryID, startDate, amount, rollover)
	if err != nil {
		return fmt.Errorf("failed to set budget: %w", err)
	}

	return nil
}

func (s *BudgetService) AnalyzeBudget(period string) (string, error) {
	// Parse the period string
	startDate, endDate, err := ParsePeriod(period)
	if err != nil {
		return "", fmt.Errorf("invalid period: %w", err)
	}

	// Query to calculate budget vs actual spending
	query := `
		SELECT c.name, bp.planned_amount, COALESCE(SUM(t.amount), 0) as actual_spending
		FROM budget_plans bp
		JOIN categories c ON bp.category_id = c.id
		LEFT JOIN transactions t ON t.category_id = c.id AND t.date BETWEEN ? AND ?
		WHERE bp.month BETWEEN ? AND ?
		GROUP BY c.name, bp.planned_amount
		ORDER BY c.name
	`

	rows, err := s.db.Query(query, startDate, endDate, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to analyze budget: %w", err)
	}
	defer rows.Close()

	var report strings.Builder
	report.WriteString(fmt.Sprintf("Budget Analysis (%s to %s)\n",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02")))
	report.WriteString("----------------------------------------\n\n")

	// Track totals
	var totalPlanned, totalActual float64

	// Category breakdown
	report.WriteString("Category Breakdown:\n")
	report.WriteString("------------------\n")
	for rows.Next() {
		var category string
		var planned, actual float64
		if err := rows.Scan(&category, &planned, &actual); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		variance := planned - actual
		percentUsed := 0.0
		if planned > 0 {
			percentUsed = (actual / planned) * 100
		}

		report.WriteString(fmt.Sprintf("Category: %s\n", category))
		report.WriteString(fmt.Sprintf("  Planned: $%.2f\n", planned))
		report.WriteString(fmt.Sprintf("  Actual:  $%.2f\n", actual))
		report.WriteString(fmt.Sprintf("  Variance: $%.2f\n", variance))
		report.WriteString(fmt.Sprintf("  Used: %.1f%%\n\n", percentUsed))

		totalPlanned += planned
		totalActual += actual
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("row iteration error: %w", err)
	}

	// Overall summary
	totalVariance := totalPlanned - totalActual
	totalPercentUsed := 0.0
	if totalPlanned > 0 {
		totalPercentUsed = (totalActual / totalPlanned) * 100
	}

	report.WriteString("Overall Summary:\n")
	report.WriteString("----------------\n")
	report.WriteString(fmt.Sprintf("Total Planned: $%.2f\n", totalPlanned))
	report.WriteString(fmt.Sprintf("Total Actual:  $%.2f\n", totalActual))
	report.WriteString(fmt.Sprintf("Total Variance: $%.2f\n", totalVariance))
	report.WriteString(fmt.Sprintf("Total Budget Used: %.1f%%\n", totalPercentUsed))

	return report.String(), nil
}

// AnalyzeBudgetWithCharts generates both text report and HTML charts
func (s *BudgetService) AnalyzeBudgetWithCharts(period string) (string, error) {
	startDate, endDate, err := ParsePeriod(period)
	if err != nil {
		return "", err
	}

	// Query data (same as before)
	query := `
		SELECT c.name, bp.planned_amount, COALESCE(SUM(t.amount), 0) as actual_spending
		FROM budget_plans bp
		JOIN categories c ON bp.category_id = c.id
		LEFT JOIN transactions t ON t.category_id = c.id AND t.date BETWEEN ? AND ?
		WHERE bp.month BETWEEN ? AND ?
		GROUP BY c.name, bp.planned_amount
		ORDER BY c.name
	`

	rows, err := s.db.Query(query, startDate, endDate, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to analyze budget: %w", err)
	}
	defer rows.Close()

	// Collect data for charts
	var categories []string
	var plannedData []opts.BarData
	var actualData []opts.BarData
	var pieData []opts.PieData

	// Generate text report (same as before)
	var report strings.Builder
	// ... (previous report generation code) ...

	// Create bar chart comparing planned vs actual
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Budget vs Actual Spending",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
		}),
	)

	bar.SetXAxis(categories).
		AddSeries("Planned", plannedData).
		AddSeries("Actual", actualData)

	// Create pie chart for spending distribution
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Spending Distribution",
		}),
	)
	pie.AddSeries("Spending", pieData)

	// Save charts to HTML file
	f, err := os.Create("budget_analysis.html")
	if err != nil {
		return "", err
	}
	defer f.Close()

	page := charts.NewPage()
	page.AddCharts(bar, pie)
	err = page.Render(f)
	if err != nil {
		return "", err
	}

	report.WriteString("\nCharts have been generated in 'budget_analysis.html'\n")
	return report.String(), nil
}

// Helper function to generate monthly trend data
func (s *BudgetService) getMonthlyTrends(categoryID int, startDate, endDate time.Time) ([]time.Time, []float64, error) {
	query := `
		SELECT DATE_TRUNC('month', date) as month,
			   SUM(amount) as monthly_total
		FROM transactions
		WHERE category_id = ? AND date BETWEEN ? AND ?
		GROUP BY DATE_TRUNC('month', date)
		ORDER BY month
	`

	rows, err := s.db.Query(query, categoryID, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var dates []time.Time
	var amounts []float64

	for rows.Next() {
		var date time.Time
		var amount float64
		if err := rows.Scan(&date, &amount); err != nil {
			return nil, nil, err
		}
		dates = append(dates, date)
		amounts = append(amounts, amount)
	}

	return dates, amounts, nil
}
