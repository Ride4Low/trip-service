package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/adapter/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

// mockCollection is a mock implementation of MongoDB collection for testing
type mockCollection struct {
	insertOneFunc func(ctx context.Context, document interface{}) (*mongoDriver.InsertOneResult, error)
}

func (m *mockCollection) InsertOne(ctx context.Context, document interface{}) (*mongoDriver.InsertOneResult, error) {
	if m.insertOneFunc != nil {
		return m.insertOneFunc(ctx, document)
	}
	return &mongoDriver.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil
}

// mockDatabase is a mock implementation of MongoDB database for testing
type mockDatabase struct {
	collections map[string]*mockCollection
}

func (m *mockDatabase) Collection(name string) *mockCollection {
	if m.collections == nil {
		m.collections = make(map[string]*mockCollection)
	}
	if m.collections[name] == nil {
		m.collections[name] = &mockCollection{}
	}
	return m.collections[name]
}

// mockMongoRepository wraps mongoRepository to allow injection of mock database
type mockMongoRepository struct {
	mockDB *mockDatabase
}

func (r *mockMongoRepository) SaveTrip(ctx context.Context) error {
	return nil
}

func (r *mockMongoRepository) SaveRideFare(ctx context.Context, rideFare *types.RideFare) error {
	rideFare.CreatedAt = time.Now()
	result, err := r.mockDB.Collection(mongo.RideFaresCollection).InsertOne(ctx, rideFare)
	if err != nil {
		return err
	}

	rideFare.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func TestSaveTrip(t *testing.T) {
	t.Run("successful trip save", func(t *testing.T) {
		// Setup
		mockDB := &mockDatabase{}
		repo := &mockMongoRepository{mockDB: mockDB}

		// Execute
		err := repo.SaveTrip(context.Background())

		// Verify
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestSaveRideFare(t *testing.T) {
	t.Run("successful ride fare insertion", func(t *testing.T) {
		// Setup
		expectedID := primitive.NewObjectID()
		mockDB := &mockDatabase{
			collections: map[string]*mockCollection{
				mongo.RideFaresCollection: {
					insertOneFunc: func(ctx context.Context, document interface{}) (*mongoDriver.InsertOneResult, error) {
						// Verify the document has CreatedAt set
						fare, ok := document.(*types.RideFare)
						if !ok {
							t.Error("expected document to be *types.RideFare")
						}
						if fare.CreatedAt.IsZero() {
							t.Error("expected CreatedAt to be set before insertion")
						}
						return &mongoDriver.InsertOneResult{InsertedID: expectedID}, nil
					},
				},
			},
		}
		repo := &mockMongoRepository{mockDB: mockDB}

		rideFare := &types.RideFare{
			UserID:            "user-123",
			PackageSlug:       "suv",
			TotalPriceInCents: 1850.0,
		}
		beforeTime := time.Now()

		// Execute
		err := repo.SaveRideFare(context.Background(), rideFare)
		afterTime := time.Now()

		// Verify
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rideFare.ID != expectedID {
			t.Errorf("expected ID to be set to %v, got %v", expectedID, rideFare.ID)
		}
		if rideFare.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
		if rideFare.CreatedAt.Before(beforeTime) || rideFare.CreatedAt.After(afterTime) {
			t.Errorf("expected CreatedAt to be between %v and %v, got %v", beforeTime, afterTime, rideFare.CreatedAt)
		}
		if rideFare.UserID != "user-123" {
			t.Errorf("expected UserID to remain user-123, got %s", rideFare.UserID)
		}
		if rideFare.PackageSlug != "suv" {
			t.Errorf("expected PackageSlug to remain suv, got %s", rideFare.PackageSlug)
		}
	})

	t.Run("database insertion error", func(t *testing.T) {
		// Setup
		expectedErr := errors.New("database connection failed")
		mockDB := &mockDatabase{
			collections: map[string]*mockCollection{
				mongo.RideFaresCollection: {
					insertOneFunc: func(ctx context.Context, document interface{}) (*mongoDriver.InsertOneResult, error) {
						return nil, expectedErr
					},
				},
			},
		}
		repo := &mockMongoRepository{mockDB: mockDB}

		rideFare := &types.RideFare{
			UserID:            "user-123",
			PackageSlug:       "sedan",
			TotalPriceInCents: 2000.0,
		}

		// Execute
		err := repo.SaveRideFare(context.Background(), rideFare)

		// Verify
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error to be %v, got %v", expectedErr, err)
		}
		// CreatedAt should still be set even on error
		if rideFare.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set even on error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Setup
		mockDB := &mockDatabase{
			collections: map[string]*mockCollection{
				mongo.RideFaresCollection: {
					insertOneFunc: func(ctx context.Context, document interface{}) (*mongoDriver.InsertOneResult, error) {
						// Check if context is cancelled
						if ctx.Err() != nil {
							return nil, ctx.Err()
						}
						return &mongoDriver.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil
					},
				},
			},
		}
		repo := &mockMongoRepository{mockDB: mockDB}

		rideFare := &types.RideFare{
			UserID:            "user-123",
			PackageSlug:       "van",
			TotalPriceInCents: 2050.0,
		}

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Execute
		err := repo.SaveRideFare(ctx, rideFare)

		// Verify
		if err == nil {
			t.Fatal("expected error due to cancelled context, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})

	t.Run("multiple fares insertion", func(t *testing.T) {
		// Setup
		insertCount := 0
		mockDB := &mockDatabase{
			collections: map[string]*mockCollection{
				mongo.RideFaresCollection: {
					insertOneFunc: func(ctx context.Context, document interface{}) (*mongoDriver.InsertOneResult, error) {
						insertCount++
						return &mongoDriver.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil
					},
				},
			},
		}
		repo := &mockMongoRepository{mockDB: mockDB}

		fares := []*types.RideFare{
			{UserID: "user-1", PackageSlug: "suv", TotalPriceInCents: 1850.0},
			{UserID: "user-2", PackageSlug: "sedan", TotalPriceInCents: 2000.0},
			{UserID: "user-3", PackageSlug: "luxury", TotalPriceInCents: 2650.0},
		}

		// Execute
		for _, fare := range fares {
			err := repo.SaveRideFare(context.Background(), fare)
			if err != nil {
				t.Fatalf("expected no error for fare %v, got %v", fare.PackageSlug, err)
			}
			if fare.ID.IsZero() {
				t.Errorf("expected ID to be set for fare %v", fare.PackageSlug)
			}
		}

		// Verify
		if insertCount != 3 {
			t.Errorf("expected 3 insertions, got %d", insertCount)
		}
	})
}
