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
	var (
		address    string
		startBlock uint64
		endBlock   uint64
		endpoint   string
	)

	cmd := &cobra.Command{
		Use:   "full",
		Short: "Retrieve transaction sender addresses with timestamps and extracts them from logs.",
		Long: `Retrieve a list of all transaction sender addresses with timestamps and extract them from logs.
	- Saves the list of addresses to a file and can be interrupted at any time.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			log.Printf("Retrieving addresses for contract %s\n", address)

			ctx := cmd.Context()

			client := full.NewClient(endpoint)

			transactionChan, errorChan := client.GetTransactions(ctx, &full.TransactionsRequest{
				Address:    address,
				StartBlock: startBlock,
				EndBlock:   endBlock,
			})

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
	cmd.Flags().Uint64VarP(&startBlock, "start", "", 19475474, "Start block number")
	cmd.Flags().Uint64VarP(&endBlock, "end", "", 19475479, "End block number")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "https://wandering-evocative-gas.xdai.quiknode.pro/0f2525676e3ba76259ab3b72243f7f60334b0423/", "Ethereum RPC endpoint")

	c.root.AddCommand(cmd)

	return nil
}
