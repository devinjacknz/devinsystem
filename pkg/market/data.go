package market

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/logging"
	"golang.org/x/time/rate"
)

type HeliusClient struct {
	rpcEndpoint string
	httpClient  *http.Client
	limiter     *rate.Limiter
	mu          sync.RWMutex
	failures    int32
	lastFailure time.Time
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
	Context struct {
		Slot uint64 `json:"slot"`
	} `json:"context"`
	Value struct {
		Amount   string `json:"amount"`
		Decimals uint8  `json:"decimals"`
	} `json:"value"`
}

func NewHeliusClient(rpcEndpoint string) Client {
	if rpcEndpoint == "" {
		rpcEndpoint = os.Getenv("RPC_ENDPOINT")
	}
	return &HeliusClient{
		rpcEndpoint: rpcEndpoint,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		limiter:     rate.NewLimiter(rate.Every(time.Second), 1), // 1 RPS for free plan
	}
}

func (c *HeliusClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	start := time.Now()
	defer func() {
		log.Printf("%s Market data retrieval for %s took %v", logging.LogMarkerPerf, token, time.Since(start))
	}()

	if err := c.limiter.Wait(ctx); err != nil {
		log.Printf("%s Rate limit exceeded for %s: %v", logging.LogMarkerError, token, err)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	log.Printf("%s Fetching market data for %s...", logging.LogMarkerMarket, token)

	// Get token supply
	supply, err := c.getTokenSupply(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get token supply: %w", err)
	}

	// Get largest token holders
	holders, err := c.getLargestTokenHolders(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get token holders: %w", err)
	}

	// Calculate volume from holder movements
	volume := calculateVolume(holders)
	price := calculatePrice(supply, holders)
	timestamp := time.Now()

	data := &MarketData{
		Symbol:    token,
		Price:     price,
		Volume:    volume,
		Timestamp: timestamp,
	}
	log.Printf("%s Retrieved data for %s: Price=%.8f Volume=%.2f Time=%s", logging.LogMarkerMarket,
		token, price, volume, timestamp.Format(time.RFC3339))

	// Save market data
	if err := c.SaveMarketData(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to save market data: %w", err)
	}

	return data, nil
}

func (c *HeliusClient) SaveMarketData(ctx context.Context, data *MarketData) error {
	// Use file logging for market data
	f, err := os.OpenFile("/home/ubuntu/repos/devinsystem/trading.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Printf("%s %s Price: %.8f Volume: %.2f Time: %s", utils.LogMarkerMarket,
		data.Symbol, data.Price, data.Volume, data.Timestamp.Format(time.RFC3339))
	return nil
}

func (c *HeliusClient) getTokenSupply(ctx context.Context, token string) (uint64, error) {
	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getTokenSupply",
		Params:  []interface{}{token},
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return 0, err
	}

	var supply tokenAccountBalance
	if err := json.Unmarshal(response.Result, &supply); err != nil {
		return 0, fmt.Errorf("failed to unmarshal supply: %w", err)
	}

	return parseAmount(supply.Value.Amount)
}

func (c *HeliusClient) getLargestTokenHolders(ctx context.Context, token string) ([]tokenHolder, error) {
	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getTokenLargestAccounts",
		Params:  []interface{}{token},
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return nil, err
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
	log.Printf("%s Making request: method=%s", logging.LogMarkerSystem, request.Method)
	
	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("%s Failed to marshal RPC request: %v", logging.LogMarkerError, err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", c.rpcEndpoint, bytes.NewReader(body))
		if err != nil {
			log.Printf("%s Failed to create RPC request: %v", logging.LogMarkerError, err)
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf("%s Failed to send RPC request (attempt %d/3): %v", logging.LogMarkerError, attempt, err)
			if attempt < 3 {
				time.Sleep(time.Second)
				continue
			}
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("%s RPC returned non-200 status: %d, body: %s", logging.LogMarkerError, resp.StatusCode, string(body))
			if attempt < 3 {
				time.Sleep(time.Second)
				continue
			}
			return fmt.Errorf("RPC returned status %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			log.Printf("%s Failed to decode RPC response: %v", logging.LogMarkerError, err)
			return fmt.Errorf("failed to decode response: %w", err)
		}

		if response.Error != nil {
			log.Printf("%s RPC error: %s", logging.LogMarkerError, response.Error.Message)
			return fmt.Errorf("RPC error: %s", response.Error.Message)
		}

		log.Printf("%s Request successful: method=%s", logging.LogMarkerSystem, request.Method)
		return nil
	}
	return fmt.Errorf("all retry attempts failed")
}

func (c *HeliusClient) GetTopTokens(ctx context.Context) ([]Token, error) {
	tokens, err := c.GetTokenList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token list: %w", err)
	}

	var result []Token
	for _, symbol := range tokens[:30] { // Get top 30 tokens
		data, err := c.GetMarketData(ctx, symbol)
		if err != nil {
			continue
		}
		result = append(result, Token{
			Symbol: symbol,
			Price:  data.Price,
		})
	}
	return result, nil
}

func (c *HeliusClient) GetTokenList(ctx context.Context) ([]string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getProgramAccounts",
		Params: []interface{}{
			"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
			map[string]interface{}{
				"encoding": "jsonParsed",
				"filters": []map[string]interface{}{
					{
						"dataSize": 165,
					},
				},
			},
		},
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	var accounts []struct {
		Account struct {
			Data struct {
				Parsed struct {
					Info struct {
						Mint string `json:"mint"`
					} `json:"info"`
				} `json:"parsed"`
			} `json:"data"`
		} `json:"account"`
	}

	if err := json.Unmarshal(response.Result, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	tokens := make([]string, 0, len(accounts))
	seen := make(map[string]bool)

	for _, acc := range accounts {
		mint := acc.Account.Data.Parsed.Info.Mint
		if !seen[mint] {
			tokens = append(tokens, mint)
			seen[mint] = true
		}
	}

	return tokens, nil
}

type tokenHolder struct {
	Address  string `json:"address"`
	Amount   string `json:"amount"`
	Decimals uint8  `json:"decimals"`
}

func parseAmount(amount string) (uint64, error) {
	var value uint64
	if _, err := fmt.Sscanf(amount, "%d", &value); err != nil {
		return 0, fmt.Errorf("failed to parse amount: %w", err)
	}
	return value, nil
}

func calculatePrice(supply uint64, holders []tokenHolder) float64 {
	if len(holders) == 0 || supply == 0 {
		return 0
	}

	// Use largest holder's amount as reference
	largestHolder := holders[0]
	amount, _ := parseAmount(largestHolder.Amount)
	
	// Simple price calculation based on supply and largest holder
	return float64(amount) / float64(supply)
}

func calculateVolume(holders []tokenHolder) float64 {
	var volume float64
	for _, holder := range holders {
		amount, _ := parseAmount(holder.Amount)
		volume += float64(amount)
	}
	return volume
}
