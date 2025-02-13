package trading

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/devinjacknz/devinsystem/internal/ai"
	"github.com/devinjacknz/devinsystem/internal/exchange"
	"github.com/devinjacknz/devinsystem/internal/monitoring"
	"github.com/devinjacknz/devinsystem/internal/risk"
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
	exchanges   []exchange.Exchange
	aiService   ai.Service
	monitor     *monitoring.Service
}

func NewTradingEngine(riskMgr risk.Manager, exchanges []exchange.Exchange, aiService ai.Service, monitor *monitoring.Service) *tradingEngine {
	return &tradingEngine{
		orderBooks: make(map[string]*OrderBook),
		riskMgr:    riskMgr,
		exchanges:  exchanges,
		aiService:  aiService,
		monitor:    monitor,
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

	// Find the appropriate exchange
	var selectedExchange exchange.Exchange
	for _, ex := range e.exchanges {
		if ex.Name() == order.Exchange {
			selectedExchange = ex
			break
		}
	}
	if selectedExchange == nil {
		return fmt.Errorf("exchange not found: %s", order.Exchange)
	}

	// Convert trading.Order to exchange.Order
	if err := selectedExchange.ExecuteOrder(exchange.Order{
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
	// Log startup
	e.monitor.LogSystem("Trading engine starting with multiple exchanges")
	
	// Initialize exchanges
	for _, ex := range e.exchanges {
		e.monitor.LogSystem(fmt.Sprintf("Initializing exchange: %s", ex.Name()))
	}
	
	// Start market data monitoring
	go e.monitorMarkets()
	
	// Keep the main thread running
	select {}
}

func (e *tradingEngine) monitorMarkets() {
	for _, ex := range e.exchanges {
		go func(exchange exchange.Exchange) {
			for {
				time.Sleep(5 * time.Second)
				
				if exchange.Name() == "Jupiter" {
					data, err := exchange.GetMarketData()
					if err != nil {
						e.monitor.LogError(fmt.Sprintf("Failed to get market data: %v", err))
						continue
					}

					for _, d := range data {
						e.monitor.LogJupiterSwap(d.Symbol, "USDC", d.Price, d.Volume, 0.1)
						
						aiData := ai.MarketData{
							Symbol: d.Symbol,
							Price:  d.Price,
							Volume: d.Volume,
							Trend:  "",
						}
						
						analysis, err := e.aiService.AnalyzeMarket(aiData)
						if err != nil {
							e.monitor.LogError(fmt.Sprintf("Failed to analyze market data for %s: %v", d.Symbol, err))
							continue
						}
						
						e.monitor.LogAISignal(d.Symbol, analysis.Trend, analysis.Confidence)
					}
				} else {
					data, err := exchange.GetMarketData()
					if err != nil {
						e.monitor.LogError(fmt.Sprintf("Failed to get market data from %s: %v", exchange.Name(), err))
						continue
					}

					for _, d := range data {
						aiData := ai.MarketData{
							Symbol: d.Symbol,
							Price:  d.Price,
							Volume: d.Volume,
							Trend:  "",
						}
						
						analysis, err := e.aiService.AnalyzeMarket(aiData)
						if err != nil {
							e.monitor.LogError(fmt.Sprintf("Failed to analyze market data: %v", err))
							continue
						}
						
						e.monitor.LogAISignal(d.Symbol, analysis.Trend, analysis.Confidence)
					}
				}
			}
		}(ex)
	}
}
