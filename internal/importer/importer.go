package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"finance-tracker/internal/importer/parsers"
	"finance-tracker/internal/transaction"
)

// csvFormat holds the type of record and the corresponding parsing function
type csvFormat struct {
	recordType reflect.Type
	parseFunc  func(io.Reader) ([]*transaction.Transaction, error)
}

// ImportCSV imports transactions from a CSV file.
// It detects the CSV structure and uses the appropriate parsing function.
func ImportCSV(filename string) ([]*transaction.Transaction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	parseFunc, err := detectCSVStructure(file)
	if err != nil {
		return nil, err
	}

	// This is necessary because the file reader's position is at the end of the headers
	// after detectCSVStructure reads them. Reopening the file resets the reader to the start.
	file.Close()
	file, err = os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error reopening file: %w", err)
	}
	defer file.Close()

	return parseFunc(file)
}

// getHeadersFromType extracts CSV headers from a struct type based on the `csv` tags.
func getHeadersFromType(t reflect.Type) []string {
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if csvTag := field.Tag.Get("csv"); csvTag != "" {
			headers = append(headers, csvTag)
		}
	}
	return headers
}

// detectCSVStructure reads the headers from the CSV file and determines the appropriate parsing function.
func detectCSVStructure(file io.Reader) (func(io.Reader) ([]*transaction.Transaction, error), error) {
	formats := []csvFormat{
		{
			recordType: reflect.TypeOf(parsers.N26Record{}),
			parseFunc:  parsers.ParseN26CSV,
		},
		{
			recordType: reflect.TypeOf(parsers.BankStatementRecord{}),
			parseFunc:  parsers.ParseBankStatementCSV,
		},
	}

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading headers: %w", err)
	}

	for _, format := range formats {
		expectedHeaders := getHeadersFromType(format.recordType)
		if equalHeaders(headers, expectedHeaders) {
			return format.parseFunc, nil
		}
	}

	return nil, fmt.Errorf("unknown CSV structure")
}

// equalHeaders compares two slices of strings to determine if they are equal, ignoring leading and trailing spaces.
func equalHeaders(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}
