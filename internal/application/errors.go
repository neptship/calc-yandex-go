package application

import "errors"

var (
	ErrEmptyExpression  = errors.New("empty expression")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrServerError      = errors.New("internal server error")
)
