package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trade struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"`
	Token       string                `bson:"token"`
	Direction   string                `bson:"direction"`
	Amount      float64               `bson:"amount"`
	Price       float64               `bson:"price"`
	Timestamp   time.Time             `bson:"timestamp"`
	Confidence  float64               `bson:"confidence"`
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
	ID       primitive.ObjectID     `bson:"_id,omitempty"`
	Token    string                `bson:"token"`
	Start    time.Time             `bson:"start"`
	End      time.Time             `bson:"end"`
	Metrics  map[string]interface{} `bson:"metrics"`
}

type RiskEvent struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty"`
	Token     string                `bson:"token"`
	Type      string                `bson:"type"`
	Timestamp time.Time             `bson:"timestamp"`
	Details   map[string]interface{} `bson:"details"`
}
