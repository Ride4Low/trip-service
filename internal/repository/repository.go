package repository

import "context"

// Repository interface
type Repository interface {
	SaveTrip(ctx context.Context) error
}

// repository struct implementing Repository
type repository struct {
}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) SaveTrip(ctx context.Context) error {
	return nil
}
