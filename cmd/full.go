package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gacevicljubisa/swaplist/pkg/filestore"
	"github.com/gacevicljubisa/swaplist/pkg/full"

	"github.com/spf13/cobra"
)

func (c *command) initFullCmd() (err error) {
	var address string

	cmd := &cobra.Command{
		Use:   "full",
		Short: "Retrieve transaction sender addresses with timestamps and extracts them from logs.",
		Long: `Retrieve a list of all transaction sender addresses with timestamps and extract them from logs.
	- Uses Chainstack API and Gnois RPC`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			log.Printf("Retrieving addresses for contract %s\n", address)

			ctx := cmd.Context()

			transactionChan, errorChan := full.GetTransactions(ctx, address)
			var wg sync.WaitGroup
			wg.Add(1)

			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			go func() {
				defer wg.Done()
				if err := filestore.SaveTransactionsAsync(ctx, transactionChan, "transactions.txt"); err != nil {
					if errors.Is(err, context.Canceled) {
						log.Fatalf("not all transactions have been saved: %v", err)
					}
					log.Fatalf("failed to save transactions: %v", err)
				}
			}()

			for {
				select {
				case err, ok := <-errorChan:
					if !ok {
						errorChan = nil
					} else {
						return fmt.Errorf("error retrieving transactions: %w", err)
					}
				case <-ticker.C:
					log.Println("processing...")
				case <-ctx.Done():
					log.Println("shutting down...")
					wg.Wait()
					return ctx.Err()
				}

				if errorChan == nil {
					break
				}
			}

			wg.Wait()
			log.Println("all transactions have been saved.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&address, "address", "a", "0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420", "Contract address on Gnosis Chain")

	c.root.AddCommand(cmd)

	return nil
}
