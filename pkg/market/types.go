package market

import (
	"context"
	"time"
)

type MarketData struct {
	Symbol    string
	Price     float64
	Volume    float64
	Timestamp time.Time
}

type Client interface {
	GetMarketData(ctx context.Context, token string) (*MarketData, error)
	GetTokenList(ctx context.Context) ([]string, error)
}
