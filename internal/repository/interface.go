package repository

import (
	"context"

	"github.com/autoparts/backend-test/internal/domain"
)

type PartRepository interface {
	Create(ctx context.Context, part *domain.Part) error
	GetByID(ctx context.Context, id string) (*domain.Part, error)
	GetAll(ctx context.Context) ([]*domain.Part, error)
	Update(ctx context.Context, part *domain.Part) error
	Delete(ctx context.Context, id string) error
}
