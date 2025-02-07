package strategies

import (
	"context"
)

type Signal struct {
	Token     string                 `json:"token"`      // Token address
	Signal    float64                `json:"signal"`     // Signal strength (0-1)
	Direction string                 `json:"direction"`  // BUY, SELL, or NEUTRAL
	Metadata  map[string]interface{} `json:"metadata"`   // Optional strategy-specific data
}

type Strategy interface {
	GenerateSignals(ctx context.Context) (*Signal, error)
	Name() string
}

type BaseStrategy struct {
	name string
}

func NewBaseStrategy(name string) *BaseStrategy {
	return &BaseStrategy{
		name: name,
	}
}

func (b *BaseStrategy) Name() string {
	return b.name
}
