package domain

import "errors"

var (
	ErrInternalServerError error = errors.New("Internal Server Error")
	ErrDataNotFound              = errors.New("Data Not Found")
)
