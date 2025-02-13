package exchange

import (
	"errors"
	"sync"
)

type PumpFun struct {
	mu      sync.RWMutex
	client  interface{}
	markets map[string]*Market
	name    string
}

func (p *PumpFun) Name() string {
	return p.name
}

func (p *PumpFun) GetMarketData() ([]*MarketData, error) {
	return []*MarketData{}, nil
}

func NewPumpFun(apiURL string) *PumpFun {
	return &PumpFun{
		markets: make(map[string]*Market),
		name:    "Pump.fun",
	}
}

func (p *PumpFun) GetMarketPrice(symbol string) (float64, error) {
	p.mu.RLock()
	market, exists := p.markets[symbol]
	p.mu.RUnlock()

	if !exists {
		return 0, errors.New("market not found")
	}

	if len(market.OrderBook.Asks) == 0 {
		return 0, errors.New("no asks in orderbook")
	}

	return market.OrderBook.Asks[0].Price, nil
}

func (p *PumpFun) ExecuteOrder(order Order) error {
	p.mu.RLock()
	_, exists := p.markets[order.Symbol]
	p.mu.RUnlock()

	if !exists {
		return errors.New("market not found")
	}

	// In production, this would interact with Pump.fun API
	return nil
}
