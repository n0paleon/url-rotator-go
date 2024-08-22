package dto

import "time"

type RequestShortURL struct {
	URL      []string `json:"urls" validate:"required,dive,min=5,max=1000,url"`
	Strategy string   `json:"strategy" validate:"required"`
}

type ResponseShortURL struct {
	URL       string    `json:"url"`
	Strategy  string    `json:"strategy"`
	CreatedAt time.Time `json:"created_at"`
}
