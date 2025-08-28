package transaction

import (
	"database/sql"
	"fmt"
	"time"
)

type sqlTransactionRepository struct {
	db *sql.DB
}

// NewSQLTransactionRepository creates a new SQL-based repository
func NewSQLTransactionRepository(db *sql.DB) TransactionRepository {
	return &sqlTransactionRepository{db: db}
}

func (r *sqlTransactionRepository) Create(t *Transaction) error {
	query := `
		INSERT INTO transactions (
			date, amount, description, category_id, 
			entity, recurring_pattern_id
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		t.Date, t.Amount, t.Description, t.CategoryID,
		t.Entity, t.RecurringPatternID,
	)
	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}
	t.ID = id
	return nil
}

func (r *sqlTransactionRepository) Get(id int64) (*Transaction, error) {
	t := &Transaction{}
	query := `
		SELECT id, date, amount, description, category_id,
			entity, recurring_pattern_id
		FROM transactions WHERE id = ?
	`
	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.Date, &t.Amount, &t.Description, &t.CategoryID,
		&t.Entity, &t.RecurringPatternID,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// GetAllTransactions retrieves all transactions from the database
func (r *sqlTransactionRepository) GetAllTransactions() ([]*Transaction, error) {
	query := `
		SELECT id, date, amount, description, category_id, entity, recurring_pattern_id
		FROM transactions
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error retrieving transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		t := &Transaction{}
		err := rows.Scan(
			&t.ID, &t.Date, &t.Amount, &t.Description,
			&t.CategoryID, &t.Entity, &t.RecurringPatternID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

func (r *sqlTransactionRepository) Update(t *Transaction) error {
	query := `
		UPDATE transactions SET
			date = ?, amount = ?, description = ?, category_id = ?,
			entity = ?, recurring_pattern_id = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(query,
		t.Date, t.Amount, t.Description, t.CategoryID,
		t.Entity, t.RecurringPatternID, t.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating transaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("transaction not found: %d", t.ID)
	}
	return nil
}

// Implement other repository methods...
func (r *sqlTransactionRepository) CreateBatch(transactions []*Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO transactions (
			date, amount, description, category_id, 
			entity, recurring_pattern_id
		) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, t := range transactions {
		result, err := stmt.Exec(
			t.Date, t.Amount, t.Description, t.CategoryID,
			t.Entity, t.RecurringPatternID,
		)
		if err != nil {
			return fmt.Errorf("error executing batch insert: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("error getting last insert ID: %w", err)
		}
		t.ID = id
	}

	return tx.Commit()
}

// Implement remaining repository methods...

// CreateRecurringPattern creates a new recurring pattern in the database
func (r *sqlTransactionRepository) CreateRecurringPattern(p *RecurringPattern) error {
	query := `INSERT INTO recurring_patterns (frequency, interval) VALUES (?, ?)`
	result, err := r.db.Exec(query, p.Frequency, p.Interval)
	if err != nil {
		return fmt.Errorf("error creating recurring pattern: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}
	p.ID = id
	return nil
}

// CreateCategory adds a new transaction category
func (r *sqlTransactionRepository) CreateCategory(name, categoryType, description string) error {
	query := `
		INSERT INTO categories (name, type, description)
		VALUES (?, ?, ?)
	`
	_, err := r.db.Exec(query, name, categoryType, description)
	if err != nil {
		return fmt.Errorf("error creating category: %w", err)
	}
	return nil
}

// CreateRule adds a new categorization rule
func (r *sqlTransactionRepository) CreateRule(pattern string, categoryID int64, priority int) error {
	query := `
		INSERT INTO category_rules (pattern, category_id, priority)
		VALUES (?, ?, ?)
	`
	_, err := r.db.Exec(query, pattern, categoryID, priority)
	if err != nil {
		return fmt.Errorf("error creating rule: %w", err)
	}
	return nil
}

// FindCategoryIDByRules attempts to find a matching category ID based on rules
func (r *sqlTransactionRepository) FindCategoryIDByRules(entity, description string) (int64, error) {
	query := `
		SELECT category_id
		FROM category_rules
		WHERE (
			? GLOB pattern OR ? GLOB pattern
		)
		ORDER BY priority DESC
		LIMIT 1
	`
	var categoryID int64
	err := r.db.QueryRow(query, entity, description).Scan(&categoryID)
	if err != nil {
		return 0, err
	}
	return categoryID, nil
}

// GenerateRules analyzes uncategorized transactions to find common patterns
func (r *sqlTransactionRepository) GenerateRules(limit int) ([]*Rule, error) {
	query := `
		SELECT 
			0 as id,
			entity as pattern,
			category_id,
			COUNT(*) as priority,
			CURRENT_TIMESTAMP as created_at,
			description as sample_description
		FROM transactions
		WHERE category_id = 1
		GROUP BY entity, description
		ORDER BY priority DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error analyzing transactions: %w", err)
	}
	defer rows.Close()

	var rules []*Rule
	for rows.Next() {
		r := &Rule{}
		var sampleDesc string
		var createdAtStr string
		err := rows.Scan(
			&r.ID, &r.Pattern, &r.CategoryID, &r.Priority, &createdAtStr,
			&sampleDesc,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning pattern: %w", err)
		}

		r.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %w", err)
		}

		r.Pattern = fmt.Sprintf("%s (%d occurrences)\nExample: %s",
			r.Pattern, r.Priority, sampleDesc)
		rules = append(rules, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating patterns: %w", err)
	}

	return rules, nil
}
