package dto

type ApiResponse struct {
	Error   bool   `json:"error"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}
