package market

import (
	"context"
	"log"
	"time"

	"golang.org/x/time/rate"
)

type FallbackClient struct {
	primary   Client
	fallback  Client
	retries   int
	backoff   time.Duration
}

func (c *FallbackClient) ValidateConnection(ctx context.Context) error {
	if err := c.primary.ValidateConnection(ctx); err != nil {
		log.Printf("[ERROR] Primary RPC validation failed: %v", err)
		return c.fallback.ValidateConnection(ctx)
	}
	return nil
}

func NewFallbackClient(endpoint string) *FallbackClient {
	// Primary client with standard settings
	primary := NewHeliusClient(endpoint)
	primary.(*HeliusClient).httpClient.Timeout = 10 * time.Second

	// Fallback client with aggressive settings
	fallbackClient := NewHeliusClient("https://eclipse.helius-rpc.com/")
	if helius, ok := fallbackClient.(*HeliusClient); ok {
		helius.httpClient.Timeout = 8 * time.Second
		helius.limiter = rate.NewLimiter(rate.Every(1*time.Second), 2)
	}

	return &FallbackClient{
		primary:  primary,
		fallback: fallbackClient,
		retries:  3,
		backoff:  2 * time.Second,
	}
}

func (c *FallbackClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	// Try primary with standard settings
	data, err := c.primary.GetMarketData(ctx, token)
	if err == nil {
		log.Printf("[MARKET] Primary RPC successful for %s", token)
		return data, nil
	}
	log.Printf("[MARKET] Primary RPC failed for %s: %v, trying aggressive retry", token, err)

	// Try with aggressive retry strategy
	for i := 0; i < c.retries; i++ {
		data, err = c.fallback.GetMarketData(ctx, token)
		if err == nil {
			log.Printf("[MARKET] Aggressive retry successful for %s on attempt %d", token, i+1)
			return data, nil
		}
		if i < c.retries-1 {
			time.Sleep(c.backoff)
			log.Printf("[MARKET] Retry %d/%d for %s", i+2, c.retries, token)
		}
	}
	log.Printf("[ERROR] All RPC attempts failed for %s", token)
	return nil, err
}

func (c *FallbackClient) GetTokenList(ctx context.Context) ([]string, error) {
	tokens, err := c.primary.GetTokenList(ctx)
	if err == nil {
		return tokens, nil
	}
	log.Printf("[MARKET] Primary source failed for token list: %v", err)
	return c.fallback.GetTokenList(ctx)
}

func (c *FallbackClient) GetTopTokens(ctx context.Context) ([]Token, error) {
	tokens, err := c.primary.GetTopTokens(ctx)
	if err == nil {
		return tokens, nil
	}
	log.Printf("[MARKET] Primary source failed for top tokens: %v", err)
	return c.fallback.GetTopTokens(ctx)
}
