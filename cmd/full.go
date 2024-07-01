package cmd

import (
	"fmt"
	"log"

	"swaplist/pkg/filestore"
	"swaplist/pkg/full"

	"github.com/spf13/cobra"
)

func (c *command) initFullCmd() (err error) {
	var address string

	cmd := &cobra.Command{
		Use:   "full",
		Short: "Retrieve transaction sender addresses with timestamps",
		Long: `Retrieve a list of all transaction sender addresses with timestamps from the Gnosis Scan API.
	- Uses Chainstack`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Printf("Retrieving addresses for contract %s\n", address)

			transactionChan, errorChan := full.GetTransactions(address)
			doneChan := make(chan struct{})

			go func() {
				if err := filestore.SaveTransactionsAsync(transactionChan, "transactions.txt", doneChan); err != nil {
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
					fmt.Println("All transactions have been saved.")
					return
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
