package filestore

import (
	"fmt"
	"os"

	"swaplist/pkg/transaction"
)

// SaveTransactions saves a slice of transactions to a file in plain text format
func SaveTransactions(transactions []transaction.Transaction, filePath string) error {
	// Open the file for writing (create if not exists, truncate if exists)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write transactions to file
	for _, transaction := range transactions {
		line := fmt.Sprintf("%s:%s\n", transaction.From, transaction.TimeStamp)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("error writing transaction to file: %w", err)
		}
	}

	return nil
}

// SaveTransactionsAsync writes transactions to a file as they are received from a channel
func SaveTransactionsAsync(transactionChan <-chan transaction.Transaction, filePath string, doneChan chan<- struct{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	for transaction := range transactionChan {
		line := fmt.Sprintf("%s:%s\n", transaction.From, transaction.TimeStamp)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("error writing transaction to file: %w", err)
		}
	}

	// Signal that writing is done
	doneChan <- struct{}{}

	return nil
}
