package ports

import (
	"URLRotatorGo/internal/core/domain"
	"context"
)

type URLRepository interface {
	Save(ctx context.Context, urls []*domain.URL) ([]*domain.URL, error)
	UpdateHit(ctx context.Context, id string) error
	GetLinks(ctx context.Context, code string) ([]*domain.URL, error)
}
