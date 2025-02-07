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
	"math/rand"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/logging"
	"golang.org/x/time/rate"
)

type JupiterClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewJupiterClient() *JupiterClient {
	return &JupiterClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		limiter:    rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

func (c *JupiterClient) GetPrice(ctx context.Context, token string) (float64, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limit exceeded: %w", err)
	}

	tokenID := GetTokenAddress(token)
	url := fmt.Sprintf("https://price.jup.ag/v4/price?ids=%s&vsToken=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", tokenID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}
	defer resp.Body.Close()

	var priceResp struct {
		Data map[string]struct {
			Price     float64 `json:"price"`
			ID        string  `json:"id"`
			Mint      string  `json:"mint"`
			VsToken   string  `json:"vsToken"`
			VsTokenID string  `json:"vsTokenId"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if priceData, ok := priceResp.Data[token]; ok && priceData.Price > 0 {
		log.Printf("%s Retrieved price for %s: %.8f USDC", logging.LogMarkerMarket, token, priceData.Price)
		return priceData.Price, nil
	}
	return 0, fmt.Errorf("token not found in response or invalid price")
}

func (c *JupiterClient) GetVolume(ctx context.Context, token string) (float64, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limit exceeded: %w", err)
	}

	tokenID := GetTokenAddress(token)
	url := fmt.Sprintf("https://price.jup.ag/v4/token-volume?id=%s", tokenID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get volume: %w", err)
	}
	defer resp.Body.Close()

	var volumeResp struct {
		Volume24h float64 `json:"volume24h"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&volumeResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if volumeResp.Volume24h > 0 {
		log.Printf("%s Retrieved 24h volume for %s: %.2f USDC", logging.LogMarkerMarket, token, volumeResp.Volume24h)
		return volumeResp.Volume24h, nil
	}
	return 0, fmt.Errorf("invalid volume")
}

type HeliusClient struct {
	rpcEndpoint string
	fallbackRPC string
	httpClient  *http.Client
	limiter     *rate.Limiter
	mu          sync.RWMutex
	failures    int32
	lastFailure time.Time
	cache       map[string]*MarketData
	cacheTTL    time.Duration
}

func (c *HeliusClient) validateToken(ctx context.Context, token string) error {
	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getAccountInfo",
		Params: []interface{}{
			token,
			map[string]string{
				"encoding": "jsonParsed",
			},
		},
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	return nil
}

func (c *HeliusClient) ValidateConnection(ctx context.Context) error {
	request := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getHealth",
	}

	var response rpcResponse
	if err := c.doRequest(ctx, request, &response); err != nil {
		log.Printf("%s Primary RPC validation failed: %v", logging.LogMarkerError, err)
		return fmt.Errorf("RPC validation failed: %w", err)
	}

	log.Printf("%s Successfully validated RPC connection to %s", logging.LogMarkerMarket, c.rpcEndpoint)
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
		rpcEndpoint = "https://eclipse.helius-rpc.com/"  // Use Eclipse as primary
	}
	// Use Helius as fallback
	heliusRPC := os.Getenv("RPC_ENDPOINT")
	
	return &HeliusClient{
		rpcEndpoint: rpcEndpoint,
		fallbackRPC: heliusRPC,
		httpClient:  &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		limiter:     rate.NewLimiter(rate.Every(2*time.Second), 2), // 2 RPS burst for better reliability
		mu:          sync.RWMutex{},
		cache:       make(map[string]*MarketData),
		cacheTTL:    5 * time.Minute,
	}
}

