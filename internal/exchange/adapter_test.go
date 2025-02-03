package exchange

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExchangeAdapter struct {
	mock.Mock
}

func (m *MockExchangeAdapter) GetMarketPrice(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockExchangeAdapter) ExecuteOrder(order Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func TestExchangeAdapter_Interface(t *testing.T) {
	var _ ExchangeAdapter = (*SolanaDEX)(nil)
	var _ ExchangeAdapter = (*PumpExchange)(nil)
	var _ ExchangeAdapter = (*MockExchangeAdapter)(nil)
}

func TestExchangeManager_RegisterExchange(t *testing.T) {
	manager := NewExchangeManager()
	mockAdapter := new(MockExchangeAdapter)

	tests := []struct {
		name      string
		exchange  string
		adapter   ExchangeAdapter
		wantErr   bool
	}{
		{
			name:     "register new exchange",
			exchange: "solana",
			adapter:  mockAdapter,
			wantErr:  false,
		},
		{
			name:     "register duplicate exchange",
			exchange: "solana",
			adapter:  mockAdapter,
			wantErr:  true,
		},
		{
			name:     "register with empty name",
			exchange: "",
			adapter:  mockAdapter,
			wantErr:  true,
		},
		{
			name:     "register nil adapter",
			exchange: "test",
			adapter:  nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RegisterExchange(tt.exchange, tt.adapter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				adapter, exists := manager.exchanges[tt.exchange]
				assert.True(t, exists)
				assert.Equal(t, tt.adapter, adapter)
			}
		})
	}
}

func TestExchangeManager_GetExchange(t *testing.T) {
	manager := NewExchangeManager()
	mockAdapter := new(MockExchangeAdapter)
	manager.RegisterExchange("solana", mockAdapter)

	tests := []struct {
		name      string
		exchange  string
		wantErr   bool
	}{
		{
			name:     "get registered exchange",
			exchange: "solana",
			wantErr:  false,
		},
		{
			name:     "get unregistered exchange",
			exchange: "unknown",
			wantErr:  true,
		},
		{
			name:     "get with empty name",
			exchange: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := manager.GetExchange(tt.exchange)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, adapter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, adapter)
				assert.Equal(t, mockAdapter, adapter)
			}
		})
	}
}

func TestExchangeManager_ExecuteOrder(t *testing.T) {
	manager := NewExchangeManager()
	mockAdapter := new(MockExchangeAdapter)
	manager.RegisterExchange("solana", mockAdapter)

	order := Order{
		ID:     "test-order",
		Symbol: "SOL/USD",
		Side:   "buy",
		Amount: 1.0,
		Price:  100.0,
		Type:   "limit",
	}

	tests := []struct {
		name      string
		exchange  string
		order     Order
		mockErr   error
		wantErr   bool
	}{
		{
			name:     "execute order on registered exchange",
			exchange: "solana",
			order:    order,
			mockErr:  nil,
			wantErr:  false,
		},
		{
			name:     "execute order on unregistered exchange",
			exchange: "unknown",
			order:    order,
			mockErr:  nil,
			wantErr:  true,
		},
		{
			name:     "execute order with exchange error",
			exchange: "solana",
			order:    order,
			mockErr:  ErrInsufficientLiquidity,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exchange == "solana" {
				mockAdapter.On("ExecuteOrder", tt.order).Return(tt.mockErr).Once()
			}
			
			err := manager.ExecuteOrder(tt.exchange, tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockAdapter.AssertExpectations(t)
		})
	}
}
