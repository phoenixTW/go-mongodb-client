package example

import (
	"context"
	"github.com/phoenixTW/go-mongodb-client/mongodb"
	"github.com/phoenixTW/go-mongodb-client/mongostorage"
	"go.uber.org/zap"
	"log"
	"time"
)

func example() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := mongodb.New(ctx, "mongodb://localhost:27017", "example", nil)

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	storage := mongostorage.New(client.Database("example-database"))
	retryingStorage := mongostorage.NewRetry(storage, logger)
	retryingStorage.GetDatabaseName()
}
