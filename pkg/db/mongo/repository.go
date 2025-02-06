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

func (r *Repository) CreateIndexes(ctx context.Context) error {
	// Market data indexes
	_, err := r.client.Collection("market_data").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{"timestamp", -1}}},
		{Keys: bson.D{{"token", 1}, {"timestamp", -1}}},
		{Keys: bson.D{{"volume", -1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create market data indexes: %w", err)
	}

	// AI decisions indexes
	_, err = r.client.Collection("ai_decisions").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{"timestamp", -1}}},
		{Keys: bson.D{{"token", 1}, {"final_action", 1}}},
		{Keys: bson.D{{"confidence", -1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create AI decisions indexes: %w", err)
	}

	// Performance indexes
	_, err = r.client.Collection("performance").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{"token", 1}, {"start", -1}}},
		{Keys: bson.D{{"ai_accuracy", -1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create performance indexes: %w", err)
	}

	// Risk events indexes
	_, err = r.client.Collection("risk_events").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{"timestamp", -1}}},
		{Keys: bson.D{{"token", 1}, {"type", 1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create risk events indexes: %w", err)
	}

	return nil
}
