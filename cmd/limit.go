package cmd

import (
	"fmt"
	"log"

	"github.com/gacevicljubisa/swaplist/pkg/filestore"
	"github.com/gacevicljubisa/swaplist/pkg/limit"

	"github.com/spf13/cobra"
)

func (c *command) initLimitCmd() (err error) {
	var (
		address string
		amount  uint32
		order   string
		apikey  string
	)

	cmd := &cobra.Command{
		Use:   "limit",
		Short: "Retrieve transaction sender addresses with timestamps on Gnosis Chain",
		Long: `Retrieve a list of transaction sender addresses with timestamps on Gnosis Chain from the Gnosis Scan API.
	- Limited to a maximum of 10,000 addresses.
	- Can retrieve in ascending or descending order.
	- Can specify an API key.
	- Uses Gnosis Scan API.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			log.Printf("retrieving %d addresses for contract %s in %s order...\n", amount, address, order)

			response, err := limit.GetTransactions(cmd.Context(), address, amount, order, apikey)
			if err != nil {
				return fmt.Errorf("error retrieving transactions: %w", err)
			}

			log.Println("number of transactions retrieved:", len(response))

			log.Println("saving to file...")

			filestore.SaveTransactions(response, "transactions.txt")

			log.Println("done")

			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420", "Contract address on Gnosis Chain")
	cmd.Flags().Uint32VarP(&amount, "number", "n", 1000, "Number of addresses to retrieve (0-10000)")
	cmd.Flags().StringVarP(&order, "order", "o", "asc", "Order to retrieve addresses (asc/desc)")
	cmd.Flags().StringVarP(&apikey, "apikey", "k", "DEN397GUGXKJN6T14HU2W8MTZVVMXZ57AU", "API key for Gnosis Scan")

	c.root.AddCommand(cmd)

	return nil
}
