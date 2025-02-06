package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	client *Client
}

func NewRepository(client *Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) SaveTrade(ctx context.Context, trade *Trade) error {
	collection := r.client.Collection("trades")
	_, err := collection.InsertOne(ctx, trade)
	return err
}

func (r *Repository) SaveMarketData(ctx context.Context, data *MarketData) error {
	collection := r.client.Collection("market_data")
	_, err := collection.InsertOne(ctx, data)
	return err
}

func (r *Repository) GetMarketData(ctx context.Context, token string, start, end time.Time) ([]*MarketData, error) {
	collection := r.client.Collection("market_data")
	
	filter := bson.M{
		"token": token,
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	
	opts := options.Find().SetSort(bson.D{{"timestamp", 1}})
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query market data: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*MarketData
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode market data: %w", err)
	}

	return results, nil
}

func (r *Repository) SaveRiskEvent(ctx context.Context, event *RiskEvent) error {
	collection := r.client.Collection("risk_events")
	_, err := collection.InsertOne(ctx, event)
	return err
}

func (r *Repository) SavePerformance(ctx context.Context, perf *Performance) error {
	collection := r.client.Collection("performance")
	_, err := collection.InsertOne(ctx, perf)
	return err
}
