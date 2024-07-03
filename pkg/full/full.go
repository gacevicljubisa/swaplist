package full

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/gacevicljubisa/swaplist/pkg/transaction"
	"github.com/go-playground/validator/v10"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gacevicljubisa/swaplist/pkg/ethclient"
)

type Client struct {
	validate        *validator.Validate
	client          *ethclient.Client
	lastBlock       *types.Block
	blockRangeLimit uint32
	mutex           sync.Mutex // To protect lastBlock
}

func NewClient(client *ethclient.Client, blockRangeLimit uint32) *Client {
	return &Client{
		validate:        validator.New(),
		client:          client,
		blockRangeLimit: blockRangeLimit,
	}
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

	if err := c.validateRequest(tr); err != nil {
		errorChan <- fmt.Errorf("error validating request: %w", err)
		close(transactionChan)
		close(errorChan)
		return transactionChan, errorChan
	}

	var toBlock, fromBlock *big.Int
	if tr.EndBlock != 0 {
		toBlock = big.NewInt(int64(tr.EndBlock))
	}
	if tr.StartBlock == 0 {
		fromBlock = new(big.Int).Sub(toBlock, big.NewInt(5))
	} else {
		fromBlock = big.NewInt(int64(tr.StartBlock))
	}

	go c.processTransactions(ctx, tr.Address, fromBlock, toBlock, transactionChan, errorChan)
	return transactionChan, errorChan
}

func (c *Client) processTransactions(ctx context.Context, address string, fromBlock, toBlock *big.Int, transactionChan chan transaction.Transaction, errorChan chan error) {
	defer close(transactionChan)
	defer close(errorChan)

	contractAddress := common.HexToAddress(address)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: fromBlock,
		ToBlock:   toBlock,
	}

	logsChan := make(chan types.Log, 10)
	go c.fetchLogs(ctx, query, logsChan, errorChan)

	for vLog := range logsChan {
		tx, isPending, err := c.client.TransactionByHash(ctx, vLog.TxHash)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve transaction: %w", err)
			return
		}
		if isPending {
			continue
		}

		block, err := c.getBlockByHash(ctx, vLog.BlockHash)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve block: %w", err)
			return
		}

		index, err := c.findTransactionIndex(block, tx)
		if err != nil {
			errorChan <- fmt.Errorf("failed to find transaction index: %w", err)
			return
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

func (c *Client) fetchLogs(ctx context.Context, query ethereum.FilterQuery, logsChan chan<- types.Log, errorChan chan error) {
	defer close(logsChan)

	maxBlocks := uint64(c.blockRangeLimit)
	startBlock := query.FromBlock.Uint64()
	endBlock := query.ToBlock.Uint64()

	for start := startBlock; start <= endBlock; start += maxBlocks {
		end := start + maxBlocks - 1
		if end > endBlock {
			end = endBlock
		}

		chunkQuery := ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(start),
			ToBlock:   new(big.Int).SetUint64(end),
			Addresses: query.Addresses,
			Topics:    query.Topics,
		}

		fmt.Printf("querying logs from block %d to block %d\n", chunkQuery.FromBlock.Uint64(), chunkQuery.ToBlock.Uint64())

		logs, err := c.client.FilterLogs(ctx, chunkQuery)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve logs: %w", err)
			return
		}

		for _, log := range logs {
			select {
			case logsChan <- log:
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			}
		}
	}
}

func (c *Client) getBlockByHash(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lastBlock != nil && c.lastBlock.Hash() == blockHash {
		return c.lastBlock, nil
	}

	block, err := c.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	c.lastBlock = block
	return block, nil
}

func (c *Client) findTransactionIndex(block *types.Block, tx *types.Transaction) (uint, error) {
	for idx, bTx := range block.Transactions() {
		if bTx.Hash() == tx.Hash() {
			return uint(idx), nil
		}
	}
	return 0, fmt.Errorf("transaction not found in block")
}
