package ethclient

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/time/rate"
)

type Client struct {
	ethclient *ethclient.Client
	limiter   *rate.Limiter
	rawURL    string
	mu        sync.Mutex
}

type ClientOption func(*Client)

// WithRateLimit sets the rate limit for the Ethereum client.
func WithRateLimit(requestsPerSecond int) ClientOption {
	return func(c *Client) {
		c.limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond)
	}
}

// // WithRateLimit sets the rate limit for the Ethereum client.
// func WithRateLimit(limiter *rate.Limiter) ClientOption {
// 	return func(c *Client) {
// 		c.limiter = limiter
// 	}
// }

// NewClient creates a new Ethereum client with possible rate limiting.
func NewClient(ctx context.Context, rawURL string, opts ...ClientOption) (*Client, error) {
	ethclient, err := ethclient.DialContext(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		ethclient: ethclient,
		rawURL:    rawURL,
		limiter:   nil,
	}

	for _, option := range opts {
		option(c)
	}

	return c, nil
}

// Close closes the underlying Ethereum client.
func (c *Client) Close() {
	c.ethclient.Close()
}

func (c *Client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.applyRateLimit(ctx); err != nil {
		return nil, err
	}

	return c.ethclient.FilterLogs(ctx, q)
}

func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.applyRateLimit(ctx); err != nil {
		return nil, false, err
	}

	return c.ethclient.TransactionByHash(ctx, hash)
}

func (c *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.applyRateLimit(ctx); err != nil {
		return nil, err
	}

	return c.ethclient.BlockByHash(ctx, hash)
}

func (c *Client) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.applyRateLimit(ctx); err != nil {
		return common.Address{}, err
	}

	return c.ethclient.TransactionSender(ctx, tx, block, index)
}

// applyRateLimit checks if the limiter is set and applies the rate limit.
func (c *Client) applyRateLimit(ctx context.Context) error {
	if c.limiter != nil {
		return c.limiter.Wait(ctx)
	}
	return nil
}
