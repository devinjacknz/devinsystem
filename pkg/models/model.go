package models

import (
	"context"
	"time"
)

type Model interface {
	GenerateTradeDecision(ctx context.Context, data interface{}) (*TradeDecision, error)
	IsAvailable() bool
	ModelType() string
}

type TradeDecision struct {
	Action     string    `json:"action"`      // BUY, SELL, or NOTHING
	Confidence float64   `json:"confidence"`  // 0.0 to 1.0
	Reasoning  string    `json:"reasoning"`   // Explanation for the decision
	Model      string    `json:"model"`       // Model used for decision
	Timestamp  time.Time `json:"timestamp"`   // Time of decision
}

type BaseModel struct {
	modelName string
	baseURL  string
}

func NewBaseModel(modelName, baseURL string) *BaseModel {
	return &BaseModel{
		modelName: modelName,
		baseURL:  baseURL,
	}
}

func (b *BaseModel) ModelName() string {
	return b.modelName
}
