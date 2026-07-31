package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/autoparts/backend-test/internal/domain"
	"github.com/autoparts/backend-test/internal/repository"
)

type Repo struct {
	mu    sync.RWMutex
	parts map[string]*domain.Part
}

var _ repository.PartRepository = (*Repo)(nil)

func New() *Repo {
	return &Repo{parts: make(map[string]*domain.Part)}
}

func (r *Repo) Create(_ context.Context, part *domain.Part) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parts[part.ID]; exists {
		return fmt.Errorf(domain.MsgPartExists, part.ID)
	}

	clone := *part
	r.parts[part.ID] = &clone
	return nil
}

func (r *Repo) GetByID(_ context.Context, id string) (*domain.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	part, exists := r.parts[id]
	if !exists {
		return nil, fmt.Errorf(domain.MsgPartNotFound, id)	}

	clone := *part
	return &clone, nil
}

func (r *Repo) GetAll(_ context.Context) ([]*domain.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Part, 0, len(r.parts))
	for _, p := range r.parts {
		clone := *p
		result = append(result, &clone)
	}
	return result, nil
}

func (r *Repo) Update(_ context.Context, part *domain.Part) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parts[part.ID]; !exists {
		return fmt.Errorf(domain.MsgPartNotFound, part.ID)	}

	clone := *part
	r.parts[part.ID] = &clone
	return nil
}

func (r *Repo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parts[id]; !exists {
		return fmt.Errorf(domain.MsgPartNotFound, id)	}

	delete(r.parts, id)
	return nil
}
