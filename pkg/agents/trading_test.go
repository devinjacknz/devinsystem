package agents

import (
	"context"
	"testing"
	"time"

	"github.com/devinjacknz/devinsystem/pkg/market"
	"github.com/devinjacknz/devinsystem/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMarketData struct {
	mock.Mock
}

func (m *MockMarketData) GetMarketData(ctx context.Context, token string) (*market.MarketData, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*market.MarketData), args.Error(1)
}

func (m *MockMarketData) GetTokenList(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockMarketData) GetTopTokens(ctx context.Context) ([]market.Token, error) {
	args := m.Called(ctx)
	return args.Get(0).([]market.Token), args.Error(1)
}

type MockAIModel struct {
	mock.Mock
}

func (m *MockAIModel) GenerateTradeDecision(ctx context.Context, data interface{}) (*models.TradeDecision, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*models.TradeDecision), args.Error(1)
}

func TestTradingAgent_Run(t *testing.T) {
	ctx := context.Background()
	marketData := new(MockMarketData)
	aiModel := new(MockAIModel)

	agent := NewTradingAgent(marketData, aiModel)

	// Setup test data
	testToken := market.Token{
		Symbol: "TEST",
		Price:  100.0,
	}
	marketData.On("GetTopTokens", ctx).Return([]market.Token{testToken}, nil)

	testMarketData := &market.MarketData{
		Symbol:    "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}
	marketData.On("GetMarketData", ctx, "TEST").Return(testMarketData, nil)

	decision := &models.TradeDecision{
		Action:     "BUY",
		Confidence: 0.8,
		Reasoning:  "Test decision",
		Model:      "test-model",
		Timestamp:  time.Now(),
	}
	aiModel.On("GenerateTradeDecision", ctx, testMarketData).Return(decision, nil)

	// Run agent
	err := agent.Run(ctx)
	assert.NoError(t, err)

	marketData.AssertExpectations(t)
	aiModel.AssertExpectations(t)
}

func TestTradingAgent_HandleMarketData(t *testing.T) {
	ctx := context.Background()
	marketData := new(MockMarketData)
	aiModel := new(MockAIModel)

	agent := NewTradingAgent(marketData, aiModel)

	testData := &market.MarketData{
		Symbol:    "TEST",
		Price:     100.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
	}

	decision := &models.TradeDecision{
		Action:     "BUY",
		Confidence: 0.8,
		Reasoning:  "Test decision",
		Model:      "test-model",
		Timestamp:  time.Now(),
	}
	aiModel.On("GenerateTradeDecision", ctx, testData).Return(decision, nil)

	err := agent.HandleMarketData(ctx, testData)
	assert.NoError(t, err)

	aiModel.AssertExpectations(t)
}
