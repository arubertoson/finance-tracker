package parsers

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/gocarina/gocsv"
)

// ParseCSV reads CSV data into a slice of structs using the csv tags
func ParseCSV[T any](reader io.Reader) ([]T, error) {
	var records []T

	// Set a custom CSV reader with specific configurations
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.LazyQuotes = true    // Handle quoted fields more flexibly
		r.Comma = ','          // Explicitly set comma as separator
		r.FieldsPerRecord = -1 // Allow variable number of fields
		return r
	})

	// Attempt to unmarshal the CSV data into the records slice
	if err := gocsv.Unmarshal(reader, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CSV: %w", err)
	}

	// Debug: print the number of records parsed
	fmt.Printf("Successfully parsed %d records\n", len(records))

	return records, nil
}
