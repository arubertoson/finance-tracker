package parsers

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"finance-tracker/internal/transaction"
)

type N26Record struct {
	Date                  string `csv:"Date"`
	Payee                 string `csv:"Payee"`
	AccountNumber         string `csv:"Account number"`
	TransactionType       string `csv:"Transaction type"`
	PaymentReference      string `csv:"Payment reference"`
	AmountEUR             string `csv:"Amount (EUR)"`
	AmountForeignCurrency string `csv:"Amount (Foreign Currency)"`
	ForeignCurrencyType   string `csv:"Type Foreign Currency"`
	ExchangeRate          string `csv:"Exchange Rate"`
}

func ParseN26CSV(reader io.Reader) ([]*transaction.Transaction, error) {
	records, err := ParseCSV[N26Record](reader)
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

func (r *N26Record) ToTransaction() (*transaction.Transaction, error) {
	if r.Date == "" {
		return nil, fmt.Errorf("empty date field")
	}

	date, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return nil, err
	}

	amount, err := strconv.ParseFloat(r.AmountEUR, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %s", r.AmountEUR)
	}

	return &transaction.Transaction{
		Date:        date,
		Amount:      amount,
		Description: strings.TrimSpace(r.PaymentReference),
		Entity:      r.Payee,
		CategoryID:  1, // Default to "Uncategorized"
	}, nil
}
