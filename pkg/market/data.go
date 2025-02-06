package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HeliusClient struct {
	rpcEndpoint string
	httpClient  *http.Client
}

func NewHeliusClient(rpcEndpoint string) *HeliusClient {
	return &HeliusClient{
		rpcEndpoint: rpcEndpoint,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HeliusClient) GetMarketData(ctx context.Context, token string) (*MarketData, error) {
	// Market data collection will be implemented here
	return nil, nil
}

func (c *HeliusClient) GetTokenList(ctx context.Context) ([]string, error) {
	// Token list retrieval will be implemented here
	return nil, nil
}
