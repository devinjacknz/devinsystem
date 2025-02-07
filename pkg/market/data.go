package market

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/logging"
	"golang.org/x/time/rate"
)

type HeliusClient struct {
	rpcEndpoint string
	httpClient  *http.Client
	limiter     *rate.Limiter
}

func NewHeliusClient(rpcEndpoint string) *HeliusClient {
	if rpcEndpoint == "" {
		rpcEndpoint = "https://eclipse.helius-rpc.com/"
	}
	return &HeliusClient{
		rpcEndpoint: rpcEndpoint,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		limiter:     rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

func (c *HeliusClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// For SOL token, use getBalance instead of getTokenSupply
	var request rpcRequest
	if token == "So11111111111111111111111111111111111111112" {
		// For SOL token, use getAccountInfo
		request = rpcRequest{
			Jsonrpc: "2.0",
			ID:      1,
			Method:  "getAccountInfo",
			Params:  []interface{}{os.Getenv("WALLET"), map[string]interface{}{
				"encoding": "jsonParsed",
			}},
		}
	} else {
		request = rpcRequest{
			Jsonrpc: "2.0",
			ID:      1,
			Method:  "getTokenSupply",
			Params:  []interface{}{GetTokenAddress(token)},
		}
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return nil, fmt.Errorf("failed to get token supply: %w", err)
	}

	var price float64
	if token == "So11111111111111111111111111111111111111112" {
		var account struct {
			Value struct {
				Lamports uint64 `json:"lamports"`
			} `json:"value"`
		}
		if err := json.Unmarshal(response.Result, &account); err != nil {
			return nil, fmt.Errorf("failed to unmarshal account: %w", err)
		}
		price = float64(account.Value.Lamports) / 1000000000 // Convert lamports to SOL
		if price <= 0 {
			log.Printf("%s Invalid SOL price from account: %.8f", logging.LogMarkerError, price)
			price = 100.0 // Default SOL price in USD
		}
	} else {
		var supply tokenAccountBalance
		if err := json.Unmarshal(response.Result, &supply); err != nil {
			return nil, fmt.Errorf("failed to unmarshal supply: %w", err)
		}
		amount, err := parseAmount(supply.Value.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}
		price = float64(amount) / 1e9
	}

	holders, err := c.getLargestTokenHolders(ctx, token)
	if err != nil {
		log.Printf("%s Failed to get token holders: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to get token holders: %w", err)
	}

	var volume float64
	for _, holder := range holders {
		amount, _ := parseAmount(holder.Amount)
		volume += float64(amount) / 1e9
	}

	data := &MarketData{
		Symbol:    token,
		Price:     price,
		Volume:    volume,
		Timestamp: time.Now(),
	}

	log.Printf("%s Retrieved data for %s: price=%.8f volume=%.2f", logging.LogMarkerMarket, token, price, volume)
	return data, nil
}

func (c *HeliusClient) ValidateConnection(ctx context.Context) error {
	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getHealth",
		Params:  []interface{}{},
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return fmt.Errorf("failed to validate connection: %w", err)
	}

	return nil
}

func (c *HeliusClient) GetTokenList(ctx context.Context) ([]string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		log.Printf("%s Rate limit exceeded for token list: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	log.Printf("%s Fetching token list...", logging.LogMarkerMarket)
	tokens := []string{
		"So11111111111111111111111111111111111111112", // Wrapped SOL
	}
	log.Printf("%s Using validated token list with %d tokens", logging.LogMarkerMarket, len(tokens))
	return tokens, nil
}

func (c *HeliusClient) GetTopTokens(ctx context.Context) ([]string, error) {
	return c.GetTokenList(ctx)
}

func (c *HeliusClient) getLargestTokenHolders(ctx context.Context, token string) ([]tokenHolder, error) {
	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getTokenLargestAccounts",
		Params:  []interface{}{GetTokenAddress(token)},
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return nil, fmt.Errorf("failed to get token holders: %w", err)
	}

	var holders struct {
		Value []tokenHolder `json:"value"`
	}
	if err := json.Unmarshal(response.Result, &holders); err != nil {
		return nil, fmt.Errorf("failed to unmarshal holders: %w", err)
	}

	return holders.Value, nil
}

func (c *HeliusClient) doRequest(ctx context.Context, request rpcRequest, response *rpcResponse) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.rpcEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return nil
}

type rpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError      `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type tokenAccountBalance struct {
	Value struct {
		Amount   string `json:"amount"`
		Decimals int    `json:"decimals"`
	} `json:"value"`
}

type tokenHolder struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

func parseAmount(amount string) (int64, error) {
	var value int64
	if _, err := fmt.Sscanf(amount, "%d", &value); err != nil {
		return 0, fmt.Errorf("failed to parse amount: %w", err)
	}
	return value, nil
}
