package filestore

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/gacevicljubisa/swaplist/pkg/transaction"
)

// SaveTransactions saves a slice of transactions to a file in plain text format
func SaveTransactions(transactions []transaction.Transaction, filePath string) error {
	// Open the file for writing (create if not exists, truncate if exists)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, transaction := range transactions {
		line := fmt.Sprintf("%s:%s\n", transaction.From, transaction.TimeStamp)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("error writing transaction to file: %w", err)
		}
	}

	return nil
}

// SaveTransactionsAsync writes transactions to a file as they are received from a channel
func SaveTransactionsAsync(ctx context.Context, transactionChan <-chan transaction.Transaction, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case transaction, ok := <-transactionChan:
			if !ok {
				return nil
			}

			line := fmt.Sprintf("%s:%s\n", transaction.From, transaction.TimeStamp)
			if _, err := file.WriteString(line); err != nil {
				return fmt.Errorf("error writing transaction to file: %w", err)
			}
		}
	}
}
