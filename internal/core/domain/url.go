package domain

import (
	"time"
)

type URL struct {
	ID        int
	ShortCode string
	TotalHit  int
	Original  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
