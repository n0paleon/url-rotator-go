package ports

import (
	"URLRotatorGo/internal/core/domain"
	"context"
)

type CacheRepository interface {
	SaveShortCode(ctx context.Context, shortcode *domain.ShortCode) error
	GetShortCode(ctx context.Context, code string) (*domain.ShortCode, error)
	IncrShortCode(ctx context.Context, code string) error
	SaveLinks(ctx context.Context, links []*domain.URL) error
	GetLinks(ctx context.Context, code string) ([]*domain.URL, error)
	IncrLink(ctx context.Context, code, id string) error
}
