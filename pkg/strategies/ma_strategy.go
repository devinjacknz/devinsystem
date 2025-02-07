package strategies

import (
	"context"
	"fmt"
)

type MAStrategy struct {
	*BaseStrategy
	fastMA  int
	slowMA  int
	market  MarketDataProvider
}

type MarketDataProvider interface {
	GetMarketData(ctx context.Context, token string) (*MarketData, error)
	GetTopTokens(ctx context.Context) ([]Token, error)
}

type MarketData struct {
	Close []float64
	Time  []int64
}

type Token struct {
	Symbol string
	Price  float64
}

func NewMAStrategy(market MarketDataProvider) *MAStrategy {
	return &MAStrategy{
		BaseStrategy: NewBaseStrategy("MA Crossover"),
		fastMA:      20,
		slowMA:      50,
		market:      market,
	}
}

func (s *MAStrategy) GenerateSignals(ctx context.Context) (*Signal, error) {
	tokens, err := s.market.GetTopTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top tokens: %w", err)
	}

	for _, token := range tokens {
		data, err := s.market.GetMarketData(ctx, token.Symbol)
		if err != nil {
			continue
		}

		if len(data.Close) < s.slowMA {
			continue
		}

		// Calculate moving averages
		fastMA := calculateMA(data.Close, s.fastMA)
		slowMA := calculateMA(data.Close, s.slowMA)

		if len(fastMA) < 2 || len(slowMA) < 2 {
			continue
		}

		// Get latest values
		currentFast := fastMA[len(fastMA)-1]
		currentSlow := slowMA[len(slowMA)-1]
		prevFast := fastMA[len(fastMA)-2]
		prevSlow := slowMA[len(slowMA)-2]

		signal := &Signal{
			Token:     token.Symbol,
			Signal:    0,
			Direction: "NEUTRAL",
			Metadata: map[string]interface{}{
				"strategy_type": "ma_crossover",
				"fast_ma":      currentFast,
				"slow_ma":      currentSlow,
				"current_price": data.Close[len(data.Close)-1],
			},
		}

		// Bullish crossover (fast crosses above slow)
		if prevFast <= prevSlow && currentFast > currentSlow {
			signal.Signal = 1.0
			signal.Direction = "BUY"
			return signal, nil
		}

		// Bearish crossover (fast crosses below slow)
		if prevFast >= prevSlow && currentFast < currentSlow {
			signal.Signal = 1.0
			signal.Direction = "SELL"
			return signal, nil
		}
	}

	return nil, nil
}

func calculateMA(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}

	ma := make([]float64, len(data)-period+1)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ma[0] = sum / float64(period)

	for i := period; i < len(data); i++ {
		sum = sum - data[i-period] + data[i]
		ma[i-period+1] = sum / float64(period)
	}

	return ma
}
