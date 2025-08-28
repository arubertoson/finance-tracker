package report

import (
	"fmt"
	"time"

	"database/sql"
)

type CategorySummary struct {
	Category    string
	TotalAmount float64
	Count       int
}

type MonthlyReport struct {
	Month         time.Time
	Categories    []CategorySummary
	TotalIncome   float64
	TotalExpenses float64
	NetAmount     float64
}

func GenerateMonthlyReport(db *sql.DB, year int, month int) (*MonthlyReport, error) {
	query := `
		SELECT 
			c.name,
			SUM(t.amount) as total,
			COUNT(*) as count
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE strftime('%Y-%m', t.date) = ?
		GROUP BY c.name
	`

	rows, err := db.Query(query, fmt.Sprintf("%04d-%02d", year, month))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &MonthlyReport{
		Month: time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local),
	}

	for rows.Next() {
		var cs CategorySummary
		if err := rows.Scan(&cs.Category, &cs.TotalAmount, &cs.Count); err != nil {
			return nil, err
		}

		report.Categories = append(report.Categories, cs)
		if cs.TotalAmount > 0 {
			report.TotalIncome += cs.TotalAmount
		} else {
			report.TotalExpenses += -cs.TotalAmount
		}
	}

	report.NetAmount = report.TotalIncome - report.TotalExpenses
	return report, nil
}
