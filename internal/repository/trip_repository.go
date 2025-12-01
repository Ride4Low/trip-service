package repository

import (
	"context"
	"time"

	"github.com/ride4Low/trip-service/internal/adapter/mongo"
	"github.com/ride4Low/trip-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
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

func (r *mongoRepository) SaveRideFare(ctx context.Context, rideFare *domain.RideFare) error {
	rideFare.CreatedAt = time.Now()
	result, err := r.db.Collection(mongo.RideFaresCollection).InsertOne(ctx, rideFare)
	if err != nil {
		return err
	}

	rideFare.ID = result.InsertedID.(primitive.ObjectID)

	return nil
}

func (r *mongoRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFare, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := r.db.Collection(mongo.RideFaresCollection).FindOne(ctx, bson.M{"_id": _id})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var fare domain.RideFare
	err = result.Decode(&fare)
	if err != nil {
		return nil, err
	}

	return &fare, nil
}

func (r *mongoRepository) CreateTrip(ctx context.Context, trip *domain.Trip) (*domain.Trip, error) {
	result, err := r.db.Collection(mongo.TripsCollection).InsertOne(ctx, trip)
	if err != nil {
		return nil, err
	}

	trip.ID = result.InsertedID.(primitive.ObjectID)

	return trip, nil
}
