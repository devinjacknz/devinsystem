package market

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/logging"
	"golang.org/x/time/rate"
)

type FallbackClient struct {
	primary  Client
	fallback Client
	retries  int
	backoff  time.Duration
}

func NewFallbackClient(endpoint string) Client {
	primary := &HeliusClient{
		rpcEndpoint: endpoint,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		limiter:     rate.NewLimiter(rate.Every(time.Second), 1),
	}

	fallback := &HeliusClient{
		rpcEndpoint: "https://eclipse.helius-rpc.com/",
		httpClient:  &http.Client{Timeout: 8 * time.Second},
		limiter:     rate.NewLimiter(rate.Every(time.Second), 2),
	}

	return &FallbackClient{
		primary:  primary,
		fallback: fallback,
		retries:  3,
		backoff:  2 * time.Second,
	}
}

func (c *FallbackClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	data, err := c.primary.GetMarketData(ctx, token)
	if err == nil {
		log.Printf("%s Primary RPC successful for %s", logging.LogMarkerMarket, token)
		return data, nil
	}
	log.Printf("%s Primary RPC failed for %s: %v, trying aggressive retry", logging.LogMarkerMarket, token, err)

	for i := 0; i < c.retries; i++ {
		data, err = c.fallback.GetMarketData(ctx, token)
		if err == nil {
			log.Printf("%s Aggressive retry successful for %s on attempt %d", logging.LogMarkerMarket, token, i+1)
			return data, nil
		}
		if i < c.retries-1 {
			time.Sleep(c.backoff)
			log.Printf("%s Retry %d/%d for %s", logging.LogMarkerRetry, i+2, c.retries, token)
		}
	}
	log.Printf("%s All RPC attempts failed for %s", logging.LogMarkerError, token)
	return nil, err
}

func (c *FallbackClient) GetTokenList(ctx context.Context) ([]string, error) {
	tokens, err := c.primary.GetTokenList(ctx)
	if err == nil {
		return tokens, nil
	}
	log.Printf("%s Primary source failed for token list: %v", logging.LogMarkerError, err)
	return c.fallback.GetTokenList(ctx)
}

func (c *FallbackClient) GetTopTokens(ctx context.Context) ([]string, error) {
	tokens, err := c.primary.GetTopTokens(ctx)
	if err == nil {
		return tokens, nil
	}
	log.Printf("%s Primary source failed for top tokens: %v", logging.LogMarkerError, err)
	return c.fallback.GetTopTokens(ctx)
}

func (c *FallbackClient) ValidateConnection(ctx context.Context) error {
	if err := c.primary.ValidateConnection(ctx); err != nil {
		log.Printf("%s Primary RPC validation failed: %v", logging.LogMarkerError, err)
		return c.fallback.ValidateConnection(ctx)
	}
	return nil
}
