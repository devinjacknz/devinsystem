package exchange

import (
	"errors"
	"sync"
)

type SolanaDEX struct {
	mu      sync.RWMutex
	client  interface{}
	markets map[string]*Market
}

type Market struct {
	Symbol        string
	BaseDecimals  uint8
	QuoteDecimals uint8
	OrderBook     OrderBook
}

type OrderBook struct {
	Bids []PriceLevel
	Asks []PriceLevel
}

type PriceLevel struct {
	Price  float64
	Size   float64
	Orders int
}

func NewSolanaDEX() *SolanaDEX {
	return &SolanaDEX{
		markets: make(map[string]*Market),
	}
}

func (dex *SolanaDEX) AddMarket(symbol string, baseDecimals, quoteDecimals uint8) error {
	dex.mu.Lock()
	defer dex.mu.Unlock()

	if _, exists := dex.markets[symbol]; exists {
		return errors.New("market already exists")
	}

	dex.markets[symbol] = &Market{
		Symbol:        symbol,
		BaseDecimals:  baseDecimals,
		QuoteDecimals: quoteDecimals,
		OrderBook: OrderBook{
			Bids: make([]PriceLevel, 0),
			Asks: make([]PriceLevel, 0),
		},
	}
	return nil
}

func (dex *SolanaDEX) UpdateOrderBook(symbol string, bids, asks []PriceLevel) error {
	dex.mu.Lock()
	defer dex.mu.Unlock()

	market, exists := dex.markets[symbol]
	if !exists {
		return errors.New("market not found")
	}

	market.OrderBook.Bids = bids
	market.OrderBook.Asks = asks
	return nil
}

func (dex *SolanaDEX) GetMarketPrice(symbol string) (float64, error) {
	dex.mu.RLock()
	market, exists := dex.markets[symbol]
	dex.mu.RUnlock()

	if !exists {
		return 0, errors.New("market not found")
	}

	if len(market.OrderBook.Asks) == 0 {
		return 0, errors.New("no asks in orderbook")
	}

	return market.OrderBook.Asks[0].Price, nil
}

func (dex *SolanaDEX) ExecuteOrder(order Order) error {
	dex.mu.RLock()
	_, exists := dex.markets[order.Symbol]
	dex.mu.RUnlock()

	if !exists {
		return errors.New("market not found")
	}

	// In production, this would interact with Solana blockchain
	return nil
}
