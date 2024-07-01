package full

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gacevicljubisa/swaplist/pkg/transaction"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetTransactions fetches transactions and sends them to a channel
func GetTransactions(ctx context.Context, address string) (<-chan transaction.Transaction, <-chan error) {
	transactionChan := make(chan transaction.Transaction, 10)
	errorChan := make(chan error)

	go func() {
		defer close(transactionChan)
		defer close(errorChan)

		// Connect to the Gnosis Chain endpoint
		client, err := ethclient.DialContext(ctx, "https://rpc.gnosischain.com")
		if err != nil {
			errorChan <- fmt.Errorf("failed to connect to the Ethereum client: %w", err)
			return
		}

		header, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve the latest block number: %w", err)
			return
		}

		contractAddress := common.HexToAddress(address)

		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
			FromBlock: big.NewInt(0),
			ToBlock:   header.Number,
		}

		logs, err := client.FilterLogs(ctx, query)
		if err != nil {
			errorChan <- fmt.Errorf("failed to retrieve logs: %w", err)
			return
		}

		httpClient := &http.Client{}

		for _, vLog := range logs {
			response, err := getTransactionByHash(ctx, httpClient, vLog.TxHash.Hex())
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve transaction: %w", err)
				return
			}

			block, err := client.BlockByHash(ctx, vLog.BlockHash)
			if err != nil {
				errorChan <- fmt.Errorf("failed to retrieve block: %w", err)
				return
			}

			transactionChan <- transaction.Transaction{
				From:      response,
				TimeStamp: strconv.FormatUint(block.Time(), 10),
			}
		}
	}()

	return transactionChan, errorChan
}

func getTransactionByHash(ctx context.Context, client *http.Client, transactionHash string) (string, error) {
	url := "https://nd-500-249-268.p2pify.com/512e720763b369ed620657f84d38d2af/"

	payload := fmt.Sprintf(`{"id":1,"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["%s"]}`, transactionHash)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create a new HTTP request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer res.Body.Close()

	var transactionResponse transactionResponse

	if err = json.NewDecoder(res.Body).Decode(&transactionResponse); err != nil {
		return "", fmt.Errorf("error decoding response body: %w", err)
	}

	return transactionResponse.Result.From, nil
}

type transactionResult struct {
	Hash             string `json:"hash"`
	Nonce            string `json:"nonce"`
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex string `json:"transactionIndex"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	GasPrice         string `json:"gasPrice"`
	Gas              string `json:"gas"`
	Input            string `json:"input"`
	Type             string `json:"type"`
	V                string `json:"v"`
	S                string `json:"s"`
	R                string `json:"r"`
}

type transactionResponse struct {
	Jsonrpc string            `json:"jsonrpc"`
	Result  transactionResult `json:"result"`
	Id      int               `json:"id"`
}
