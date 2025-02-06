package models

import (
	"context"
)

type TradeDecision struct {
	Action      string
	Confidence  float64
	Reasoning   string
	Metadata    map[string]interface{}
}

type Client interface {
	GenerateTradeDecision(ctx context.Context, data interface{}) (*TradeDecision, error)
}
