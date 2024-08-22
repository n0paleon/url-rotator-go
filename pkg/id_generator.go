package pkg

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func GenerateRandomID() string {
	uid, err := uuid.NewV7()
	if err != nil {
		uid = uuid.New()
	}

	return uid.String()
}

func GenerateRandomNumber(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}
