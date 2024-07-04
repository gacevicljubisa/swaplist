package full_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gacevicljubisa/swaplist/pkg/ethclient/simulated"
	"github.com/gacevicljubisa/swaplist/pkg/full"
)

func TestGetTransactions(t *testing.T) {
	t.Parallel()

	ethClient := simulated.New()
	toAddress := common.HexToAddress("0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420")
	value := big.NewInt(1000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(fmt.Errorf("error suggesting gas price: %w", err))
	}
	var data []byte
	tx := types.NewTransaction(0, toAddress, value, gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ethClient.ChainID()), ethClient.Key())
	if err != nil {
		t.Fatal(fmt.Errorf("error signing transaction 1: %w", err))
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		t.Fatal(fmt.Errorf("error sending transaction 1: %w", err))
	}

	ethClient.Commit()

	tx = types.NewTransaction(1, toAddress, value, gasLimit, gasPrice, data)

	signedTx, err = types.SignTx(tx, types.NewEIP155Signer(ethClient.ChainID()), ethClient.Key())
	if err != nil {
		t.Fatal(fmt.Errorf("error signing transaction 2: %w", err))
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		t.Fatal(fmt.Errorf("error sending transaction 2: %w", err))
	}

	tx = types.NewTransaction(2, toAddress, value, gasLimit, gasPrice, data)

	signedTx, err = types.SignTx(tx, types.NewEIP155Signer(ethClient.ChainID()), ethClient.Key())
	if err != nil {
		t.Fatal(fmt.Errorf("error signing transaction 3: %w", err))
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		t.Fatal(fmt.Errorf("error sending transaction 4: %w", err))
	}

	ethClient.Commit()

	c := full.NewClient(ethClient, 5)
	ctx := context.Background()
	tr := &full.TransactionsRequest{
		Address:    "0xc2d5a532cf69aa9a1378737d8ccdef884b6e7420",
		StartBlock: 1,
		EndBlock:   5,
	}
	txChan, errChan := c.GetTransactions(ctx, tr)
	for {
		select {
		case val, ok := <-txChan:
			if !ok {
				t.Log("transaction channel closed")
				return
			}
			t.Log(val)
		case val, ok := <-errChan:
			if !ok {
				t.Log("error channel closed")
				return
			}
			t.Error(val)
		case <-ctx.Done():
			t.Log("context done")
			return
		}
	}
}
