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
	"github.com/gacevicljubisa/swaplist/pkg/ethclient"
)

type Client struct {
	validate        *validator.Validate
	client          *ethclient.Client
	blockRangeLimit uint32
}

func NewClient(client *ethclient.Client, blockRangeLimit uint32) *Client {
	c := &Client{
		validate:        validator.New(),
		client:          client,
		blockRangeLimit: blockRangeLimit,
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

	var toBlock *big.Int
	var fromBlock *big.Int

	if tr.EndBlock != 0 {
		toBlock = big.NewInt(int64(tr.EndBlock))
	}

	if tr.StartBlock == 0 {
		fromBlock = &big.Int{}
		fromBlock = fromBlock.Sub(toBlock, big.NewInt(5))
	} else {
		fromBlock = big.NewInt(int64(tr.StartBlock))
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

		contractAddress := common.HexToAddress(tr.Address)

		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
			FromBlock: fromBlock,
			ToBlock:   toBlock,
		}

		fmt.Printf("querying logs for address %s, from block %d to block %d\n", tr.Address, query.FromBlock, query.ToBlock)

		logs, err := c.client.FilterLogs(ctx, query)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve logs: %w", err)
			return
		}

		for _, vLog := range logs {
			tx, isPending, err := c.client.TransactionByHash(ctx, vLog.TxHash)
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve transaction: %w", err)
				return
			}

			if isPending {
				continue
			}

			block, err := c.client.BlockByHash(ctx, vLog.BlockHash)
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

			sender, err := c.client.TransactionSender(ctx, tx, vLog.BlockHash, index)
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve sender: %w", err)
				return
			}

			transactionChan <- transaction.Transaction{
				From:      sender.Hex(),
				TimeStamp: strconv.FormatUint(block.Time(), 10),
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
