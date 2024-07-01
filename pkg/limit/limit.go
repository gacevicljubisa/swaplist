package limit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gacevicljubisa/swaplist/pkg/transaction"
)

var (
	ErrEmptyAddress = errors.New("address should not be empty")
	ErrZeroAmount   = errors.New("amount should be greater than 0")
	ErrMaxAmount    = errors.New("amount should not exceed 10000")
	ErrEmptyAPIKey  = errors.New("apiKey should not be empty")
	ErrInvalidOrder = errors.New("order should be either 'asc' or 'desc'")
)

type transactionsResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Result  []transaction.Transaction `json:"result"`
}

func GetTransactions(address string, amount uint32, order, apiKey string) ([]transaction.Transaction, error) {
	if err := validateInputs(address, amount, order, apiKey); err != nil {
		return nil, err
	}

	page := 0
	if amount != 10000 {
		page = 1
	}

	requestURL := fmt.Sprintf("https://api.gnosisscan.io/api?module=account&action=txlist&address=%s&page=%v&offset=%v&sort=%s&apikey=%s", address, page, amount, order, apiKey)

	fmt.Println(requestURL)

	res, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", res.Status)
	}

	var response transactionsResponse

	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return response.Result, nil
}

// validateInputs validates the input parameters
func validateInputs(address string, amount uint32, order, apiKey string) error {
	if address == "" {
		return ErrEmptyAddress
	}
	if amount == 0 {
		return ErrZeroAmount
	}
	if amount > 10000 {
		return ErrMaxAmount
	}
	if apiKey == "" {
		return ErrEmptyAPIKey
	}
	order = strings.ToLower(order)
	if order != "asc" && order != "desc" {
		return ErrInvalidOrder
	}
	return nil
}
