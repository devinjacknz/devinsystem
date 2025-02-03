package trading

import (
	"errors"
	"fmt"
	"sync"

	"github.com/devinjacknz/devintrade/internal/exchange"
	"github.com/devinjacknz/devintrade/internal/risk"
	"github.com/devinjacknz/devintrade/internal/wallet"
)

type Order struct {
	ID        string
	Symbol    string
	Side      string
	Amount    float64
	Price     float64
	OrderType string
	Exchange  string
}

type Engine interface {
	PlaceOrder(order Order) error
	CancelOrder(orderID string, symbol string) error
	Start() error
}

type tradingEngine struct {
	mu          sync.RWMutex
	orderBooks  map[string]*OrderBook
	riskMgr     risk.Manager
	walletMgr   wallet.Manager
	exchangeMgr *exchange.ExchangeManager
}

func NewTradingEngine(riskMgr risk.Manager, walletMgr wallet.Manager) *tradingEngine {
	return &tradingEngine{
		orderBooks:  make(map[string]*OrderBook),
		riskMgr:     riskMgr,
		walletMgr:   walletMgr,
		exchangeMgr: exchange.NewExchangeManager(),
	}
}

func (e *tradingEngine) PlaceOrder(order Order) error {
	riskOrder := risk.Order{
		Symbol:    order.Symbol,
		Side:      order.Side,
		Amount:    order.Amount,
		Price:     order.Price,
		OrderType: order.OrderType,
	}
	if err := e.riskMgr.ValidateOrder(riskOrder); err != nil {
		return fmt.Errorf("risk validation failed: %w", err)
	}

	exchangeImpl, err := e.exchangeMgr.GetExchange(order.Exchange)
	if err != nil {
		return fmt.Errorf("failed to get exchange: %w", err)
	}

	// Convert trading.Order to exchange.Order
	if err := exchangeImpl.ExecuteOrder(exchange.Order{
		Symbol:    order.Symbol,
		Side:      order.Side,
		Amount:    order.Amount,
		Price:     order.Price,
		OrderType: order.OrderType,
	}); err != nil {
		return fmt.Errorf("failed to execute order: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	orderBook, exists := e.orderBooks[order.Symbol]
	if !exists {
		orderBook = NewOrderBook()
		e.orderBooks[order.Symbol] = orderBook
	}

	return orderBook.AddOrder(order)
}

func (e *tradingEngine) CancelOrder(orderID string, symbol string) error {
	e.mu.RLock()
	orderBook, exists := e.orderBooks[symbol]
	e.mu.RUnlock()

	if !exists {
		return errors.New("market not found")
	}

	return orderBook.RemoveOrder(orderID)
}

func (e *tradingEngine) Start() error {
	return nil
}
