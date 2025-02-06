package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ModelDecision struct {
	Name       string    `bson:"name"`
	Action     string    `bson:"action"`
	Confidence float64   `bson:"confidence"`
	Reasoning  string    `bson:"reasoning"`
}

type AIDecision struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"`
	Token       string                `bson:"token"`
	Timestamp   time.Time             `bson:"timestamp"`
	Models      []ModelDecision       `bson:"models"`
	FinalAction string                `bson:"final_action"`
	Confidence  float64               `bson:"confidence"`
}

type Trade struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"`
	Token       string                `bson:"token"`
	Direction   string                `bson:"direction"`
	Amount      float64               `bson:"amount"`
	Price       float64               `bson:"price"`
	Timestamp   time.Time             `bson:"timestamp"`
	Confidence  float64               `bson:"confidence"`
	AIDecision  *AIDecision           `bson:"ai_decision,omitempty"`
	RiskMetrics map[string]interface{} `bson:"risk_metrics"`
}

type MarketData struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty"`
	Token     string                `bson:"token"`
	Price     float64               `bson:"price"`
	Volume    float64               `bson:"volume"`
	Timestamp time.Time             `bson:"timestamp"`
	Metrics   map[string]interface{} `bson:"metrics"`
}

type Performance struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty"`
	Token      string                `bson:"token"`
	Start      time.Time             `bson:"start"`
	End        time.Time             `bson:"end"`
	AIAccuracy float64               `bson:"ai_accuracy"`
	Metrics    map[string]interface{} `bson:"metrics"`
}

type RiskEvent struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty"`
	Token      string                `bson:"token"`
	Type       string                `bson:"type"`
	Timestamp  time.Time             `bson:"timestamp"`
	AIDecision *AIDecision           `bson:"ai_decision,omitempty"`
	Details    map[string]interface{} `bson:"details"`
}
