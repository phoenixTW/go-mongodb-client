package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// New creates new instance of the MongoDB client
func New(ctx context.Context, dsn string, name string, logger *zap.Logger) *mongo.Client {
	clientOptions := options.Client().ApplyURI(dsn).SetAppName(name)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Fatal("failed to initiate a mongo client", zap.Error(err))
	}

	return client
}
