package trading

import (
	"context"
)

type Strategy interface {
	Name() string
	GenerateSignals(ctx context.Context, marketData interface{}) (*Signal, error)
}

type Signal struct {
	Token     string
	Direction string
	Strength  float64
	Metadata  map[string]interface{}
}

type BaseStrategy struct {
	name string
}

func (s *BaseStrategy) Name() string {
	return s.name
}
