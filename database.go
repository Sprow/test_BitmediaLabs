package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"test_BitmediaLabs/core/settings"
)

func connectToMongoDatabase(
	ctx context.Context,
	conf settings.MongoDBConfig,
) (*mongo.Database, error) {
	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(conf.ConnectionURL()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return mongoClient.Database(conf.DatabaseName), nil
}
