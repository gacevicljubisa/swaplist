package limit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gacevicljubisa/swaplist/pkg/transaction"
	"github.com/go-playground/validator/v10"
)

type transactionsResponse struct {
	Status  string                    `json:"status"`
	Message string                    `json:"message"`
	Result  []transaction.Transaction `json:"result"`
}

type Client struct {
	validate   *validator.Validate
	httpClient *http.Client
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		validate: validator.New(),
	}

	for _, option := range opts {
		option(c)
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}

	return c
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

type TransactionsRequest struct {
	Address    string `validate:"required"`
	Amount     uint32 `validate:"required,gte=1,lte=10000"`
	Order      string `validate:"required,oneof=asc desc"`
	StartBlock uint64
	EndBlock   uint64
	APIKey     string `validate:"required"`
}

func (c *Client) GetTransactions(ctx context.Context, tr *TransactionsRequest) ([]transaction.Transaction, error) {
	if err := c.validate.Struct(tr); err != nil {
		return nil, fmt.Errorf("error validating request: %w", err)
	}

	page := 0
	if tr.Amount != 10000 {
		page = 1
	}

	if tr.EndBlock == 0 {
		tr.EndBlock = 99999999
	}

	if tr.StartBlock > tr.EndBlock {
		return nil, fmt.Errorf("start block should be less than or equal to end block")
	}

	requestURL := fmt.Sprintf(`https://api.gnosisscan.io/api?module=account&action=txlist&address=%s&startblock=%v&endblock=%v&page=%v&offset=%v&sort=%s&apikey=%s`,
		tr.Address, tr.StartBlock, tr.EndBlock, page, tr.Amount, tr.Order, tr.APIKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	res, err := c.httpClient.Do(req)
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
