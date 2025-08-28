package parsers

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"finance-tracker/internal/transaction"
)

// BankStatementRecord represents a record in the updated bank statement CSV
type BankStatementRecord struct {
	AccountNumber   string `csv:"Kontonummer"`
	BookingDate     string `csv:"Buchungsdatum"`
	ValueDate       string `csv:"Valuta"`
	Recipient1      string `csv:"Empfaenger 1"`
	Recipient2      string `csv:"Empfaenger 2"`
	Purpose         string `csv:"Verwendungszweck"`
	Amount          string `csv:"Betrag"`
	Currency        string `csv:"Waehrung"`
	Entity          string `csv:"Entity"`
	TransactionType string `csv:"Transaction Type"`
	Category        string `csv:"Category"`
}

// ParseBankStatementCSV parses the updated bank statement CSV into transactions
func ParseBankStatementCSV(reader io.Reader) ([]*transaction.Transaction, error) {
	records, err := ParseCSV[BankStatementRecord](reader)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Parsed %d records\n", len(records))

	var transactions []*transaction.Transaction
	for i, record := range records {
		transaction, err := record.ToTransaction()
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", i, err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransaction converts an UpdatedBankStatementRecord to a Transaction
func (r *BankStatementRecord) ToTransaction() (*transaction.Transaction, error) {
	if r.BookingDate == "" {
		return nil, fmt.Errorf("empty booking date field")
	}

	date, err := time.Parse("02.01.2006", r.BookingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid booking date: %s", r.BookingDate)
	}

	// Replace comma with dot for decimal conversion
	amountStr := strings.Replace(r.Amount, ".", "", -1) // Remove thousand separator
	amountStr = strings.Replace(amountStr, ",", ".", 1) // Replace decimal separator

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %s", r.Amount)
	}

	return &transaction.Transaction{
		Date:        date,
		Amount:      amount,
		Description: strings.TrimSpace(r.Purpose),
		Entity:      r.Entity,
		CategoryID:  1, // Default to "Uncategorized"
	}, nil
}
