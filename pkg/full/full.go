package full

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/gacevicljubisa/swaplist/pkg/transaction"
	"github.com/go-playground/validator/v10"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	validate *validator.Validate
}

func NewClient() *Client {
	c := &Client{
		validate: validator.New(),
	}

	return c
}

type TransactionsRequest struct {
	Address    string `validate:"required"`
	StartBlock uint64
	EndBlock   uint64
}

// GetTransactions fetches transactions and sends them to a channel
func (c *Client) GetTransactions(ctx context.Context, tr *TransactionsRequest) (<-chan transaction.Transaction, <-chan error) {
	transactionChan := make(chan transaction.Transaction, 10)
	errorChan := make(chan error)

	if tr.EndBlock == 0 {
		tr.EndBlock = 99999999
	}

	if err := c.validateRequest(tr); err != nil {
		defer close(transactionChan)
		defer close(errorChan)
		errorChan <- fmt.Errorf("error validating request: %w", err)
		return nil, nil
	}

	go func() {
		defer close(transactionChan)
		defer close(errorChan)

		// Connect to the QuickNode endpoint
		client, err := ethclient.DialContext(ctx, "https://wandering-evocative-gas.xdai.quiknode.pro/0f2525676e3ba76259ab3b72243f7f60334b0423/")
		if err != nil {
			errorChan <- fmt.Errorf("failed to connect to the Ethereum client: %w", err)
			return
		}

		contractAddress := common.HexToAddress(tr.Address)

		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
			FromBlock: big.NewInt(19475474),
			ToBlock:   big.NewInt(19475479),
		}

		logs, err := client.FilterLogs(ctx, query)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve logs: %w", err)
			return
		}

		for i, vLog := range logs {
			fmt.Println(vLog.TxHash.Hex())

			tx, isPending, err := client.TransactionByHash(context.Background(), vLog.TxHash)
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve transaction: %w", err)
				return
			}

			if isPending {
				continue
			}

			block, err := client.BlockByHash(context.Background(), vLog.BlockHash)
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve block: %w", err)
				return
			}

			var index uint

			for idx, tr := range block.Transactions() {
				if tr.Hash() == tx.Hash() {
					index = uint(idx)
					break
				}
			}

			sender, err := client.TransactionSender(context.Background(), tx, vLog.BlockHash, index)
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve sender: %w", err)
				return
			}

			transactionChan <- transaction.Transaction{
				From:      sender.Hex(),
				TimeStamp: strconv.FormatUint(block.Time(), 10),
			}

			if i > 10 {
				break
			}
		}
	}()

	return transactionChan, errorChan
}

func (c *Client) validateRequest(tr *TransactionsRequest) error {
	if err := c.validate.Struct(tr); err != nil {
		return err
	}

	if tr.StartBlock > tr.EndBlock {
		return fmt.Errorf("start block should be less than or equal to end block")
	}

	return nil
}
