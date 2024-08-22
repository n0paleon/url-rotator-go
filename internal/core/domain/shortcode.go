package domain

import "time"

type Strategy string

const (
	RoundRobin Strategy = "RR"
	Random     Strategy = "RNDM"
)

type ShortCode struct {
	ID        string
	Code      string
	TotalHit  int
	Strategy  Strategy
	CreatedAt time.Time
	UpdatedAt time.Time
}
