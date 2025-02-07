package trading

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/devinjacknz/devinsystem/internal/risk"
	"github.com/devinjacknz/devinsystem/internal/wallet"
	"github.com/devinjacknz/devinsystem/internal/exchange"
)

type MockRiskManager struct {
	mock.Mock
}

func (m *MockRiskManager) ValidateOrder(order risk.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockRiskManager) CheckExposure(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRiskManager) UpdateStopLoss(symbol string, currentPrice float64) error {
	args := m.Called(symbol, currentPrice)
	return args.Error(0)
}

type MockWalletManager struct {
	mock.Mock
}

func (m *MockWalletManager) CreateWallet(walletType wallet.WalletType) error {
	args := m.Called(walletType)
	return args.Error(0)
}

func (m *MockWalletManager) GetWallet(walletType wallet.WalletType) (*wallet.SolanaWallet, error) {
	args := m.Called(walletType)
	return args.Get(0).(*wallet.SolanaWallet), args.Error(1)
}

type MockExchange struct {
	mock.Mock
}

func (m *MockExchange) ExecuteOrder(order exchange.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func TestTradingEngine_PlaceOrder(t *testing.T) {
	mockRisk := new(MockRiskManager)
	mockWallet := new(MockWalletManager)
	engine := NewTradingEngine(mockRisk, mockWallet)

	tests := []struct {
		name      string
		order     Order
		setupMock func()
		wantErr   bool
	}{
		{
			name: "successful order placement",
			order: Order{
				ID:        "test-order",
				Symbol:    "SOL/USD",
				Side:      "buy",
				Amount:    1.0,
				Price:     100.0,
				OrderType: "limit",
				Exchange:  "solana",
			},
			setupMock: func() {
				mockRisk.On("ValidateOrder", mock.Anything).Return(nil)
				mockExchange := new(MockExchange)
				mockExchange.On("ExecuteOrder", mock.Anything).Return(nil)
				engine.exchangeMgr.exchanges["solana"] = mockExchange
			},
			wantErr: false,
		},
		{
			name: "risk validation failure",
			order: Order{
				ID:        "test-order",
				Symbol:    "SOL/USD",
				Side:      "buy",
				Amount:    1000.0,
				Price:     100.0,
				OrderType: "limit",
				Exchange:  "solana",
			},
			setupMock: func() {
				mockRisk.On("ValidateOrder", mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "invalid exchange",
			order: Order{
				ID:        "test-order",
				Symbol:    "SOL/USD",
				Side:      "buy",
				Amount:    1.0,
				Price:     100.0,
				OrderType: "limit",
				Exchange:  "invalid",
			},
			setupMock: func() {
				mockRisk.On("ValidateOrder", mock.Anything).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := engine.PlaceOrder(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingEngine_CancelOrder(t *testing.T) {
	mockRisk := new(MockRiskManager)
	mockWallet := new(MockWalletManager)
	engine := NewTradingEngine(mockRisk, mockWallet)

	// Setup test order book
	order := Order{
		ID:        "test-order",
		Symbol:    "SOL/USD",
		Side:      "buy",
		Amount:    1.0,
		Price:     100.0,
		OrderType: "limit",
	}
	engine.orderBooks["SOL/USD"] = NewOrderBook()
	engine.orderBooks["SOL/USD"].AddOrder(order)

	tests := []struct {
		name      string
		orderID   string
		symbol    string
		wantErr   bool
	}{
		{
			name:    "cancel existing order",
			orderID: "test-order",
			symbol:  "SOL/USD",
			wantErr: false,
		},
		{
			name:    "cancel non-existent order",
			orderID: "invalid-order",
			symbol:  "SOL/USD",
			wantErr: true,
		},
		{
			name:    "cancel order in non-existent market",
			orderID: "test-order",
			symbol:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.CancelOrder(tt.orderID, tt.symbol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTradingEngine_Start(t *testing.T) {
	mockRisk := new(MockRiskManager)
	mockWallet := new(MockWalletManager)
	engine := NewTradingEngine(mockRisk, mockWallet)

	err := engine.Start()
	assert.NoError(t, err)
}
