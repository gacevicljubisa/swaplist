package simulated

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

type Client struct {
	simulated.Client
	backend    *simulated.Backend
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
}

func New() *Client {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	c := &Client{
		privateKey: key,
	}

	c.chainID = big.NewInt(1337)

	auth, err := bind.NewKeyedTransactorWithChainID(key, c.chainID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Auth: ", auth.From.Hex())

	alloc := map[common.Address]types.Account{
		auth.From: {
			Balance: big.NewInt(1000000000000000000),
		},
	}

	backend := simulated.NewBackend(alloc, simulated.WithBlockGasLimit(4712388))

	c.Client = backend.Client()
	c.backend = backend

	return c
}

func (c *Client) Close() {
	c.backend.Close()
}

func (c *Client) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	// return c.Client.TransactionSender(ctx, tx, block, index)
	// TODO: check if this is correct, why is it missing from the original implementation?
	return common.Address{}, nil
}

func (c *Client) Commit() {
	c.backend.Commit()
}

func (c *Client) Key() *ecdsa.PrivateKey {
	return c.privateKey
}

func (c *Client) ChainID() *big.Int {
	return c.chainID
}
