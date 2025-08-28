package transaction

import (
	"database/sql"
	"fmt"
	"time"
)

// Transaction represents a financial transaction in the database
type Transaction struct {
	ID                 int64
	Date               time.Time
	Amount             float64
	Description        string
	CategoryID         int64
	Entity             string
	RecurringPatternID *int64
}

// Error types and constants
type TransactionError struct {
	Code    string
	Message string
	Err     error
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *TransactionError) Unwrap() error {
	return e.Err
}

const (
	ErrNotFound   = "NOT_FOUND"
	ErrValidation = "VALIDATION"
	ErrDatabase   = "DATABASE"
)

// RecurringPattern represents a template for recurring transactions
type RecurringPattern struct {
	ID        int64
	Frequency string
	Interval  int
	CreatedAt time.Time
}

// Rule represents a categorization rule in the database
type Rule struct {
	ID         int64
	Pattern    string
	CategoryID int64
	Priority   int
	CreatedAt  time.Time
}

// TransactionRepository defines the interface for transaction storage operations
type TransactionRepository interface {
	Create(t *Transaction) error
	Get(id int64) (*Transaction, error)
	Update(t *Transaction) error
	CreateBatch(transactions []*Transaction) error
	CreateRecurringPattern(p *RecurringPattern) error
	CreateCategory(name, categoryType, description string) error
	CreateRule(pattern string, categoryID int64, priority int) error
	FindCategoryIDByRules(entity, description string) (int64, error)
	GetAllTransactions() ([]*Transaction, error)
	GenerateRules(limit int) ([]*Rule, error)
}

// TransactionService handles transaction-related business logic
type TransactionService struct {
	repo TransactionRepository
}

// NewTransactionService creates a new TransactionService instance
func NewTransactionService(db *sql.DB) *TransactionService {
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &TransactionService{
		repo: NewSQLTransactionRepository(db),
	}
}

// Create creates a new transaction
func (s *TransactionService) Create(t *Transaction) error {
	if err := s.repo.Create(t); err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to create transaction",
			Err:     err,
		}
	}
	return nil
}

// Get retrieves a transaction by its ID
func (s *TransactionService) Get(id int64) (*Transaction, error) {
	t, err := s.repo.Get(id)
	if err == sql.ErrNoRows {
		return nil, &TransactionError{
			Code:    ErrNotFound,
			Message: fmt.Sprintf("transaction not found: %d", id),
			Err:     err,
		}
	}
	return t, err
}

// Update modifies an existing transaction
func (s *TransactionService) Update(t *Transaction) error {
	if err := s.repo.Update(t); err != nil {
		if err == sql.ErrNoRows {
			return &TransactionError{
				Code:    ErrNotFound,
				Message: fmt.Sprintf("transaction not found: %d", t.ID),
				Err:     err,
			}
		}
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to update transaction",
			Err:     err,
		}
	}
	return nil
}

// CreateRecurringPattern creates a new recurring pattern
func (s *TransactionService) CreateRecurringPattern(p *RecurringPattern) error {
	if err := s.repo.CreateRecurringPattern(p); err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to create recurring pattern",
			Err:     err,
		}
	}
	return nil
}

// CreateCategory adds a new transaction category
func (s *TransactionService) CreateCategory(name, categoryType, description string) error {
	if err := s.repo.CreateCategory(name, categoryType, description); err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to create category",
			Err:     err,
		}
	}
	return nil
}

// CreateRule adds a new categorization rule
func (s *TransactionService) CreateRule(pattern string, categoryID int64, priority int) error {
	if err := s.repo.CreateRule(pattern, categoryID, priority); err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to create rule",
			Err:     err,
		}
	}
	return nil
}

// GenerateRules generates new categorization rules
func (s *TransactionService) GenerateRules(limit int) ([]*Rule, error) {
	rules, err := s.repo.GenerateRules(limit)
	if err != nil {
		return nil, &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to generate rules",
			Err:     err,
		}
	}
	return rules, nil
}

// ApplyRules attempts to automatically categorize a transaction
func (s *TransactionService) ApplyRules(t *Transaction) error {
	categoryID, err := s.repo.FindCategoryIDByRules(t.Entity, t.Description)
	if err == sql.ErrNoRows {
		return nil // No matching rule found
	}
	if err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "error applying rules",
			Err:     err,
		}
	}

	if t.CategoryID == categoryID {
		return nil // Category already set correctly
	}

	t.CategoryID = categoryID
	return s.Update(t)
}

// ApplyRulesToAll applies categorization rules to all transactions
func (s *TransactionService) ApplyRulesToAll() error {
	transactions, err := s.repo.GetAllTransactions()
	if err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "error retrieving transactions",
			Err:     err,
		}
	}

	for _, t := range transactions {
		if err := s.ApplyRules(t); err != nil {
			return &TransactionError{
				Code:    ErrDatabase,
				Message: fmt.Sprintf("error applying rules to transaction %d", t.ID),
				Err:     err,
			}
		}
	}

	return nil
}

// CreateBatch creates multiple transactions in a batch
func (s *TransactionService) CreateBatch(transactions []*Transaction) error {
	if err := s.repo.CreateBatch(transactions); err != nil {
		return &TransactionError{
			Code:    ErrDatabase,
			Message: "failed to create batch transactions",
			Err:     err,
		}
	}
	return nil
}
