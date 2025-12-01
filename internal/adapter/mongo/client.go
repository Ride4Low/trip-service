package mongo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ride4Low/contracts/env"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TripsCollection     = "trips"
	RideFaresCollection = "ride_fares"
)

type MongoConfig struct {
	URI      string
	Database string
}

func NewMongoDefaultConfig() *MongoConfig {
	return &MongoConfig{
		URI:      env.GetString("MONGODB_URI", ""),
		Database: env.GetString("MONGODB_DATABASE", ""),
	}
}

func NewMongoClient(cfg *MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("mongodb database is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	err = CreateTTLIndex(ctx, GetDatabase(client, cfg.Database))
	if err != nil {
		return nil, err
	}
	log.Println("Successfully connected to MongoDB")
	return client, nil
}

func GetDatabase(client *mongo.Client, database string) *mongo.Database {
	return client.Database(database)
}

func CreateTTLIndex(ctx context.Context, db *mongo.Database) error {
	// Create TTL index that expires documents after 24 hours (86400 seconds)
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"created_at": 1, // index on the created_at field
		},
		Options: options.Index().SetExpireAfterSeconds(86400).SetName("created_at_1"), // 24 hours TTL
	}

	collection := db.Collection(RideFaresCollection)
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		// Check for IndexOptionsConflict
		if strings.Contains(err.Error(), "IndexOptionsConflict") {
			log.Printf("Index conflict detected for 'created_at_1'. Dropping and recreating...")
			if _, dropErr := collection.Indexes().DropOne(ctx, "created_at_1"); dropErr != nil {
				return fmt.Errorf("failed to drop conflicting index: %w", dropErr)
			}
			// Retry creating the index
			_, err = collection.Indexes().CreateOne(ctx, indexModel)
			return err
		}
		return err
	}
	return nil
}
