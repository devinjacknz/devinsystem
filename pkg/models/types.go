package models

import (
	"context"
)

// TradeDecision is defined in model.go

type Client interface {
	GenerateTradeDecision(ctx context.Context, data interface{}) (*TradeDecision, error)
}
