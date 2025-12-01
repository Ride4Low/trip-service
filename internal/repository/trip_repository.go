package repository

import (
	"context"
	"time"

	"github.com/ride4Low/trip-service/internal/adapter/mongo"
	"github.com/ride4Low/trip-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

// repository struct implementing Repository
type mongoRepository struct {
	db *mongoDriver.Database
}

func NewRepository(db *mongoDriver.Database) domain.Repository {
	return &mongoRepository{
		db: db,
	}
}

func (r *mongoRepository) SaveTrip(ctx context.Context) error {
	return nil
}

func (r *mongoRepository) SaveRideFare(ctx context.Context, rideFare *domain.RideFare) error {
	rideFare.CreatedAt = time.Now()
	result, err := r.db.Collection(mongo.RideFaresCollection).InsertOne(ctx, rideFare)
	if err != nil {
		return err
	}

	rideFare.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}
