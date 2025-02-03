package exchange

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPumpClient struct {
	mock.Mock
}

func (m *MockPumpClient) GetMarketPrice(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockPumpClient) PlaceOrder(order Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func TestPumpExchange_GetMarketPrice(t *testing.T) {
	mockClient := new(MockPumpClient)
	exchange := &PumpExchange{
		client: mockClient,
	}

	tests := []struct {
		name      string
		symbol    string
		mockPrice float64
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "get valid price",
			symbol:    "PEPE/USD",
			mockPrice: 0.00001234,
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "market not found",
			symbol:    "INVALID/USD",
			mockPrice: 0.0,
			mockErr:   ErrMarketNotFound,
			wantErr:   true,
		},
		{
			name:      "api error",
			symbol:    "DOGE/USD",
			mockPrice: 0.0,
			mockErr:   ErrAPIError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.On("GetMarketPrice", tt.symbol).Return(tt.mockPrice, tt.mockErr).Once()
			price, err := exchange.GetMarketPrice(tt.symbol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockPrice, price)
			}
		})
	}
}

func TestPumpExchange_ExecuteOrder(t *testing.T) {
	mockClient := new(MockPumpClient)
	exchange := &PumpExchange{
		client: mockClient,
	}

	tests := []struct {
		name    string
		order   Order
		mockErr error
		wantErr bool
	}{
		{
			name: "execute buy order",
			order: Order{
				ID:     "order1",
				Symbol: "PEPE/USD",
				Side:   "buy",
				Amount: 1000000,
				Price:  0.00001234,
				Type:   "limit",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "execute sell order",
			order: Order{
				ID:     "order2",
				Symbol: "PEPE/USD",
				Side:   "sell",
				Amount: 500000,
				Price:  0.00001345,
				Type:   "market",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "insufficient liquidity",
			order: Order{
				ID:     "order3",
				Symbol: "PEPE/USD",
				Side:   "buy",
				Amount: 100000000000,
				Price:  0.00001234,
				Type:   "limit",
			},
			mockErr: ErrInsufficientLiquidity,
			wantErr: true,
		},
		{
			name: "invalid order type",
			order: Order{
				ID:     "order4",
				Symbol: "PEPE/USD",
				Side:   "buy",
				Amount: 1000000,
				Price:  0.00001234,
				Type:   "invalid",
			},
			mockErr: ErrInvalidOrderType,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.On("PlaceOrder", tt.order).Return(tt.mockErr).Once()
			err := exchange.ExecuteOrder(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