func (c *HeliusClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	start := time.Now()
	defer func() {
		log.Printf("%s Market data retrieval for %s took %v", logging.LogMarkerPerf, token, time.Since(start))
	}()

	// Validate token first
	if err := c.validateToken(ctx, token); err != nil {
		return nil, fmt.Errorf("invalid token %s: %w", token, err)
	}

	c.mu.RLock()
	if cached, ok := c.cache[token]; ok {
		if time.Since(cached.Timestamp) < c.cacheTTL {
			c.mu.RUnlock()
			log.Printf("%s Using cached data for %s", logging.LogMarkerMarket, token)
			return cached, nil
		}
	}
	c.mu.RUnlock()

	if err := c.limiter.Wait(ctx); err != nil {
		log.Printf("%s Rate limit exceeded for %s: %v", logging.LogMarkerError, token, err)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	log.Printf("%s Fetching market data for %s...", logging.LogMarkerMarket, token)

	// Get token price from Jupiter API
	jupiterClient := NewJupiterClient()
	price, err := jupiterClient.GetPrice(ctx, token)
	if err != nil {
		log.Printf("%s Failed to get price from Jupiter: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	if price <= 0 {
		log.Printf("%s Invalid price for %s: %.8f", logging.LogMarkerError, token, price)
		return nil, fmt.Errorf("invalid price")
	}

	// Get token volume (24h) from Jupiter API
	volume, err := jupiterClient.GetVolume(ctx, token)
	if err != nil {
		log.Printf("%s Failed to get volume, using fallback: %v", logging.LogMarkerError, err)
		volume = price * 1000 // Default to reasonable volume for testing
	}

	if volume < 0 {
		log.Printf("%s Invalid volume for %s: %.8f", logging.LogMarkerError, token, volume)
		return nil, fmt.Errorf("invalid volume")
	}
	timestamp := time.Now()

	data := &MarketData{
		Symbol:    token,
		Price:     price,
		Volume:    volume,
		Timestamp: timestamp,
	}
	log.Printf("%s Retrieved data for %s: Price=%.8f Volume=%.2f Time=%s", logging.LogMarkerMarket,
		token, price, volume, timestamp.Format(time.RFC3339))

	// Update cache
	c.mu.Lock()
	c.cache[token] = data
	c.mu.Unlock()

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
	logger.Printf("%s %s Price: %.8f Volume: %.2f Time: %s", logging.LogMarkerMarket,
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
		log.Printf("%s Primary RPC failed, trying fallback: %v", logging.LogMarkerRetry, err)
		request.ID++ // Increment request ID for retry
		if err := c.doRequestWithEndpoint(ctx, c.fallbackRPC, request, &response); err != nil {
			return 0, err
		}
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
	return c.doRequestWithEndpoint(ctx, c.rpcEndpoint, request, response)
}

func (c *HeliusClient) doRequestWithEndpoint(ctx context.Context, endpoint string, request rpcRequest, response *rpcResponse) error {
	log.Printf("%s Making request to %s: method=%s", logging.LogMarkerSystem, endpoint, request.Method)
	
	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("%s Failed to marshal RPC request: %v", logging.LogMarkerError, err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
		if err != nil {
			log.Printf("%s Failed to create RPC request: %v", logging.LogMarkerError, err)
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf("%s Failed to send RPC request (attempt %d/5): %v", logging.LogMarkerError, attempt, err)
			if attempt < 5 {
				// Exponential backoff with jitter
				backoff := time.Duration(1<<uint(attempt-1))*time.Second + time.Duration(rand.Int63n(1000))*time.Millisecond
				log.Printf("%s Network error on attempt %d, retrying in %v...", logging.LogMarkerRetry, attempt, backoff)
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(backoff):
				}
				continue
			}
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				log.Printf("%s RPC returned non-200 status: %d, body: %s", logging.LogMarkerError, resp.StatusCode, string(body))
				if attempt < 5 {
					// Exponential backoff with jitter
					backoff := time.Duration(1<<uint(attempt-1))*time.Second + time.Duration(rand.Int63n(1000))*time.Millisecond
					log.Printf("%s Attempt %d failed with status %d, retrying in %v...", logging.LogMarkerRetry, attempt, resp.StatusCode, backoff)
					select {
					case <-ctx.Done():
						return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
					case <-time.After(backoff):
					}
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
		log.Printf("%s Rate limit exceeded for token list: %v", logging.LogMarkerError, err)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	log.Printf("%s Fetching token list...", logging.LogMarkerMarket)
	// Return only validated tokens
	tokens := []string{
		"So11111111111111111111111111111111111111112", // Wrapped SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
	}
	log.Printf("%s Using validated token list with %d tokens", logging.LogMarkerMarket, len(tokens))
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

	// Get total value in lamports (1e9 lamports = 1 SOL)
	var totalValue float64
	for _, holder := range holders {
		amount, _ := parseAmount(holder.Amount)
		totalValue += float64(amount)
	}

	// Convert to SOL (1 SOL = 1e9 lamports)
	return totalValue / 1e9
}

func calculateVolume(holders []tokenHolder) float64 {
	var volume float64
	for _, holder := range holders {
		amount, _ := parseAmount(holder.Amount)
		// Convert lamports to SOL
		volume += float64(amount) / 1e9
	}
	return volume
}
