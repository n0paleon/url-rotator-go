package ports

import (
	"URLRotatorGo/internal/core/domain"
	"context"
)

type ShortenerService interface {
	ShortURL(ctx context.Context, urls []string, strategy string) (*domain.ShortCode, error)
	GetRedirectURL(ctx context.Context, code string) (string, error)
}
