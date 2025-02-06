package trading

import (
	"context"
	"testing"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/devinjacknz/devinsystem/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMarketClient struct {
	mock.Mock
}

func (m *MockMarketClient) GetMarketData(ctx context.Context, token string) (*market.MarketData, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*market.MarketData), args.Error(1)
}

func (m *MockMarketClient) GetTokenList(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

type MockOllamaClient struct {
	mock.Mock
}

func (m *MockOllamaClient) GenerateTradeDecision(ctx context.Context, data *market.MarketData) (*models.TradeDecision, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*models.TradeDecision), args.Error(1)
}

func TestTradeExecution(t *testing.T) {
	ctx := context.Background()
	marketClient := new(MockMarketClient)
	ollamaClient := new(MockOllamaClient)
	riskMgr := NewRiskManager()
	tokenCache := utils.NewTokenCache(time.Hour, 30, nil)

	engine := NewEngine(marketClient, ollamaClient, riskMgr, tokenCache)

	marketData := &market.MarketData{
		Symbol:    "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	decision := &models.TradeDecision{
		Action:     "BUY",
		Confidence: 0.8,
		Reasoning:  "Test reasoning",
	}

	marketClient.On("GetMarketData", ctx, "TEST").Return(marketData, nil)
	ollamaClient.On("GenerateTradeDecision", ctx, marketData).Return(decision, nil)

	err := engine.ExecuteTrade(ctx, "TEST", 1.0)
	assert.NoError(t, err)

	marketClient.AssertExpectations(t)
	ollamaClient.AssertExpectations(t)
}

func TestRiskValidation(t *testing.T) {
	ctx := context.Background()
	marketClient := new(MockMarketClient)
	ollamaClient := new(MockOllamaClient)
	riskMgr := NewRiskManager()
	tokenCache := utils.NewTokenCache(time.Hour, 30, nil)

	engine := NewEngine(marketClient, ollamaClient, riskMgr, tokenCache)

	marketData := &market.MarketData{
		Symbol:    "TEST",
		Price:     1_000_000.0, // High price to trigger risk limit
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	decision := &models.TradeDecision{
		Action:     "BUY",
		Confidence: 0.8,
		Reasoning:  "Test reasoning",
	}

	marketClient.On("GetMarketData", ctx, "TEST").Return(marketData, nil)
	ollamaClient.On("GenerateTradeDecision", ctx, marketData).Return(decision, nil)

	err := engine.ExecuteTrade(ctx, "TEST", 10.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trade would exceed max exposure")

	marketClient.AssertExpectations(t)
	ollamaClient.AssertExpectations(t)
}
