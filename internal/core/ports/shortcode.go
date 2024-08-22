package ports

import (
	"URLRotatorGo/internal/core/domain"
	"context"
)

type ShortCodeRepository interface {
	Save(ctx context.Context, url *domain.ShortCode) (*domain.ShortCode, error)
	UpdateHit(ctx context.Context, code string) error
	GetShortCode(ctx context.Context, code string) (*domain.ShortCode, error)
}
