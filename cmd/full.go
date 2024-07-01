package cmd

import (
	"log"
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
			doneChan := make(chan struct{})
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			go func() {
				if err := filestore.SaveTransactionsAsync(ctx, transactionChan, "transactions.txt", doneChan); err != nil {
					log.Fatalf("Failed to save transactions: %v", err)
				}
			}()

			for {
				select {
				case err, ok := <-errorChan:
					if !ok {
						errorChan = nil
					} else {
						log.Printf("Error: %v\n", err)
					}
				case <-doneChan:
					log.Println("All transactions have been saved.")
					return
				case <-ticker.C:
					log.Println("processing...")
				case <-ctx.Done():
					log.Println("Shutting down...")
					return ctx.Err()
				}

				if errorChan == nil {
					break
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&address, "address", "a", "0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420", "Contract address on Gnosis Chain")

	c.root.AddCommand(cmd)

	return nil
}
