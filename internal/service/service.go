package service

import "context"

// Service interface
type Service interface {
	CreateTrip(ctx context.Context) error
}

// service struct implementing Service
type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) CreateTrip(ctx context.Context) error {
	return nil
}
