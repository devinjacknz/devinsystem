package market

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type HeliusClient struct {
	endpoint   string
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewHeliusClient(endpoint string) Client {
	return &HeliusClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
		limiter: rate.NewLimiter(rate.Every(1*time.Second), 2),
	}
}

func (c *HeliusClient) ValidateConnection(ctx context.Context) error {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getHealth",
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Result string `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Result != "ok" {
		return fmt.Errorf("unhealthy RPC endpoint: %s", result.Result)
	}

	log.Printf("[MARKET] Successfully validated RPC connection to %s", c.endpoint)
	return nil
}

func (c *HeliusClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenSupply",
		"params":  []interface{}{token},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Result struct {
			Amount float64 `json:"amount"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &MarketData{
		Symbol:    token,
		Price:     result.Result.Amount,
		Volume:    0,
		Timestamp: time.Now(),
	}, nil
}

func (c *HeliusClient) GetTokenList(ctx context.Context) ([]string, error) {
	return []string{
		"So11111111111111111111111111111111111111112", // Wrapped SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT
	}, nil
}

func (c *HeliusClient) GetTopTokens(ctx context.Context) ([]Token, error) {
	tokens := []Token{
		{Symbol: "So11111111111111111111111111111111111111112", Price: 0},
		{Symbol: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", Price: 0},
		{Symbol: "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", Price: 0},
	}
	return tokens, nil
}
