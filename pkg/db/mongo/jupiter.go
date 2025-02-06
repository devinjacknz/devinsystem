package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JupiterTrade struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Token     string            `bson:"token"`
	Amount    float64           `bson:"amount"`
	Quote     float64           `bson:"quote"`
	Timestamp time.Time         `bson:"timestamp"`
	Status    string            `bson:"status"`
	TxHash    string           `bson:"tx_hash"`
}

func (r *Repository) SaveJupiterTrade(ctx context.Context, trade *JupiterTrade) error {
	collection := r.client.Collection("jupiter_trades")
	_, err := collection.InsertOne(ctx, trade)
	return err
}

func (r *Repository) GetJupiterTrades(ctx context.Context, token string) ([]*JupiterTrade, error) {
	collection := r.client.Collection("jupiter_trades")
	filter := bson.M{"token": token}
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trades []*JupiterTrade
	if err := cursor.All(ctx, &trades); err != nil {
		return nil, err
	}

	return trades, nil
}

func (r *Repository) CreateJupiterIndexes(ctx context.Context) error {
	collection := r.client.Collection("jupiter_trades")
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{"timestamp", -1}}},
		{Keys: bson.D{{"token", 1}, {"timestamp", -1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"tx_hash", 1}}, Options: options.Index().SetUnique(true)},
	})
	return err
}
